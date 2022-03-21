package cronjob

import "log"

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
