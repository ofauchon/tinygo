//go:build scheduler.none

package runtime

import "internal/task"

const hasScheduler = false

// No goroutines are allowed, so there's no parallelism anywhere.
const hasParallelism = false

// Set to true after main.main returns.
var mainExited bool

// run is called by the program entry point to execute the go program.
// With the "none" scheduler, init and the main function are invoked directly.
func run() {
	initHeap()
	initRand()
	initAll()
	callMain()
	mainExited = true
}

//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	sleepTicks(nanosecondsToTicks(duration))
}

func deadlock() {
	// The only goroutine available is deadlocked.
	runtimePanic("all goroutines are asleep - deadlock!")
}

func scheduleTask(t *task.Task) {
	// Pause() will panic, so this should not be reachable.
}

func Gosched() {
	// There are no other goroutines, so there's nothing to schedule.
}

func addTimer(tim *timerNode) {
	runtimePanic("timers not supported without a scheduler")
}

func removeTimer(tim *timer) bool {
	runtimePanic("timers not supported without a scheduler")
	return false
}

func schedulerRunQueue() *task.Queue {
	// This function is not actually used, it is only called when hasScheduler
	// is true.
	runtimePanic("unreachable: no runqueue without a scheduler")
	return nil
}

func scheduler(returnAtDeadlock bool) {
	// The scheduler should never be run when using -scheduler=none. Meaning,
	// this code should be unreachable.
	runtimePanic("unreachable: scheduler must not be called with the 'none' scheduler")
}

// getSystemStackPointer returns the current stack pointer of the system stack.
// This is always the current stack pointer.
func getSystemStackPointer() uintptr {
	return getCurrentStackPointer()
}
