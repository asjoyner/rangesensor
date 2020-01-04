// Package rangesensor facilitates measuring distance with an HC-SR04
// ultrasonic ranging module.
package rangesensor

import (
	"fmt"
	"os"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

// Measurement expresses a sensor measurement and facilitates conversion to
// various units.
//
// It uses constants, found in the datasheet, which are based on the speed of
// sound in air at sea level (~340 m/s).
//
// TODO: Encode the 340m/s assumption as a constant, and derive the other
// constants from that value.
type Measurement struct {
	timeOfFlight time.Duration
}

// InCentimeters converts the time of flight measurement into centimeters.
func (r *Measurement) InCentimeters() float32 {
	return float32(r.timeOfFlight.Microseconds()) / 58
}

// InInches converts the time of flight measurement into inches.
func (r *Measurement) InInches() float32 {
	return float32(r.timeOfFlight.Microseconds()) / 148
}

// InMicroseconds returns the raw time of flight measurement.
func (r *Measurement) InMicroseconds() int64 {
	return r.timeOfFlight.Microseconds()
}

// InMilliseconds returns the raw time of flight measurement.
func (r *Measurement) InMilliseconds() int64 {
	return r.timeOfFlight.Milliseconds()
}

// Sensor represents an HC-SR04 ultrasonic ranging module.
//
// Datasheet: https://cdn.sparkfun.com/datasheets/Sensors/Proximity/HCSR04.pdf
type Sensor struct {
	EchoPin    gpio.PinIO
	TriggerPin gpio.PinIO
}

// New initializes and returns a Sensor object.
//
// Echo is the name of the GPIO pin connected to the module's "Echo" pin.
// Trigger is the name of the GPIO pin connected to the module's "Trig" pin.
//
// Both names should be in the format expected by this module's ByName function
// periph.io/x/periph/conn/gpio/gpioreg
//
// For a RaspberryPi, this corresponds to the BCM pin number as a string.
func New(echo, trigger string) (*Sensor, error) {
	if _, err := host.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var s Sensor
	s.EchoPin = gpioreg.ByName(echo)
	if s.EchoPin == nil {
		return nil, fmt.Errorf("no GPIO echo pin named: %s", echo)
	}
	s.TriggerPin = gpioreg.ByName(trigger)
	if s.TriggerPin == nil {
		return nil, fmt.Errorf("no GPIO trigger pin named: %s", trigger)
	}

	if err := s.TriggerPin.Out(gpio.Low); err != nil {
		return nil, err
	}
	if err := s.EchoPin.In(gpio.PullDown, gpio.BothEdges); err != nil {
		return nil, err
	}

	return &s, nil
}

// MeasureDistance returns a distance measurement from the sensor.
func (s *Sensor) MeasureDistance() (*Measurement, error) {
	// Clear the existing edge values in preparation for the timing signal's
	// rising edge
	if err := s.EchoPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		return nil, err
	}

	// Briefly raise the TriggerPin to cause the HC-SR04 to take a measurement.
	s.TriggerPin.Out(gpio.High)
	time.Sleep(10 * time.Microsecond)
	s.TriggerPin.Out(gpio.Low)

	// Await the EchoPin going High, which signals the start of the duration
	var start, end time.Time
	if ok := s.EchoPin.WaitForEdge(1 * time.Second); !ok {
		return nil, fmt.Errorf("no timing signal detected")
	}
	start = time.Now()

	// Switch to watching for a falling edge
	if err := s.EchoPin.In(gpio.PullDown, gpio.FallingEdge); err != nil {
		return nil, err
	}

	// Await the EchoPin going Low, which signals the end of the duration
	if ok := s.EchoPin.WaitForEdge(1 * time.Second); !ok {
		return nil, fmt.Errorf("timing signal exceed valid duration")
	}
	end = time.Now()

	return &Measurement{end.Sub(start)}, nil
}