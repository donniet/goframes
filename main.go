package main

import (
	"encoding/json"
	"fmt"
	"os"

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
)

const (
	Gravity = 32.174048554 // ft/s^2 = 9.80665 m/s^2
)

type SimpleFrame struct {
	Width     float64
	Height    float64
	Length    float64
	TieHeight float64
	BraceRise float64
	RoofRise  float64
	RoofRun   float64
	Bents     int

	RoofSnowLoad float64
	RoofLiveLoad float64
	RoofDeadLoad float64
	WindLiveLoad float64
}

func WindPressure(airDensity, windSpeed float64) float64 {
	return 0.5 * airDensity / Gravity * windSpeed * windSpeed
}

func windAreaLoad(m *model.Skyciv, windSpeed float64, loadGroup string) {
	presureMag := WindPressure(AirDensity, windSpeed)

	// sides
	for i := 0; i < 2; i++ {
		z := Length / 2 * float64(i)

		// roofTop := Height * Width / 2 * RoofRise / RoofRun

		// left side of building
		nl := []*model.Node{
			m.FindNearestNode(-Width/2, 0, z),
			m.FindNearestNode(-Width/2, Height, z),
			m.FindNearestNode(-Width/2, Height, z+Length/2),
			m.FindNearestNode(-Width/2, 0, z+Length/2),
		}
		if al, err := m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "X"
			al.Mag = presureMag
		}
	}

	// roof
	for i := 0; i < 4; i++ {
		z := Length / 4 * float64(i)

		roofTop := Height + Width/2*RoofRise/RoofRun

		// left side of building
		nl := []*model.Node{
			m.FindNearestNode(-Width/2, Height, z),
			m.FindNearestNode(0, roofTop, z),
			m.FindNearestNode(0, roofTop, z+Length/4),
			m.FindNearestNode(-Width/2, Height, z+Length/4),
		}
		if al, err := m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "X"
			al.Mag = presureMag
		}
	}
}

func roofAreaLoad(m *model.Skyciv, magnitude float64, loadGroup string) {
	for i := 0; i < 4; i++ {
		z := Length / 4 * float64(i)

		roofTop := Height + Width/2*RoofRise/RoofRun

		// left side
		nl := []*model.Node{
			m.FindNearestNode(-Width/2, Height, z),
			m.FindNearestNode(0, roofTop, z),
			m.FindNearestNode(0, roofTop, z+Length/4),
			m.FindNearestNode(-Width/2, Height, z+Length/4),
		}

		if al, err := m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "Y"
			al.Mag = magnitude
		}

		// right side
		nl = []*model.Node{
			m.FindNearestNode(Width/2, Height, z),
			m.FindNearestNode(0, roofTop, z),
			m.FindNearestNode(0, roofTop, z+Length/4),
			m.FindNearestNode(Width/2, Height, z+Length/4),
		}

		if al, err := m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "Y"
			al.Mag = magnitude
		}
	}
}

func bent(m *model.Skyciv, post, tie, rafter, brace, plate *model.Section, z, betweenBents float64) {
	first := len(m.Nodes) == 0

	post00 := post.NewContinuousMember(-Width/2, 0, z, -Width/2, Height, z)
	post01 := post.NewContinuousMember(Width/2, 0, z, Width/2, Height, z)

	topA0 := post00.B()
	topA1 := post01.B()

	post00.A().FixedSupport()
	post01.A().FixedSupport()

	rooftop := m.NewNode(0, Height+Width/2*RoofRise/RoofRun, z)
	rafter.NewContinuousMemberBetweenNodes(post00.B(), rooftop)
	rafter.NewContinuousMemberBetweenNodes(post01.B(), rooftop)

	var tieBeam *model.Member

	if tieNode0, _, err := post00.Split(TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %e\n", err)
		return
	} else if tieNode1, _, err := post01.Split(TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %e\n", err)
		return
	} else {
		tieBeam = tie.NewContinuousMemberBetweenNodes(tieNode0, tieNode1)
	}

	// add braces
	if postBrace0, _, err := post00.Split(-BraceRise); err != nil {
		panic(err)
	} else if postBrace1, _, err := post01.Split(-BraceRise); err != nil {
		panic(err)
	} else if tieBrace0, tieBeamR, err := tieBeam.Split(BraceRise); err != nil {
		panic(err)
	} else if tieBrace1, _, err := tieBeamR.Split(-BraceRise); err != nil {
		panic(err)
	} else {
		brace.NewContinuousMemberBetweenNodes(postBrace0, tieBrace0)
		brace.NewContinuousMemberBetweenNodes(postBrace1, tieBrace1)
	}

	// connect with top plates and plate braces
	if first {
		return
	}

	rtt := m.NewNode(0, Height+Width/2*RoofRise/RoofRun, z-Length/4)
	rtt0 := m.NewNode(-Width/2, Height, z-Length/4)
	rtt1 := m.NewNode(Width/2, Height, z-Length/4)

	rafter.NewContinuousMemberBetweenNodes(rtt0, rtt)
	rafter.NewContinuousMemberBetweenNodes(rtt1, rtt)

	topB0 := m.FindNearestNode(-Width/2, Height, z-betweenBents)
	topB1 := m.FindNearestNode(Width/2, Height, z-betweenBents)

	// connect with plates
	p00 := plate.NewContinuousMemberBetweenNodes(topA0, rtt0)
	p01 := plate.NewContinuousMemberBetweenNodes(topB0, rtt0)
	p10 := plate.NewContinuousMemberBetweenNodes(topA1, rtt1)
	p11 := plate.NewContinuousMemberBetweenNodes(topB1, rtt1)

	// add plate braces
	if pb0, _, err := p00.Split(BraceRise); err != nil {
		panic(err)
	} else if pb1, _, err := p10.Split(BraceRise); err != nil {
		panic(err)
	} else if postAttach0, err := m.FindNearestMemberAndSplitAt(pb0.X, pb0.Y-BraceRise, pb0.Z+BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace: %e", err))
	} else if postAttach1, err := m.FindNearestMemberAndSplitAt(pb1.X, pb1.Y-BraceRise, pb1.Z+BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace, %e", err))
	} else {
		brace.NewContinuousMemberBetweenNodes(pb0, postAttach0)
		brace.NewContinuousMemberBetweenNodes(pb1, postAttach1)
	}

	if pb0, _, err := p01.Split(BraceRise); err != nil {
		panic(err)
	} else if pb1, _, err := p11.Split(BraceRise); err != nil {
		panic(err)
	} else if postAttach0, err := m.FindNearestMemberAndSplitAt(pb0.X, pb0.Y-BraceRise, pb0.Z-BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace: %e", err))
	} else if postAttach1, err := m.FindNearestMemberAndSplitAt(pb1.X, pb1.Y-BraceRise, pb1.Z-BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace, %e", err))
	} else {
		brace.NewContinuousMemberBetweenNodes(pb0, postAttach0)
		brace.NewContinuousMemberBetweenNodes(pb1, postAttach1)
	}

}

func main() {
	m := model.NewModel()

	redPineGreen := m.NewMaterial("Red Pine (green)")
	redPineGreen.Class = model.MaterialClassWood
	redPineGreen.ElasticityModulus = 1280 // ksi
	redPineGreen.Density = 25             // lb/ft^3
	redPineGreen.PoissonsRatio = 0.27
	redPineGreen.YieldStrength = 0.26   // ksi
	redPineGreen.UltimateStrength = 0.3 // ksi

	post := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")
	tie := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")
	rafter := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")
	brace := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "4 x 8")
	plate := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")

	for z := 0.; z <= Length; z += Length / 2 {
		bent(m, post, tie, rafter, brace, plate, z, Length/2)
	}

	roofAreaLoad(m, -RoofSnowLoad, "snow")
	roofAreaLoad(m, -RoofDeadLoad, "dead")
	roofAreaLoad(m, -RoofLiveLoad, "live")

	windAreaLoad(m, WindSpeed, "wind")

	sw := m.NewSelfWeight()
	sw.LoadGroup = "SW1" // for some reason skyciv always uses SW1 for this...
	sw.Y = -1

	m.LoadCombinations.Mapping.DeadCases("dead", "SW1").LiveCases("live").SnowCases("snow")
	m.LoadCombinations.Cases = []model.Case{
		{Name: "ULS: 1. 1.4D", Dead: 1.4},
		{Name: "ULS: 2. 1.2D + 1.6L + 0.5S", Dead: 1.2, Live: 1.6, Snow: 0.5},
		{Name: "ULS: 3. 1.2D + 1.6S + L", Dead: 1.2, Snow: 1.6, Live: 1},
		{Name: "ULS: 3. 1.2D + 1.6S + 0.5W", Dead: 1.2, Snow: 1.6, Wind: 0.5},
		{Name: "ULS: 4. 1.2D + W + L + 0.5S", Dead: 1.2, Wind: 1, Live: 1, Snow: 0.5},
		{Name: "ULS: 5. 1.2D + L + 0.2S", Dead: 1.2, Live: 1, Snow: 0.2},
		{Name: "ULS: 6. 0.9D + W", Dead: 0.9, Wind: 1},
	}

	b, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%e", err)
		return
	}

	os.Stdout.Write(b)
	os.Stdout.Write([]byte("\n"))
}
