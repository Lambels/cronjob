package cronjob

import (
	"log"
	"time"
)

type CronJobConf func(*CronJob)

func WithLogger(logger *log.Logger) CronJobConf {
	return func(cj *CronJob) {
		cj.logger = logger
	}
}

func WithVerbose() CronJobConf {
	return func(cj *CronJob) {
		cj.verbose = true
	}
}

func WithLocation(loc *time.Location) CronJobConf {
	return func(cj *CronJob) {
		cj.location = loc
	}
}

type JobConf func(*Job)

func WithRunOnStart() JobConf {
	return func(j *Job) {
		j.runOnStart = true
	}
}

func WithChain(chain Chain) JobConf {
	return func(j *Job) {
		j.chain = chain
	}
}
