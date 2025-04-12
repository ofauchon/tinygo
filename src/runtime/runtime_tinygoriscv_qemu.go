//go:build tinygo.riscv && virt && qemu

package runtime

import (
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

// This file implements the VirtIO RISC-V interface implemented in QEMU, which
// is an interface designed for emulation.

//export main
func main() {
	preinit()

	// Set the interrupt address.
	// Note that this address must be aligned specially, otherwise the MODE bits
	// of MTVEC won't be zero.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&handleInterruptASM)))

	// Enable global interrupts now that they've been set up.
	// This is currently only for timer interrupts.
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

	run()
	exit(0)
}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	code := uint(cause &^ (1 << 31))
	if cause&(1<<31) != 0 {
		// Topmost bit is set, which means that it is an interrupt.
		switch code {
		case riscv.MachineTimerInterrupt:
			// Signal timeout.
			timerWakeup.Set(1)
			// Disable the timer, to avoid triggering the interrupt right after
			// this interrupt returns.
			riscv.MIE.ClearBits(riscv.MIE_MTIE)
		}
	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(code)
	}

	// Zero MCAUSE so that it can later be used to see whether we're in an
	// interrupt or not.
	riscv.MCAUSE.Set(0)
}

// One tick is 100ns by default in QEMU.
// (This is not a standard, just the default used by QEMU).
func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 100 // one tick is 100ns
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 100) // one tick is 100ns
}

var timerWakeup volatile.Register8

func sleepTicks(d timeUnit) {
	// Enable the timer.
	target := uint64(ticks() + d)
	aclintMTIMECMP.Set(target)
	riscv.MIE.SetBits(riscv.MIE_MTIE)

	// Wait until it fires.
	for {
		if timerWakeup.Get() != 0 {
			timerWakeup.Set(0)
			// Disable timer.
			break
		}
		riscv.Asm("wfi")
	}
}

func ticks() timeUnit {
	// Combining the low bits and the high bits (at a rate of 100ns per tick)
	// yields a time span of over 59930 years without counter rollover.
	highBits := aclintMTIME.high.Get()
	for {
		lowBits := aclintMTIME.low.Get()
		newHighBits := aclintMTIME.high.Get()
		if newHighBits == highBits {
			// High bits stayed the same.
			return timeUnit(lowBits) | (timeUnit(highBits) << 32)
		}
		// Retry, because there was a rollover in the low bits (happening every
		// 429 days).
		highBits = newHighBits
	}
}

// Memory-mapped I/O as defined by QEMU.
// Source: https://github.com/qemu/qemu/blob/master/hw/riscv/virt.c
// Technically this is an implementation detail but hopefully they won't change
// the memory-mapped I/O registers.
var (
	// UART0 output register.
	stdoutWrite = (*volatile.Register8)(unsafe.Pointer(uintptr(0x10000000)))
	// SiFive test finisher
	testFinisher = (*volatile.Register32)(unsafe.Pointer(uintptr(0x100000)))

	// RISC-V Advanced Core Local Interruptor.
	// It is backwards compatible with the SiFive CLINT.
	// https://github.com/riscvarchive/riscv-aclint/blob/main/riscv-aclint.adoc
	aclintMTIME = (*struct {
		low  volatile.Register32
		high volatile.Register32
	})(unsafe.Pointer(uintptr(0x0200_bff8)))
	aclintMTIMECMP = (*volatile.Register64)(unsafe.Pointer(uintptr(0x0200_4000)))
)

func putchar(c byte) {
	stdoutWrite.Set(uint8(c))
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

func abort() {
	exit(1)
}

func exit(code int) {
	// Make sure the QEMU process exits.
	if code == 0 {
		testFinisher.Set(0x5555) // FINISHER_PASS
	} else {
		// Exit code is stored in the upper 16 bits of the 32 bit value.
		testFinisher.Set(uint32(code)<<16 | 0x3333) // FINISHER_FAIL
	}

	// Lock up forever (as a fallback).
	for {
		riscv.Asm("wfi")
	}
}

// handleException is called from the interrupt handler for any exception.
// Exceptions can be things like illegal instructions, invalid memory
// read/write, and similar issues.
func handleException(code uint) {
	// For a list of exception codes, see:
	// https://content.riscv.org/wp-content/uploads/2019/08/riscv-privileged-20190608-1.pdf#page=49
	print("fatal error: exception with mcause=")
	print(code)
	print(" pc=")
	print(riscv.MEPC.Get())
	println()
	abort()
}
