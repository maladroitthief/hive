package hive

import (
	"errors"
	"sync"
)

type Worker struct {
	mutex     sync.Mutex
	waitGroup *sync.WaitGroup
	stop      chan struct{}
	done      chan struct{}
}

func (w *Worker) Run(fn func(stop <-chan struct{})) (done func()) {
	if w == nil || fn == nil {
		panic(errors.New("worker cannot be nil"))
	}
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.stop == nil && w.done == nil {
		w.stop, w.done = make(chan struct{}), make(chan struct{})
		go w.wait()
		go w.run(fn)
	}

	if w.waitGroup == nil {
		w.waitGroup = new(sync.WaitGroup)
	}

	w.waitGroup.Add(1)
	return w.waitGroup.Done
}

func (w *Worker) wait() {
	for {
		w.mutex.Lock()
		wg := w.waitGroup
		if wg == nil {
			break
		}
		w.waitGroup = nil
		w.mutex.Unlock()
		wg.Wait()
	}

	close(w.stop)
	<-w.done
	w.stop, w.done = nil, nil
	w.mutex.Unlock()
}

func (w *Worker) run(fn func(stop <-chan struct{})) {
	fn(w.stop)
	close(w.done)
}
