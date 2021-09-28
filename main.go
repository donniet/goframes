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

func main() {
	m := model.NewModel()

	redPineGreen := m.NewMaterial("Red Pine (green)")
	redPineGreen.Class = model.MaterialClassWood
	redPineGreen.ElasticityModulus = 1280
	redPineGreen.Density = 25
	redPineGreen.PoissonsRatio = 0.27
	redPineGreen.UltimateStrength = 0.3

	redPine7x9 := redPineGreen.NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")

	post00 := redPine7x9.NewContinuousMember(-Width/2, 0, 0, -Width/2, Height, 0)
	post01 := redPine7x9.NewContinuousMember(Width/2, 0, 0, Width/2, Height, 0)

	// var tieBeam *model.Member

	if tieNode0, _, err := post00.Split(TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %e\n", err)
		return
	} else if tieNode1, _, err := post01.Split(TieHeight); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting post for tie beam: %e\n", err)
		return
	} else {
		_ = redPine7x9.NewContinuousMemberBetweenNodes(tieNode0, tieNode1)
	}

	b, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%e", err)
		return
	}

	os.Stdout.Write(b)
}
