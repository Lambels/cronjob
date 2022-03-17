package cronjob

import (
	"log"
	"os"
	"sync"
	"time"
)

type CronJob struct {
	scheduler Scheduler
	logger    *log.Logger
	idCount   int
	location  *time.Location
	add       chan *Node
	remove    chan int
	stop      chan struct{}
	runningMu sync.Mutex
	isRunning bool
}

type Schedule interface {
	// Calculate calculates the duartion till the next activation time.
	Calculate(time.Time) time.Duration
}

type Scheduler interface {
	// NextCycle returns the duration to sleep before next activation cycle.
	NextCycle(time.Time) time.Duration

	// GetNow returns the jobs that need to be ran now.
	GetNow(time.Time) []*Job

	// ScheduleJob schedules a new job.
	ScheduleJob(*Node) int

	// RemoveJob removes job with id provided.
	RemoveJob(int) error
}

type FuncJob func() error

type Job struct {
	job FuncJob

	chain Chain

	runOnStart bool
}

func New(confs ...CronJobConf) *CronJob {
	cronJob := &CronJob{
		scheduler: &linkedList{},
		logger:    log.New(os.Stdout, "[CronJob]", log.Flags()),
		location:  time.Local,
		add:       make(chan *Node),
		remove:    make(chan int),
		stop:      make(chan struct{}),
	}

	for _, conf := range confs {
		conf(cronJob)
	}
	return cronJob
}

func (c *CronJob) Now() time.Time {
	return time.Now().In(c.location)
}

func (c *CronJob) AddFunc(at time.Time, cmd FuncJob, confs ...JobConf) int {
	return c.addJob(&Job{job: cmd}, &ConstantSchedule{at}, confs...)
}

func (c *CronJob) AddCyclicFunc(every time.Duration, cmd FuncJob, confs ...JobConf) int {
	return c.addJob(&Job{job: cmd}, &CyclicSchedule{every: every}, confs...)
}

func (c *CronJob) AddFixedCyclicFunc(every time.Duration, cmd FuncJob, confs ...JobConf) int {
	return c.addJob(&Job{job: cmd}, &FixedCyclicSchedule{every}, confs...)
}

func (c *CronJob) RemoveJob(id int) {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if c.isRunning {
		c.scheduler.RemoveJob(id)
	} else {
		c.remove <- id
	}
}

func (c *CronJob) Location() *time.Location {
	return c.location
}

func (c *CronJob) addJob(job *Job, schedule Schedule, confs ...JobConf) int {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	c.idCount++
	for _, conf := range confs {
		conf(job)
	}

	// add a job which will be ran on the first execution cycle (negative time.Duration).
	if job.runOnStart {
		node := &Node{
			Schedule: &ConstantSchedule{time.Time{}},
			Job:      job,
		}

		if c.isRunning {
			c.add <- node
		} else {
			c.scheduler.ScheduleJob(node)
		}
	}

	node := &Node{
		Id:       c.idCount,
		Schedule: schedule,
		Job:      job,
	}
	if c.isRunning {
		c.scheduler.ScheduleJob(node)
	} else {
		c.add <- node
	}
	return node.Id
}

func (j *Job) Run() {
	j.chain.Run(j.job)
}
