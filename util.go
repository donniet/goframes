package main

import (
	"fmt"
	"math"

	"github.com/donniet/goframes/model"
)

type Slope struct {
	Rise float64
	Run  float64
}

type nodeItem struct {
	x, y, z float64
	id      int
}

type nodeSet struct {
	nodes     map[string]*nodeItem
	precision int
	count     int
}

func newNodeSet(precision int) *nodeSet {
	return &nodeSet{
		nodes:     make(map[string]*nodeItem),
		precision: precision,
		count:     0,
	}
}

// use precision to collapse nearby points in case of rounding errors
func (n *nodeSet) p2s(x, y, z float64) string {
	fac := float64(math.Pow10(n.precision))
	return fmt.Sprintf("%d,%d,%d", int(x*fac), int(y*fac), int(z*fac))
}

func (n *nodeSet) node(x, y, z float64) int {
	key := n.p2s(x, y, z)
	if item, ok := n.nodes[key]; !ok {
		// create a new one
		n.count++
		n.nodes[key] = &nodeItem{x, y, z, n.count}
		return n.count
	} else {
		return item.id
	}
}

type sectionSet struct {
	set            map[string]*model.Section
	sectionVersion int
}

func newSectionSet(sectionVersion int) *sectionSet {
	return &sectionSet{
		set:            make(map[string]*model.Section),
		sectionVersion: sectionVersion,
	}
}

func torsionConstant(breadth, depth float64) float64 {
	a := depth / 2
	b := breadth / 2

	if a < b {
		t := a
		a = b
		b = t
	}

	return a * b * b * b * (16./3. - 3.36*b/a*(1.-b*b*b*b/12./a/a/a/a))
}

func (sects *sectionSet) rectangular(breadth, depth float64, material int) (s *model.Section) {
	key := fmt.Sprintf("%fx%f,%d", breadth, depth, material)

	s, ok := sects.set[key]
	if !ok {
		name := fmt.Sprintf("%fx%f", breadth, depth)
		s = &model.Section{
			Version:    sects.sectionVersion,
			Name:       name,
			Area:       breadth * depth,
			Iz:         breadth * depth * depth * depth / 12,
			Iy:         depth * breadth * breadth * breadth / 12,
			MaterialId: material,
			J:          torsionConstant(breadth, depth),
			Id:         len(sects.set) + 1,
		}
		sects.set[key] = s
	}
	return
}

type memberSet struct {
	set map[string]*model.Member
}

func newMemberSet() *memberSet {
	return &memberSet{
		set: make(map[string]*model.Member),
	}
}

func (ms *memberSet) member(nodeA, nodeB int, section int) *model.Member {
	key := fmt.Sprintf("%d,%d", nodeA, nodeB)
	m, ok := ms.set[key]
	if !ok {
		m = &model.Member{
			Type:          "normal_continuous",
			CableLength:   nil,
			NodeA:         nodeA,
			NodeB:         nodeB,
			SectionId:     section,
			RotationAngle: 0,
			FixityA:       "FFFFFF",
			FixityB:       "FFFFFF",
			OffsetAx:      0,
			OffsetAy:      0,
			OffsetAz:      0,
			OffsetBx:      0,
			OffsetBy:      0,
			OffsetBz:      0,
			Id:            len(ms.set) + 1,
		}
		ms.set[key] = m
	}
	return m
}

type SimplePostBeam struct {
	Width     float64
	Length    float64
	Height    float64
	Bents     int
	Roof      Slope
	TieOffset float64
	BraceRise float64

	Precision int

	nodes     *nodeSet
	sections  *sectionSet
	members   *memberSet
	materials map[string]*model.Material
}

func (m *SimplePostBeam) init() {
	m.materials = make(map[string]*model.Material)
	m.sections = newSectionSet(model.SectionVersion)
	m.nodes = newNodeSet(3)
	m.members = newMemberSet()
}

type SimpleBent struct {
	structure                                            *SimplePostBeam
	position                                             float64
	postSection, tieSection, braceSection, rafterSection int
}

func (b *SimpleBent) Bent() {
	b.postGroundToTieBrace()
	b.postTieBraceToPlateBrace()
	b.postPlateBraceToTie()
	b.postTieToPlate()
	b.rafter()
	b.tiePostToBrace()
	b.tieBraceToBrace()
	b.tieBrace()
}

func (b *SimpleBent) postGroundToTieBrace() {
	x := -b.structure.Width / 2
	y0 := 0.
	y1 := b.structure.Height - b.structure.TieOffset - b.structure.BraceRise
	z := b.position

	b.structure.Member(x, y0, z, x, y1, z, b.postSection)
	b.structure.Member(-x, y0, z, -x, y1, z, b.postSection)
}
func (b *SimpleBent) postTieBraceToPlateBrace() {
	x := -b.structure.Width / 2
	y0 := b.structure.Height - b.structure.TieOffset - b.structure.BraceRise
	y1 := b.structure.Height - b.structure.BraceRise
	z := b.position

	b.structure.Member(x, y0, z, x, y1, z, b.postSection)
	b.structure.Member(-x, y0, z, -x, y1, z, b.postSection)
}
func (b *SimpleBent) postPlateBraceToTie() {
	x := -b.structure.Width / 2
	y0 := b.structure.Height - b.structure.BraceRise
	y1 := b.structure.Height - b.structure.TieOffset
	z := b.position

	b.structure.Member(x, y0, z, x, y1, z, b.postSection)
	b.structure.Member(-x, y0, z, -x, y1, z, b.postSection)
}
func (b *SimpleBent) postTieToPlate() {
	x := -b.structure.Width / 2
	y0 := b.structure.Height - b.structure.BraceRise
	y1 := b.structure.Height - b.structure.TieOffset
	z := b.position

	b.structure.Member(x, y0, z, x, y1, z, b.postSection)
	b.structure.Member(-x, y0, z, -x, y1, z, b.postSection)
}
func (b *SimpleBent) rafter() {
	x0 := -b.structure.Width / 2
	x1 := 0.
	y0 := b.structure.Height
	y1 := b.structure.Height + b.structure.Width/2*b.structure.Roof.Rise/b.structure.Roof.Run
	z := b.position

	b.structure.Member(x0, y0, z, x1, y1, z, b.rafterSection)
	b.structure.Member(-x0, y0, z, -x1, y1, z, b.rafterSection)
}
func (b *SimpleBent) tiePostToBrace() {
	x0 := -b.structure.Width / 2
	x1 := -b.structure.Width/2 + b.structure.BraceRise
	y := b.structure.Height - b.structure.TieOffset
	z := b.position

	b.structure.Member(x0, y, z, x1, y, z, b.tieSection)
	b.structure.Member(-x0, y, z, -x1, y, z, b.tieSection)
}
func (b *SimpleBent) tieBraceToBrace() {
	x := -b.structure.Width/2 + b.structure.BraceRise
	y := b.structure.Height - b.structure.TieOffset
	z := b.position

	b.structure.Member(x, y, z, -x, y, z, b.tieSection)
}
func (b *SimpleBent) tieBrace() {
	x0 := -b.structure.Width / 2
	y0 := b.structure.Height - b.structure.TieOffset - b.structure.BraceRise
	x1 := -b.structure.Width/2 + b.structure.BraceRise
	y1 := b.structure.Height - b.structure.TieOffset
	z := b.position

	b.structure.Member(x0, y0, z, x1, y1, z, b.braceSection)
	b.structure.Member(-x0, y0, z, -x1, y1, z, b.braceSection)
}

func (m *SimplePostBeam) Model() (r *model.Skyciv) {
	r = model.NewModel()

	for _, m := range m.materials {
		r.Materials[m.Id] = m
	}
	for _, s := range m.sections.set {
		r.Sections[s.Id] = s
	}
	for _, n := range m.nodes.nodes {
		r.Nodes[n.id] = &model.Node{X: n.x, Y: n.y, Z: n.z}
	}
	for _, m := range m.members.set {
		r.Members[m.Id] = m
	}

	return
}

func (b *SimplePostBeam) getMaterial(name string) *model.Material {
	m, ok := b.materials[name]
	if !ok {
		return nil
	}
	return m
}

func (b *SimplePostBeam) Material(name string) (m *model.Material, empty bool) {
	m, ok := b.materials[name]
	if !ok {
		m = &model.Material{
			Name: name,
			Id:   len(b.materials) + 1,
		}
		b.materials[name] = m
		return m, true
	}
	return m, false

}

func (b *SimplePostBeam) Section(breadth, depth float64, material string) (int, error) {
	m := b.getMaterial(material)
	if m == nil {
		return 0, fmt.Errorf("material '%s' not found", material)
	}
	s := b.sections.rectangular(breadth, depth, m.Id)
	return s.Id, nil
}

func (m *SimplePostBeam) Node(x, y, z float64) int {
	return m.nodes.node(x, y, z)
}

func (b *SimplePostBeam) Member(x0, y0, z0, x1, y1, z1 float64, sectionId int) int {
	n0 := b.Node(x0, y0, z0)
	n1 := b.Node(x1, y1, z1)

	m := b.members.member(n0, n1, sectionId)
	return m.Id
}
