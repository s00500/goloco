# GoLoco

Simple webui for interfacing with LGB MTS modeltrains via an LGB 55060 MTS PC Interface
Protocol is very simple and has been reverse engineered, _this project is work in progress_

## TODO

- [ ] Add feedback from locos to webui
- [ ] add state of accessories to webui
- [ ] style UI

### Useful Links

http://getskeleton.com/#grid
https://elmassian.com/index.php?option=com_content&view=article&id=505&Itemid=614

# Contribute

Liked this Programm? You can **support** me by sending me a :coffee:
[Coffee](https://paypal.me/lukasbachschwell/5).

Otherwise I really welcome **Pull Requests**.

# Notes

LGB Commands

Emergency stop:
07 00 80 87

Emergency release
07 00 81 86

Switch 17 right
03 11 00 12

Left:
03 11 01 13

Switch 15 right
03 0F 00 0C

Switch 15 left
03 0F 01 0D

Loko 3 stop
01 03 20 22

(Short after: 06 03 01 04 ?)

Loko 3 forward 1
01 03 22 20

forward 2
01 03 23 21

forward 3
01 03 24 26

Max:
01 03 2F 2D (8 steps ?)

Backward 1 or 2
01 03 03 01

3
01 03 04 06

Loko 3 light toggl
02 03 80 80

Interface to PC

—————————————
PC to interface

Loko 6 Light on
01 06 00 07 02 86 80 04

01 06 00 07 02 86 80 04

Loko 6 trigger function 1

01 06 00 07 02 86 01 05

loko 6 trigger function 5
01 06 00 07 02 86 05 81

Loko 6 emergency break

01 06 21 26 01 06 20 27

Genral handbreak:

01 06 20 27 for all running lokos!
