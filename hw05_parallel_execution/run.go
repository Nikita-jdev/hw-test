package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	taskCh := make(chan Task)
	var wg sync.WaitGroup
	var errorsCount int32
	stopCh := make(chan struct{})
	var once sync.Once

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				case task, ok := <-taskCh:
					if !ok {
						return
					}
					if err := task(); err != nil {
						if atomic.AddInt32(&errorsCount, 1) >= int32(m) {
							once.Do(func() { close(stopCh) })
						}
					}
				}
			}
		}()
	}

	for _, task := range tasks {
		select {
		case <-stopCh:
			close(taskCh)
			wg.Wait()
			return ErrErrorsLimitExceeded
		case taskCh <- task:
		}
	}

	close(taskCh)
	wg.Wait()
	return nil
}
