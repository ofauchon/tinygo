//go:build stm32
// +build stm32

package machine

import (
	"device/stm32"
)

var iwdgInitDone = false

const (
	IWDG_RELOAD = 0xFFF
)

func IWDGStart(timeoutMs uint32) {
	if !iwdgInitDone {
		initIWDG()
	}
	// TODO : Enable LSI

	// Reload the watchdog
	stm32.IWDG.KR.Set(0xaaaa)
	stm32.IWDG.KR.Set(0x5555)
	stm32.IWDG.PR.Set(0x10)
	stm32.IWDG.RLR.Set(IWDG_RELOAD)
	for stm32.IWDG.SR.Get() != 0 {
	}

}

func IWDGStop() {
}

func IWDGKick() {

}

func IWDGReloadValue() {
}
