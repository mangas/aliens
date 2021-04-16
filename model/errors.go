package model

import "fmt"

var (
	ErrNoDirectionsLeft      = fmt.Errorf("no directions left")
	ErrCityHasBeenDestroyed  = fmt.Errorf("city has been destryed")
	ErrAlienDestroyed        = fmt.Errorf("alien was destroyed")
	ErrWorldHasBeenDestroyed = fmt.Errorf("world has been destroyed")
)
