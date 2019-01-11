package crontab

import (
	"log"
)

// Schedule represents a Job to be executed from a high-level point of view
// i.e is a unified interface that can be treated in a standard way by other
// parts, but allows to inject any kind of custom logic via the constructor
// `NewSchedule` which expect just a function of type `InvokeFunc` that could
// be provided containing any logic.
type Schedule interface {
	CrontabSyntax() string
	Invoke()
	Description() string
}

// InvokeFunc represent the function containing the logic to schedule as a job.
// It is handy because it allows to submit stats about the job execution
// in a unified way via the channel as definded in the `crontab` interface.
type InvokeFunc func(statsChan chan ExecStats)

// schedule represents (as a struct) the implementation of the Schedule interface.
type schedule struct {
	crontabSyntax     string
	description       string
	invokeFunc        InvokeFunc
	injectedStatsChan chan ExecStats
}

func (s *schedule) CrontabSyntax() string {
	return s.crontabSyntax
}

func (s *schedule) Description() string {
	return s.description
}

// Invoke allows to expose the job execution. It is going to invoke the
// injected function of type `InvokeFunc`.
func (s *schedule) Invoke() {
	// TODO perhaps move this to "Validate()" to be called by ScheduleAll
	if &s.invokeFunc == nil {
		log.Printf("WARNING, this schedule is not properly configured, the schedule description is '%s' and crontab syntax is '%s'", s.Description(), s.CrontabSyntax())
		return
		// otherwise runtime error: Crontab error runtime error: invalid memory address or nil pointer dereference
	}
	s.invokeFunc(s.injectedStatsChan)
}

// NewSchedule creates a new Schedule object containing common configuration parameters like
// the crontab syntax (e.g. `* * * * *`), a description for the schedule, the Invoke Func
// containing the job implementation and a global channel for stats shared between all the jobs.
func NewSchedule(crontabSyntax string, description string, invokeFunction InvokeFunc, statsChan chan ExecStats) Schedule {
	currSchedule := schedule{
		crontabSyntax:     crontabSyntax,
		description:       description,
		invokeFunc:        invokeFunction,
		injectedStatsChan: statsChan,
	}
	log.Printf("Returning a new schedule with description '%s', crontab syntax '%s' and expected-not-to-be-nil address to func: %v", currSchedule.Description(), currSchedule.CrontabSyntax(), &invokeFunction)
	return &currSchedule
}

// ScheduleAll is a static function invoking all the schedules implementing the interface.
// These schedule functions are invoked against the crontab instance
func ScheduleAll(ctab *Crontab, schedules ...Schedule) error {

	for idx, currSche := range schedules {
		currScheduleIdx := 1 + idx
		log.Printf("Processing schedule %d out of %d with crontab %s and description '%s'", currScheduleIdx, len(schedules), currSche.CrontabSyntax(), currSche.Description())
		ctab.MustAddJob(currSche.CrontabSyntax(), currSche.Invoke)
		log.Printf("Done scheduling the schedule %d out of %d", currScheduleIdx, len(schedules))
	}

	return nil
}

// StstsConsumerFunc provides a unified way of consuming execution stats from
// various independent schedules.
type StstsConsumerFunc func(stats ExecStats)

// BlockingStatsConsumer is a unified way of consuming stats
func BlockingStatsConsumer(ctab *Crontab, statsConsumerFunc StstsConsumerFunc) {
	for {
		currExecStats := <-ctab.StatsChan()
		statsConsumerFunc(currExecStats)
	}
}
