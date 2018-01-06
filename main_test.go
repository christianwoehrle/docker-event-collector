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
	fmt.Print(oneMinuteAgo.Hour())
	starttime := getFirstAlertTime("" + strconv.Itoa(oneMinuteAgo.Hour()) + ":" + strconv.Itoa(oneMinuteAgo.Minute()))
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
	starttime := getFirstAlertTime("" + string(inOneMinute.Hour()) + ":" + string(inOneMinute.Minute()))
	if starttime.Before(now) {
		t.Error("Starttime not in the future")
	}
	shouldBeStarttime := inOneMinute
	if !starttime.Equal(shouldBeStarttime) {
		t.Error("Wrong Starttime, expected:" + shouldBeStarttime.String() + " was: " + starttime.String())
	}
}
