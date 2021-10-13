package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/donniet/goframes/frames"
	"github.com/donniet/goframes/model"
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

	materialFile string
)

const (
	Gravity = 32.174048554 // ft/s^2 = 9.80665 m/s^2
)

func init() {
	flag.StringVar(&materialFile, "materials", "materials.json", "path to materials json file")
}

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

	var mats *model.MaterialFile
	if f, err := os.Open(materialFile); err != nil {
		panic(err)
	} else if mats, err = model.ReadMaterials(f); err != nil {
		panic(err)
	}

	f := &frames.Yurt{
		Diameter:       24,
		CrownDiameter:  3,
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
		MaterialFile: mats,
	}
	f.Build("Red Pine")

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	if err := enc.Encode(f.Model()); err != nil {
		panic(err)
	}
	fmt.Println()
}
