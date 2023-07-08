package frames

import (
	"github.com/donniet/goframes/model"
)

type Geodesic struct {
	m        *model.Skyciv
	Diameter float64
	Chord    float64

	mats *model.MaterialFile
}

type polar struct {
	theta float64
	fie   float64
}

/* finds the number of triangularuzation steps of an icosohedron before the chord length is below Chord */
func (g *Geodesic) findTriangularization() {
	phi := (1 + math.Sqrt(5))/2
	
	// first start wih an icosohedron
	nodes = []polar{}{
		polar{}
	}
}
