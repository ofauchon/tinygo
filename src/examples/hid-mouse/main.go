package main

import (
	"machine"
	"machine/usb/hid/mouse"
	"time"
)

func main() {
	println("1")
	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	println("2")
	mouse := mouse.New()

	println("3")

	for {
		if !button.Get() {
			for j := 0; j < 5; j++ {
				for i := 0; i < 100; i++ {
					mouse.Move(1, 0)
					time.Sleep(1 * time.Millisecond)
				}

				for i := 0; i < 100; i++ {
					mouse.Move(0, 1)
					time.Sleep(1 * time.Millisecond)
				}

				for i := 0; i < 100; i++ {
					mouse.Move(-1, -1)
					time.Sleep(1 * time.Millisecond)
				}
			}

			time.Sleep(1 * time.Second)
			println("X")
		}
			time.Sleep(1 * time.Second)
			println("X")
	}

}
