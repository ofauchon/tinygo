//go:build rp2350b

package machine

// RP2350B has additional pins.

const (
	GPIO30 Pin = 30 // peripherals: PWM7 channel A
	GPIO31 Pin = 31 // peripherals: PWM7 channel B
	GPIO32 Pin = 32 // peripherals: PWM8 channel A
	GPIO33 Pin = 33 // peripherals: PWM8 channel B
	GPIO34 Pin = 34 // peripherals: PWM9 channel A
	GPIO35 Pin = 35 // peripherals: PWM9 channel B
	GPIO36 Pin = 36 // peripherals: PWM10 channel A
	GPIO37 Pin = 37 // peripherals: PWM10 channel B
	GPIO38 Pin = 38 // peripherals: PWM11 channel A
	GPIO39 Pin = 39 // peripherals: PWM11 channel B
	GPIO40 Pin = 40 // peripherals: PWM8 channel A
	GPIO41 Pin = 41 // peripherals: PWM8 channel B
	GPIO42 Pin = 42 // peripherals: PWM9 channel A
	GPIO43 Pin = 43 // peripherals: PWM9 channel B
	GPIO44 Pin = 44 // peripherals: PWM10 channel A
	GPIO45 Pin = 45 // peripherals: PWM10 channel B
	GPIO46 Pin = 46 // peripherals: PWM11 channel A
	GPIO47 Pin = 47 // peripherals: PWM11 channel B
)

// Analog pins on 2350b.
const (
	ADC0 Pin = GPIO40
	ADC1 Pin = GPIO41
	ADC2 Pin = GPIO42
	ADC3 Pin = GPIO43
	ADC4 Pin = GPIO44
	ADC5 Pin = GPIO45
	ADC6 Pin = GPIO46
	ADC7 Pin = GPIO47
	// Ninth ADC channel.
	thermADC = 48
)

// Additional PWMs on the RP2350B.
var (
	PWM8  = getPWMGroup(8)
	PWM9  = getPWMGroup(9)
	PWM10 = getPWMGroup(10)
	PWM11 = getPWMGroup(11)
)
