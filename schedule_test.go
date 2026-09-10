package emailverifier

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Waits for firings rather than sleeping a fixed window: the third tick of a
// three-period sleep lands as the sleep expires, so an exact count is a coin
// flip. The count after stop() is settled by construction -- stop() sends on
// an unbuffered channel and the goroutine receives only while parked in its
// select, never mid job, so nothing can fire after it returns.
func TestStartScheduleOK(t *testing.T) {
	var ops int32
	fired := make(chan struct{}, 16)
	s := newSchedule(10*time.Millisecond, func() {
		atomic.AddInt32(&ops, 1)
		select {
		case fired <- struct{}{}:
		default:
		}
	})

	s.start()
	for i := 0; i < 3; i++ {
		select {
		case <-fired:
		case <-time.After(5 * time.Second):
			t.Fatalf("fired %d times in 5s, expected at least 3", atomic.LoadInt32(&ops))
		}
	}
	s.stop()

	settled := atomic.LoadInt32(&ops)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, settled, atomic.LoadInt32(&ops), "schedule kept firing after stop")
}

func TestNewScheduleOK(t *testing.T) {
	f := func() {}
	actual := newSchedule(time.Minute, f)

	assert.NotNil(t, actual)
}

func TestNewScheduleOK_FuncWithParams(t *testing.T) {
	f := func(a, b int) bool {
		return a > b
	}
	actual := newSchedule(time.Minute, f, 3, 4)

	assert.NotNil(t, actual)
	assert.Equal(t, []interface{}{3, 4}, actual.jobParams)
}

func TestNewScheduleWithWrongFunc(t *testing.T) {
	f := "test"
	actual := newSchedule(time.Minute, f)

	assert.Nil(t, actual.jobParams)
	assert.Equal(t, actual.jobFunc, f)
}

func TestRunScheduleFailedWithWrongFunc(t *testing.T) {
	f := "test"
	actual := newSchedule(time.Minute, f)
	assert.Equal(t, actual.jobFunc, f)
	assert.NotPanics(t, actual.start)
}
