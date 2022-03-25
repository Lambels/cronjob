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
	verbose   bool
	idCount   int
	location  *time.Location
	add       chan *Node
	remove    chan int
	stop      chan struct{}
	jobs      chan chan []*Job
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

	// GetAll returns all the jobs in the scheduler.
	GetAll() []*Job

	// AddNode adds a new node to the scheduler.
	AddNode(time.Time, *Node)

	// RemoveNode removes node with id provided.
	RemoveNode(int)
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
		jobs:      make(chan chan []*Job),
	}

	for _, conf := range confs {
		conf(cronJob)
	}
	return cronJob
}

// Now returns the current time in the location used by the instance.
func (c *CronJob) Now() time.Time {
	return time.Now().In(c.location)
}

// AddFunc adds the function: cmd (field) to the execution cycle.
//
// can be called after starting the execution cycle or before.
//
//	(*CronJob).AddFunc(foo, cronjob.In(time.Now(), 4 * time.Hour))
// will schedule foo to run in 4 hours from time.Now()
func (c *CronJob) AddFunc(cmd FuncJob, schedule Schedule, confs ...JobConf) int {
	return c.addJob(&Job{job: cmd}, schedule, confs...)
}

// RemoveJob removes the job with id: id (field). (no-op if job not found)
//
// can be called after starting the execution cycle or before.
func (c *CronJob) RemoveJob(id int) {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if c.isRunning {
		c.scheduler.RemoveNode(id)
	} else {
		c.remove <- id
	}
}

// Location returns the location used by the instance.
func (c *CronJob) Location() *time.Location {
	return c.location
}

// Start the processing thread in its own gorutine.
//
// no-op if already running.
func (c *CronJob) Start() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.isRunning {
		return
	}
	c.isRunning = true
	go c.run()
}

// Start the processing thread.
//
// no-op if already running.
func (c *CronJob) Run() {
	c.runningMu.Lock()
	if c.isRunning {
		return
	}
	c.isRunning = true
	c.runningMu.Unlock()
	c.run()
}

// Jobs returns the current jobs which are registered to the scheduler.
func (c *CronJob) Jobs() []*Job {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if c.isRunning {
		replyChan := make(chan []*Job, 1)
		c.jobs <- replyChan

		return <-replyChan
	} else {
		return c.scheduler.GetAll()
	}
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
			Schedule: &constantSchedule{time.Time{}},
			Job:      job,
		}

		if c.isRunning {
			c.add <- node
		} else {
			c.scheduler.AddNode(c.Now(), node)
		}
	}

	node := &Node{
		Id:       c.idCount,
		Schedule: schedule,
		Job:      job,
	}
	if c.isRunning {
		c.scheduler.AddNode(c.Now(), node)
	} else {
		c.add <- node
	}
	return node.Id
}

func (c *CronJob) run() {
	c.logDebugf("started processing thread")
	now := c.Now()

	for {
		var timer *time.Timer
		if sleep := c.scheduler.NextCycle(now); sleep > 0 {
			timer = time.NewTimer(sleep)
		} else {
			timer = time.NewTimer(1000000 * time.Hour)
		}

		for {
			select {
			case now := <-timer.C:
				now = now.In(c.location)

				// run all jobs + clean.
				jobs := c.scheduler.GetNow(now)
				for _, job := range jobs {
					go job.Run()
				}

			case reply := <-c.jobs:
				reply <- c.scheduler.GetAll()
				continue // no need to re-calc timer.

			case node := <-c.add:
				timer.Stop()
				now = c.Now()

				c.scheduler.AddNode(now, node)
				c.logDebugf("added new node with id: %v\n", node.Id)

			case id := <-c.remove:
				timer.Stop()
				now = c.Now()

				c.scheduler.RemoveNode(id)
				c.logDebugf("atempting to remove node with id: %v\n", id)

			case <-c.stop:
				timer.Stop()

				c.logger.Println("exiticing processing thread")
				return
			}

			break
		}
	}
}

func (c *CronJob) logDebugf(format string, v ...interface{}) {
	if c.verbose {
		c.logger.Printf(format, v...)
	}
}

// Run runs the function provided to job with the chains.
func (j *Job) Run() {
	j.chain.Run(j.job)
}
