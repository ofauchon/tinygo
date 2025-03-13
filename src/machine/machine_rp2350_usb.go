//go:build rp2350

package machine

import (
	"device/rp"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

var (
	sendOnEP0DATADONE struct {
		offset int
		data   []byte
		pid    uint32
	}
)

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	// Reset usb controller
	resetBlock(rp.RESETS_RESET_USBCTRL)
	unresetBlockWait(rp.RESETS_RESET_USBCTRL)

	// Clear any previous state in dpram just in case
	_usbDPSRAM.clear()

	// Enable USB interrupt at processor
	rp.USB.INTE.Set(0)
	intr := interrupt.New(rp.IRQ_USBCTRL_IRQ, handleUSBIRQ)
	intr.SetPriority(0x00)
	intr.Enable()
	irqSet(rp.IRQ_USBCTRL_IRQ, true)

	// Mux the controller to the onboard usb phy
	rp.USB.USB_MUXING.Set(rp.USB_USB_MUXING_TO_PHY | rp.USB_USB_MUXING_SOFTCON)

	// Force VBUS detect so the device thinks it is plugged into a host
	rp.USB.USB_PWR.Set(rp.USB_USB_PWR_VBUS_DETECT | rp.USB_USB_PWR_VBUS_DETECT_OVERRIDE_EN)

	// Enable the USB controller in device mode.
	rp.USB.MAIN_CTRL.Set(rp.USB_MAIN_CTRL_CONTROLLER_EN)

	// Enable an interrupt per EP0 transaction
	rp.USB.SIE_CTRL.Set(rp.USB_SIE_CTRL_EP0_INT_1BUF)

	// Enable interrupts for when a buffer is done, when the bus is reset,
	// and when a setup packet is received
	rp.USB.INTE.Set(rp.USB_INTE_BUFF_STATUS |
		rp.USB_INTE_BUS_RESET |
		rp.USB_INTE_SETUP_REQ)

	// Present full speed device by enabling pull up on DP
	rp.USB.SIE_CTRL.SetBits(rp.USB_SIE_CTRL_PULLUP_EN)

	// 12.7.2 Disable phy isolation
	rp.USB.SetMAIN_CTRL_PHY_ISO(0x0)
}

func handleUSBIRQ(intr interrupt.Interrupt) {
	status := rp.USB.INTS.Get()

	// Setup packet received
	if (status & rp.USB_INTS_SETUP_REQ) > 0 {
		rp.USB.SIE_STATUS.Set(rp.USB_SIE_STATUS_SETUP_REC)
		setup := usb.NewSetup(_usbDPSRAM.setupBytes())

		ok := false
		if (setup.BmRequestType & usb.REQUEST_TYPE) == usb.REQUEST_STANDARD {
			// Standard Requests
			ok = handleStandardSetup(setup)
		} else {
			// Class Interface Requests
			if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
				ok = usbSetupHandler[setup.WIndex](setup)
			}
		}

		if !ok {
			// Stall endpoint?
			sendStallViaEPIn(0)
		}

	}

	// Buffer status, one or more buffers have completed
	if (status & rp.USB_INTS_BUFF_STATUS) > 0 {
		if sendOnEP0DATADONE.offset > 0 {
			ep := uint32(0)
			data := sendOnEP0DATADONE.data
			count := len(data) - sendOnEP0DATADONE.offset
			if ep == 0 && count > usb.EndpointPacketSize {
				count = usb.EndpointPacketSize
			}

			sendViaEPIn(ep, data[sendOnEP0DATADONE.offset:], count)
			sendOnEP0DATADONE.offset += count
			if sendOnEP0DATADONE.offset == len(data) {
				sendOnEP0DATADONE.offset = 0
			}
		}

		s2 := rp.USB.BUFF_STATUS.Get()

		// OUT (PC -> rp2040)
		for i := 0; i < 16; i++ {
			if s2&(1<<(i*2+1)) > 0 {
				buf := handleEndpointRx(uint32(i))
				if usbRxHandler[i] != nil {
					usbRxHandler[i](buf)
				}
				handleEndpointRxComplete(uint32(i))
			}
		}

		// IN (rp2040 -> PC)
		for i := 0; i < 16; i++ {
			if s2&(1<<(i*2)) > 0 {
				if usbTxHandler[i] != nil {
					usbTxHandler[i]()
				}
			}
		}

		rp.USB.BUFF_STATUS.Set(s2)
	}

	// Bus is reset
	if (status & rp.USB_INTS_BUS_RESET) > 0 {
		rp.USB.SIE_STATUS.Set(rp.USB_SIE_STATUS_BUS_RESET)
		//fixRP2040UsbDeviceEnumeration()

		rp.USB.ADDR_ENDP.Set(0)
		initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)
	}
}

func initEndpoint(ep, config uint32) {
	val := uint32(usbEpControlEnable) | uint32(usbEpControlInterruptPerBuff)
	offset := ep*2*usbBufferLen + 0x100
	val |= offset

	switch config {
	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn:
		val |= usbEpControlEndpointTypeInterrupt
		_usbDPSRAM.EPxControl[ep].In.Set(val)

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointOut:
		val |= usbEpControlEndpointTypeBulk
		_usbDPSRAM.EPxControl[ep].Out.Set(val)
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointOut:
		val |= usbEpControlEndpointTypeInterrupt
		_usbDPSRAM.EPxControl[ep].Out.Set(val)
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointIn:
		val |= usbEpControlEndpointTypeBulk
		_usbDPSRAM.EPxControl[ep].In.Set(val)

	case usb.ENDPOINT_TYPE_CONTROL:
		val |= usbEpControlEndpointTypeControl
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBuf0CtrlData1Pid)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	}
}

func handleUSBSetAddress(setup usb.Setup) bool {
	// Using 570μs timeout which is exactly the same as SAMD21.

	const ackTimeout = 570
	rp.USB.SIE_STATUS.Set(rp.USB_SIE_STATUS_ACK_REC)
	sendUSBPacket(0, []byte{}, 0)

	// Wait for transfer to complete with a timeout.
	t := timer.timeElapsed()
	for (rp.USB.SIE_STATUS.Get() & rp.USB_SIE_STATUS_ACK_REC) == 0 {
		if dt := timer.timeElapsed() - t; dt >= ackTimeout {
			return false
		}
	}

	// Set the device address to that requested by host.
	rp.USB.ADDR_ENDP.Set(uint32(setup.WValueL) & rp.USB_ADDR_ENDP_ADDRESS_Msk)
	return true
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data, 0)
	return true
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
	count := len(data)
	if 0 < int(maxsize) && int(maxsize) < count {
		count = int(maxsize)
	}

	if ep == 0 {
		if count > usb.EndpointPacketSize {
			count = usb.EndpointPacketSize

			sendOnEP0DATADONE.offset = count
			sendOnEP0DATADONE.data = data
		} else {
			sendOnEP0DATADONE.offset = 0
		}
		epXdata0[ep] = true
	}

	sendViaEPIn(ep, data, count)
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte
	ep := 0

	for !_usbDPSRAM.EPxBufferControl[ep].Out.HasBits(usbBuf0CtrlFull) {
		// TODO: timeout
	}

	ctrl := _usbDPSRAM.EPxBufferControl[ep].Out.Get()
	_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
	sz := ctrl & usbBuf0CtrlLenMask

	copy(b[:], _usbDPSRAM.EPxBuffer[ep].Buffer0[:sz])

	_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlData1Pid)
	_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	return b, nil
}

func handleEndpointRx(ep uint32) []byte {
	ctrl := _usbDPSRAM.EPxBufferControl[ep].Out.Get()
	_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
	sz := ctrl & usbBuf0CtrlLenMask

	return _usbDPSRAM.EPxBuffer[ep].Buffer0[:sz]
}

func handleEndpointRxComplete(ep uint32) {
	epXdata0[ep] = !epXdata0[ep]
	if epXdata0[ep] || ep == 0 {
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlData1Pid)
	}

	_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)
}

func SendZlp() {
	sendUSBPacket(0, []byte{}, 0)
}

func sendViaEPIn(ep uint32, data []byte, count int) {
	// Prepare buffer control register value
	val := uint32(count) | usbBuf0CtrlAvail

	// DATA0 or DATA1
	epXdata0[ep&0x7F] = !epXdata0[ep&0x7F]
	if !epXdata0[ep&0x7F] {
		val |= usbBuf0CtrlData1Pid
	}

	// Mark as full
	val |= usbBuf0CtrlFull

	copy(_usbDPSRAM.EPxBuffer[ep&0x7F].Buffer0[:], data[:count])
	_usbDPSRAM.EPxBufferControl[ep&0x7F].In.Set(val)
}

func sendStallViaEPIn(ep uint32) {
	// Prepare buffer control register value
	if ep == 0 {
		rp.USB.EP_STALL_ARM.Set(rp.USB_EP_STALL_ARM_EP0_IN)
	}
	val := uint32(usbBuf0CtrlFull)
	_usbDPSRAM.EPxBufferControl[ep&0x7F].In.Set(val)
	val |= uint32(usbBuf0CtrlStall)
	_usbDPSRAM.EPxBufferControl[ep&0x7F].In.Set(val)
}

type usbDPSRAM struct {
	// Note that EPxControl[0] is not EP0Control but 8-byte setup data.
	EPxControl [16]usbEndpointControlRegister

	EPxBufferControl [16]usbBufferControlRegister

	EPxBuffer [16]usbBuffer
}

type usbEndpointControlRegister struct {
	In  volatile.Register32
	Out volatile.Register32
}
type usbBufferControlRegister struct {
	In  volatile.Register32
	Out volatile.Register32
}

type usbBuffer struct {
	Buffer0 [usbBufferLen]byte
	Buffer1 [usbBufferLen]byte
}

var (
	_usbDPSRAM = (*usbDPSRAM)(unsafe.Pointer(uintptr(0x50100000)))
	epXdata0   [16]bool
	setupBytes [8]byte
)

func (d *usbDPSRAM) setupBytes() []byte {

	data := d.EPxControl[usb.CONTROL_ENDPOINT].In.Get()
	setupBytes[0] = byte(data)
	setupBytes[1] = byte(data >> 8)
	setupBytes[2] = byte(data >> 16)
	setupBytes[3] = byte(data >> 24)

	data = d.EPxControl[usb.CONTROL_ENDPOINT].Out.Get()
	setupBytes[4] = byte(data)
	setupBytes[5] = byte(data >> 8)
	setupBytes[6] = byte(data >> 16)
	setupBytes[7] = byte(data >> 24)

	return setupBytes[:]
}

func (d *usbDPSRAM) clear() {
	for i := 0; i < len(d.EPxControl); i++ {
		d.EPxControl[i].In.Set(0)
		d.EPxControl[i].Out.Set(0)
		d.EPxBufferControl[i].In.Set(0)
		d.EPxBufferControl[i].Out.Set(0)
	}
}

const (
	// DPRAM : Endpoint control register
	usbEpControlEnable                 = 0x80000000
	usbEpControlDoubleBuffered         = 0x40000000
	usbEpControlInterruptPerBuff       = 0x20000000
	usbEpControlInterruptPerDoubleBuff = 0x10000000
	usbEpControlEndpointType           = 0x0c000000
	usbEpControlInterruptOnStall       = 0x00020000
	usbEpControlInterruptOnNak         = 0x00010000
	usbEpControlBufferAddress          = 0x0000ffff

	usbEpControlEndpointTypeControl   = 0x00000000
	usbEpControlEndpointTypeISO       = 0x04000000
	usbEpControlEndpointTypeBulk      = 0x08000000
	usbEpControlEndpointTypeInterrupt = 0x0c000000

	// Endpoint buffer control bits
	usbBuf1CtrlFull     = 0x80000000
	usbBuf1CtrlLast     = 0x40000000
	usbBuf1CtrlData0Pid = 0x20000000
	usbBuf1CtrlData1Pid = 0x00000000
	usbBuf1CtrlSel      = 0x10000000
	usbBuf1CtrlStall    = 0x08000000
	usbBuf1CtrlAvail    = 0x04000000
	usbBuf1CtrlLenMask  = 0x03FF0000
	usbBuf0CtrlFull     = 0x00008000
	usbBuf0CtrlLast     = 0x00004000
	usbBuf0CtrlData0Pid = 0x00000000
	usbBuf0CtrlData1Pid = 0x00002000
	usbBuf0CtrlSel      = 0x00001000
	usbBuf0CtrlStall    = 0x00000800
	usbBuf0CtrlAvail    = 0x00000400
	usbBuf0CtrlLenMask  = 0x000003FF

	usbBufferLen = 64
)
