package crontab

import (
	"fmt"
	"log"
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

	ctab := Fake(2) // fake crontab wiht 2sec timer to speed up test

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
	case <-time.After(5 * time.Second):
	}

	if testN != 1 {
		t.Error("func 1 not executed as scheduled")
	}

	if testS != "param" {
		t.Error("func 2 not executed as scheduled")
	}
	ctab.Shutdown()
}

func TestRunAll(t *testing.T) {
	testN = 0
	testS = ""

	ctab := New(true)

	if err := ctab.AddJob("* * * * *", func() { testN++ }); err != nil {
		t.Fatal(err)
	}

	if err := ctab.AddJob("* * * * *", func(s string) { testS = s }, "param"); err != nil {
		t.Fatal(err)
	}

	ctab.RunAll()
	// log.Println("Waiting for a minute...")
	time.Sleep(time.Second)

	if testN != 1 {
		t.Error("func not executed on RunAll()")
	}

	if testS != "param" {
		t.Error("func not executed on RunAll() or arg not passed")
	}

	ctab.Clear()
	ctab.RunAll()

	if testN != 1 {
		t.Error("Jobs not cleared")
	}

	if testS != "param" {
		t.Error("Jobs not cleared")
	}

	ctab.Shutdown()
}

type myTickCustomStats struct {
	tickTime time.Time
	testN    int
}

func myFuncWithTickCustomStats(statsChan chan ExecStats) {
	statsChan <- ExecStats{
		// ID to identify the job
		JobType: "myFuncWithTickCustomStats",
		// custom execution stats
		Stats: func() interface{} {
			return &myTickCustomStats{
				tickTime: time.Now(),
				testN:    1,
			}
		},
	}
}

func TestTicksAtTheBeginningOfMinute(t *testing.T) {
	log.Println("Testing the job checks trigger at the beginning of the minute (0 seconds), so waiting AT MOST 1 minute + the 1 minute iteration period for the schedule '* * * * *'...")
	ctab := New()

	ctab.MustAddJob("* * * * *", myFuncWithTickCustomStats, ctab.StatsChan())

	for i := 1; i <= 1; i++ {
		firstStatsStruct := <-ctab.StatsChan()
		if firstStatsStruct.JobType != "myFuncWithTickCustomStats" {
			t.Errorf("Found an unexpected Job type")
		}
		customStuff := firstStatsStruct.Stats().(*myTickCustomStats)
		log.Printf("The received timestamp in the execution stats: %v\n", customStuff.tickTime)
		if customStuff.testN != 1 {
			t.Error("The func is not executed as scheduled")
		}
		if customStuff.tickTime.Second() != 0 {
			t.Errorf("The func did not trigger at the beginning of the minute, found: %v", customStuff.tickTime)
		}
		ctab.Shutdown()
	}
}
