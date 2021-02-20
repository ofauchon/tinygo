// +build nucleowl55jc

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED1 = PB15 // Red
	LED2 = PB9  // Green
	LED3 = PB11 // Blue
	LED  = LED2 // Default

	BTN1   = PA0
	BTN2   = PA1
	BTN3   = PC6
	BUTTON = BTN1

	// SubGhz (SPI3)
	SPI0_NSS_PIN = PA4
	SPI0_SCK_PIN = PA5
	SPI0_SDO_PIN = PA6
	SPI0_SDI_PIN = PA7

	//MCU USART1
	UART1_TX_PIN = PB6
	UART1_RX_PIN = PB7

	//MCU USART2
	UART2_RX_PIN = PA3
	UART2_TX_PIN = PA2

	// DEFAULT USART
	UART_RX_PIN = UART2_RX_PIN
	UART_TX_PIN = UART2_TX_PIN
)

var (
	// STM32WLE5's UART2 connected to STLINKV3 Virtual Com Port
	UART0  = &_UART0
	_UART0 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 7,
	}
	DefaultUART = UART0

	// UART1 is free
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 7,
	}

	// SPI
	SPI0 = SPI{
		Bus: stm32.SPI3,
	}
)

func init() {
	// Enable UARTs Interrupts
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART1.handleInterrupt)
}
