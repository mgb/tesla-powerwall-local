package retrysync

import (
	"sync"
)

// Once is a resetable sync.Once. This allows you to only run f() once. When it becomes invalid, you can reset it allowing it to run once again.
type Once struct {
	once *sync.Once
	sync.RWMutex
}

func (o *Once) Do(f func()) {
	once := o.getOnce()

	once.Do(func() {
		// Once f() finishes, reset
		defer o.reset()

		f()
	})
}

func (o *Once) getOnce() *sync.Once {
	o.Lock()
	defer o.Unlock()

	if o.once == nil {
		o.once = &sync.Once{}
	}

	return o.once
}

func (o *Once) reset() {
	o.Lock()
	defer o.Unlock()

	o.once = nil
}
