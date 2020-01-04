# rangesensor
A package to take a distance measurement from an HC-SR04 ultrasonic range
sensor.

# Basic Usage

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/asjoyner/rangesensor"
)

func main() {

	s, err := rangesensor.New("22", "23")
	if err != nil {
		fmt.Println("could not configure pin: ", err)
		os.Exit(1)
	}

	for {
		distance, err := s.MeasureDistance()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Distance: %5.2f cm", distance.InCentimeters())
			fmt.Printf(" (%d us)\n", distance.InMicroseconds())
		}
		time.Sleep(time.Duration(250) * time.Millisecond)
	}

}
```

# Credit
This takes inspiration from both of these repositories.

* https://github.com/jdevelop/golang-rpi-extras
* https://godoc.org/github.com/ricallinson/engine

Both of those are based on github.com/stianeikeland/go-rpio.  I chose to
reimplement a similar basic mechanic on top of the periph.io/x/periph/conn/gpio
library so it would be easier to add testing, use the more sysfs kernel API,
and leverage the associated interrupt-based edge detection.


