// +build sifive

package interrupt

import "device/sifive"

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	sifive.PLIC.ENABLE[irq.num/32].SetBits(1 << (irq.num % 32))
}

// SetPriority sets the interrupt priority for this interrupt. A lower number
// means a higher priority. Additionally, most hardware doesn't implement all
// priority bits (only the uppoer bits).
//
// Examples: 0xff (lowest priority), 0xc0 (low priority), 0x00 (highest possible
// priority).
func (irq Interrupt) SetPriority(priority uint8) {
	sifive.PLIC.PRIORITY[irq.num].Set(uint32(priority))
}
