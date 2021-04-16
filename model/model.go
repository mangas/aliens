package model

import (
	"fmt"
	"strings"
)

// CityName basically disambiguates the usage of string thoughout the app, increase clarity.
type CityName string

// AlienName basically disambiguates the usage of string thoughout the app, increase clarity.
type AlienName string

// Direction represents the multiple directions an alien can travel.
type Direction int

const (
	north = "north"
	south = "south"
	west  = "west"
	east  = "east"
)

// String does it says in the tin!
func (d Direction) String() string {
	switch d {
	case DirectionNorth:
		return north
	case DirectionSouth:
		return south
	case DirectionEast:
		return east
	case DirectionWest:
		return west
	}

	return ""
}

const (
	DirectionNorth Direction = iota
	DirectionSouth
	DirectionEast
	DirectionWest
)

// DirectionFromString parses direction from a string, I know it's shocking!
func DirectionFromString(s string) (Direction, error) {
	switch strings.ToLower(s) {
	case north:
		return DirectionNorth, nil
	case south:
		return DirectionSouth, nil
	case west:
		return DirectionWest, nil
	case east:
		return DirectionEast, nil
	default:
	}

	return -1, fmt.Errorf("%s is not a valid direction", s)
}

// AllDirections contains all the possible directions. If this is increased there are some functions that need adjusting
// on the alien package.
func AllDirections() []Direction {
	return []Direction{DirectionEast, DirectionWest, DirectionNorth, DirectionSouth}
}

// NewCity creates a new city and inits the map and array.
func NewCity(name CityName) City {
	return City{
		Name:        name,
		Borders:     make(map[Direction]CityName),
		NumVisitors: 0,
		Visitors:    [2]AlienName{},
	}
}

// City represents a city in the world map. The borders are paths that can be travelled given a certain directions.
// NumVisitors and Visitors are not a slice because this way we don't need memory allocations and we can copy the
// structs very easily and cheaply.
type City struct {
	Name CityName

	Borders map[Direction]CityName

	NumVisitors int
	Visitors    [2]AlienName
}

// CityFromString assumes the format CityName [Direction=CityName2]..., if the same direction is passed multiples times,
// the previously defined one will be overriden. There is no assumption that the cities in the directions have been previously created.
func CityFromString(s string) (City, error) {
	line := strings.TrimSpace(s)
	if s == "" {
		return City{}, fmt.Errorf("city line can't be empty")
	}

	parts := strings.Split(line, " ")
	if len(parts) < 1 {
		return City{}, fmt.Errorf("city needs at least a name")
	}

	city := NewCity(CityName(parts[0]))

	for _, border := range parts[1:] {
		bparts := strings.Split(border, "=")
		if len(bparts) != 2 {
			return City{}, fmt.Errorf("border can't be parsed %s", border)
		}

		d, err := DirectionFromString(bparts[0])
		if err != nil {
			return City{}, err
		}

		city.Borders[d] = CityName(bparts[1])
	}

	return city, nil
}

// String prints the city in the same format that is expected for input.
func (c City) String() string {
	str := string(c.Name)
	for dir, cityName := range c.Borders {
		str = fmt.Sprintf("%s %s=%s", str, dir.String(), cityName)
	}

	return str
}

// WithVisitor returns a copy of the City with one more visitor and incremented counter.
func (c City) WithVisitor(name AlienName) City {
	c.Visitors[c.NumVisitors] = name
	c.NumVisitors++

	return c
}

// WithoutVisitor removes the visitor
func (c City) WithoutVisitor(name AlienName) City {
	if c.NumVisitors == 0 {
		return c
	}

	if c.NumVisitors == 1 {
		c.NumVisitors--
		return c
	}

	// if we remove the first then move 2 to 1.
	// This is needed because we always insert at NumVisitors.
	// if visitor was on index 1 it will just be overwritten by Add and will be ignored by NumVisitors so nothing is needed.
	if c.Visitors[0] == name {
		c.Visitors[0] = c.Visitors[1]
		c.Visitors[1] = ""
		c.NumVisitors--
	} else if c.Visitors[1] == name {
		c.NumVisitors--
	}

	return c
}

// Alien is deadly, beware!
type Alien struct {
	Name     AlienName
	Position CityName
}
