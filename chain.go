package cronjob

import (
	"time"
)

// Chain is a slice of decorator functions.
type Chain []func(FuncJob) FuncJob

// NewChain returns a chain of decorators which run in FIFO order.
func NewChain(c ...func(FuncJob) FuncJob) Chain {
	return Chain(c)
}

// MergeChains merges n (field) chains into 1 parent chain.
func MergeChains(n ...Chain) (merged Chain) {
	for _, chain := range n {
		merged = append(merged, chain...)
	}
	return
}

// Run runs job (field) with the chains.
func (c Chain) Run(job FuncJob) {
	// decorate job.
	for i := range c {
		job = c[len(c)-i-1](job)
	}

	// run decorated job.
	job()
}

// Retry will retry your job decorated with past chains max (field) times with a timeout (field)
// delay.
//
// Retry MUST only be added first in cronjob.NewChain() if you want all the chains to run properly.
// Not adding retry as the first argument in NewChain will cause unexpected behaviour.
func Retry(timeout time.Duration, max int) func(FuncJob) FuncJob {
	if max <= 0 {
		max = 1
	}
	if timeout < time.Second {
		timeout = time.Second
	}

	return func(fj FuncJob) FuncJob {
		// call chain from here.
		err := fj()

		if err != nil {

			ticker := time.NewTicker(timeout)

			// use 1 to compensate for first error checking call.
			for i := 1; i < max; i++ {
				select {
				case <-ticker.C:
					if err := fj(); err == nil {
						ticker.Stop()
						break
					}
					continue
				}
				break
			}
		}

		// ends chain.
		return func() error {
			return nil
		}
	}
}
