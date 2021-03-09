package main

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"within.website/ln"
)

type Planet struct {
	Name      string `json:"name"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Owner     int    `json:"owner"`
	ShipCount int    `json:"ship_count"`
}

type Expedition struct {
	ID             int    `json:"id"`
	Origin         string `json:"origin"`
	Destination    string `json:"destination"`
	TurnsRemaining int    `json:"turns_remaining"`
	Owner          int    `json:"owner"`
	ShipCount      int    `json:"ship_count"`
}

type State struct {
	Planets     []Planet     `json:"planets"`
	Expeditions []Expedition `json:"expeditions"`
}

type Move struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	ShipCount   int    `json:"ship_count"`
}

func process(states <-chan State, out chan<- []Move) {
	for state := range states {
		// setup the empty move array
		moves := make([]Move, 0)

		// Get the current game state in maps
		status, planets, fleet := ParseGameState(state)

		// go through our fleet to determine targets
		for name, planet := range fleet {
			for 1 < planet.ShipCount {
				targetPl := ""
				selected := 100000000
				// Iterate over the possible targets
				for target, owned := range status {
					// if we own the planet, skip it
					if owned {
						continue
					}

					// determine the score of the planet
					score := planet.CalcDist(planets[target])
					if score < selected {
						selected = score
						targetPl = target
					}
				}

				send := 2
				if send <= planet.ShipCount {
					planet.ShipCount -= send
					tempMove := CreateMove(name, targetPl, send)
					if 0 < tempMove.ShipCount {
						moves = append(moves, tempMove)
					}
				}
			}
		}
		out<-moves
	}
	close(out)
}

func main() {
	var state State
	ctx := context.Background()

	stateChan := make(chan State)
	moveChan := make(chan []Move)

	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		// decode the incoming JSON, skip EOF errors
		err := decoder.Decode(&state)
		switch err {
		case io.EOF:
			break
		case nil:
		default:
			ln.Error(ctx, err)
			break
		}

		go process(stateChan, moveChan)
		stateChan<-state

		out := struct {
			Moves []Move `json:"moves"`
		}{
			Moves: <-moveChan,
		}
		encoder.Encode(out)
	}
	close(stateChan)
}

func (src *Planet) CalcDist(dest *Planet) int {
	diffX := src.X - dest.X
	diffY := src.Y - dest.Y

	if diffX < 0 {
		diffX *= -1
	}
	if diffY < 0 {
		diffY *= -1
	}

	dist := (diffX * diffX) + (diffY * diffY)
	score := dist * dest.ShipCount
	return score
}

func CreateMove(src, dest string, count int) Move {
	return Move{
		Origin:      src,
		Destination: dest,
		ShipCount:   count,
	}
}

func MaxFleet(fleet map[string]*Planet) *Planet {
	max := &Planet{ShipCount: -1}
	for _, planet := range fleet {
		if 2 < planet.ShipCount && max.ShipCount < planet.ShipCount {
			max = planet
		}
	}
	return max
}

func ParseGameState(state State) (map[string]bool, map[string]*Planet, map[string]*Planet) {
	status := make(map[string]bool)
	planets := make(map[string]*Planet)
	fleet := make(map[string]*Planet)
	for _, planet := range state.Planets {

		pl := Planet{}
		pl.Name = planet.Name
		pl.Owner = planet.Owner
		pl.ShipCount = planet.ShipCount
		pl.X = planet.X
		pl.Y = planet.X
		planets[planet.Name] = &pl

		if planet.Owner == 1 {
			status[planet.Name] = true
			fleet[planet.Name] = &pl
		} else {
			status[planet.Name] = false
		}

	}
	return status, planets, fleet
}
