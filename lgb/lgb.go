package lgb

import (
	"fmt"
	"sync"
	"time"

	log "github.com/s00500/env_logger"
	"github.com/tarm/serial"
)

const controlLocoSpeed byte = 0x01
const controlLocoFunction byte = 0x02
const controlAccessory byte = 0x03
const locoRelease byte = 0x06
const systemStatus byte = 0x07

/*
var switch17left = []byte{0x03, 0x11, 0x00}
var switch17right = []byte{0x03, 0x11, 0x01}
var emergencyStop = []byte{0x07, 0x00, 0x80}
var emergencyGo = []byte{0x07, 0x00, 0x81}

var tutu = []byte{0x02, 0x86, 0x08}

var loco2light = []byte{0x02, 0x82, 0x80}
var loco2dampf = []byte{0x02, 0x82, 0x03}
*/

type StateChange struct {
	Number uint8
	Loco   *Locomotive
	Acc    *Accessory
}
type Accessory struct {
	State          bool
	hasInitialized bool
}
type Locomotive struct {
	Light           bool
	Speed           int8
	isControlled    bool
	controlledSince time.Time
}

type System struct {
	PortName    string
	s           *serial.Port
	sm          sync.Mutex
	locos       []Locomotive
	accessories []Accessory
	OutChannel  chan StateChange
}

func init() {
	log.ConfigureDefaultLogger()
}

func (lgb *System) Start(resetAll bool) error {
	// open serial port
	lgb.locos = make([]Locomotive, 24)
	lgb.accessories = make([]Accessory, 256)
	c := &serial.Config{Name: lgb.PortName, Baud: 9600}
	var err error
	lgb.s, err = serial.OpenPort(c)
	if err != nil {
		return err
	}

	if resetAll {
		for i := 1; i < 255; i++ {
			log.Info("Reset acc ", i)
			lgb.SwitchFunction(uint8(i), false)
			time.Sleep(50 * time.Millisecond)
		}
	}

	lgb.OutChannel = make(chan StateChange)

	go lgb.CheckControlledLocos()
	go lgb.CheckIncoming()
	return nil
}

func (lgb *System) send(data []byte) error {
	lgb.sm.Lock()
	defer lgb.sm.Unlock()
	log.Info(chkSum(data))
	_, err := lgb.s.Write(chkSum(data))
	return err
}

func (lgb *System) SwitchFunction(switchNumber uint8, direction bool) {
	// Later on accessory state could also be saved
	var directionByte byte = 0x00
	if direction {
		directionByte = 0x01
	}
	lgb.send([]byte{controlAccessory, switchNumber, directionByte})
	lgb.accessories[switchNumber].State = direction
	log.Info("Sent Command to accessory ", switchNumber)
	if lgb.OutChannel == nil {
		return
	}
	lgb.OutChannel <- StateChange{Number: switchNumber, Acc: &Accessory{State: direction}}
}

func (lgb *System) LocoLight(loco uint8) error {
	if loco > 24 && loco != 0 {
		return fmt.Errorf("Loco number too high")
	}
	err := lgb.send([]byte{controlLocoFunction, loco + 0x80, 0x80})

	lgb.locos[loco-1].Light = !lgb.locos[loco-1].Light
	lgb.OutChannel <- StateChange{Number: loco, Loco: &lgb.locos[loco-1]}

	log.Info("Sent Light to loco ", loco)
	return err
}

func (lgb *System) LocoFunction(number uint8, loco uint8) error {
	if loco > 24 && loco != 0 {
		return fmt.Errorf("Loco number too high")
	}
	err := lgb.send([]byte{controlLocoFunction, loco + 0x80, number})
	log.Info("Sent Function ", number, " to loco ", loco)
	return err

}

func (lgb *System) LocoStop(loco uint8) error {
	if loco > 24 && loco != 0 {
		return fmt.Errorf("Loco number too high")
	}

	lgb.locoMarkControlled(loco)

	lgb.locos[loco-1].Speed = 0
	err := lgb.send([]byte{controlLocoSpeed, loco, 0x20})
	lgb.OutChannel <- StateChange{Number: loco, Loco: &lgb.locos[loco-1]}

	log.Info("Sent STOP to loco ", loco)
	return err
}

func (lgb *System) LocoForward(loco uint8) error {
	if loco > 24 && loco != 0 {
		return fmt.Errorf("Loco number too high")
	}
	if lgb.locos[loco-1].Speed >= 14 {
		return nil
	}

	lgb.locoMarkControlled(loco)

	lgb.locos[loco-1].Speed = lgb.locos[loco-1].Speed + 1

	var err error
	if lgb.locos[loco-1].Speed > 0 {
		err = lgb.send([]byte{controlLocoSpeed, loco, byte(0x20 + lgb.locos[loco-1].Speed)})
	} else {
		err = lgb.send([]byte{controlLocoSpeed, loco, byte(-lgb.locos[loco-1].Speed)})
	}
	lgb.OutChannel <- StateChange{Number: loco, Loco: &lgb.locos[loco-1]}

	log.Info("Sent Speed ", lgb.locos[loco-1].Speed, " to loco ", loco)
	return err
}

func (lgb *System) LocoBackward(loco uint8) error {
	if loco > 24 && loco != 0 {
		return fmt.Errorf("Loco number too high")
	}
	if lgb.locos[loco-1].Speed <= -14 {
		return nil
	}

	lgb.locoMarkControlled(loco)

	lgb.locos[loco-1].Speed = lgb.locos[loco-1].Speed - 1

	var err error
	if lgb.locos[loco-1].Speed > 0 {
		err = lgb.send([]byte{controlLocoSpeed, loco, byte(0x20 + lgb.locos[loco-1].Speed)})
	} else {
		err = lgb.send([]byte{controlLocoSpeed, loco, byte(-lgb.locos[loco-1].Speed)})
	}
	lgb.OutChannel <- StateChange{Number: loco, Loco: &lgb.locos[loco-1]}

	log.Info("Sent Speed ", lgb.locos[loco-1].Speed, " to loco ", loco)
	return err
}

func (lgb *System) locoMarkControlled(loco uint8) {
	if loco > 24 && loco != 0 {
		return
	}
	lgb.locos[loco-1].isControlled = true
	lgb.locos[loco-1].controlledSince = time.Now()
}

func (lgb *System) LocoRelease(loco uint8) error {
	if loco > 24 && loco != 0 {
		return fmt.Errorf("Loco number too high")
	}
	err := lgb.send([]byte{06, loco, 0x01})
	lgb.locos[loco-1].isControlled = false
	log.Info("Release loco ", loco)
	return err
}

func (lgb *System) GetLocoStates() (locoStates []StateChange) {
	for num := range lgb.locos {
		locoStates = append(locoStates, StateChange{Number: uint8(num) + 1, Loco: &lgb.locos[num]})
	}
	return locoStates
}

func (lgb *System) GetAccessoryStates() (accStates []bool) {
	for num := range lgb.accessories {
		accStates = append(accStates, lgb.accessories[num].State)
	}
	return accStates
}

func (lgb *System) CheckControlledLocos() {
	for {
		for index, loco := range lgb.locos {
			if loco.isControlled && time.Now().After(loco.controlledSince.Add(time.Second*3)) {
				lgb.LocoRelease(uint8(index) + 1)
			}
		}
		time.Sleep(time.Second)
	}
}

func (lgb *System) CheckIncoming() {
	for {

		buf := make([]byte, 4) // Should use 3 or 4 or so
		for i := 0; i < 4; {
			n, err := lgb.s.Read(buf[i:])
			i = i + n
			if err != nil {
				log.Warn("Error reading !", err)
				break
			}
		}
		lgb.ParseCommand(buf)
	}
}
func (lgb *System) ParseCommand(data []byte) {
	if len(data) != 4 {
		log.Warn("Invalid data length", data)
		return
	}

	if chkSum(data[:3])[3] != data[3] {
		log.Warn("Invalid checksum recieved", data)
		return
	}

	// switch type
	switch data[0] {
	case controlLocoSpeed:
		log.Info("Loco speed", data)
		if data[1] > 24 && data[1] != 0 {
			log.Warn("Loco number bugger than 24 or 0, could not parse speed")
		}
		log.Info("Loco: ", data[1], " number: ", data[2])

		if data[2] == 0x20 {
			lgb.locos[data[1]-1].Speed = 0
		} else if data[2] > 0x20 {
			lgb.locos[data[1]-1].Speed = int8(data[2] - 0x20)
		} else {
			lgb.locos[data[1]-1].Speed = -int8(data[2])
		}

		log.Info("Speed->: ", lgb.locos[data[1]-1].Speed)

		lgb.OutChannel <- StateChange{Number: uint8(data[1]), Loco: &lgb.locos[data[1]-1]}
		break
	case controlLocoFunction:
		log.Info("Loco function", data)
		if data[1] > 24 && data[1] != 0 {
			log.Warn("Loco number bugger than 24 or 0, could not parse speed")
		}
		if data[2] == 128 {
			lgb.locos[data[1]-1].Light = !lgb.locos[data[1]-1].Light
		}
		lgb.OutChannel <- StateChange{Number: uint8(data[1]), Loco: &lgb.locos[data[1]-1]}
		break
	case controlAccessory:
		dirString := "left"
		if data[2] == 0x01 {
			dirString = "right"

		}
		log.Info("Accessory Number: ", int(data[1]), " turning ", dirString)

		break
	case locoRelease:
		log.Info("Loco release !", data)
		break
	case systemStatus:
		sysString := ""
		if data[2] == 128 {
			sysString = "Emergency STOP"
		} else if data[2] == 129 {
			sysString = "Emergency Release"
		}
		log.Info("System status ! ", sysString, " ", data)
		break
	}

}

func (lgb *System) EmergencyStop() {
	lgb.send([]byte{0x07, 0x00, 0x80})
	log.Info("Sent Emergency STOP")
}

func (lgb *System) EmergencyRelease() {
	lgb.send([]byte{0x07, 0x00, 0x81})
	log.Info("Sent Emergency Release")
}

func chkSum(data []byte) []byte {
	var checksum byte = 0
	for _, b := range data {
		checksum = checksum ^ b
	}
	data = append(data, checksum)
	return data
}
