package world

import (
	"testing"

	"github.com/mangas/aliens/model"
	"github.com/stretchr/testify/require"
)

func TestAddVisitor(t *testing.T) {
	city := model.NewCity(model.CityName("city1"))

	m, err := NewMap([]model.City{city})
	require.NoError(t, err)

	c, err := m.addVisitor(city, model.AlienName("alien1"))
	require.NoError(t, err)
	require.Contains(t, c.Visitors, model.AlienName("alien1"))

	c, err = m.addVisitor(city, model.AlienName("alien2"))
	require.Error(t, err)
	require.Contains(t, c.Visitors, model.AlienName("alien2"))
}
