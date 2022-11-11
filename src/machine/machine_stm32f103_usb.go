//go:build stm32 && stm32f103
// +build stm32,stm32f103

package machine

import (
	"device/stm32"
	"machine/usb"
	"runtime/interrupt"
	//
	//	"unsafe"
)

const (
	dbg bool = true
	//USB      = (*uint16)(unsafe.Pointer(uintptr(0x40006000)))
)

func dp(a string) {
	if dbg {
		println(a)
	}
}

// Configure the USB peripheral. The config is here for compatibility with the UART interface
func (dev *USBDevice) Configure(config UARTConfig) {
	dp("Configure")

	// Enable USB Clock
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USBEN)    // Enable USB Clock
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_USBRST) // Reset USB
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_USBRST)
	stm32.USB.CNTR.Set(stm32.USB_CNTR_CTRM | stm32.USB_CNTR_RESETM | stm32.USB_CNTR_ERRM | stm32.USB_CNTR_SOFM | stm32.USB_CNTR_SUSPM | stm32.USB_CNTR_WKUPM)

	// Enable Interrupt
	intr := interrupt.New(stm32.IRQ_USB_LP_CAN_RX0, handleUSBIRQ)
	intr.SetPriority(0xc0)
	intr.Enable()
}

// handleUSBIRQ will handle USB hardware interrupts
func handleUSBIRQ(intr interrupt.Interrupt) {
	istr := stm32.USB.ISTR.Get()
	ep := istr & stm32.USB_ISTR_EP_ID_Msk // Endpoint identifier
	//ev := 0

	if (istr & stm32.USB_ISTR_CTR) > 0 {
		reg := getEPR(uint8(ep)) & 0xFF
		if (reg & stm32.USB_EPR_CTR_TX) > 0 {
			ep = ep | 0x80
			//	ev = usbd_evt_eptx
			println("ep ", ep, " CTR_TX")
		} else {
			if (reg & stm32.USB_EPR_SETUP) > 0 {
				println("ep ", ep, " CTR_RX SETUP")
				//		ev = usbd_evt_epsetup
			} else {
				println("ep ", ep, " CTR_RX NOSETUP")
				//		ev = usbd_evt_eprx
			}
		}
		stm32.USB.ISTR.ClearBits(stm32.USB_ISTR_CTR)
		//println("ep: ", ep, " ev: ", ev)
		println("ep: ", ep)
	}

	if (istr & stm32.USB_ISTR_RESET) > 0 {
		initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)
		usbConfiguration = 0

		stm32.USB.ISTR.ClearBits(stm32.USB_ISTR_RESET)
		println("ep ", ep, " RESET")
	}
	if (istr & stm32.USB_ISTR_ERR) > 0 {
		stm32.USB.ISTR.ClearBits(stm32.USB_ISTR_ERR)
		println("ep ", ep, " ERROR")
	}

	if (istr & stm32.USB_ISTR_SOF) > 0 {
		stm32.USB.ISTR.ClearBits(stm32.USB_ISTR_SOF)
		//println("ep ", ep, " SOF")
	}

	if (istr & stm32.USB_ISTR_ESOF) > 0 {
		stm32.USB.ISTR.ClearBits(stm32.USB_ISTR_ESOF)
		//println("ep ", ep, " ESOF")
	}

}

func getEPR(ep uint8) uint32 {
	switch ep {
	case 0:
		return stm32.USB.EP0R.Get()
	case 1:
		return stm32.USB.EP1R.Get()
	case 2:
		return stm32.USB.EP2R.Get()
	case 3:
		return stm32.USB.EP3R.Get()
	case 4:
		return stm32.USB.EP4R.Get()
	case 5:
		return stm32.USB.EP5R.Get()
	case 6:
		return stm32.USB.EP6R.Get()
	case 7:
		return stm32.USB.EP7R.Get()
	default:
		return 0
	}
}

// initEndpoint
func initEndpoint(ep, config uint32) {
	dp("initEndpoint")
	switch config {
	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn:
	case usb.ENDPOINT_TYPE_BULK | usb.EndpointOut:
	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointOut:
	case usb.ENDPOINT_TYPE_BULK | usb.EndpointIn:
	case usb.ENDPOINT_TYPE_CONTROL:
	}
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	dp("SendUSBInPacket")
	return true
}

func handleUSBSetAddress(setup usb.Setup) bool {
	dp("handleUSBSetAddress")
	return true
}

func SendZlp() {
	dp("SendZlp")
}

func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
	dp("sendUSBPacket")
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	dp("ReceiveUSBControlPacket")
	return [7]byte{0, 0, 0, 0, 0, 0, 0}, nil
}

// pmaAddr returns memory address from offset
// Packet Memory Area
/*
func getPMATable(addr uint16) *uint16 {
	return (*uint16)(0)
}
*/
func epWrite(ep uint8) {
	_ = ep
}

func getFrameNumber() uint32 {
	return stm32.USB.GetFNR_FN()
}

//func handlePadCalibration()                 {}
//func handleUSBIRQ(intr interrupt.Interrupt) {}

/*
func handleEndpointRx(ep uint32) []byte  { return nil }
func handleEndpointRxComplete(ep uint32) {}
func epPacketSize(size uint16) uint32     { return 0 }
func getEPCFG(ep uint32) uint8            { return 0 }
func setEPCFG(ep uint32, val uint8)       {}
func setEPSTATUSCLR(ep uint32, val uint8) {}
func setEPSTATUSSET(ep uint32, val uint8) {}
func getEPSTATUS(ep uint32) uint8         { return 0 }
func getEPINTFLAG(ep uint32) uint8        { return 0 }
func setEPINTFLAG(ep uint32, val uint8)   {}
func setEPINTENCLR(ep uint32, val uint8)  {}
func setEPINTENSET(ep uint32, val uint8)  {}
*/

/*

func getEPCFG(ep uint32) uint8 {
	switch ep {
	case 0:
		return stm32.USB..Get()
	case 1:
		return sam.USB_DEVICE.EPCFG1.Get()
	case 2:
		return sam.USB_DEVICE.EPCFG2.Get()
	case 3:
		return sam.USB_DEVICE.EPCFG3.Get()
	case 4:
		return sam.USB_DEVICE.EPCFG4.Get()
	case 5:
		return sam.USB_DEVICE.EPCFG5.Get()
	case 6:
		return sam.USB_DEVICE.EPCFG6.Get()
	case 7:
		return sam.USB_DEVICE.EPCFG7.Get()
	default:
		return 0
	}
}

func setEPCFG(ep uint32, val uint8) {
	switch ep {
	case 0:
		sam.USB_DEVICE.EPCFG0.Set(val)
	case 1:
		sam.USB_DEVICE.EPCFG1.Set(val)
	case 2:
		sam.USB_DEVICE.EPCFG2.Set(val)
	case 3:
		sam.USB_DEVICE.EPCFG3.Set(val)
	case 4:
		sam.USB_DEVICE.EPCFG4.Set(val)
	case 5:
		sam.USB_DEVICE.EPCFG5.Set(val)
	case 6:
		sam.USB_DEVICE.EPCFG6.Set(val)
	case 7:
		sam.USB_DEVICE.EPCFG7.Set(val)
	default:
		return
	}
}

*/
