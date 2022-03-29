package cronjob_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Lambels/cronjob"
)

func TestChain(t *testing.T) {
	t.Parallel()
	var nums []int

	chain1 := func(num int) func(job cronjob.FuncJob) cronjob.FuncJob {
		return func(job cronjob.FuncJob) cronjob.FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(1)

	chain2 := func(num int) func(job cronjob.FuncJob) cronjob.FuncJob {
		return func(job cronjob.FuncJob) cronjob.FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(2)

	chain3 := func(num int) func(job cronjob.FuncJob) cronjob.FuncJob {
		return func(job cronjob.FuncJob) cronjob.FuncJob {
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

	cronjob.NewChain(chain1, chain2, chain3).Run(job)

	if got, want := nums, []int{1, 2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("want: %v got: %v\n", want, got)
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()
	t.Run("Test Timeout", func(t *testing.T) {
		t.Parallel()
		var count int

		job := func() error {
			count++
			return fmt.Errorf("error")
		}

		cronjob.NewChain(cronjob.Retry(100*time.Hour, 10)).Run(job)
		time.Sleep(1 * time.Second)

		if got, want := count, 1; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	})

	t.Run("Test Exit On Success", func(t *testing.T) {
		t.Parallel()
		var count int

		job := func() error {
			count++
			time.Sleep(2 * time.Second)
			if count%2 != 0 {
				return fmt.Errorf("error")
			}
			return nil
		}

		cronjob.NewChain(cronjob.Retry(2*time.Second, 4)).Run(job)
		// each retry takes 4 seconds (retry timeout + sleep in job)
		// we allow technically 3 retries to happen, although only 2 should happen.
		time.Sleep(12 * time.Second)

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

		cronjob.NewChain(cronjob.Retry(0, 4)).Run(job)
		// give time for technically more then 4 retries.
		time.Sleep(6 * time.Second)

		if got, want := count, 4; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	})

	t.Run("Test Retry With Decorators", func(t *testing.T) {
		t.Parallel()
		var count int

		incrementChain := func(fj cronjob.FuncJob) cronjob.FuncJob {
			return func() error {
				count++
				return fj()
			}
		}

		incrementChain2 := func(fj cronjob.FuncJob) cronjob.FuncJob {
			return func() error {
				count++
				count++
				return fj()
			}
		}

		job := func() error {
			return fmt.Errorf("error")
		}

		cronjob.NewChain(cronjob.Retry(0, 5), incrementChain, incrementChain2).Run(job)
		time.Sleep(5 * time.Second)

		if got, want := count, 15; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	})

}

func TestMergeChains(t *testing.T) {
	var nums []int

	chain1 := func(num int) func(job cronjob.FuncJob) cronjob.FuncJob {
		return func(job cronjob.FuncJob) cronjob.FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(1)

	chain2 := func(num int) func(job cronjob.FuncJob) cronjob.FuncJob {
		return func(job cronjob.FuncJob) cronjob.FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(2)

	chain3 := func(num int) func(job cronjob.FuncJob) cronjob.FuncJob {
		return func(job cronjob.FuncJob) cronjob.FuncJob {
			return func() error {
				nums = append(nums, num)
				return job()
			}
		}
	}(3)

	chain4 := func(num int) func(job cronjob.FuncJob) cronjob.FuncJob {
		return func(job cronjob.FuncJob) cronjob.FuncJob {
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

	merge1 := cronjob.NewChain(chain1, chain2)
	merge2 := cronjob.NewChain(chain3, chain4)

	cronjob.MergeChains(merge1, merge2).Run(job)

	if got, want := nums, []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("want: %v got: %v\n", want, got)
	}
}
