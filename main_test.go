package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestGetFirstAlertTimePastTime(t *testing.T) {
	now := time.Now()
	oneMinuteAgo := now.Add(-1 * time.Minute)
	oneMinuteAgo = time.Date(oneMinuteAgo.Year(), oneMinuteAgo.Month(), oneMinuteAgo.Day(), oneMinuteAgo.Hour(), oneMinuteAgo.Minute(), 0, 0, oneMinuteAgo.Location())
	fmt.Println(oneMinuteAgo.Hour())
	hour := strconv.Itoa(oneMinuteAgo.Hour())
	if len(hour) == 1 {
		hour = "0" + hour
	}
	minute := strconv.Itoa(oneMinuteAgo.Minute())
	if len(minute) == 1 {
		minute = "0" + minute
	}
	fmt.Println(hour + ":" + minute)
	starttime := getFirstAlertTime(hour + ":" + minute)
	fmt.Println(starttime.Hour())
	if starttime.Before(now) {
		t.Error("Starttime not in the future")
	}
	shouldBeStarttime := oneMinuteAgo.Add(24 * time.Hour)
	if !starttime.Equal(shouldBeStarttime) {
		t.Error("Wrong Starttime, expected:" + shouldBeStarttime.String() + " was: " + starttime.String())
	}
}

func TestGetFirstAlertTimeFutureTime(t *testing.T) {
	now := time.Now()
	inOneMinute := now.Add(1 * time.Minute)
	inOneMinute = time.Date(inOneMinute.Year(), inOneMinute.Month(), inOneMinute.Day(), inOneMinute.Hour(), inOneMinute.Minute(), 0, 0, inOneMinute.Location())
	fmt.Print(inOneMinute.Hour())
	hour := strconv.Itoa(inOneMinute.Hour())
	if len(hour) == 1 {
		hour = "0" + hour
	}
	minute := strconv.Itoa(inOneMinute.Minute())
	if len(minute) == 1 {
		minute = "0" + minute
	}
	fmt.Println(hour + ":" + minute)
	starttime := getFirstAlertTime(hour + ":" + minute)
	if starttime.Before(now) {
		t.Error("Starttime not in the future")
	}
	shouldBeStarttime := inOneMinute
	if !starttime.Equal(shouldBeStarttime) {
		t.Error("Wrong Starttime, expected:" + shouldBeStarttime.String() + " was: " + starttime.String())
	}
}
