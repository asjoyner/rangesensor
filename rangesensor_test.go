package rangesensor

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

var (
	quietEdge = make(chan gpio.Level, 2)
	quietPin  = gpiotest.Pin{N: "quietpin", L: gpio.Low, EdgesChan: quietEdge}

	triggerLevel gpio.Level
	echoLevel    gpio.Level
	echoEdge     = make(chan gpio.Level, 2)
	echoPin      = gpiotest.Pin{N: "TestEchoPin", L: echoLevel, EdgesChan: echoEdge}
	triggerPin   = gpiotest.Pin{N: "TestTriggerPin", L: triggerLevel, EdgesChan: quietEdge}
)

func init() {
	gpioreg.Register(&quietPin)
	gpioreg.Register(&echoPin)
	gpioreg.Register(&triggerPin)
	go watchPin()
}

func TestNoSensor(t *testing.T) {
	s, err := New("quietpin", "quietpin")
	if err != nil {
		t.Fatalf("initializing sensor: %s", err)
	}
	if _, err := s.MeasureDistance(); err == nil {
		t.Errorf("no error returned when no sensor is attached")
	}
}

func TestSensor(t *testing.T) {
	s, err := New("TestEchoPin", "TestTriggerPin")
	if err != nil {
		t.Fatalf("initializing sensor: %s", err)
	}
	distance, err := s.MeasureDistance()
	if err != nil {
		t.Fatalf("MeasureDistance: %s", err)
	}
	d := distance.InMicroseconds()
	t.Logf("Timing result: %d", d)
	if d < 100 {
		t.Errorf("MeasureDistance: want: >100, got: %d", d)
	}
	// I would like this to be tighter, but the realities of scheduling on a
	// multiuser system mean occasionally the system is busy, and if this was
	// ~400 like it should be, the test would be flaky.
	if d > 20000 {
		t.Errorf("MeasureDistance: want: <=20000, got: %d", d)
	}
}

func TestTeardownParallel(t *testing.T) {
	if err := quietPin.Halt(); err != nil {
		t.Errorf("quietPin.Halt(): %s", err)
	}
	if err := echoPin.Halt(); err != nil {
		t.Errorf("echoPin.Halt(): %s", err)
	}
	if err := triggerPin.Halt(); err != nil {
		t.Errorf("triggerPin.Halt(): %s", err)
	}
}

func watchPin() {
	// This just melts the processor to try to catch the 10us pulse on the test
	// pin.  It reduces the likelihood the test flakes, but in my testing it's
	// still only a little better than 95% reliable.
	for {
		triggerPin.Lock()
		value := triggerPin.L
		triggerPin.Unlock()
		if value == gpio.High {
			// Now wait to see the Trigger pin go back down
			for {
				triggerPin.Lock()
				value := triggerPin.L
				triggerPin.Unlock()
				if value == gpio.Low {
					break
				}
				time.Sleep(20 * time.Microsecond)
			}

			// simulate the delay of sending the sonic burst
			time.Sleep(200 * time.Microsecond)

			// Raise the timing edge
			echoPin.Lock()
			echoPin.L = gpio.High
			echoPin.EdgesChan <- true
			echoPin.Unlock()

			// Hold the echo pin high to simulate a 100us time of flight
			time.Sleep(100 * time.Microsecond)

			// Lowering the timing edge.
			echoPin.Lock()
			echoPin.L = gpio.Low
			echoPin.EdgesChan <- true
			echoPin.Unlock()
		}
	}
}

func TestTimeToCentimeters(t *testing.T) {
	want := float32(speedOfSound * 100) // in cm
	got := TimeToCentimeters(2e6)
	if got != want {
		t.Errorf("TimeToCentimeters(1s), got: %f, want: %f)", got, want)
	}
}
