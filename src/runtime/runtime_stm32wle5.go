// +build stm32,stm32wle5

package runtime

import (
	"device/stm32"
	"machine"
)

const (
/*
	TICK_RATE        = 1000 // 1 KHz
	TICK_TIMER_IRQ   = stm32.IRQ_TIM1
	TICK_TIMER_FREQ  = 4000000 // 32 MHz
	SLEEP_TIMER_IRQ  = stm32.IRQ_TIM2
	SLEEP_TIMER_FREQ = 4000000 // 32 MHz
*/
)

type arrtype = uint32

func init() {
	// Main Clock
	initCLK()

	// UART init
	machine.Serial.Configure(machine.UARTConfig{})

	// Timers init
	initTickTimer(&machine.TIM1)

	// Init Lora Radio
	SubGhzInit()
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

const (
	FLASH_ACR_LATENCY_WS2 = 0x2
	RCC_CFGR_PPRE1_Div2   = 0x4
	RCC_CFGR_PPRE2_Div1   = 0x0

	RCC_CFGR_SW_MASK   = 0x3
	RCC_CFGR_SW_POS    = 0x0
	RCC_CFGR_SW_MSI    = 0x0
	RCC_CFGR_SW_HSI    = 0x1
	RCC_CFGR_SW_HSE    = 0x2
	RCC_CFGR_SW_PLLCLK = 0x3

	RCC_CFGR_SWS_PLLCLK = 0x3
	RCC_CFGR_SWS_HSI    = 0x1

	HSE_STARTUP_TIMEOUT = 0x0500
	/* PLL Options - See RMN0461 Reference Manual pg. 247 */
	PLL_M = 2
	PLL_N = 6
	PLL_R = 2
	PLL_P = 2
	PLL_Q = 2
)

func initCLK() {

	if machine.OSC_PLLHSE == true { // HSE

		// Set Power Voltage Regulator Range 2
		stm32.PWR.CR1.ReplaceBits(0b10, stm32.PWR_CR1_VOS_Msk, stm32.PWR_CR1_VOS_Pos)

		// Set HSE division factor : HSE clock not divided
		stm32.RCC.CR.ReplaceBits(0b000, 0b1111, stm32.RCC_CR_HSEPRE_Pos)

		// enable external Clock HSE32 TXCO (RM0461p226)
		stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEBYPPWR)
		stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
		for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
		}

		// Disable PLL
		stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
		for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
		}

		// Configure PLL
		//stm32.RCC.PLLCFGR.Set(0x22020613)
		stm32.RCC.PLLCFGR.Set(0x23020613) // Same with PLLQ enabled to enable RNG

		// Enable PLL
		stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
		for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
		}

		// Enable PLL System Clock output.
		stm32.RCC.PLLCFGR.SetBits(stm32.RCC_PLLCFGR_PLLREN)
		for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
		}

		// Set Flash Latency of 2 and wait until it's set properly
		stm32.FLASH.ACR.ReplaceBits(0b010, 0b111, stm32.Flash_ACR_LATENCY_Pos)
		for (stm32.FLASH.ACR.Get() & 0b11) != 0x2 {
		}

	}

	//****************** CLOCK Dividers

	// HCLK1 Configuration (DIV1)
	stm32.RCC.CFGR.ReplaceBits(0x0000, 0b1111, stm32.RCC_CFGR_HPRE_Pos)
	for !stm32.RCC.CFGR.HasBits(stm32.RCC_CFGR_HPREF) {
	}

	// HCLK3 Configuration (DIV1)
	stm32.RCC.EXTCFGR.ReplaceBits(0x0000, 0b1111, stm32.RCC_EXTCFGR_SHDHPRE_Pos)
	for !stm32.RCC.EXTCFGR.HasBits(stm32.RCC_EXTCFGR_SHDHPREF) {
	}

	// PCLK1 Configuration (DIV1)
	stm32.RCC.CFGR.ReplaceBits(0x000, 0b111, stm32.RCC_CFGR_PPRE1_Pos)
	for !stm32.RCC.CFGR.HasBits(stm32.RCC_CFGR_PPRE1F) {
	}

	// PCLK2 Configuration (DIV1)
	stm32.RCC.CFGR.ReplaceBits(0x000, 0b111, stm32.RCC_CFGR_PPRE2_Pos)
	for !stm32.RCC.CFGR.HasBits(stm32.RCC_CFGR_PPRE2F) {
	}

	// Switch Clock source
	if machine.OSC_PLLHSE == true { // HSE
		// Set clock source to PLL (0x3)
		stm32.RCC.CFGR.ReplaceBits(0b11, 0b11, stm32.RCC_CFGR_SW_Pos)
		for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != 0xc {
		}
	} else {
		// Set clock source to MSI (0x00)
		stm32.RCC.CFGR.ReplaceBits(0b00, 0b11, stm32.RCC_CFGR_SW_Pos)
		for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != 0x00 {
		}

	}

}

// SubGhzInit enable radio module
func SubGhzInit() error {

	// Enable APB3 Periph clock
	stm32.RCC.APB3ENR.SetBits(stm32.RCC_APB3ENR_SUBGHZSPIEN)
	_ = stm32.RCC.APB3ENR.Get() //Delay after RCC periph clock enable

	// Enable TXCO and HSE
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEBYPPWR)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	// Disable radio reset and wait it's ready
	stm32.RCC.CSR.ClearBits(stm32.RCC_CSR_RFRST)
	for stm32.RCC.CSR.HasBits(stm32.RCC_CSR_RFRSTF) {
	}

	// Disable radio NSS=1
	stm32.PWR.SUBGHZSPICR.SetBits(stm32.PWR_SUBGHZSPICR_NSS)

	// Enable Exti Line 44: Radio IRQ ITs for CPU1 (RM0461-14.3.1)
	stm32.EXTI.IMR2.SetBits(0x1000) // IM44 ===> TEST

	// Enable radio busy wakeup from Standby for CPU
	stm32.PWR.CR3.SetBits(stm32.PWR_CR3_EWRFBUSY)

	// Clear busy flag
	stm32.PWR.SCR.Set(stm32.PWR_SCR_CWRFBUSYF)

	//  SUBGHZSPI configuration
	stm32.SPI3.CR1.ClearBits(stm32.SPI_CR1_SPE)                                                   // Disable SPI
	stm32.SPI3.CR1.Set(stm32.SPI_CR1_MSTR | stm32.SPI_CR1_SSI | (0b010 << 3) | stm32.SPI_CR1_SSM) // Software Slave Management (NSS) + /8 prescaler
	stm32.SPI3.CR2.Set(stm32.SPI_CR2_FRXTH | (0b111 << 8))                                        // FIFO Threshold and 8bit size
	stm32.SPI3.CR1.SetBits(stm32.SPI_CR1_SPE)                                                     // Enable SPI

	return nil
}
