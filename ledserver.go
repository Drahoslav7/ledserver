package main

import (
	"fmt"
	"github.com/drahoslav7/go-nec"
	"github.com/drahoslav7/go-rpio"
	"html/template"
	"image/color"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const httpPort = ":8080"

var (
	addr uint16 = 61184 // 24 key ir remote control
)

var (
	grey  = color.RGBA{187, 187, 187, 255}
	black = color.RGBA{8, 8, 8, 255}
	red   = color.RGBA{204, 0, 0, 255}
)

func c(r, g, b uint8) color.RGBA {
	return color.RGBA{r * 51, g * 51, b * 51, 255}
}

type button struct {
	Name  string
	Color color.RGBA
}

var buttons = []button{
	{"\u2795", grey}, {"\u2796", grey}, {"OFF", black}, {"ON", red},
	{"", c(5, 0, 0)}, {"", c(0, 5, 0)}, {"", c(0, 0, 5)}, {"", c(5, 5, 5)},
	{"", c(4, 1, 0)}, {"", c(0, 4, 1)}, {"", c(1, 0, 4)}, {"FLASH", grey},
	{"", c(3, 2, 0)}, {"", c(0, 3, 2)}, {"", c(2, 0, 3)}, {"STROBE", grey},
	{"", c(2, 3, 0)}, {"", c(0, 2, 3)}, {"", c(3, 0, 2)}, {"FADE", grey},
	{"", c(1, 4, 0)}, {"", c(0, 1, 4)}, {"", c(4, 0, 1)}, {"SMOOTH", grey},
}

func init() {
	runtime.LockOSThread()
	syscall.Setpriority(syscall.PRIO_PROCESS, 0, -5) // lower niceness to get more timing precission (theoretically)
}

func main() {
	// init gpio
	err := rpio.Open()
	if err != nil {
		os.Exit(1)
	}
	defer rpio.Close()

	// set output pin
	led := rpio.Pin(26)
	led.Mode(rpio.Output)
	led.Write(rpio.Low)

	// set carry signal at 38kHz with 1/3 duty cycle
	carry := rpio.Pin(19)
	carry.Mode(rpio.Pwm)
	carry.DutyCycle(2, 3) // 2/3 actually because connected to neg end of LED
	carry.Freq(38000 * 3)

	toLED := func(v bool) {
		if v {
			led.Write(rpio.High)
		} else {
			led.Write(rpio.Low)
		}
	}

	// use single channel for handling commands to ensure only one signal is transmited at time
	cmdChan := make(chan uint8)

	// load html template
	indexTmpl := template.Must(template.ParseFiles("index.html"))

	// http handlers
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		err := indexTmpl.Execute(w, &buttons)
		if err != nil {
			fmt.Printf("err: %s", err)
		}
	})

	http.HandleFunc("/cmd/", func(w http.ResponseWriter, req *http.Request) {
		path := strings.SplitN(req.URL.Path[1:], "/", 2)
		if len(path) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		cmd, err := strconv.Atoi(path[1])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		go func () {
			cmdChan <- uint8(cmd)
		}()
	})

	go http.ListenAndServe(httpPort, nil)
	fmt.Printf("Serving at %s\n", httpPort)

	for cmd := range cmdChan {
		nec.EncodeExt(addr, cmd).TransmitTimes(toLED, 3) // 3 times just to be sure
	}
}
