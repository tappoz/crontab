package crontab

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func myFunc() {
	fmt.Println("Helo, world")
}

func myFunc3() {
	fmt.Println("Noon!")
}

func myFunc2(s string, n int) {
	fmt.Printf("We have params here, string `%s` and nymber %d\n", s, n)
}

func TestJobError(t *testing.T) {

	ctab := New(true)

	if err := ctab.AddJob("* * * * *", myFunc, 10); err == nil {
		t.Error("This AddJob should return Error, wrong number of args")
	}

	if err := ctab.AddJob("* * * * *", nil); err == nil {
		t.Error("This AddJob should return Error, fn is nil")
	}

	var x int
	if err := ctab.AddJob("* * * * *", x); err == nil {
		t.Error("This AddJob should return Error, fn is not func kind")
	}

	if err := ctab.AddJob("* * * * *", myFunc2, "s", 10, 12); err == nil {
		t.Error("This AddJob should return Error, wrong number of args")
	}

	if err := ctab.AddJob("* * * * *", myFunc2, "s", "s2"); err == nil {
		t.Error("This AddJob should return Error, args are not the correct type")
	}

	if err := ctab.AddJob("* * * * * *", myFunc2, "s", "s2"); err == nil {
		t.Error("This AddJob should return Error, syntax error")
	}

	ctab.Shutdown()
}

var testN int
var testS string

func TestCrontab(t *testing.T) {
	testN = 0
	testS = ""

	ctab := Fake(2) // fake crontab wiht 5sec timer to speed up test

	var wg sync.WaitGroup
	wg.Add(2)

	if err := ctab.AddJob("* * * * *", func() { testN++; wg.Done() }); err != nil {
		t.Fatal(err)
	}

	if err := ctab.AddJob("* * * * *", func(s string) { testS = s; wg.Done() }, "param"); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
	}

	if testN != 1 {
		t.Error("func 1 not executed as scheduled")
	}

	if testS != "param" {
		t.Error("func 2 not executed as scheduled")
	}
	ctab.Shutdown()
}

type myRunAllStats struct {
	Counter int `json:"counter"`
}

func (mras *myRunAllStats) JSONString() string {
	jsonBytes, _ := json.Marshal(mras)
	return string(jsonBytes)
}

func (mras *myRunAllStats) ErrorMessage() string {
	// TODO
	return ""
}

func (mras *myRunAllStats) StatsMessage() string {
	// TODO
	return ""
}

var globalRunAllCounter int

func myRunAllFunc(statsChan chan ExecStats) {
	globalRunAllCounter++
	statsChan <- ExecStats{
		// ID to identify the job
		JobType: "myRunAllFunc",
		// custom execution stats
		Stats: &myRunAllStats{
			Counter: 1,
		},
	}
}

func TestRunAll(t *testing.T) {
	testN = 0
	testS = ""

	ctab := New(true)

	if err := ctab.AddJob("* * * * *", myRunAllFunc, ctab.StatsChan()); err != nil {
		t.Fatal(err)
	}

	ctab.RunAll()

	for i := 1; i <= 1; i++ {
		firstStatsStruct := <-ctab.StatsChan()
		if firstStatsStruct.JobType != "myRunAllFunc" {
			t.Errorf("Found an unexpected Job type")
		}
		customStuff := firstStatsStruct.Stats.JSONString()
		if !strings.Contains(customStuff, "\"counter\":1") {
			t.Error("func not executed on RunAll()")
		}
		if globalRunAllCounter != 1 {
			t.Error("func not executed on RunAll()")
		}
	}

	ctab.Clear()
	ctab.RunAll()

	if globalRunAllCounter != 1 {
		t.Error("Jobs not cleared!")
	}

	ctab.Shutdown()
}

type myTickCustomStats struct {
	TickTime time.Time `json:"tick_time"`
	TestN    int       `json:"test_n"`
}

func (mtcs *myTickCustomStats) JSONString() string {
	// jsonBytes, _ := json.MarshalIndent(mtcs, "", "  ")
	jsonBytes, _ := json.Marshal(mtcs)
	return string(jsonBytes)
}

func (mtcs *myTickCustomStats) ErrorMessage() string {
	// TODO
	return ""
}

func (mtcs *myTickCustomStats) StatsMessage() string {
	// TODO
	return ""
}

func myFuncWithTickCustomStats(statsChan chan ExecStats) {
	statsChan <- ExecStats{
		// ID to identify the job
		JobType: "myFuncWithTickCustomStats",
		// custom execution stats
		Stats: &myTickCustomStats{
			TickTime: time.Now(),
			TestN:    1,
		},
	}
}

func TestTicksAtTheBeginningOfMinute(t *testing.T) {
	log.Println("Testing the job checks trigger at the beginning of the minute (0 seconds, 000 milliseconds), so waiting AT MOST 1 minute for '* * * * *'...")
	ctab := New()

	ctab.MustAddJob("* * * * *", myFuncWithTickCustomStats, ctab.StatsChan())

	for i := 1; i <= 1; i++ {
		firstStatsStruct := <-ctab.StatsChan()
		if firstStatsStruct.JobType != "myFuncWithTickCustomStats" {
			t.Errorf("Found an unexpected Job type")
		}
		customStuff := firstStatsStruct.Stats.JSONString()
		log.Printf("The received stats: %v\n", customStuff)
		if !strings.Contains(customStuff, "\"test_n\":1") {
			t.Error("The func is not executed as scheduled")
		}
		if !strings.Contains(customStuff, ":00") { // TODO this is the seoncds: make it more robust
			t.Errorf("The func did not trigger at the beginning of the minute, found: %v", customStuff)
		}
		ctab.Shutdown()
	}
}
