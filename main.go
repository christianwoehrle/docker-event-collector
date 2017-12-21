package main

import (
	"fmt"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/alecthomas/kingpin.v2"

	"os"
	"strconv"
)

type container struct {
	name   string
	deaths int
}

func (c container) String() string {
	return fmt.Sprintf("{Name:%s, Deaths:%d}", c.name, c.deaths)
}

var mutex = &sync.Mutex{}

type containers []container

func (a containers) Len() int           { return len(a) }
func (a containers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a containers) Less(i, j int) bool { return a[i].deaths < a[j].deaths }

var containerDeaths map[string]container

func main() {

	var (
		interval  = kingpin.Flag("interval", "Statistics every <interval> minutes.").Default("3").Int()
		starttime = kingpin.Flag("starttime", "Time when report should be printed [hh:mm]").String()
		logLevel  = kingpin.Flag("logLevel", "LogLevel for Program").Default("INFO").Enum("DEBUG", "INFO", "WARNING", "ERROR")
	)
	kingpin.Parse()

	containerDeaths = make(map[string]container)
	fmt.Println(*interval, *logLevel)

	//reporttimer := time.Now()

	//fmt.Println(reporttimer, starttime)
	switch *logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARNING":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)

	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		log.Info("start interrupt listener")
		for {
			sig := <-signalChan
			fmt.Println("\nReceived an interrupt, showstats... :\n", sig)
			showStatistics()
		}
	}()

	firstAlert := getFirstAlertTime(*starttime)
	log.Info("FirstAlert: ", firstAlert)
	timer := time.NewTimer(time.Until(firstAlert))

	log.Info("TimeUntil: ", time.Until(firstAlert))
	log.Info("TimeSub: ", firstAlert.Sub(time.Now()))

	go func() {
		<-timer.C
		log.Info("Startzeitpunkt erreicht")
		showStatistics()
		ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
		go func() {
			log.Info("Start Statisctics")
			for {
				for {
					select {
					case <-ticker.C:
						showStatistics()
					}
				}
			}
		}()

	}()

	//pattern für services, containername ist vorne.number.id
	servicePattern := regexp.MustCompile("\\.([0-9]+)\\.([0-9a-z]+)$")

	endpoint := "unix:///var/run/docker.sock"

	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	log.Info("Start Event Listener für Docker Events...")
	events := make(chan *docker.APIEvents)
	client.AddEventListener(events)

	quit := make(chan struct{})

	numContainerDeaths := 0

	// Process Docker events
	for msg := range events {
		switch msg.Status {
		case "die":
			numContainerDeaths++
			//log.Println("Die event #", numContainerDeaths, "...", msg)
			id := msg.ID
			var dc *docker.Container
			var err error
			if id != "" {
				dc, err = client.InspectContainer(id)
				if err != nil {
					fmt.Println(err)
				} else {
					log.Debug("Container died, name:", dc.Name, " Id:", id)
				}
			}
			cname := servicePattern.ReplaceAllString(dc.Name, "")
			cname = strings.TrimPrefix(cname, "/")
			mutex.Lock()
			c, ok := containerDeaths[cname]
			if ok {
				c.deaths = c.deaths + 1
				// TODO Why can´t i just increment
				containerDeaths[cname] = c
			} else {
				c = container{name: cname, deaths: 1}
				containerDeaths[cname] = c
			}
			mutex.Unlock()

		case "stop", "kill":
			log.Debug("Stop event ...", msg)
		case "start":
			log.Debug("Start event ...", msg)
		case "create":
			log.Debug("Create event ...", msg)
		case "destroy":
			log.Debug("Destroy event ...", msg)
		default:
			log.Debug("Default Event, network connect?:", msg.Status, ",", msg.ID, ";", msg.From, msg)

		}

	}
	close(quit)
	log.Info("Docker event loop closed") // todo: reconnect?

}

func showStatistics() {

	log.Info("Start Statisctics")

	var cs containers
	mutex.Lock()
	for i := range containerDeaths {
		cs = append(cs, containerDeaths[i])
	}
	mutex.Unlock()
	fmt.Println(cs)
	sort.Sort(cs)
	fmt.Println("Stats:")
	for _, k := range cs {
		fmt.Println(k.deaths, ";", k.name)
	}

}

func getFirstAlertTime(starttime string) (alarmTime time.Time) {

	hour, _ := strconv.Atoi(string([]rune(starttime)[0:2]))
	//	fmt.Println(err)

	minute, _ := strconv.Atoi(string([]rune(starttime)[3:]))
	//	fmt.Println(err)

	//	fmt.Println(string(hour))
	//fmt.Println(string(minute))
	now := time.Now()
	log.Info("Time Now:", now)

	alarmTime = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	//fmt.Println(now)
	//fmt.Println(alarmTime)
	log.Info("AlarmTime: ", alarmTime)
	if alarmTime.Before(now) {
		log.Info("alarmTime before now")
		alarmTime = alarmTime.Add(24 * time.Hour)
		//fmt.Println(now)

	}
	log.Info("AlarmTime: ", alarmTime)
	return

}
