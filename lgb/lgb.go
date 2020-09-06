package lgb

import (
	"fmt"
	"log"
	"sync"

	"github.com/tarm/serial"
	//"github.com/tarm/serial"
)

const controlLocoSpeed byte = 0x01
const controlLocoFunction byte = 0x02
const controlAccessory byte = 0x03

/*
var switch17left = []byte{0x03, 0x11, 0x00}
var switch17right = []byte{0x03, 0x11, 0x01}
var emergencyStop = []byte{0x07, 0x00, 0x80}
var emergencyGo = []byte{0x07, 0x00, 0x81}

var tutu = []byte{0x02, 0x86, 0x08}

var loco2light = []byte{0x02, 0x82, 0x80}
var loco2dampf = []byte{0x02, 0x82, 0x03}
*/

type Locomotive struct {
	light bool
	speed int8
}

type System struct {
	PortName string
	s        *serial.Port
	sm       sync.Mutex
	locos    []Locomotive
}

func (lgb *System) Start() error {
	// open serial port
	c := &serial.Config{Name: lgb.PortName, Baud: 9600}
	var err error
	lgb.s, err = serial.OpenPort(c)
	if err != nil {
		return err
	}
	return nil
}

func (lgb *System) send(data []byte) error {
	lgb.sm.Lock()
	defer lgb.sm.Unlock()
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
	log.Printf("Sent Command to accessory %d\n", switchNumber)
}

func (lgb *System) LocoLight(loco uint8) error {
	if loco >= 24 {
		return fmt.Errorf("Loco number too high")
	}
	err := lgb.send([]byte{controlLocoFunction, loco + 0x80, 0x80})
	lgb.locos[loco].light = !lgb.locos[loco].light
	log.Printf("Sent Light to loco %d\n", loco)
	return err
}

func (lgb *System) LocoFunction(number uint8, loco uint8) error {
	if loco >= 24 {
		return fmt.Errorf("Loco number too high")
	}
	err := lgb.send([]byte{controlLocoFunction, loco + 0x80, number})
	log.Printf("Sent Function %d to loco %s\n", number, loco)
	return err

}

func (lgb *System) LocoStop() {

}

func (lgb *System) EmergencyStop() {
	lgb.send([]byte{0x07, 0x00, 0x80})
	log.Println("Sent Emergency STOP")
}

func (lgb *System) EmergencyRelease() {
	lgb.send([]byte{0x07, 0x00, 0x81})
	log.Println("Sent Emergency Release")
}

func chkSum(data []byte) []byte {
	var checksum byte = 0
	for _, b := range data {
		checksum = checksum ^ b
	}
	data = append(data, checksum)
	return data
}
