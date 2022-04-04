package cronjob

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAddFunc(t *testing.T) {
	t.Parallel()
	t.Run("While Running", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(1)

		c := New()
		c.Start()
		defer c.Stop()
		c.AddFunc(func() error { wg.Done(); return nil }, In(c.Now(), 1*time.Second))

		select {
		case <-wait(wg):
			// job ran.
		case <-time.After(2 * time.Second):
			t.Fatal("no job ran.")
		}
	})

	t.Run("While Stopped", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(1)

		c := New()
		c.AddFunc(func() error { wg.Done(); return nil }, In(c.Now(), 1*time.Second))
		c.Start()
		defer c.Stop()

		select {
		case <-wait(wg):
			// job ran.
		case <-time.After(2 * time.Second):
			t.Fatal("no job ran.")
		}
	})
}

func TestRemoveJob(t *testing.T) {
	t.Parallel()
	t.Run("While Running", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(1)

		c := New()
		id := c.AddFunc(func() error { wg.Done(); return nil }, In(c.Now(), 1*time.Second))
		c.Start()
		defer c.Stop()
		c.RemoveJob(id)

		select {
		case <-wait(wg):
			t.Fatal("job ran.")
		case <-time.After(2 * time.Second):
			// job was removed
		}
	})

	t.Run("While Stopped", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(1)

		c := New()
		id := c.AddFunc(func() error { wg.Done(); return nil }, In(c.Now(), 1*time.Second))
		c.RemoveJob(id)
		c.Start()
		defer c.Stop()

		select {
		case <-wait(wg):
			t.Fatal("job ran.")
		case <-time.After(2 * time.Second):
			// job was removed
		}
	})
}

func TestStopStopsJobsFromRunning(t *testing.T) {
	t.Parallel()
	t.Run("Before", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(1)

		c := New()
		c.AddFunc(func() error { wg.Done(); return nil }, In(c.Now(), 1*time.Second))
		c.Start()
		c.Stop()

		select {
		case <-wait(wg):
			t.Fatal("job ran.")
		case <-time.After(2 * time.Second):
			// job was removed
		}
	})

	t.Run("After", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(1)

		c := New()
		c.Start()
		c.AddFunc(func() error { wg.Done(); return nil }, In(c.Now(), 1*time.Second))
		c.Stop()

		select {
		case <-wait(wg):
			t.Fatal("job ran.")
		case <-time.After(2 * time.Second):
			// job was removed
		}
	})
}

func TestStopWithFlushWait(t *testing.T) {
	t.Parallel()
	t.Run("No Jobs", func(t *testing.T) {
		t.Parallel()
		c := New()

		c.Start()
		ctx := c.StopWithFlush()

		select {
		case <-ctx.Done():
			// canceled immediately

		case <-time.After(20 * time.Microsecond):
			t.Fatal("ctx wasnt canceled.")
		}
	})

	t.Run("One Job With Chains", func(t *testing.T) {
		t.Parallel()

		c := New()
		c.Start()
		c.AddFunc(
			func() error { return fmt.Errorf("ERR!") },
			In(c.Now(), 10*time.Second),
			WithChain(
				NewChain(Retry(2*time.Second, 3)),
			),
		)
		ctx := c.StopWithFlush()

		start := time.Now()
		select {
		case <-ctx.Done():
			if time.Since(start).Round(time.Second) < (4 * time.Second) {
				t.Fatal("ctx was cancelled to early.")
			}

		case <-time.After(5 * time.Second):
			t.Fatal("ctx was cancelled to late.")
		}
	})

	t.Run("2 Jobs, 1 fast 1 slow, waits for slow", func(t *testing.T) {
		t.Parallel()

		c := New()
		c.Start()
		c.AddFunc(
			func() error { time.Sleep(5 * time.Second); return nil },
			In(c.Now(), 5*time.Second),
		)
		c.AddFunc(
			func() error { time.Sleep(2 * time.Second); return nil },
			In(c.Now(), 5*time.Second),
		)
		ctx := c.StopWithFlush()

		start := time.Now()
		select {
		case <-ctx.Done():
			if time.Since(start) < (5 * time.Second) {
				t.Fatal("ctx cancelled to early.")
			}

		case <-time.After(6 * time.Second):
			t.Fatal("ctx cancelled to late.")
		}
	})

	t.Run("Stopped cronjob returns cancelled ctx", func(t *testing.T) {
		t.Parallel()

		c := New()
		c.Start()
		c.Stop()
		ctx := c.StopWithFlush()

		select {
		case <-ctx.Done():

		case <-time.After(20 * time.Microsecond):
			t.Fatal("ctx wasnt canceled.")
		}
	})
}

func wait(wg *sync.WaitGroup) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		ch <- struct{}{}
	}()
	return ch
}
