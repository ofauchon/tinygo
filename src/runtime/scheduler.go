package runtime

import "internal/task"

const schedulerDebug = false

var mainExited bool

// Simple logging, for debugging.
func scheduleLog(msg string) {
	if schedulerDebug {
		println("---", msg)
	}
}

// Simple logging with a task pointer, for debugging.
func scheduleLogTask(msg string, t *task.Task) {
	if schedulerDebug {
		println("---", msg, t)
	}
}

// Simple logging with a channel and task pointer.
func scheduleLogChan(msg string, ch *channel, t *task.Task) {
	if schedulerDebug {
		println("---", msg, ch, t)
	}
}

// Goexit terminates the currently running goroutine. No other goroutines are affected.
func Goexit() {
	panicOrGoexit(nil, panicGoexit)
}

//go:linkname fips_getIndicator crypto/internal/fips140.getIndicator
func fips_getIndicator() uint8 {
	return task.Current().FipsIndicator
}

//go:linkname fips_setIndicator crypto/internal/fips140.setIndicator
func fips_setIndicator(indicator uint8) {
	// This indicator is stored per goroutine.
	task.Current().FipsIndicator = indicator
}
