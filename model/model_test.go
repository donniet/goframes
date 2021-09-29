package model

import (
	"testing"
)

func TestDistanceTo(T *testing.T) {
	m := NewModel()
	mat := m.NewMaterial("test")
	sec := mat.NewSectionFromLibrary("test")
	mem := sec.NewContinuousMember(0, 0, 0, 1, 0, 0)

	if t, d := mem.distanceTo(0, 0, 0); d != 0 || t != 0 {
		T.Errorf("distance to origin non-zero")
	}
	if t, d := mem.distanceTo(1, 0, 0); d != 0 || t != 1 {
		T.Errorf("distance to 1,0,0 not zero, but %f %f", d, t)
	}
}
