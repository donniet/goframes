package frames

import (
	"math"

	"github.com/donniet/goframes/model"
)

type Yurt struct {
	m              *model.Skyciv
	Diameter       float64
	CrownDiameter  float64
	MaxPostSpacing float64
	Height         float64
	RoofRise       float64
	RoofRun        float64
	BraceRise      float64

	RoofSnowLoad float64
	RoofLiveLoad float64
	RoofDeadLoad float64
	// WindLiveLoad float64
	WindSpeed    float64
	AirDensity   float64
	MaterialFile *model.MaterialFile

	posts     []*model.ContinuousMember
	splits    []*model.Node
	tops      []*model.Node
	topsplits []*model.Node
}

func (y *Yurt) Model() *model.Skyciv {
	return y.m
}

func (y *Yurt) roofAreaLoad(mag float64, loadGroup string) {
	for i, j := 0, 1; i < len(y.posts); i, j = i+1, j+1 {
		p0, p1 := y.posts[i], y.posts[j%len(y.posts)]

		if al, err := y.m.NewAreaLoad(y.tops[i], y.topsplits[i], y.splits[i], p0.End()); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "Y"
			al.Mag = mag
		}

		if al, err := y.m.NewAreaLoad(y.topsplits[i], y.tops[j%len(y.posts)], p1.End(), y.splits[i]); err != nil {
			panic(err)
		} else {
			al.LoadGroup = loadGroup
			al.Direction = "Y"
			al.Mag = mag
		}
	}
}

func (y *Yurt) Build(materialName string) error {
	y.m = model.NewModel(y.MaterialFile)

	post := y.m.NewSectionFromLibrary(y.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "8 x 10")
	tie := y.m.NewSectionFromLibrary(y.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "4 x 8")
	rafter := y.m.NewSectionFromLibrary(y.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "4 x 8")
	// brace := y.m.NewSectionFromLibrary(y.m.Materials[materialName], "American", "NDS", "Sawn Lumber", "4 x 8")

	// determine number of posts
	count := math.Ceil(y.Diameter * math.Pi / y.MaxPostSpacing)
	r, cr := 0.5*y.Diameter, 0.5*y.CrownDiameter

	// top of roof node (maybe this should be a circle or something instead?)
	// y.top = y.m.NewNode(0, y.Height+y.RoofRise*0.5*y.Diameter/y.RoofRun, 0)

	// create posts
	for i := 0.; i < count; i += 1. {
		theta := 2. * i * math.Pi / count

		x, z := r*math.Cos(theta), r*math.Sin(theta)
		p := y.m.NewContinuousMember(post, x, 0, z, x, y.Height, z)
		p.RotationAngle = -theta * 180. / math.Pi
		// supported at the base
		p.Begin().FixedSupport()

		y.posts = append(y.posts, p)

		x, z = cr*math.Cos(theta), cr*math.Sin(theta)

		top := y.m.NewNode(x, y.Height+y.RoofRise*0.5*y.Diameter/y.RoofRun, z)

		y.m.NewContinuousMemberBetweenNodes(rafter, p.End(), top)

		y.tops = append(y.tops, top)

	}

	// connect all the posts with ties and then brace
	for i, j := 0, 1; i < len(y.posts); i, j = i+1, j+1 {
		p0, p1 := y.posts[i], y.posts[j%len(y.posts)]
		t0, t1 := y.tops[i], y.tops[j%len(y.tops)]

		t := y.m.NewContinuousMemberBetweenNodes(tie, p0.End(), p1.End())
		tt := y.m.NewContinuousMemberBetweenNodes(tie, t0, t1)

		if s, err := t.SplitPercent(0.5); err != nil {
			panic(err)
		} else if ts, err := tt.SplitPercent(0.5); err != nil {
			panic(err)
		} else {
			y.m.NewContinuousMemberBetweenNodes(rafter, s, ts)
			y.splits = append(y.splits, s)
			y.topsplits = append(y.topsplits, ts)
		}

		// // why not brace the posts while we are here?
		// if _, err := p0.Brace(t, brace, y.BraceRise, model.QuadrantNP); err != nil {
		// 	panic(err)
		// }
		// if _, err := p1.Brace(t, brace, y.BraceRise, model.QuadrantNN); err != nil {
		// 	panic(err)
		// }
	}

	y.roofAreaLoad(-y.RoofDeadLoad, "dead")
	y.roofAreaLoad(-y.RoofLiveLoad, "live")
	y.roofAreaLoad(-y.RoofSnowLoad, "snow")

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
	return nil
}
