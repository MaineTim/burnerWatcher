// burnerWatcher

package main

import "fmt"
import "github.com/warthog618/gpio"
import "github.com/docopt/docopt-go"
import "os"
import "os/signal"
import "syscall"
import "time"
import log "github.com/Sirupsen/logrus"

const version = ".01a-2017Nov15"

const usage = `
burnerWatcher

Usage: burnerWatcher <url>
`

var PinState gpio.Level
var LastState gpio.Level
var StartTime = time.Time{}

/*
func storeEntry(db *sql.DB, startTime time.Time, endTime time.Time) {
	sql_additem := `
INSERT INTO runtimes(
StartTime,
EndTime,
Duration,
InsertedDatetime
) values(?, ?, ?, CURRENT_TIMESTAMP)
`
	stmt, err := db.Prepare(sql_additem)
	if err != nil { panic(err) }
	defer stmt.Close()

	_, err2 := stmt.Exec(startTime, endTime, endTime.Sub(startTime))
	if err2 != nil { panic(err2) }
	log.Println ("Logged run time.")
}
*/

func sendTemperatures() {
	fmt.Println("Sending temp data.")
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
					postp = 6
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
	fmt.Println("Entry to send: ", startTime, endTime, endTime.Sub(startTime))
	log.Info("Sent entry to log.")
}

func mainloop() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}

func main() {
	arguments, _ := docopt.Parse(usage, nil, true, version, false)
	url := arguments["<url>"].(string)

	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-Jan-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)

	log.Info("Logging data to ", url)

	signal := make(chan int)
	go tempLogger(signal)

	if err := gpio.Open(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer gpio.Close()

	pin := gpio.NewPin(23)
	pin.Input()

	pin.Watch(gpio.EdgeBoth, func(pin *gpio.Pin) {
		LastState = pin.Read()
		time.Sleep(500 * time.Millisecond)
		newState := pin.Read()
		if newState == LastState {
			PinState = newState
		}
		if PinState == gpio.High {
			StartTime = time.Now()
			log.Info("Started time")
			signal <- 1
		} else if (PinState == gpio.Low) && (StartTime != time.Time{}) {
			endTime := time.Now()
			sendEntry(url, StartTime, endTime)
			StartTime = time.Time{}
			signal <- 0
		}
	})
	defer pin.Unwatch()

	mainloop()
	signal <- 2
}
