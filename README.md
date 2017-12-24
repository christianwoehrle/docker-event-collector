# christianwoehrle/docker-event-collector

Simple Collector and Reporter of Docker events

[![GoDoc](https://godoc.org/github.com/christianwoehrle/docker-event-collector?status.svg)](https://godoc.org/github.com/christianwoehrle/docker-event-collector)
[![CircleCI](https://img.shields.io/circleci/project/github/christianwoehrle/docker-event-collector.svg)](https://circleci.com/gh/christianwoehrle/docker-event-collector)
[![Go Report Card](https://goreportcard.com/badge/github.com/christianwoehrle/docker-event-collector)](https://goreportcard.com/report/github.com/christianwoehrle/docker-event-collector)

I had the problem that a development server was very busy with instable containers that were restarted all the time.

This programm collects all the docker events, stores the names of containers that are restarted and prints a
summary of the containers and number of restarts at a given interval.

Program can be called with Parameters
--interval minutes --starttime [hh:mm|now]

i.e.
--interval 60 --starttime [15:00|now]

Default is starttime now and interval 10 Minutes.

To run the docker-container, you have to mount in the docker-socket so that the container can collect the events:

docker run -v /var/run/docker.sock:/var/run/docker.sock christianwoehrle/docker-event-collector:latest --interval 1

