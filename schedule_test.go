package emailverifier

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartScheduleOK(t *testing.T) {
	var ops uint32
	f := func() {
		atomic.AddUint32(&ops, 1)
	}

	s := newSchedule(time.Second, f)
	s.start()
	time.Sleep(time.Second * 3)
	s.stop()

	actual := atomic.LoadUint32(&ops)
	assert.Equal(t, uint32(3), actual)
	// assert.True(t, uint32(2) <= actual && actual <= uint32(3))
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
	assert.Equal(t, actual.jobParams, []interface{}{3, 4})
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
