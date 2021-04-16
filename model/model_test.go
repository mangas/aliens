package model_test

import (
	"testing"

	"github.com/mangas/aliens/model"
	"github.com/stretchr/testify/require"
)

func TestRemoveVisitor(t *testing.T) {
	const alien1 = model.AlienName("alien1")

	city := model.NewCity("city1")
	city = city.WithVisitor(alien1)
	require.Contains(t, city.Visitors, alien1)
	require.Equal(t, 1, city.NumVisitors)

	city = city.WithoutVisitor(alien1)
	require.Zero(t, city.NumVisitors)
}
