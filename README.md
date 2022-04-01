# CronJob ![build](https://github.com/Lambels/cronjob/workflows/build/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/Lambels/cronjob)](https://goreportcard.com/report/github.com/Lambels/cronjob)
CronJob is like cron but uses golang [time](https://pkg.go.dev/time) specification.

### go get it:
```bash
go get github.com/Lambels/cronjob@latest
```

# Code Examples

## Scheduling functions.

Its very simple, the cronjob object has `AddFunc` method exposed, you use this to add your function to the scheduler.
`AddFunc` takes 3 parameters respectively: `FuncJob`, `Schedule`, `...JobConf`.
You can schedule functions before or after starting the processing thread.

### FuncJob:

`Job1` and `Job2` implement the type `FuncJob`

```go
func Job1() error {
    fmt.Println("Hello World, im a FuncJob")
    return nil
}

func Job2() error {
    fmt.Println("Hello World, im also a FuncJob")
    return fmt.Errorf("ERR")
}
```

### Schedules:

Schedules determine the time at which your `FunJob` runs at.

```go
package main

func main() {
    // runs in 5 seconds
    sched1 := cronjob.In(time.Now(), 5 * time.Second)

    // runs in 1 second
    sched2 := cronjob.In(time.Now(), 4 * time.Microsecond)

    // runs on 2022, 03, 16 at 16:18:59
    sched3 := cronjob.At(time.Date(
        2022,
        time.March,
        26,
        16,
        18,
        59,
        0,
        time.Local, // use cronjob location.
    ))

    // runs every hour.
    sched4 := cronjob.Every(1 * time.Hour)

    // runs on each 3 hour intervals: 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00, 24:00
    sched5 := cronjob.EveryFixed(3 * time.Hour)
}
```

### JobConf:

Job Configurations configure the behaviour of the job. Examples of such functions are found [here.](https://github.com/Lambels/cronjob/blob/main/conf.go)

A job configuration is a function with the signature `JobConf func(*Job)`.

```go
func Job1() error {
    fmt.Println("Hello World, im a FuncJob")
    return nil
}

func Job2() error {
    fmt.Println("Im a FuncJob which returns an error")
    return fmt.Errorf("ERR")
}

func main() {
    cron := cronjob.New()

    cron.AddFunc(
        Job1,
        cronjob.In(cron.Now(), 5 * time.Second),
        

        // configs:
        cronjob.WithRunOnStart(), // runs job on start.
    )

    cron.AddFunc(
        Job2,
        cronjob.In(cron.Now(), 10 * time.Second),

        // configs:
        cronjob.WithChain(
            // retry Job2 5 times in 5 second intervals.
            // always add cornjob.Retry() the first in the chain.
            cronjob.NewChain(cronjob.Retry(5 * time.Second, 5)),
        ),
    )
}
```

### All together:

```go
func Job1() error {
    fmt.Println("Hello World, im a FuncJob")
    return nil
}

func Job2() error {
    fmt.Println("Hello World, im also a FuncJob")
    return fmt.Errorf("ERR")
}

func main() {
    cron := cronjob.New()

    cron.AddFunc(Job1, cronjob.In(cron.Now(), 5 * time.Second), cronjob.WithRunOnStart())

    cron.Start()

    cron.AddFunc(Job2, cronjob.EveryFixed(cron.Now(), 1 * time.Hour))

    time.Sleep(1 * time.Hour)

    cron.Close()
}
```

## Chains:

Chains allow you to customize the behavour of the Jobs when the jobs are running. A chain function making up a chain has the following signature: `func(FuncJob) FuncJob`. A chain is just a slice of these functions: `type Chain []func(FuncJob) FuncJob`.

**cronjob.Retry() should always be added first in the chain to keep expected behaviour.**
> *Inspiration from [cron](https://github.com/robfig/cron)*

### Creating A Chain:
```go
func SomeChain(fj cronjob.FuncJob) cronjob.FuncJob {
	return func() error {
		log.Println("Hello from SomeChain")
        return fj() // call next function in chain.
	}
}

func SomeOtherChain(fj cronjob.FuncJob) cronjob.FuncJob {
	return func() error {
		log.Println("Hello from SomeOtherChain")
        return fj() // call next function in chain.
	}
}

func Job() error {
    log.Println("Hello from Job")
    return nil
}

func main() {
    chain := cronjob.NewChain(SomeChain, SomeOtherChain)
    chain.Run(job)

    // output:
    // "Hello from SomeChain"
    // "Hello from SomeOtherChain"
    // "Hello from Job"
}
```

### Merging n Chains:
```go
func SomeChain(fj cronjob.FuncJob) cronjob.FuncJob {
	return func() error {
		log.Println("Hello from SomeChain")
        return fj() // call next function in chain.
	}
}

func SomeOtherChain(fj cronjob.FuncJob) cronjob.FuncJob {
	return func() error {
		log.Println("Hello from SomeOtherChain")
        return fj() // call next function in chain.
	}
}

func SomeOtherOtherChain(fj cronjob.FuncJob) cronjob.FuncJob {
	return func() error {
		log.Println("Hello from SomeOtherOtherChain")
        return fj() // call next function in chain.
	}
}

func Job() error {
    log.Println("Hello from Job")
    return nil
}

func main() {
    chain1 := cronjob.NewChain(SomeChain, SomeOtherChain)
    chain2 := cronjob.NewChain(SomeOtherOtherChain)

    chain3 := cronjob.MergeChains(chain1, chain2)
    chain3.Run(Job)

    // output:
    // "Hello from SomeChain"
    // "Hello from SomeOtherChain"
    // "Hello from SomeOtherOtherChain"
    // "Hello from Job"
}
```

### Fault Tolerance:
**always add cronjob.Retry() first in the chain!**
```go
func SomeChain(fj cronjob.FuncJob) cronjob.FuncJob {
	return func() error {
		log.Println("Hello from SomeChain")
        return fj() // call next function in chain.
	}
}

func Job() error {
    log.Println("Hello from Job")
    return fmt.Errorf("ERR")
}

func main() {
    chain1 := cronjob.NewChain(cronjob.Retry(5*time.Second, 5), SomeChain)

    chain1.Run(Job)

    // output:
    // "Hello from SomeChain"
    // "Hello from Job"
    // "Hello from SomeChain"
    // "Hello from Job"
    // "Hello from SomeChain"
    // "Hello from Job"
    // "Hello from SomeChain"
    // "Hello from Job"
    // "Hello from SomeChain"
    // "Hello from Job"
}
```

## Removing Jobs:

The cronjob object has `RemoveJob` method exposed, it takes the job id as a parameter. `RemoveJob` will no-op if no job matches the id. You can call `RemoveJob` either after starting the processing thread or before.

### Get Job ID:
```go
func Job1() error {
    fmt.Println("Hello World, im a FuncJob")
    return nil
}

func main() {
    cron := cronjob.New()

    id := cron.AddFunc(Job1, cronjob.In(cron.Now(), 5 * time.Second))

    // store id ...
}
```

### Remove Job:
```go
func Job1() error {
    fmt.Println("Hello World, im a FuncJob")
    return nil
}

func main() {
    cron := cronjob.New()

    id := cron.AddFunc(Job1, cronjob.In(cron.Now(), 5 * time.Second))

    cron.RemoveJob(id)

    id := cron.AddFunc(Job1, cronjob.In(cron.Now(), 5 * time.Second))

    cron.Start()
    defer cron.Stop()
    
    cron.RemoveJob(id)
}
```

## Stopping:
There are 2 ways to stop a cronjob's processing thread, `Stop` and `StopWithFlush`. `Stop` exits the processing thread and `StopWithFlush` exits the processing thread and runs the remaining jobs providing a context to wait for their completion.

### Stop:
```go
func Job1() error {
    fmt.Println("Hello World, im a FuncJob")
    return nil
}

func main() {
    cron := cronjob.New()

    cron.AddFunc(Job1, cronjob.In(cron.Now(), 5 * time.Second))

    cron.Start()
    cron.Stop()

    cron.AddFunc(Job1, cronjob.In(cron.Now(), 2 * time.Second)) // still works.
}
```

### StopWithFlush:
```go
func Job1() error {
    time.Sleep(1 * time.Hour)
    return nil
}
func Job2() error {
    time.Sleep(5 * time.Second)
    return nil
}

func main() {
    cron := cronjob.New()

    cron.AddFunc(Job1, cronjob.In(cron.Now(), 5 * time.Second))
    cron.AddFunc(Job1, cronjob.In(cron.Now(), 5 * time.Second))

    cron.Start()
    ctx := cron.StopWithFlush()
    <-ctx.Done() // waits for Job1 and Job2 to finish. (1 hour)

    cron.AddFunc(Job1, cronjob.In(cron.Now(), 2 * time.Second)) // still works.
}
```