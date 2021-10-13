package materials

import (
	"github.com/donniet/goframes/model"
)

func Create(m *model.Skyciv) map[string]*model.Material {
	r := make(map[string]*model.Material)

	redPineGreen := m.NewMaterial("Red Pine (green)")
	redPineGreen.Class = model.MaterialClassWood
	redPineGreen.ElasticityModulus = 1280 // ksi
	redPineGreen.Density = 25             // lb/ft^3
	redPineGreen.PoissonsRatio = 0.27
	redPineGreen.YieldStrength = 0.26   // ksi
	redPineGreen.UltimateStrength = 0.3 // ksi
	r[redPineGreen.Name] = redPineGreen

	redPineDry := m.NewMaterial("Red Pine")
	redPineDry.Class = model.MaterialClassWood
	redPineDry.ElasticityModulus = 1630 // ksi
	redPineDry.Density = 28.7168619     // lb/ft^3
	redPineDry.PoissonsRatio = 0.27
	redPineDry.YieldStrength = 0.300    // ksi  0.6 according to matweb??
	redPineDry.UltimateStrength = 0.460 // ksi
	r[redPineDry.Name] = redPineDry

	aspen := m.NewMaterial("Aspen")
	aspen.Class = model.MaterialClassWood
	aspen.ElasticityModulus = 1170 // ksi
	aspen.Density = 23.7           // lb/ft^3
	aspen.PoissonsRatio = 0.37

	// aspen.YieldStrength =

	return r
}
