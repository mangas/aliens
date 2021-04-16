package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/mangas/aliens"
	"github.com/mangas/aliens/alien"
	"github.com/mangas/aliens/model"
)

func main() {
	var (
		cityFile string
		n        int
	)
	flag.StringVar(&cityFile, "file", "./cities", "specify the path to the file with all the cities")
	flag.StringVar(&cityFile, "f", "./cities", "specify the path to the file with all the cities")
	flag.IntVar(&n, "n", 10, "specifies the number of aliens that will be spawned")
	flag.Parse()

	citiesStr, err := ioutil.ReadFile(cityFile)
	if err != nil {
		log.Fatalf("unable to read file %s", cityFile)
	}

	cityLines := strings.Split(string(citiesStr), "\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cities []model.City
	for _, line := range cityLines {
		if line == "" {
			continue
		}

		c, err := model.CityFromString(line)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		cities = append(cities, c)
	}

	cities, err = aliens.Invade(ctx, n, cities, alien.RandomDirGen, func(e model.Event) {
		fmt.Println(e.String())
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, c := range cities {
		fmt.Println(c.String())
	}
}
