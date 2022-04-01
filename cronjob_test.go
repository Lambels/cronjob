package cronjob

import "sync"

func wait(wg *sync.WaitGroup) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		ch <- struct{}{}
	}()
	return ch
}
