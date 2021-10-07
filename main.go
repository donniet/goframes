package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/donniet/goframes/frames"
)

var (
	Width     = 12.
	Height    = 10.
	Length    = 20.
	TieHeight = 8.5
	BraceRise = 3.
	RoofRise  = 8.
	RoofRun   = 12.
	Bents     = 3

	RoofLiveLoad = 0.02 // ksf == 20 psf
	RoofSnowLoad = 0.06 // ksf == 60 psf
	RoofDeadLoad = 0.02 // ksf == 20 psf
	WindLiveLoad = 0.

	AirDensity = 0.000075 // kip/ft^3 = 0.075 lb/ft^3 = 1.2 kg/m^3
	WindSpeed  = 177.     // ft/s = 120 mph = 54 m/s
)

const (
	Gravity = 32.174048554 // ft/s^2 = 9.80665 m/s^2
)

func main() {
	// f := &frames.SimpleFrame{
	// 	Width:        Width,
	// 	Height:       Height,
	// 	Length:       Length,
	// 	TieHeight:    TieHeight,
	// 	BraceRise:    BraceRise,
	// 	RoofRise:     RoofRise,
	// 	RoofRun:      RoofRun,
	// 	Bents:        Bents,
	// 	RoofSnowLoad: RoofSnowLoad,
	// 	RoofLiveLoad: RoofLiveLoad,
	// 	RoofDeadLoad: RoofDeadLoad,
	// 	WindSpeed:    WindSpeed,
	// 	AirDensity:   AirDensity,
	// }
	// f.Build()

	f := &frames.Yurt{
		Diameter:       24,
		Height:         10,
		RoofRise:       4,
		RoofRun:        12,
		MaxPostSpacing: 12,
		BraceRise:      3,

		RoofSnowLoad: RoofSnowLoad,
		RoofLiveLoad: RoofLiveLoad,
		RoofDeadLoad: RoofDeadLoad,
		WindSpeed:    WindSpeed,
		AirDensity:   AirDensity,
	}
	f.Build()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	if err := enc.Encode(f.Model()); err != nil {
		panic(err)
	}
	fmt.Println()
}
