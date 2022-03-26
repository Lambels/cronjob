# CronJob

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

Job Configurations configure the behaviour of the job. Examples of such functions are found [here](https://github.com/Lambels/cronjob/blob/main/conf.go)

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
    time.Sleep(5 * time.Hour)
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
    <-ctx // waits for Job1 and Job2 to finish. (1 hour)

    cron.AddFunc(Job1, cronjob.In(cron.Now(), 2 * time.Second)) // still works.
}
```