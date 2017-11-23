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
	white = color.RGBA{255, 255, 255, 255}
)

func c(h float64) color.RGBA {
	l := 0.5
	s := 1.0
	hue2rgb := func(p, q, t float64) float64 {
		if t < 0 {
			t += 1.0
		}
		if t > 1 {
			t -= 1.0
		}
		switch {
		case t < 1./6:
			return p + (q-p)*6*t
		case t < 3./6:
			return q
		case t < 4./6:
			return p + (q-p)*(2/3-t)*6
		default:
			return p
		}
	}
	q := l + s - l*s
	p := 2*l - q
	r := uint8(hue2rgb(p, q, h+1./3) * 255)
	g := uint8(hue2rgb(p, q, h) * 255)
	b := uint8(hue2rgb(p, q, h-1./3) * 255)
	return color.RGBA{r, g, b, 255}
}

type button struct {
	Name  string
	Color color.RGBA
}

var buttons = []button{
	{"\u2795", grey}, {"\u2796", grey}, {"OFF", black}, {"ON", red},
	{"", c(0. / 15)}, {"", c(5. / 15)}, {"", c(10. / 15)}, {"", white},
	{"", c(1. / 15)}, {"", c(6. / 15)}, {"", c(11. / 15)}, {"STROBE", grey},
	{"", c(2. / 15)}, {"", c(7. / 15)}, {"", c(12. / 15)}, {"FADE", grey},
	{"", c(3. / 15)}, {"", c(8. / 15)}, {"", c(13. / 15)}, {"SMOOTH", grey},
	{"", c(4. / 15)}, {"", c(9. / 15)}, {"", c(14. / 15)}, {"FLASH", grey},
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
		go func() {
			cmdChan <- uint8(cmd)
		}()
	})

	go http.ListenAndServe(httpPort, nil)
	fmt.Printf("Serving at %s\n", httpPort)

	for cmd := range cmdChan {
		nec.EncodeExt(addr, cmd).TransmitTimes(toLED, 3) // 3 times just to be sure
	}
}
