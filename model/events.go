package model

import "fmt"

type Event interface {
	String() string
}

type EventAliensFought struct {
	Atacker  AlienName
	Defender AlienName
	City     CityName
}

func (e EventAliensFought) String() string {
	return fmt.Sprintf("%s has been destroyed by alien %s and alien %s", e.City, e.Atacker, e.Defender)
}

type EventAlienTrapped struct {
	Name AlienName
}

func (e EventAlienTrapped) String() string {
	return fmt.Sprintf("%s was trapped and died", e.Name)
}

type EventAlienExpired struct {
	Name AlienName
}

func (e EventAlienExpired) String() string {
	return fmt.Sprintf("%s's reign of terror is over!", e.Name)
}
