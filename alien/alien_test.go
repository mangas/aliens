package alien_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/mangas/aliens/alien"
	"github.com/mangas/aliens/model"
	"github.com/mangas/aliens/world"
	"github.com/mangas/aliens/world/mocks"

	"github.com/stretchr/testify/require"
)

func TestRandomGen(t *testing.T) {
	cases := []struct {
		name     string
		weights  [4]int32
		expected []model.Direction
	}{
		{name: "ascending", weights: [4]int32{1, 2, 3, 4}, expected: model.AllDirections()},
		{name: "descending", weights: [4]int32{4, 3, 2, 1}, expected: []model.Direction{
			model.DirectionSouth,
			model.DirectionNorth,
			model.DirectionWest,
			model.DirectionEast,
		}},
		{name: "alternate", weights: [4]int32{3, 1, 4, 2}, expected: []model.Direction{
			model.DirectionWest,
			model.DirectionEast,
			model.DirectionSouth,
			model.DirectionNorth,
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dirs := alien.RandomDirGen(c.weights)

			require.Equal(t, c.expected, dirs)
		})
	}
}

func infiniteMap() []model.City {
	city1 := model.NewCity(model.CityName("city1"))
	city2 := model.NewCity(model.CityName("city2"))
	city1.Borders[model.DirectionNorth] = city2.Name
	city2.Borders[model.DirectionSouth] = city1.Name

	return []model.City{city1, city2}
}

func TestActorExpires(t *testing.T) {
	const alienName = "alien1"

	worldMap, err := world.NewMap(infiniteMap())
	require.NoError(t, err)

	a := alien.New(alienName, worldMap, gen)

	ctx, cancel := context.WithCancel(context.Background())
	eventC := make(chan model.Event)
	var events []model.Event
	go func() {
		events = serialiseEvents(ctx, eventC)
		close(eventC)
	}()

	err = a.Start(ctx, eventC)
	require.NoError(t, err)
	cancel()
	<-eventC

	require.Len(t, events, 1)
	evt, ok := events[0].(model.EventAlienExpired)
	require.True(t, ok)
	require.Equal(t, evt.Name, model.AlienName(alienName))
}

func TestActor(t *testing.T) {
	const alienName = "alien1"

	var count int
	worldMap := &mocks.Map{}
	worldMap.TryLandStub = func(an model.AlienName) (model.City, error) {
		return model.City{
			Name: model.CityName(strconv.Itoa(count)),
		}, nil
	}
	worldMap.TryMoveStub = func(cn model.CityName, an model.AlienName, d ...model.Direction) (model.City, error) {
		require.Equal(t, model.AllDirections(), d)
		count++
		return model.City{
			Name: model.CityName(strconv.Itoa(count)),
		}, nil
	}

	a := alien.New(alienName, worldMap, gen)

	ctx, cancel := context.WithCancel(context.Background())
	eventC := make(chan model.Event)
	var events []model.Event
	go func() {
		events = serialiseEvents(ctx, eventC)
		close(eventC)
	}()

	err := a.Start(ctx, eventC)
	require.NoError(t, err)
	cancel()
	<-eventC

	require.Equal(t, 10000, count)
	require.Len(t, events, 1)
	evt, ok := events[0].(model.EventAlienExpired)
	require.True(t, ok)
	require.Equal(t, evt.Name, model.AlienName(alienName))
}

func TestActorAlienDestroyed(t *testing.T) {
	worldMap := &mocks.Map{}
	worldMap.TryLandReturns(model.City{}, nil)

	worldMap.TryMoveReturns(model.City{
		NumVisitors: 2,
		Visitors:    [2]model.AlienName{"1", "2"},
	}, model.ErrAlienDestroyed)

	a := alien.New("alien1", worldMap, gen)

	ctx, cancel := context.WithCancel(context.Background())
	eventC := make(chan model.Event)
	var events []model.Event
	go func() {
		events = serialiseEvents(ctx, eventC)
		close(eventC)
	}()

	err := a.Start(ctx, eventC)
	require.NoError(t, err)
	cancel()
	<-eventC

	require.Len(t, events, 1)
	evt, ok := events[0].(model.EventAliensFought)
	require.True(t, ok)
	require.Equal(t, evt.Atacker, model.AlienName("2"))
	require.Equal(t, evt.Defender, model.AlienName("1"))
}

func TestActorNoDirections(t *testing.T) {
	worldMap := &mocks.Map{}
	worldMap.TryLandReturns(model.City{}, nil)
	worldMap.TryMoveReturns(model.City{}, model.ErrNoDirectionsLeft)

	a := alien.New("alien1", worldMap, gen)

	ctx, cancel := context.WithCancel(context.Background())
	eventC := make(chan model.Event)
	var events []model.Event
	go func() {
		events = serialiseEvents(ctx, eventC)
		close(eventC)
	}()

	err := a.Start(ctx, eventC)
	require.NoError(t, err)
	cancel()
	<-eventC

	require.Len(t, events, 1)
	evt, ok := events[0].(model.EventAlienTrapped)
	require.True(t, ok)
	require.Equal(t, evt.Name, model.AlienName("alien1"))
}

func TestActorNoCitiesLeftLand(t *testing.T) {
	worldMap := &mocks.Map{}
	worldMap.TryLandReturns(model.City{}, model.ErrWorldHasBeenDestroyed)

	ctx := context.Background()
	a := alien.New("alien1", worldMap, gen)

	eventC := make(chan model.Event)

	err := a.Start(ctx, eventC)
	require.Error(t, err)
	require.ErrorIs(t, err, model.ErrWorldHasBeenDestroyed)
}

func TestActorCurrentCityDestroyed(t *testing.T) {
	worldMap := &mocks.Map{}
	worldMap.TryLandReturns(model.City{}, nil)
	worldMap.TryMoveReturns(model.City{}, model.ErrCityHasBeenDestroyed)

	a := alien.New("alien1", worldMap, gen)

	ctx, cancel := context.WithCancel(context.Background())
	eventC := make(chan model.Event)
	var events []model.Event
	go func() {
		events = serialiseEvents(ctx, eventC)
		close(eventC)
	}()

	err := a.Start(ctx, eventC)
	require.NoError(t, err)
	cancel()
	<-eventC

	require.Len(t, events, 0)
}

func serialiseEvents(ctx context.Context, events <-chan model.Event) []model.Event {
	var evs []model.Event

loop:
	for {
		select {
		case m := <-events:
			fmt.Println(m)
			evs = append(evs, m)
		case <-ctx.Done():
			break loop
		default:
			// don't spin
			time.Sleep(10 * time.Millisecond)
		}
	}

	return evs
}

func gen(ws [4]int32) []model.Direction {
	return model.AllDirections()
}
