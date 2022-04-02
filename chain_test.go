package cronjob

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestChain(t *testing.T) {
	t.Parallel()
	var nums []int

	chain1 := func(num int) func(job FuncJob) FuncJob {
		return func(job FuncJob) FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(1)

	chain2 := func(num int) func(job FuncJob) FuncJob {
		return func(job FuncJob) FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(2)

	chain3 := func(num int) func(job FuncJob) FuncJob {
		return func(job FuncJob) FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(3)

	job := func() error {
		nums = append(nums, 4)
		return nil
	}

	NewChain(chain1, chain2, chain3).Run(job)

	if got, want := nums, []int{1, 2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("want: %v got: %v\n", want, got)
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()
	t.Run("Test Timeout", func(t *testing.T) {
		t.Parallel()
		wg := &sync.WaitGroup{}
		wg.Add(2)

		job := func() error {
			wg.Done()
			return fmt.Errorf("error")
		}
		go NewChain(Retry(2*time.Second, 2)).Run(job)

		select {
		case <-wait(wg):
			t.Fatal("job ran.")

		case <-time.After(1 * time.Second):
			// timeout works and second job got delayed.
		}
	})

	t.Run("Test Exit On Success", func(t *testing.T) {
		t.Parallel()
		var count int

		job := func() error {
			count++
			if count%2 != 0 {
				return fmt.Errorf("error")
			}
			return nil
		}
		NewChain(Retry(0, 4)).Run(job)

		if got, want := count, 2; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	})

	t.Run("Test Max Retries", func(t *testing.T) {
		t.Parallel()
		var count int

		job := func() error {
			count++
			return fmt.Errorf("error")
		}
		NewChain(Retry(0, 4)).Run(job)

		if got, want := count, 4; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	})

	t.Run("Test Retry With Decorators", func(t *testing.T) {
		t.Parallel()
		var count int

		incrementChain := func(fj FuncJob) FuncJob {
			return func() error {
				count++
				return fj()
			}
		}
		incrementChain2 := func(fj FuncJob) FuncJob {
			return func() error {
				count++
				count++
				return fj()
			}
		}
		job := func() error {
			return fmt.Errorf("error")
		}
		NewChain(Retry(0, 5), incrementChain, incrementChain2).Run(job)

		if got, want := count, 15; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	})

}

func TestMergeChains(t *testing.T) {
	var nums []int

	chain1 := func(num int) func(job FuncJob) FuncJob {
		return func(job FuncJob) FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(1)
	chain2 := func(num int) func(job FuncJob) FuncJob {
		return func(job FuncJob) FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(2)
	chain3 := func(num int) func(job FuncJob) FuncJob {
		return func(job FuncJob) FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(3)
	chain4 := func(num int) func(job FuncJob) FuncJob {
		return func(job FuncJob) FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(4)
	job := func() error {
		nums = append(nums, 5)
		return nil
	}

	merge1 := NewChain(chain1, chain2)
	merge2 := NewChain(chain3, chain4)
	MergeChains(merge1, merge2).Run(job)

	if got, want := nums, []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("want: %v got: %v\n", want, got)
	}
}
