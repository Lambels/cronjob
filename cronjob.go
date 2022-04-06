package cronjob

import (
	"context"
	"log"
	"os"
	"sync"
	"sync/atomic"
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
	nodes     chan chan []*Node
	runningMu sync.Mutex
	isRunning bool
}

type Schedule interface {
	// Calculate calculates the duartion till the next activation time.
	Calculate(time.Time) time.Duration
}

type CyclicSchedule interface {
	// MoveNextAvtivation re-calculates the next time the schedule will be activated
	// at.
	MoveNextAvtivation(time.Time)

	Schedule
}

type Scheduler interface {
	// NextCycle returns the duration to sleep before next activation cycle.
	NextCycle(time.Time) time.Duration

	// RunNow returns the jobs that need to be ran now and cleans the scheduler.
	RunNow(time.Time) []*Node

	// GetAll returns all the nodes in the scheduler.
	GetAll() []*Node

	// AddNode adds a new node to the scheduler.
	AddNode(time.Time, *Node)

	// RemoveNode removes node with id provided.
	RemoveNode(int)

	// Clean removes the node (field) and re-calculates appropriate nodes.
	Clean(time.Time, []*Node)
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
		nodes:     make(chan chan []*Node),
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

	if !c.isRunning {
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

// Stop stops the cronjobs processing thread.
//
// no-op if not running.
func (c *CronJob) Stop() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if !c.isRunning {
		return
	}

	c.stop <- struct{}{}
	c.isRunning = false
}

// StopWithFlush stops the cronjobs processing thread.
//
// runs all the current jobs and cleans the scheduler.
//
// no-op if not running.
func (c *CronJob) StopWithFlush() context.Context {
	c.runningMu.Lock()
	if !c.isRunning {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}

	c.stop <- struct{}{}
	c.isRunning = false
	c.runningMu.Unlock()

	// run jobs.
	nodes := c.Jobs()

	ctx, cancel := context.WithCancel(context.Background())
	if len(nodes) == 0 { // no nodes.
		cancel()
		return ctx
	}

	var runningWorkerCount int64 = int64(len(nodes))
	for _, node := range nodes {
		go func(node *Node) {
			node.Job.Run()
			c := atomic.AddInt64(&runningWorkerCount, -1)
			if c == 0 { // last job, cancel.
				cancel()
			}
		}(node)
	}

	// clean nodes.
	c.scheduler.Clean(c.Now(), nodes)

	return ctx
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

// Jobs returns the current nodes which are registered to the scheduler.
func (c *CronJob) Jobs() []*Node {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if c.isRunning {
		replyChan := make(chan []*Node, 1)
		c.nodes <- replyChan

		return <-replyChan
	} else {
		return c.scheduler.GetAll()
	}
}

func (c *CronJob) addJob(job *Job, schedule Schedule, confs ...JobConf) int {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	for _, conf := range confs {
		conf(job)
	}

	// add a job which will be ran on the first execution cycle (negative time.Duration).
	if job.runOnStart {
		node := &Node{
			Schedule: &constantSchedule{c.Now().Add(-1 * time.Second)},
			Job:      job,
		}

		if c.isRunning {
			go job.Run()
		} else {
			c.idCount++

			node.Id = c.idCount
			c.scheduler.AddNode(c.Now(), node)
		}
	}

	c.idCount++
	node := &Node{
		Id:       c.idCount,
		Schedule: schedule,
		Job:      job,
	}
	if c.isRunning {
		c.add <- node
	} else {
		c.scheduler.AddNode(c.Now(), node)
	}
	return node.Id
}

func (c *CronJob) run() {
	c.logger.Println("starting processing thread")
	now := c.Now()

	for {
		var timer *time.Timer
		if sleep := c.scheduler.NextCycle(now); sleep >= 0 {
			timer = time.NewTimer(sleep)
		} else {
			timer = time.NewTimer(1000000 * time.Hour)
		}

		for {
			select {
			case woke := <-timer.C:
				now = woke.In(c.location)

				// run all jobs.
				nodes := c.scheduler.RunNow(now)
				for _, node := range nodes {
					go node.Job.Run()
				}

				// clean nodes after running.
				c.scheduler.Clean(now, nodes)
				c.logDebugf("woke up at: %v\n", woke)

			case reply := <-c.nodes:
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
				c.logDebugf("attempting to remove node with id: %v\n", id)

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
