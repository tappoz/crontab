package crontab

import (
	"log"
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
		customStuff := myExecStats.Stats().(*myCustomStats)
		if customStuff.strParam != "foo" {
			t.Errorf("Found an unexpected string parameter in the stats")
		}
		if customStuff.intParam != 42 {
			t.Errorf("Found an unexpected integer parameter in the stats")
		}
		ctab.Shutdown()
		log.Printf("Done with the test, the received stats: %+v", customStuff)
	}
}

// custom execution stats depending on the scheduled function
type myCustomStats struct {
	strParam string
	intParam int
}

func myFuncWithStats(statsChan chan ExecStats) {
	// work a bit...
	time.Sleep(1 * time.Second)
	// publish the execution stats...
	statsChan <- ExecStats{
		// ID to identify the job
		JobType: "myFuncWithStats",
		// custom execution stats
		Stats: func() interface{} {
			return &myCustomStats{
				strParam: "foo",
				intParam: 42,
			}
		},
	}
}
