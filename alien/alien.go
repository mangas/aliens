package alien

import (
	"context"
	"math/rand"
	"sort"

	"github.com/pkg/errors"

	"github.com/mangas/aliens/model"
	"github.com/mangas/aliens/world"
)

const maxMoves = 10000

// DirGen should return a slice of directions ordered using the input as priority.
type DirGen func([4]int32) []model.Direction

// RandomWeights creates random weights to be used with directions.
func RandomWeights() [4]int32 {
	return [...]int32{rand.Int31n(100), rand.Int31n(100), rand.Int31n(100), rand.Int31n(100)}
}

// RandomDirGen returns the content of AllDirections, ordered by the weights passed in. See tests usage examples.
func RandomDirGen(weights [4]int32) []model.Direction {
	dirs := model.AllDirections()

	sort.Slice(dirs, func(i, j int) bool {
		return weights[i] < weights[j]
	})

	return dirs
}

// New creates a new Actor.
func New(name model.AlienName, wm world.Map, dirGen DirGen) *Actor {
	return &Actor{
		wm:    wm,
		moves: 0,
		alien: model.Alien{
			Name: name,
		},
		dirgen: dirGen,
	}
}

// Actor is responsible for managing the Alien lifecycle.
type Actor struct {
	alien model.Alien

	dirgen  DirGen
	wm      world.Map
	moves   int
	running bool
}

// Start will land the alien and move it around until an error is found or maxMoves is reached.
// Start will respect context cancellation and will use the channel to pass certains events defined in the model package.
// These events will allow the caller to be notified of certain important actions about a specific alien.
func (a *Actor) Start(ctx context.Context, eventC chan model.Event) error {
	a.running = true

	city, err := a.wm.TryLand(a.alien.Name)
	if err != nil {
		a.running = false
		return a.handleError(ctx, city, err, eventC)
	}

	a.alien.Position = city.Name

	return a.loop(ctx, eventC)
}

func (a *Actor) loop(ctx context.Context, eventC chan model.Event) error {
	defer func() {
		a.running = false
	}()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if a.moves >= maxMoves {
			eventC <- model.EventAlienExpired{
				Name: a.alien.Name,
			}
			return nil
		}

		city, err := a.wm.TryMove(a.alien.Position, a.alien.Name, a.dirgen(RandomWeights())...)
		if err != nil {
			return a.handleError(ctx, city, err, eventC)
		}

		a.alien.Position = city.Name
		a.moves++
	}
}

func (a *Actor) handleError(ctx context.Context, city model.City, err error, eventC chan model.Event) error {
	switch {
	case errors.Is(err, model.ErrAlienDestroyed):
		eventC <- model.EventAliensFought{
			Atacker:  city.Visitors[1],
			Defender: city.Visitors[0],
			City:     city.Name,
		}

		return nil
	case errors.Is(err, model.ErrNoDirectionsLeft):
		eventC <- model.EventAlienTrapped{
			Name: a.alien.Name,
		}

		return nil
	case errors.Is(err, model.ErrCityHasBeenDestroyed):
		return nil
	default:
	}

	return errors.Wrap(err, "unexpected error")
}
