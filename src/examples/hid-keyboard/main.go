package main

import (
	"machine"
	"machine/usb/hid/keyboard"
	"time"
)

func main() {


	cnt:=0

	println("Hello")
	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	kb := keyboard.New()
	_=kb

	for {
		cnt++
		println("OK",cnt)
		if !button.Get() {
		//	kb.Write([]byte("tinygo"))
			time.Sleep(200 * time.Millisecond)
		}
	time.Sleep(time.Second)
	}
}
