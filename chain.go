package cronjob

type Chain []func(FuncJob) FuncJob

func NewChain(c ...func(FuncJob) FuncJob) Chain {
	return Chain(c)
}

func (c Chain) Run(job FuncJob) {
	// decorate job.
	for i := range c {
		job = c[len(c)-i-1](job)
	}

	// run decorated job.
	job()
}
