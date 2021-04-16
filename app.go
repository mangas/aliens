package aliens

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/mangas/aliens/alien"
	"github.com/mangas/aliens/model"
	"github.com/mangas/aliens/world"
)

type EventHandler func(model.Event)

// Invade glues everything together, will create the map, start the AlienActors and ensure all of them will stop.
// Any messages not consumed in the 2 seconds after all the aliens terminate will be lost.
func Invade(ctx context.Context, numberOfAlients int, cities []model.City, dirGen alien.DirGen, evtHandler EventHandler) ([]model.City, error) {
	if len(cities) == 0 {
		return nil, fmt.Errorf("we need to invade one or more cities")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rand.Seed(time.Now().UnixNano())

	eventsC := make(chan model.Event)
	wg := sync.WaitGroup{}
	worldMap, err := world.NewMap(cities)
	if err != nil {
		return nil, err
	}

	als := make(map[model.AlienName]*alien.Actor)
	for i := 0; i < numberOfAlients; i++ {
		name := model.AlienName(fmt.Sprintf("Alien%d", i))

		actor := alien.New(name, worldMap, dirGen)
		als[name] = actor

		wg.Add(1)
		go func() {
			err := actor.Start(ctx, eventsC)
			if err != nil {
				fmt.Printf("alien %s is not happy: %s", name, err.Error())
			}

			wg.Done()
		}()
	}

	endC := make(chan bool)

	go func() {
		for m := range eventsC {
			evtHandler(m)
		}
		endC <- true
	}()

	go func() {
		wg.Wait()
		time.Sleep(2 * time.Second)
		cancel()
		close(eventsC)
	}()

	wg.Wait()
	<-ctx.Done()
	<-endC

	return worldMap.Cities(), nil
}
