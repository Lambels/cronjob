package cronjob

import (
	"log"
	"time"
)

// CronJobConf represents a function to configure the behaviour of cronjob.
type CronJobConf func(*CronJob)

// WithLogger overwrites the default logger.
func WithLogger(logger *log.Logger) CronJobConf {
	return func(cj *CronJob) {
		cj.logger = logger
	}
}

// WithVerbose puts cronjob in verbose mode.
func WithVerbose() CronJobConf {
	return func(cj *CronJob) {
		cj.verbose = true
	}
}

// WithLocation sets the location used by cronjob.
func WithLocation(loc *time.Location) CronJobConf {
	return func(cj *CronJob) {
		cj.location = loc
	}
}

// JobConf represents a function to configure the behaviour of a job.
type JobConf func(*Job)

// WithRunOnStart makes the job run on start.
//
// if running: run when added.
//
// if not running: run when cronjob starts.
//
// note: the job will run with any previous configuration provided.
func WithRunOnStart() JobConf {
	return func(j *Job) {
		j.runOnStart = true
	}
}

// WithChain sets the chains to run with the job.
func WithChain(chain Chain) JobConf {
	return func(j *Job) {
		j.chain = chain
	}
}
