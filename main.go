// Package detects and reports a summary about dying containers
package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

// container is a struct with the name and number of deaths of that container
// the name can be the actual name (useful when docker run --name <> has been used or the image name
type container struct {
	name   string
	deaths int
}

// Stringer
func (c container) String() string {
	return fmt.Sprintf("{Name:%s, Deaths:%d}", c.name, c.deaths)
}

var mutex = &sync.Mutex{}

// containers is a type used for storing and sorting the containers
type containers []container

func (a containers) Len() int           { return len(a) }
func (a containers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a containers) Less(i, j int) bool { return a[i].deaths < a[j].deaths }

var containerDeathsByContainerName map[string]*container
var containerDeathsByImageName map[string]*container

// main starts the even collection loop and prints the statistics in a given intervall
// starts a signal handler that catches ctrl-c and prints the statistics
//
// Parameters
//    --interval <minutes>     interval in which the statistics is printed, default: 10
//    -- starttime [hh:mm]|now     time when the statistics is printed the first time, default: now
func main() {

	var (
		interval  = kingpin.Flag("interval", "Statistics every <interval> minutes.").Default("10").Int()
		starttime = kingpin.Flag("starttime", "Time when report should be printed [hh:mm|now]").Default("now").String()
		logLevel  = kingpin.Flag("logLevel", "LogLevel for Program").Default("INFO").Enum("DEBUG", "INFO", "WARNING", "ERROR")
	)
	kingpin.Parse()

	containerDeathsByContainerName = make(map[string]*container)
	containerDeathsByImageName = make(map[string]*container)
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

	log.Debug("TimeUntil: ", time.Until(firstAlert))

	go func() {
		<-timer.C
		log.Debug("Startzeitpunkt erreicht")
		showStatistics()
		ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
		go func() {
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

	endpoint := "/var/run/docker.sock"

	_, err := os.Stat(endpoint)
	if err != nil {
		log.Error("no docker socket available, shutting down ...")
		return
	}
	client, err := docker.NewClient("unix://" + endpoint)
	if err != nil {
		log.Error("cannot connect to docker, shutting down ...")
		panic(err)
	}

	log.Info("Start Event Listener für Docker Events...")
	events := make(chan *docker.APIEvents)
	err = client.AddEventListener(events)

	if err != nil {
		log.Error("cannot add event listener, shutting down ...")
		panic(err)
	}
	quit := make(chan struct{})

	numContainerDeaths := 0

	// Process Docker events
	for msg := range events {
		switch msg.Status {
		case "die":
			numContainerDeaths++
			log.Debug("Die event #", numContainerDeaths, "...", msg)
			id := msg.ID
			var dc *docker.Container
			var err error
			if id != "" {
				dc, err = client.InspectContainer(id)
				if err != nil {
					fmt.Println(err)
				} else {
					log.Debug("Container died, name:", dc.Name, " Id:", id, " Image: ", dc.Config.Image)
				}
			}
			cname := servicePattern.ReplaceAllString(dc.Name, "")
			cname = strings.TrimPrefix(cname, "/")
			mutex.Lock()
			c, ok := containerDeathsByContainerName[cname]
			if ok {
				c.deaths += 1
				//containerDeathsByContainerName[cname] = c
			} else {
				c = &container{name: cname, deaths: 1}
				containerDeathsByContainerName[cname] = c
			}
			imageName := string(dc.Config.Image)
			c, ok = containerDeathsByImageName[imageName]
			if ok {
				c.deaths += 1
				containerDeathsByImageName[imageName] = c
			} else {
				c = &container{name: imageName, deaths: 1}
				containerDeathsByImageName[imageName] = c
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
	log.Info("Docker event loop closed")

}

// Print the statistics about died containers to stdout
func showStatistics() {

	log.Info("Start Statisctics")

	var cs containers
	mutex.Lock()
	for i := range containerDeathsByContainerName {
		cs = append(cs, *containerDeathsByContainerName[i])
	}
	mutex.Unlock()
	sort.Sort(cs)
	fmt.Println("Stats:")
	for _, k := range cs {
		fmt.Println(k.deaths, ";", k.name)
	}

	cs = nil
	mutex.Lock()
	for i := range containerDeathsByImageName {
		cs = append(cs, *containerDeathsByImageName[i])
	}
	mutex.Unlock()
	sort.Sort(cs)
	fmt.Println("Stats:")
	for _, k := range cs {
		fmt.Println(k.deaths, ";", k.name)
	}
}

// compute the time when statistics should be printed
// if [hh:mm] today has already passed, the the time to tomorrow:[hh:mm]
func getFirstAlertTime(starttime string) (alarmTime time.Time) {

	if starttime == "now" {
		log.Info("Start immediately")
		alarmTime = time.Now()
		return
	}

	hour, _ := strconv.Atoi(string([]rune(starttime)[0:2]))
	//	fmt.Println(err)

	minute, _ := strconv.Atoi(string([]rune(starttime)[3:]))
	//	fmt.Println(err)

	//	fmt.Println(string(hour))
	//fmt.Println(string(minute))
	now := time.Now()
	alarmTime = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	//log.Info("AlarmTime: ", alarmTime)
	if alarmTime.Before(now) {
		log.Info("alarmTime before now")
		alarmTime = alarmTime.Add(24 * time.Hour)

	}
	return

}
