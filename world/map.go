//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -fake-name Map -o mocks/mocks.go ./ Map

package world

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/imdario/mergo"
	"github.com/mangas/aliens/model"
)

// Map defines the world coordinator.
type Map interface {
	Cities() []model.City
	TryLand(name model.AlienName) (model.City, error)
	TryMove(from model.CityName, name model.AlienName, directions ...model.Direction) (model.City, error)
}

const MaxAliens = 2

// NewMap creates a map populated with the cities passed in. Cities are not assumed to have two way connections.
// If city A has B in the north border, this does not mean that B has A has south. This needs to be explicitly passed
// in the world map.
func NewMap(cities []model.City) (*MMap, error) {
	m := &MMap{
		cities: make(map[model.CityName]model.City),
	}

	for _, c := range cities {
		if err := m.tryAddCity(c); err != nil {
			return nil, err
		}
	}

	return m, nil
}

var _ Map = (*MMap)(nil)

// MMap is the in-memory implementation of Map
type MMap struct {
	cities map[model.CityName]model.City

	lock sync.Mutex
}

// Cities returns the current state of the world map.
func (m *MMap) Cities() []model.City {
	m.lock.Lock()
	defer m.lock.Unlock()

	var cities []model.City
	for _, v := range m.cities {
		cities = append(cities, v)
	}

	return cities
}

// tryAddCity will add a city, will merge the record is it exists.
func (m *MMap) tryAddCity(city model.City) error {
	c, ok := m.cities[city.Name]
	if !ok {
		c = city
	}

	for _, v := range c.Borders {
		if v == city.Name {
			return fmt.Errorf("city %s has a border with itself, that's not allowed", v)
		}

		if err := m.tryAddCity(model.NewCity(v)); err != nil {
			return err
		}
	}

	err := mergo.Merge(&c, city, mergo.WithOverride)
	if err != nil {
		return err
	}

	m.cities[city.Name] = c

	return nil
}

// TryLand will land a new alien in a city.
func (m *MMap) TryLand(name model.AlienName) (model.City, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.cities) == 0 {
		return model.City{}, model.ErrWorldHasBeenDestroyed
	}

	landIndex := rand.Int31n(int32(len(m.cities)))

	var i int
	for _, v := range m.cities {
		if i != int(landIndex) {
			i++
			continue
		}

		return m.addVisitor(v, name)
	}

	return model.City{}, fmt.Errorf("%d is not a valid index", landIndex)
}

// TryMove is an expensive and atomic operation, it will try to find a direction guided by the  priority given by
// the directions argument. If one of the directions is valid, the map will be updated and the new position returned.
func (m *MMap) TryMove(from model.CityName, name model.AlienName, directions ...model.Direction) (model.City, error) {
	if len(directions) == 0 {
		return model.City{}, fmt.Errorf("no direction provided")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	c, ok := m.cities[from]
	if !ok {
		return model.City{}, model.ErrCityHasBeenDestroyed
	}

	// Update old city
	c = c.WithoutVisitor(name)
	m.cities[c.Name] = c

	for _, d := range directions {
		newCityName, ok := c.Borders[d]
		if !ok {
			continue
		}

		// city has been destroyed, it's too expensive to update all the borders and we don't need to optimize prematurely.
		newCity, ok := m.cities[newCityName]
		if !ok {
			continue
		}

		return m.addVisitor(newCity, name)
	}

	return model.City{}, model.ErrNoDirectionsLeft
}

// addVisitor will encapsulate the logic for counting and managing the state.
func (m *MMap) addVisitor(city model.City, alienName model.AlienName) (model.City, error) {
	c, ok := m.cities[city.Name]
	if !ok {
		return model.City{}, model.ErrCityHasBeenDestroyed
	}

	c = c.WithVisitor(alienName)

	m.cities[c.Name] = c

	if c.NumVisitors < MaxAliens {
		return c, nil
	}

	// remove the city, the copy will be returned in case more actions need to be performed.
	if c.NumVisitors == MaxAliens {
		delete(m.cities, c.Name)
		return c, model.ErrAlienDestroyed
	}

	return model.City{}, fmt.Errorf("invalid number of visitors")
}
