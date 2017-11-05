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

	// Process Docker events
	for msg := range events {
		switch msg.Status {
		case "start":
			log.Println("Start event ...")
			log.Println("sleeping 3s ...")


		case "die":
			log.Println("Die event ...")

		case "stop", "kill":
			log.Println("Stop event ...")
		default:
			log.Println("Default Event, was ist denn das:" , msg.Status, msg.ID, msg.From, "duh")

		}

	}
	close(quit)
	log.Fatal("Docker event loop closed") // todo: reconnect?

}
