// burnerWatcher

package main

import (
	"bytes"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
	"github.com/spf13/viper"
	"io/ioutil"
	//	"github.com/pkg/errors"
	"github.com/warthog618/gpio"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const version = ".01-2017Nov19"

const usage = `
burnerWatcher

Usage: httpLogger [options]

Options:
 -d LEVEL  Set logging level.
             i = Info
             e = Error
             d = Debug
             [default: e]
 -v         Show version.
`

type RunEntry struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type ConfigFile struct {
	urlTempServer string
	urlTimeServer string
}

var (
	PinState   gpio.Level
	LastState  gpio.Level
	StartTime  = time.Time{}
	configFile ConfigFile
)

func sendTemperatures() {
	var status int
	var body []byte

	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}
	log.Debugf("Sending GET to: %s", configFile.urlTempServer)
	response, err := netClient.Get(configFile.urlTempServer)
	if err != nil {
		log.Errorf("The HTTP request failed with error %s\n", err)
	} else {
		status = response.StatusCode
		body, _ = ioutil.ReadAll(response.Body)
	}
	log.Info(status, " - "+string(body))

}

func tempLogger(signal chan int) {
	active := 0
	postp := 0
	loop := true
	for loop == true {
		select {
		case active = <-signal:
		default:
			if active == 1 || postp > 0 {
				sendTemperatures()
				time.Sleep(5 * time.Second)
				if active == 1 {
					postp = 6 // 5 passes
				}
			} else if active == 2 {
				loop = false
			}
			if postp > 0 {
				postp--
			}
		}
	}
}

func sendEntry(url string, startTime time.Time, endTime time.Time) {

	var entry RunEntry
	var status int

	entry.StartTime = startTime.Format(time.RFC3339)
	entry.EndTime = endTime.Format(time.RFC3339)
	body, _ := json.Marshal(entry)
	log.Debugf("Sending POST to: %s", configFile.urlTimeServer)
	response, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Errorf("The HTTP request failed with error %s\n", err)
	} else {
		status = response.StatusCode
		body, _ = ioutil.ReadAll(response.Body)
	}
	log.Info(status, " - "+string(body))
}

func mainloop() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}

func main() {
	var (
		err error
	)
	defer os.Exit(0)

	viper.SetConfigFile("burnerWatcher.toml")
	//  viper.AddConfigPath(".")
	if err = viper.ReadInConfig(); err != nil {
		log.Errorf("Config file error: %s", err)
		runtime.Goexit()
	} else {
		configFile.urlTempServer = viper.GetString("Servers.temperatures")
		configFile.urlTimeServer = viper.GetString("Servers.time")
	}
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-Jan-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	arguments, _ := docopt.Parse(usage, nil, true, version, false)
	logLevel := arguments["-d"]
	switch logLevel {
	case "d":
		log.SetLevel(log.DebugLevel)
	case "i":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}

	log.Info("burnerWatcher " + version + " starting")
	signal := make(chan int)
	go tempLogger(signal)

	if err = gpio.Open(); err != nil {
		log.Errorf("GPIO pin open failed: %s", err)
		runtime.Goexit()
	}
	defer gpio.Close()

	pin := gpio.NewPin(23)
	pin.Input()

	pin.Watch(gpio.EdgeBoth, func(pin *gpio.Pin) {
		LastState = pin.Read()
		time.Sleep(1 * time.Second)
		newState := pin.Read()
		if newState == LastState {
			PinState = newState
		}
		if PinState == gpio.High {
			StartTime = time.Now().UTC()
			log.Info("Started timing")
			signal <- 1
		} else if (PinState == gpio.Low) && (StartTime != time.Time{}) {
			log.Info("Ended timing")
			endTime := time.Now().UTC()
			sendEntry(configFile.urlTimeServer, StartTime, endTime)
			StartTime = time.Time{}
			signal <- 0
		}
	})
	defer pin.Unwatch()

	mainloop()
	signal <- 2
}
