package frames

import (
	"fmt"
	"os"

	"github.com/donniet/goframes/model"
)

type SimpleFrame struct {
	m            *model.Skyciv
	Width        float64
	Height       float64
	Length       float64
	TieHeight    float64
	BraceRise    float64
	RoofRise     float64
	RoofRun      float64
	Bents        int
	MaterialFile *model.MaterialFile

	RoofSnowLoad float64
	RoofLiveLoad float64
	RoofDeadLoad float64
	// WindLiveLoad float64
	WindSpeed  float64
	AirDensity float64

	posts []*model.ContinuousMember
}

func (f *SimpleFrame) Model() *model.Skyciv {
	return f.m
}

func (f *SimpleFrame) Build(materialName string) {
	f.m = model.NewModel(f.MaterialFile)

	post := f.m.NewSectionFromLibrary(f.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "8 x 10")
	tie := f.m.NewSectionFromLibrary(f.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "8 x 10")
	rafter := f.m.NewSectionFromLibrary(f.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "8 x 10")
	brace := f.m.NewSectionFromLibrary(f.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "4 x 8")
	plate := f.m.NewSectionFromLibrary(f.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "8 x 10")

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
		if al, err := f.m.NewAreaLoad(nl...); err != nil {
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
		if al, err := f.m.NewAreaLoad(nl...); err != nil {
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

		if al, err := f.m.NewAreaLoad(nl...); err != nil {
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

		if al, err := f.m.NewAreaLoad(nl...); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "Y"
			al.Mag = magnitude
		}
	}
}

func (f *SimpleFrame) bent(post, tie, rafter, brace, plate *model.Section, z, betweenBents float64) {
	first := len(f.posts) == 0

	post00 := f.m.NewContinuousMember(post, -f.Width/2, 0, z, -f.Width/2, f.Height, z)
	post01 := f.m.NewContinuousMember(post, f.Width/2, 0, z, f.Width/2, f.Height, z)

	post00.Begin().FixedSupport()
	post01.Begin().FixedSupport()

	rooftop := f.m.NewNode(0, f.Height+f.Width/2*f.RoofRise/f.RoofRun, z)
	f.m.NewContinuousMemberBetweenNodes(rafter, post00.End(), rooftop)
	f.m.NewContinuousMemberBetweenNodes(rafter, post01.End(), rooftop)

	var tieBeam *model.ContinuousMember

	if tieNode0, err := post00.Split(f.TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %v\n", err)
		return
	} else if tieNode1, err := post01.Split(f.TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %v\n", err)
		return
	} else {
		tieBeam = f.m.NewContinuousMemberBetweenNodes(tie, tieNode0, tieNode1)
	}

	// add braces
	if _, err := post00.Brace(tieBeam, brace, f.BraceRise, model.QuadrantNP); err != nil {
		panic(err)
	}
	if _, err := post01.Brace(tieBeam, brace, f.BraceRise, model.QuadrantNN); err != nil {
		panic(err)
	}

	// connect with top plates and plate braces
	if !first {
		// roof middle nodes
		rtt := f.m.NewNode(0, f.Height+f.Width/2*f.RoofRise/f.RoofRun, z-f.Length/4)
		rtt0 := f.m.NewNode(-f.Width/2, f.Height, z-f.Length/4)
		rtt1 := f.m.NewNode(f.Width/2, f.Height, z-f.Length/4)

		// middle rafters
		f.m.NewContinuousMemberBetweenNodes(rafter, rtt0, rtt)
		f.m.NewContinuousMemberBetweenNodes(rafter, rtt1, rtt)

		prev0 := f.posts[0]
		prev1 := f.posts[1]

		// connect with plates
		plateA0 := f.m.NewContinuousMemberBetweenNodes(plate, post00.End(), rtt0)
		plateB0 := f.m.NewContinuousMemberBetweenNodes(plate, prev0.End(), rtt0)
		plateA1 := f.m.NewContinuousMemberBetweenNodes(plate, post01.End(), rtt1)
		plateB1 := f.m.NewContinuousMemberBetweenNodes(plate, prev1.End(), rtt1)

		// add plate braces
		if _, err := post00.Brace(plateA0, brace, f.BraceRise, model.QuadrantNP); err != nil {
			panic(err)
		}
		if _, err := post01.Brace(plateA1, brace, f.BraceRise, model.QuadrantNP); err != nil {
			panic(err)
		}
		if _, err := f.posts[0].Brace(plateB0, brace, f.BraceRise, model.QuadrantNP); err != nil {
			panic(err)
		}
		if _, err := f.posts[1].Brace(plateB1, brace, f.BraceRise, model.QuadrantNP); err != nil {
			panic(err)
		}
	}
	f.posts = append([]*model.ContinuousMember{post00, post01}, f.posts...)
}
