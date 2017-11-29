# LEDserver
Web application to remotely control RGB LEDs via IR signal - for Raspberry Pi - written in Go

<img src="/../images/piir.jpg?raw=true" height=256 /> **➕**
<img src="/../images/web.png?raw=true" height=256 /> **❯**
<img src="/../images/irr.jpg?raw=true" height=256 />

It uses my other go packeges:
   - [go-rpio](//github.com/drahoslav7/go-rpio) - to control gpio pins of Raspberri Pi (forked from [stianeikeland](//github.com/stianeikeland/go-rpio) and extended by clock and pwm modes)
   - [go-nec](//github.com/drahoslav7/go-nec) - for basic encoding and transmission of NEC - Infrared transmission protocol
   

## Requirements
- RGB LED strip with 24-button IR remote controller
   - for 44-button controller some tweaks in code are neccesary
   - the remote itself is not actually needed - it will be replaced with pi
- Raspberry Pi A+/B+/2B/3B/ZERO
  - with network connection
  - with golang installed
- 940nm IR LED
- suitable resistor and jumper wires

## Installation
### Software
- Run:
  ```bash
  go get -u -v github.com/drahoslav7/ledserver
  cd $GOPATH/src/github.com/drahoslav7/ledserver
  sudo ledserver
  ```
- Open in your browser:
  ```
  http://<ip of your pi>:8080/
  ```
### Hardware
- connect IR LED between BCM pins 19 and 26
  - cathode (-) to pin [19](https://pinout.xyz/pinout/pin35_gpio19 "physical pin 35")
  - anode (+) to pin [26](https://pinout.xyz/pinout/pin37_gpio26 "physical pin 37")
  - usage of proper resistor highly recommended
