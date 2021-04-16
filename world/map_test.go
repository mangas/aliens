package world_test

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mangas/aliens/model"
	"github.com/mangas/aliens/world"
	"github.com/stretchr/testify/require"
)

func TestCityLoop(t *testing.T) {
	city := model.NewCity("city1")
	city.Borders[model.DirectionEast] = "city1"

	_, err := world.NewMap([]model.City{city})
	require.Error(t, err)
	require.Contains(t, err.Error(), "has a border with itself")
}

func TestLandEmpty(t *testing.T) {
	m, err := world.NewMap(nil)
	require.NoError(t, err)
	_, err = m.TryLand("alien1")
	require.Error(t, err)
	require.ErrorIs(t, err, model.ErrWorldHasBeenDestroyed)
}

func TestMoveNonExistent(t *testing.T) {
	m, err := world.NewMap(nil)
	require.NoError(t, err)
	_, err = m.TryMove(model.CityName("somename"), model.AlienName("some alien"), model.AllDirections()...)
	require.Error(t, err)
	require.ErrorIs(t, err, model.ErrCityHasBeenDestroyed)
}

func TestMoveNoDirections(t *testing.T) {
	cityName := model.CityName("city1")
	m, err := world.NewMap([]model.City{model.NewCity(cityName)})
	require.NoError(t, err)
	_, err = m.TryMove(cityName, model.AlienName("some alien"), model.AllDirections()...)
	require.Error(t, err)
	require.ErrorIs(t, err, model.ErrNoDirectionsLeft)
}

func TestMoveEmptyDirections(t *testing.T) {
	cityName := model.CityName("city1")
	m, err := world.NewMap([]model.City{model.NewCity(cityName)})
	require.NoError(t, err)
	_, err = m.TryMove(cityName, model.AlienName("some alien"))
	require.Error(t, err)
	require.Equal(t, err.Error(), "no direction provided")
}

func TestMaxAliens(t *testing.T) {
	alien1 := model.AlienName("alien1")
	alien2 := model.AlienName("alien2")
	cityName := model.CityName("city1")
	m, err := world.NewMap([]model.City{model.NewCity(cityName)})
	require.NoError(t, err)

	_, err = m.TryLand(alien1)
	require.NoError(t, err)

	c, err := m.TryLand(alien2)
	require.Error(t, err)
	require.ErrorIs(t, err, model.ErrAlienDestroyed)
	require.Equal(t, alien1, c.Visitors[0])
	require.Equal(t, alien2, c.Visitors[1])
}

func TestMove(t *testing.T) {
	alien1 := model.AlienName("alien1")

	city1 := model.NewCity(model.CityName("city1"))
	city2 := model.NewCity(model.CityName("city2"))
	city1.Borders[model.DirectionNorth] = city2.Name
	city2.Borders[model.DirectionSouth] = city1.Name

	m, err := world.NewMap([]model.City{city1, city2})
	require.NoError(t, err)

	c, err := m.TryLand(alien1)
	require.NoError(t, err)

	target := city1
	if c.Name == city1.Name {
		target = city2
	}

	c, err = m.TryMove(c.Name, alien1, model.DirectionEast, model.DirectionWest, model.DirectionSouth, model.DirectionNorth)
	require.NoError(t, err)
	require.Equal(t, target.Name, c.Name)
	require.Equal(t, [2]model.AlienName{alien1}, c.Visitors)
}

func TestCities(t *testing.T) {
	city1 := model.NewCity(model.CityName("city1"))
	city2 := model.NewCity(model.CityName("city2"))
	city1.Borders[model.DirectionNorth] = city2.Name
	city2.Borders[model.DirectionSouth] = city1.Name

	cities := []model.City{city1, city2}
	m, err := world.NewMap(cities)
	require.NoError(t, err)

	res := m.Cities()
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	sort.Slice(cities, func(i, j int) bool {
		return cities[i].Name < cities[j].Name
	})
	require.Empty(t, cmp.Diff(cities, res))
}

func TestVisitorCount(t *testing.T) {
	city1 := model.NewCity(model.CityName("city1"))
	city2 := model.NewCity(model.CityName("city2"))
	city1.Borders[model.DirectionNorth] = city2.Name
	city2.Borders[model.DirectionSouth] = city1.Name

	cities := []model.City{city1, city2}
	m, err := world.NewMap(cities)
	require.NoError(t, err)

	alienName := model.AlienName("alien1")
	city, err := m.TryLand(alienName)
	require.NoError(t, err)

	city, err = m.TryMove(city.Name, alienName, model.AllDirections()...)
	require.NoError(t, err)
	city, err = m.TryMove(city.Name, alienName, model.AllDirections()...)
	require.NoError(t, err)

	var count int
	for _, c := range m.Cities() {
		count += c.NumVisitors
	}
	require.Equal(t, 1, count)
}
