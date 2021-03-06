package cronjob

import (
	"bufio"
	"bytes"
	"log"
	"sync"
	"testing"
	"time"
)

func TestWithLogger(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	logger := log.New(buf, "[Test]", log.Flags())

	cron := New(WithLogger(logger))

	// start and stop should generate messages into the logger.
	cron.Start()
	cron.Stop()
	time.Sleep(1 * time.Second)

	if got, want := buf.Len(), 0; got <= want {
		t.Fatalf("got: %v want: non zero + positive", got)
	}
}

func TestWithVerbose(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	logger := log.New(buf, "[Test]", log.Flags())

	cron := New(WithLogger(logger), WithVerbose())

	// start and stop should generate messages into the logger.
	cron.Start()
	cron.AddFunc(func() error { return nil }, In(cron.Now(), 5*time.Hour))
	cron.Stop()
	time.Sleep(1 * time.Second)

	reader := bufio.NewReader(buf)
	// advance first 2 lines.
	reader.ReadLine()
	reader.ReadLine()

	// expected 3rd line with verbose mode.
	if _, _, err := reader.ReadLine(); err != nil {
		t.Fatal(err)
	}
}

func TestWithLocation(t *testing.T) {
	t.Parallel()
	cron := New(WithLocation(time.UTC))

	if got, want := cron.Location(), time.UTC; got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

/* Job Confs --------------------------------------------------------------------------- */

func TestWithChain(t *testing.T) {
	t.Parallel()
	var count int

	cron := New()
	cron.AddFunc(
		func() error { return nil },
		In(cron.Now(), 1*time.Second),
		WithChain(
			NewChain(func(fj FuncJob) FuncJob {
				return func() error {
					count++
					return fj()
				}
			}),
		),
	)
	cron.Start()
	defer cron.Stop()

	time.Sleep(2 * time.Second)
	if got, want := count, 1; got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

func TestWithRunOnStart(t *testing.T) {
	t.Parallel()
	t.Run("Without Chains", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(1)

		cron := New()
		cron.AddFunc(
			func() error { wg.Done(); return nil },
			In(
				cron.Now(),
				10*time.Second,
			),
			WithRunOnStart(),
		)
		cron.Start()
		defer cron.Stop()

		select {
		case <-wait(wg):
			// ran.

		case <-time.After(1 * time.Second):
			t.Fatal("job didn't run.")
		}

	})

	t.Run("With Chains", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(2)

		cron := New()
		cron.AddFunc(
			func() error { wg.Done(); return nil },
			In(
				cron.Now(),
				10*time.Second,
			),
			WithRunOnStart(),
			WithChain(
				NewChain(func(fj FuncJob) FuncJob {
					return func() error {
						wg.Done()
						return fj()
					}
				}),
			),
		)
		cron.Start()
		defer cron.Stop()

		select {
		case <-wait(wg):
			// ran.

		case <-time.After(1 * time.Second):
			t.Fatal("job didn't run.")
		}
	})
}
