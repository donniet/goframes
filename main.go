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
	BraceRise = 2.5
	RoofRise  = 8.
	RoofRun   = 12.
)

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
			al.LoadGroup = "snow"
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
			al.LoadGroup = "snow"
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
	} else if pb1, _, err := p11.Split(-BraceRise); err != nil {
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
	redPineGreen.ElasticityModulus = 1280
	redPineGreen.Density = 25
	redPineGreen.PoissonsRatio = 0.27
	redPineGreen.UltimateStrength = 0.3

	redPine7x9 := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")
	redPine4x8 := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "4 x 8")

	bent(m, redPine7x9, redPine7x9, redPine7x9, redPine4x8, redPine7x9, 0, Length/2)
	bent(m, redPine7x9, redPine7x9, redPine7x9, redPine4x8, redPine7x9, Length/2, Length/2)
	bent(m, redPine7x9, redPine7x9, redPine7x9, redPine4x8, redPine7x9, Length, Length/2)

	roofAreaLoad(m, -0.06, "snow")

	m.NewSelfWeight().LoadGroup = "LG"

	b, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%e", err)
		return
	}

	os.Stdout.Write(b)
	os.Stdout.Write([]byte("\n"))
}
