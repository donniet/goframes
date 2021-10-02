package frames

import (
	"fmt"
	"os"

	"github.com/donniet/goframes/model"
)

const (
	Gravity = 32.174048554
)

type SimpleFrame struct {
	m         *model.Skyciv
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
	// WindLiveLoad float64
	WindSpeed  float64
	AirDensity float64
}

func (f *SimpleFrame) Model() *model.Skyciv {
	return f.m
}

func (f *SimpleFrame) Build() {
	f.m = model.NewModel()

	redPineGreen := f.m.NewMaterial("Red Pine (green)")
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

	for z := 0.; z <= f.Length; z += f.Length / 2 {
		f.bent(post, tie, rafter, brace, plate, z, f.Length/2)
	}

	f.roofAreaLoad(-f.RoofSnowLoad, "snow")
	f.roofAreaLoad(-f.RoofDeadLoad, "dead")
	f.roofAreaLoad(-f.RoofLiveLoad, "live")

	f.windAreaLoad(f.WindSpeed, "wind")

	sw := f.m.NewSelfWeight()
	sw.LoadGroup = "SW1" // for some reason skyciv always uses SW1 for this...
	sw.Y = -1

	f.m.LoadCombinations.Mapping.DeadCases("dead", "SW1").LiveCases("live").SnowCases("snow")
	f.m.LoadCombinations.Cases = []model.Case{
		{Name: "ULS: 1. 1.4D", Dead: 1.4},
		{Name: "ULS: 2. 1.2D + 1.6L + 0.5S", Dead: 1.2, Live: 1.6, Snow: 0.5},
		{Name: "ULS: 3. 1.2D + 1.6S + L", Dead: 1.2, Snow: 1.6, Live: 1},
		{Name: "ULS: 3. 1.2D + 1.6S + 0.5W", Dead: 1.2, Snow: 1.6, Wind: 0.5},
		{Name: "ULS: 4. 1.2D + W + L + 0.5S", Dead: 1.2, Wind: 1, Live: 1, Snow: 0.5},
		{Name: "ULS: 5. 1.2D + L + 0.2S", Dead: 1.2, Live: 1, Snow: 0.2},
		{Name: "ULS: 6. 0.9D + W", Dead: 0.9, Wind: 1},
	}

}

func WindPressure(airDensity, windSpeed float64) float64 {
	return 0.5 * airDensity / Gravity * windSpeed * windSpeed
}

func (f *SimpleFrame) windAreaLoad(windSpeed float64, loadGroup string) {
	presureMag := WindPressure(f.AirDensity, windSpeed)

	// sides
	for i := 0; i < 2; i++ {
		z := f.Length / 2 * float64(i)

		// roofTop := Height * Width / 2 * RoofRise / RoofRun

		// left side of building
		nl := []*model.Node{
			f.m.FindNearestNode(-f.Width/2, 0, z),
			f.m.FindNearestNode(-f.Width/2, f.Height, z),
			f.m.FindNearestNode(-f.Width/2, f.Height, z+f.Length/2),
			f.m.FindNearestNode(-f.Width/2, 0, z+f.Length/2),
		}
		if al, err := f.m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "X"
			al.Mag = presureMag
		}
	}

	// roof
	for i := 0; i < 4; i++ {
		z := f.Length / 4 * float64(i)

		roofTop := f.Height + f.Width/2*f.RoofRise/f.RoofRun

		// left side of building
		nl := []*model.Node{
			f.m.FindNearestNode(-f.Width/2, f.Height, z),
			f.m.FindNearestNode(0, roofTop, z),
			f.m.FindNearestNode(0, roofTop, z+f.Length/4),
			f.m.FindNearestNode(-f.Width/2, f.Height, z+f.Length/4),
		}
		if al, err := f.m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "X"
			al.Mag = presureMag
		}
	}
}

func (f *SimpleFrame) roofAreaLoad(magnitude float64, loadGroup string) {
	for i := 0; i < 4; i++ {
		z := f.Length / 4 * float64(i)

		roofTop := f.Height + f.Width/2*f.RoofRise/f.RoofRun

		// left side
		nl := []*model.Node{
			f.m.FindNearestNode(-f.Width/2, f.Height, z),
			f.m.FindNearestNode(0, roofTop, z),
			f.m.FindNearestNode(0, roofTop, z+f.Length/4),
			f.m.FindNearestNode(-f.Width/2, f.Height, z+f.Length/4),
		}

		if al, err := f.m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "Y"
			al.Mag = magnitude
		}

		// right side
		nl = []*model.Node{
			f.m.FindNearestNode(f.Width/2, f.Height, z),
			f.m.FindNearestNode(0, roofTop, z),
			f.m.FindNearestNode(0, roofTop, z+f.Length/4),
			f.m.FindNearestNode(f.Width/2, f.Height, z+f.Length/4),
		}

		if al, err := f.m.NewAreaLoad(nl); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "Y"
			al.Mag = magnitude
		}
	}
}

func (f *SimpleFrame) bent(post, tie, rafter, brace, plate *model.Section, z, betweenBents float64) {
	first := len(f.m.Nodes) == 0

	post00 := post.NewContinuousMember(-f.Width/2, 0, z, -f.Width/2, f.Height, z)
	post01 := post.NewContinuousMember(f.Width/2, 0, z, f.Width/2, f.Height, z)

	topA0 := post00.B()
	topA1 := post01.B()

	post00.A().FixedSupport()
	post01.A().FixedSupport()

	rooftop := f.m.NewNode(0, f.Height+f.Width/2*f.RoofRise/f.RoofRun, z)
	rafter.NewContinuousMemberBetweenNodes(post00.B(), rooftop)
	rafter.NewContinuousMemberBetweenNodes(post01.B(), rooftop)

	var tieBeam *model.Member

	if tieNode0, _, err := post00.Split(f.TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %e\n", err)
		return
	} else if tieNode1, _, err := post01.Split(f.TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %e\n", err)
		return
	} else {
		tieBeam = tie.NewContinuousMemberBetweenNodes(tieNode0, tieNode1)
	}

	// add braces
	if _, err := post00.Brace(tieBeam, brace, f.BraceRise); err != nil {
		panic(err)
	}
	if _, err := post01.Brace(tieBeam, brace, f.BraceRise); err != nil {
		panic(err)
	}
	// if postBrace0, _, err := post00.Split(-f.BraceRise); err != nil {
	// 	panic(err)
	// } else if postBrace1, _, err := post01.Split(-f.BraceRise); err != nil {
	// 	panic(err)
	// } else if tieBrace0, tieBeamR, err := tieBeam.Split(f.BraceRise); err != nil {
	// 	panic(err)
	// } else if tieBrace1, _, err := tieBeamR.Split(-f.BraceRise); err != nil {
	// 	panic(err)
	// } else {
	// 	brace.NewContinuousMemberBetweenNodes(postBrace0, tieBrace0)
	// 	brace.NewContinuousMemberBetweenNodes(postBrace1, tieBrace1)
	// }

	// connect with top plates and plate braces
	if first {
		return
	}

	rtt := f.m.NewNode(0, f.Height+f.Width/2*f.RoofRise/f.RoofRun, z-f.Length/4)
	rtt0 := f.m.NewNode(-f.Width/2, f.Height, z-f.Length/4)
	rtt1 := f.m.NewNode(f.Width/2, f.Height, z-f.Length/4)

	rafter.NewContinuousMemberBetweenNodes(rtt0, rtt)
	rafter.NewContinuousMemberBetweenNodes(rtt1, rtt)

	topB0 := f.m.FindNearestNode(-f.Width/2, f.Height, z-betweenBents)
	topB1 := f.m.FindNearestNode(f.Width/2, f.Height, z-betweenBents)

	// connect with plates
	p00 := plate.NewContinuousMemberBetweenNodes(topA0, rtt0)
	p01 := plate.NewContinuousMemberBetweenNodes(topB0, rtt0)
	p10 := plate.NewContinuousMemberBetweenNodes(topA1, rtt1)
	p11 := plate.NewContinuousMemberBetweenNodes(topB1, rtt1)

	// add plate braces
	if pb0, _, err := p00.Split(f.BraceRise); err != nil {
		panic(err)
	} else if pb1, _, err := p10.Split(f.BraceRise); err != nil {
		panic(err)
	} else if postAttach0, err := f.m.FindNearestMemberAndSplitAt(pb0.X, pb0.Y-f.BraceRise, pb0.Z+f.BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace: %e", err))
	} else if postAttach1, err := f.m.FindNearestMemberAndSplitAt(pb1.X, pb1.Y-f.BraceRise, pb1.Z+f.BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace, %e", err))
	} else {
		brace.NewContinuousMemberBetweenNodes(pb0, postAttach0)
		brace.NewContinuousMemberBetweenNodes(pb1, postAttach1)
	}

	if pb0, _, err := p01.Split(f.BraceRise); err != nil {
		panic(err)
	} else if pb1, _, err := p11.Split(f.BraceRise); err != nil {
		panic(err)
	} else if postAttach0, err := f.m.FindNearestMemberAndSplitAt(pb0.X, pb0.Y-f.BraceRise, pb0.Z-f.BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace: %e", err))
	} else if postAttach1, err := f.m.FindNearestMemberAndSplitAt(pb1.X, pb1.Y-f.BraceRise, pb1.Z-f.BraceRise); err != nil {
		panic(fmt.Sprintf("could not find the post attachment for brace, %e", err))
	} else {
		brace.NewContinuousMemberBetweenNodes(pb0, postAttach0)
		brace.NewContinuousMemberBetweenNodes(pb1, postAttach1)
	}

}
