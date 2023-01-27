package retrysync

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOnce_noReset(t *testing.T) {
	t.Parallel()

	var o Once
	var i int64

	f := func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt64(&i, 1)
	}

	var wg sync.WaitGroup
	for n := 0; n < 5; n++ {
		wg.Add(1)
		go func() {
			o.Do(f)

			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(1), i)
}

func TestOnce_multipleResets(t *testing.T) {
	t.Parallel()

	var o Once
	var i int64

	f := func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt64(&i, 1)
	}

	for x := int64(1); x <= int64(5); x++ {
		var wg sync.WaitGroup
		for n := 0; n < 5; n++ {
			wg.Add(1)
			go func() {
				o.Do(f)

				wg.Done()
			}()
		}

		wg.Wait()

		assert.Equal(t, x, i)
	}
}
