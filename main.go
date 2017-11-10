package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"regexp"
	"gopkg.in/alecthomas/kingpin.v2"
	"time"
	log "github.com/Sirupsen/logrus"
)

func main() {



	var (
		verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
		interval    = kingpin.Arg("interval", "Statistics every <interval> minutes.").Default("3").Int()
        logLevel    = kingpin.Flag("logLevel", "LogLevel for Program").Default("DEBUG").Enum("DEBUG", "WARNING", "ERROR")

	)


	kingpin.Parse()
	var containerDeaths map[string]int
	containerDeaths = make(map[string]int)
    fmt.Println("", *verbose, *interval)

	switch *logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)

	case "WARNING":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)

	}


    showStatitics(*interval, containerDeaths)
	//pattern für services, containername ist vorne.number.id
	servicePattern := regexp.MustCompile("\\.([0-9]+)\\.([0-9a-z]+)$")

	endpoint := "unix:///var/run/docker.sock"

	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	log.Println("Start Event Listener für Docker Events...")
	events := make(chan *docker.APIEvents)
	client.AddEventListener(events)

	quit := make(chan struct{})


	numContainerDeaths := 0

	// Process Docker events
	for msg := range events {
		switch msg.Status {
		case "start":
			//log.Println("Start event ...", msg)

		case "die":
			numContainerDeaths++
			//log.Println("Die event #", numContainerDeaths, "...", msg)
			id := msg.ID
			var c *docker.Container
			var err error
			if id != "" {
				c, err = client.InspectContainer(id)
				if err != nil {
					fmt.Println(err)
				} else
				{
					fmt.Println("Container died, name:", c.Name, " Id:", id)
				}
			}
			name := servicePattern.ReplaceAllString(c.Name, "")
			i, ok := containerDeaths[name]
			if ok {
				containerDeaths[name] = i + 1
			} else {
				containerDeaths[name] = 1
			}
			fmt.Println("Container ", name, ": Deaths:", containerDeaths[name])

		case "stop", "kill":
			//			log.Println("Stop event ...", msg)
		case "create":
			//			log.Println("Create event ...", msg)
		case "destroy":
			//log.Println("Destroy event ...", msg)
		default:
			//			log.Println("Default Event, was ist denn das:", msg.Status, ",", msg.ID, ";", msg.From, "duh", msg)

		}

	}
	close(quit)
	log.Fatal("Docker event loop closed") // todo: reconnect?

}


func showStatitics(interval int, containerDeaths map[string]int) {

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	go func() {
		for {
			select {
			    case <- ticker.C:
			    	fmt.Println("Stats:")
			    	if containerDeaths != nil {
						for i,j := range containerDeaths {
							fmt.Println(i,j)
						}
						fmt.Println(containerDeaths)
					}
			    	fmt.Println("empty")

			}
		}
	}()
}