package crontab

import (
	"encoding/json"
	"log"
	"strings"
	"testing"
	"time"
)

// go test -v ./... -run TestExampleExecStats
func TestExampleExecStats(t *testing.T) {
	ctab := Fake(2) // fake crontab wiht 2sec timer to speed up test

	ctab.MustAddJob("* * * * *", myFuncWithStats, ctab.StatsChan())

	for i := 1; i <= 1; i++ {
		myExecStats := <-ctab.StatsChan()
		if myExecStats.JobType != "myFuncWithStats" {
			t.Errorf("Found an unexpected Job type")
		}
		customStuff := myExecStats.Stats.JSONString()
		if !strings.Contains(customStuff, "\"str_param\":\"foo\"") {
			t.Errorf("Found an unexpected string parameter in the stats: %s", customStuff)
		}
		if !strings.Contains(customStuff, "\"int_param\":42") {
			t.Errorf("Found an unexpected integer parameter in the stats: %s", customStuff)
		}
		ctab.Shutdown()
		log.Printf("Done with the test, the received stats: %+v", customStuff)
	}
}

// custom execution stats depending on the scheduled function
type myCustomStats struct {
	StrParam string `json:"str_param"`
	IntParam int    `json:"int_param"`
}

func (mcs *myCustomStats) JSONString() string {
	jsonBytes, _ := json.Marshal(mcs)
	jsonStr := string(jsonBytes)
	return jsonStr
}

func (mcs *myCustomStats) ErrorMessage() string {
	// TODO
	return ""
}

func (mcs *myCustomStats) StatsMessage() string {
	// TODO
	return ""
}

func myFuncWithStats(statsChan chan ExecStats) {
	// work a bit...
	time.Sleep(1 * time.Second)
	// publish the execution stats...
	statsChan <- ExecStats{
		// ID to identify the job
		JobType: "myFuncWithStats",
		// custom execution stats
		Stats: &myCustomStats{
			StrParam: "foo",
			IntParam: 42,
		},
	}
}
