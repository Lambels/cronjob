package cronjob_test

import (
	"bufio"
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/Lambels/cronjob"
)

func TestWithLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "[Test]", log.Flags())

	cron := cronjob.New(cronjob.WithLogger(logger))

	// start and stop should generate messages into the logger.
	cron.Start()
	cron.Stop()
	time.Sleep(1 * time.Second)

	if got, want := buf.Len(), 0; got <= want {
		t.Fatalf("got: %v want: non zero + positive", got)
	}
}

func TestWithVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "[Test]", log.Flags())

	cron := cronjob.New(cronjob.WithLogger(logger), cronjob.WithVerbose())

	// start and stop should generate messages into the logger.
	cron.Start()
	cron.AddFunc(func() error { return nil }, cronjob.In(cron.Now(), 5*time.Hour))
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
	cron := cronjob.New(cronjob.WithLocation(time.UTC))

	if got, want := cron.Location(), time.UTC; got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

/* Job Confs --------------------------------------------------------------------------- */

func TestWithChain(t *testing.T) {
	var count int

	cron := cronjob.New()

	cron.AddFunc(
		func() error { return nil },
		cronjob.In(cron.Now(), 1*time.Second),
		cronjob.WithChain(
			cronjob.NewChain(func(fj cronjob.FuncJob) cronjob.FuncJob {
				return func() error {
					count++
					return fj()
				}
			}),
		),
	)

	cron.Start()
	time.Sleep(1 * time.Second)
	cron.Stop()

	if got, want := count, 1; got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

func TestWithRunOnStart(t *testing.T) {
	t.Run("Without Chains", func(t *testing.T) {
		var count int

		cron := cronjob.New()

		cron.AddFunc(
			func() error { count++; return nil },
			cronjob.In(
				cron.Now(),
				1*time.Second,
			),
			cronjob.WithRunOnStart(),
		)

		cron.Start()
		time.Sleep(2 * time.Second)
		cron.Stop()

		if got, want := count, 2; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})

	t.Run("With Chains", func(t *testing.T) {
		var count int

		cron := cronjob.New()

		cron.AddFunc(
			func() error { return nil },
			cronjob.In(
				cron.Now(),
				1*time.Second,
			),
			cronjob.WithRunOnStart(),
			cronjob.WithChain(
				cronjob.NewChain(func(fj cronjob.FuncJob) cronjob.FuncJob {
					return func() error {
						count++
						return fj()
					}
				}),
			),
		)

		cron.Start()
		time.Sleep(2 * time.Second)
		cron.Stop()

		if got, want := count, 2; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})
}
