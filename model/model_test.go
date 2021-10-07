package model

import (
	"testing"
)

func TestDistanceTo(T *testing.T) {
	m := NewModel()
	mat := m.NewMaterial("test")
	sec := mat.NewSectionFromLibrary("test")
	mem := sec.NewContinuousMember(0, 0, 0, 1, 0, 0)

	if t, d := mem.DistanceTo(0, 0, 0); d != 0 || t != 0 {
		T.Errorf("distance to origin non-zero")
	}
	if t, d := mem.DistanceTo(1, 0, 0); d != 0 || t != 1 {
		T.Errorf("distance to 1,0,0 not zero, but %f %f", d, t)
	}
}

func TestMemberParent(t *testing.T) {

}

// func TestSplit(T *testing.T) {
// 	m := NewModel()
// 	mat := m.NewMaterial("test")
// 	sec := mat.NewSectionFromLibrary("test")
// 	mem := sec.NewContinuousMember(0, 0, 0, 2, 0, 0)

// 	if l := len(mem.parent.children); l != 1 {
// 		T.Errorf("children length %d instead of 1", l)
// 	}

// 	n, m2, err := mem.Split(1.)

// 	if err != nil {
// 		T.Error(err)
// 	}
// 	if n.X != 1. {
// 		T.Errorf("split does not cut member at the right spot")
// 	}
// 	if mem.parent != m2.parent {
// 		T.Errorf("parents are not equal")
// 	}
// 	if len(mem.parent.children) != 2 {
// 		T.Errorf("children not equal to 2")
// 		return
// 	}
// 	if mem.parent.children[0] != mem || mem.parent.children[1] != m2 {
// 		T.Errorf("children are incorrect")
// 	}
// }
