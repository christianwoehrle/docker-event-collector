# christianwoehrle/docker-event-collector

Simple Collector and Reporter of Docker events

[![GoDoc](https://godoc.org/github.com/christianwoehrle/docker-event-collector?status.svg)](https://godoc.org/github.com/christianwoehrle/docker-event-collector)
[![CircleCI](https://img.shields.io/circleci/project/github/christianwoehrle/docker-event-collector.svg)](https://circleci.com/gh/christianwoehrle/docker-event-collector)

I had the problem that a development server was very busy with instable containers that were restarted all the time.

THis programm gets all the docker events, collects the names of containers that are restarted and prints a
summary of the containers and #restarts at a given interval.

