package main

import (
	"fmt"
	"log"
	"github.com/fsouza/go-dockerclient"
)

func main() {
	endpoint := "unix:///var/run/docker.sock"

	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	log.Println("Start Event Listener für Docker Events...")
	events := make(chan *docker.APIEvents)
	client.AddEventListener(events)

	fmt.Println("Duh")

	quit := make(chan struct{})

	var containerDeaths map[string]int
	numContainerDeaths := 0

	containerDeaths = make(map[string]int)
	// Process Docker events
	for msg := range events {
		switch msg.Status {
		case "start":
			//log.Println("Start event ...", msg)

		case "die":
			numContainerDeaths++
			log.Println("Die event #", numContainerDeaths, "...", msg)
			id := msg.ID
			fmt.Println("ID:", id)
			var c *docker.Container
			var err error
			if id != "" {
				c, err = client.InspectContainer(id)
				if err != nil {
					fmt.Println(err)
				} else
				{
					fmt.Println("Container:", c.Name)
				}
			}
			i, ok := containerDeaths[c.Name]
			if ok {
				containerDeaths[c.Name] = i + 1
			} else {
				containerDeaths[c.Name] = 1
			}
			fmt.Println("Container ", c.Name, ": Deaths:", containerDeaths[c.Name])

		case "stop", "kill":
			//			log.Println("Stop event ...", msg)
		case "create":
			//			log.Println("Create event ...", msg)
		case "destroy":
			log.Println("Destroy event ...", msg)
		default:
			//			log.Println("Default Event, was ist denn das:", msg.Status, ",", msg.ID, ";", msg.From, "duh", msg)

		}

	}
	close(quit)
	log.Fatal("Docker event loop closed") // todo: reconnect?

}
