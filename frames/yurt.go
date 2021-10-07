package frames

import (
	"math"

	"github.com/donniet/goframes/materials"
	"github.com/donniet/goframes/model"
)

type Yurt struct {
	m              *model.Skyciv
	Diameter       float64
	MaxPostSpacing float64
	Height         float64
	RoofRise       float64
	RoofRun        float64
	BraceRise      float64

	RoofSnowLoad float64
	RoofLiveLoad float64
	RoofDeadLoad float64
	// WindLiveLoad float64
	WindSpeed  float64
	AirDensity float64

	posts  []*model.ContinuousMember
	splits []*model.Node
}

func (y *Yurt) Model() *model.Skyciv {
	return y.m
}

func (y *Yurt) Build() {
	y.m = model.NewModel()

	mats := materials.Create(y.m)

	post := mats["Red Pine"].NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")
	tie := mats["Red Pine"].NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")
	rafter := mats["Red Pine"].NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "8 x 10")
	brace := mats["Red Pine"].NewSectionFromLibrary("American", "NDS", "Sawn Lumber", "4 x 8")

	// determine number of posts
	count := math.Ceil(y.Diameter * math.Pi / y.MaxPostSpacing)
	r := 0.5 * y.Diameter

	// create posts
	for i := 0.; i < count; i += 1. {
		x, z := r*math.Cos(2.*i*math.Pi/count), r*math.Sin(2.*i*math.Pi/count)
		p := post.NewContinuousMember(x, 0, z, x, y.Height, z)
		// supported at the base
		p.Begin().FixedSupport()

		y.posts = append(y.posts, p)
	}

	// connect all the posts with ties and then brace
	for i, j := 0, 1; i < len(y.posts); i, j = i+1, j+1 {
		p0, p1 := y.posts[i], y.posts[j%len(y.posts)]
		t := tie.NewContinuousMemberBetweenNodes(p0.End(), p1.End())

		if s, err := t.SplitPercent(0.5); err != nil {
			panic(err)
		} else {
			y.splits = append(y.splits, s)
		}

		// why not brace the posts while we are here?
		if _, err := p0.Brace(t, brace, y.BraceRise, model.QuadrantNP); err != nil {
			panic(err)
		}
		if _, err := p1.Brace(t, brace, y.BraceRise, model.QuadrantNN); err != nil {
			panic(err)
		}
	}

	// top of roof node (maybe this should be a circle or something instead?)
	top := y.m.NewNode(0, y.Height+y.RoofRise*0.5*y.Diameter/y.RoofRun, 0)

	// add rafters
	for _, p := range y.posts {
		rafter.NewContinuousMemberBetweenNodes(p.End(), top)
	}
	for _, s := range y.splits {
		rafter.NewContinuousMemberBetweenNodes(s, top)
	}

	sw := y.m.NewSelfWeight()
	sw.LoadGroup = "SW1"
	sw.Y = -1

	y.m.LoadCombinations.Mapping.DeadCases("dead", "SW1").LiveCases("live").SnowCases("snow")
	y.m.LoadCombinations.Cases = []model.Case{
		{Name: "ULS: 1. 1.4D", Dead: 1.4},
		{Name: "ULS: 2. 1.2D + 1.6L + 0.5S", Dead: 1.2, Live: 1.6, Snow: 0.5},
		{Name: "ULS: 3. 1.2D + 1.6S + L", Dead: 1.2, Snow: 1.6, Live: 1},
		{Name: "ULS: 3. 1.2D + 1.6S + 0.5W", Dead: 1.2, Snow: 1.6, Wind: 0.5},
		{Name: "ULS: 4. 1.2D + W + L + 0.5S", Dead: 1.2, Wind: 1, Live: 1, Snow: 0.5},
		{Name: "ULS: 5. 1.2D + L + 0.2S", Dead: 1.2, Live: 1, Snow: 0.2},
		{Name: "ULS: 6. 0.9D + W", Dead: 0.9, Wind: 1},
	}
}
