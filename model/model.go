package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

type Auth struct {
	UserName  string `json:"username"`
	Key       string `json:"key"`
	SessionId string `json:"session_id"`
}

type Options struct {
	ValidateInput    *bool `json:"validate_input"`
	ResponseDataOnly *bool `json:"response_data_only"`
}

type Function struct {
	Function  string                 `json:"function"`
	Arguments map[string]interface{} `json:"arguments"`
}

type Request struct {
	Auth      Auth       `json:"auth"`
	Options   Options    `json:"options"`
	Functions []Function `json:"functions"`
}

type Units struct {
	Length           string `json:"length"`
	SectionLength    string `json:"section_length"`
	MaterialStrength string `json:"material_strength"`
	Density          string `json:"density"`
	Force            string `json:"force"`
	Moment           string `json:"moment"`
	Pressure         string `json:"pressure"`
	Mass             string `json:"mass"`
	Translation      string `json:"translation"`
	Stress           string `json:"stress"`
	Name             string `json:"-"`
}

type Settings struct {
	Units                                interface{} `json:"units,omitempty"`
	Precision                            string      `json:"precision,omitempty"`
	PrecisionValues                      int         `json:"precision_values,omitempty"`
	EvaluationPoints                     int         `json:"evaluation_points,omitempty"`
	VerticalAxis                         string      `json:"vertical_axis,omitempty"`
	MemberOffsetsAxis                    string      `json:"member_offsets_axis,omitempty"`
	ProjectionSystem                     string      `json:"projection_system,omitempty"`
	SolverTimeout                        int         `json:"solver_timeout,omitempty"`
	AccurateBucklingShape                *bool       `json:"accurate_buckling_shape,omitempty"`
	BucklingJohnson                      *bool       `json:"buckling_johnson,omitempty"`
	NonLinearTolerance                   string      `json:"non_linear_tolerance,omitempty"`
	NonLinearTheory                      string      `json:"small,omitempty"`
	AutoStabilizeModel                   *bool       `json:"auto_stabilize_model,omitempty"`
	OnlySolveUserDefinedLoadCombinations *bool       `json:"only_solve_user_defined_load_combinations,omitempty"`
	IncludeRigidLinksForRealAreaLoads    *bool       `json:"include_rigid_links_for_area_loads,omitempty"`
}

type Details struct{}

type Vector struct {
	X, Y, Z float64
}

func (v Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector) Dot(w Vector) float64 {
	return v.X*w.X + v.Y*w.Y + v.Z*w.Z
}

func (A Vector) Diff(B Vector) (v Vector) {
	v.X = A.X - B.X
	v.Y = A.Y - B.Y
	v.Z = A.Z - B.Z
	return
}

type Node struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Z       float64 `json:"z"`
	Id      int     `json:"-"`
	model   *Skyciv
	support *Support
}

func (n *Node) ToVector() (v Vector) {
	v.X = n.X
	v.Y = n.Y
	v.Z = n.Z
	return
}

func (n *Node) FixedSupport() {
	if n.support != nil {
		return
	}

	s := &Support{
		DirectionCode: "BBBBBB",
		Node:          n.Id,
		RestraintCode: "FFFFFF",
		Id:            len(n.model.Supports) + 1,
	}
	n.support = s
	n.model.Supports[s.Id] = s
}

type Member struct {
	Type          string  `json:"type"`
	CableLength   *int    `json:"cable_length"`
	NodeA         int     `json:"node_A"`
	NodeB         int     `json:"node_B"`
	SectionId     int     `json:"section_id"`
	RotationAngle float64 `json:"rotation_angle"`
	FixityA       string  `json:"fixity_A"`
	FixityB       string  `json:"fixity_B"`
	OffsetAx      float64 `json:"offset_Ax,string"`
	OffsetAy      float64 `json:"offset_Ay,string"`
	OffsetAz      float64 `json:"offset_Az,string"`
	OffsetBx      float64 `json:"offset_Bx,string"`
	OffsetBy      float64 `json:"offset_By,string"`
	OffsetBz      float64 `json:"offset_Bz,string"`
	Id            int     `json:"-"`
	model         *Skyciv
}

func Distance(A, B *Node) float64 {
	return math.Sqrt((A.X-B.X)*(A.X-B.X) + (A.Y-B.Y)*(A.Y-B.Y) + (A.Z-B.Z)*(A.Z-B.Z))
}

func (m *Member) distanceTo(x, y, z float64) (t float64, d float64) {
	A := m.A().ToVector()
	B := m.B().ToVector()
	C := Vector{x, y, z}

	ab := B.Diff(A).Length()
	ac := C.Diff(A).Length()
	// bc := C.Diff(A).Length()
	CA := C.Diff(A)
	BA := B.Diff(A)

	t = CA.Dot(BA) / ab / ab
	d = (ac*ac*ab*ab - CA.Dot(BA)*CA.Dot(BA)) / ab / ab
	return
}

func (m *Member) A() *Node {
	return m.model.Nodes[m.NodeA]
}
func (m *Member) B() *Node {
	return m.model.Nodes[m.NodeB]
}

var (
	ErrColocated = errors.New("Colocated with Node")
)

// splits the member as close as it can to the specified point
func (m *Member) SplitAt(x, y, z float64) (*Node, *Member, error) {
	A := m.A()
	B := m.B()
	t, _ := m.distanceTo(x, y, z)

	if t < 0 || t > 1 {
		return nil, nil, fmt.Errorf("cannot split a node outside of it's length")
	}

	// adjust to the real splitting spot
	x = A.X + (B.X-A.X)*t
	y = A.Y + (B.Y-A.Y)*t
	z = A.Z + (B.Z-A.Z)*t

	if A.Colocated(x, y, z) {
		return A, nil, ErrColocated
	} else if B.Colocated(x, y, z) {
		return B, nil, ErrColocated
	}

	C := m.model.NewNode(x, y, z)
	m2 := &Member{}
	*m2 = *m // duplicate m
	m.NodeB = C.Id
	m2.NodeA = C.Id
	m2.Id = len(m.model.Members) + 1
	// TODO: what to do about offsets??
	m.model.Members[m2.Id] = m2

	return C, m2, nil
}

// negative numbers start from B
func (m *Member) Split(distance float64) (*Node, *Member, error) {
	A := m.A()
	B := m.B()

	length := Distance(A, B)
	if distance < -length || distance > length {
		return nil, nil, fmt.Errorf("distance %f is less than 0 or greater than member length %f", distance, length)
	}

	distance /= length // as a percentage
	var x, y, z float64

	if distance >= 0 {
		x = (B.X-A.X)*distance + A.X
		y = (B.Y-A.Y)*distance + A.Y
		z = (B.Z-A.Z)*distance + A.Z
	} else {
		x = (B.X-A.X)*distance + B.X
		y = (B.Y-A.Y)*distance + B.Y
		z = (B.Z-A.Z)*distance + B.Z
	}

	if A.Colocated(x, y, z) || B.Colocated(x, y, z) {
		return nil, nil, fmt.Errorf("distance %f is too close to zero or length %f", distance*length, length)
	}

	C := m.model.NewNode(x, y, z)
	m2 := &Member{}
	*m2 = *m // duplicate m
	if distance >= 0 {
		m.NodeB = C.Id
		m2.NodeA = C.Id
	} else {
		m.NodeA = C.Id
		m2.NodeB = C.Id
	}
	m2.Id = len(m.model.Members) + 1
	// TODO: what to do about offsets??
	m.model.Members[m2.Id] = m2

	return C, m2, nil
}

type Plate struct{}
type MeshedPlate struct{}

type PointsCalc struct {
	X    int
	Y    int
	Type string
}

func (c *PointsCalc) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{c.X, c.Y, c.Type})
}

type Coord struct {
	X float64
	Y float64
}

func (c *Coord) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{c.X, c.Y})
}

type Locat struct {
	Coords      []Coord
	Placeholder string
	DimensionId string
	Dimension   float64
}

func (l *Locat) MarshalJSON() ([]byte, error) {
	arr := make([]interface{}, len(l.Coords)+1)
	for i := 0; i < len(l.Coords); i++ {
		arr[i] = l.Coords[i]
	}
	arr[len(l.Coords)] = map[string]interface{}{
		"placeholder":  l.Placeholder,
		"dimension_id": l.DimensionId,
		"dimension":    l.Dimension,
	}

	return json.Marshal(arr)
}

type Dimension struct {
	Value float64 `json:"value"`
	Locat Locat   `json:"locat"`
}

type Dimensions struct {
	H Dimension `json:"h"`
	B Dimension `json:"b"`
}

type Operations struct {
	Rotation    float64 `json:"rotation"`
	Translation Coord   `json:"translation"`
	MirrorZ     bool    `json:"mirror_z"`
	MirrorY     bool    `json:"mirror_y"`
}

type Material struct {
	Name               string  `json:"name"`
	ElasticityModulus  float64 `json:"elasticity_modulus"`
	Density            float64 `json:"density"`
	PoissonsRatio      float64 `json:"poissons_ratio"`
	YieldStrength      float64 `json:"yield_strength,omitempty"`
	UltimateStrength   float64 `json:"ultimate_strength"`
	Class              string  `json:"class"`
	ElasticityModulusX float64 `json:"elasticity_modulus_x,omitempty"`
	ElasticityModulusY float64 `json:"elasticity_modulus_y,omitempty"`
	ShearModulusXY     float64 `json:"shear_modulus_xy,omitempty"`
	ShearModulusXZ     float64 `json:"shear_modulus_xz,omitempty"`
	ShearModulusYZ     float64 `json:"shear_modulus_yz,omitempty"`
	Id                 int     `json:"id"`
	model              *Skyciv
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

func (m *Material) NewSectionFromLibrary(path ...string) *Section {
	s := &Section{
		MaterialId:  m.Id,
		Id:          len(m.model.Sections) + 1,
		model:       m.model,
		LoadSection: path,
	}
	m.model.Sections[s.Id] = s
	return s
}

func (m *Material) NewRectangularSection(breadth, depth float64) *Section {
	s := &Section{
		MaterialId: m.Id,
		Id:         len(m.model.Sections) + 1,
		model:      m.model,
		Version:    SectionVersion,
		Name:       fmt.Sprintf("%s %fx%f", m.Name, breadth, depth),
		Area:       breadth * depth,
		Iz:         breadth * depth * depth * depth / 12.,
		Iy:         depth * breadth * breadth * breadth / 12.,
		J:          torsionConstant(breadth, depth),
		Aux: &SectionAux{
			Composite:      false,
			CentroidPoint:  []float64{breadth / 2, depth / 2},
			CentroidLength: []float64{breadth / 2, depth / 2},
			Depth:          depth,
			Width:          breadth,
			Alpha:          0,
		},
	}
	m.model.Sections[s.Id] = s
	return s
}

type Polygon struct {
	Name                  string        `json:"string"`
	GroupId               int           `json:"group_id"`
	PointsCalc            []PointsCalc  `json:"points_calc"`
	PointsCustomOrig      []interface{} `json:"points_custom_orig"`
	Shape                 string        `json:"shape"`
	DimensionsShow        bool          `json:"dimensions_show"`
	Dimensions            Dimensions    `json:"dimensions"`
	Operations            Operations    `json:"operations"`
	Cutout                bool          `json:"cutout"`
	Material              Material      `json:"material"`
	Type                  string        `json:"type"`
	PointsCentroidShifted []PointsCalc  `json:"points_centroid_shifted"`
	sectionAux            *SectionAux
}

type SectionAux struct {
	Composite       bool `json:"composite"`
	Qz              float64
	Qy              float64
	CentroidPoint   []float64 `json:"centroid_point"`
	CentroidLength  []float64 `json:"centroid_length"`
	Depth           float64   `json:"depth"`
	Width           float64   `json:"width"`
	Alpha           float64   `json:"alpha"`
	Zy              float64
	Zz              float64
	Polygons        []Polygon `json:"polygons"`
	WarpingConstant float64   `json:"warping_constant"`
	ShearAreaZ      float64   `json:"shear_area_z"`
	ShearAreaY      float64   `json:"shear_area_y"`
	TorsionRadius   float64   `json:"torsion_radius"`
	NonPrismatic    *int      `json:"non_prismatic"`
	section         *Section
}

type Section struct {
	Version     int         `json:"version,omitempty"`
	Name        string      `json:"name,omitempty"`
	Area        float64     `json:"area,omitempty"`
	Iz          float64     `json:"Iz,omitempty"`
	Iy          float64     `json:"Iy,omitempty"`
	MaterialId  int         `json:"material_id"`
	Aux         *SectionAux `json:"aux,omitempty"`
	J           float64     `json:"J,omitempty"`
	Id          int         `json:"-"`
	LoadSection []string    `json:"load_section,omitempty"`
	model       *Skyciv
}

func (s *Section) NewContinuousMember(x0, y0, z0, x1, y1, z1 float64) *Member {
	n0 := s.model.NewNode(x0, y0, z0)
	n1 := s.model.NewNode(x1, y1, z1)
	return s.NewContinuousMemberBetweenNodes(n0, n1)
}

func (s *Section) NewContinuousMemberBetweenNodes(n0, n1 *Node) *Member {
	m := &Member{
		Type:  "normal_continuous",
		NodeA: n0.Id, NodeB: n1.Id,
		FixityA: "FFFFFF", FixityB: "FFFFFF",
		SectionId: s.Id,
		Id:        len(s.model.Members) + 1,
		model:     s.model,
	}
	s.model.Members[m.Id] = m
	return m
}

type Support struct {
	DirectionCode string  `json:"direction_code"`
	Tx            float64 `json:"tx"`
	Ty            float64 `json:"ty"`
	Tz            float64 `json:"tz"`
	Rx            float64 `json:"rx"`
	Ry            float64 `json:"ry"`
	Rz            float64 `json:"rz"`
	Node          int     `json:"node"`
	RestraintCode string  `json:"restraint_code"`
	model         *Skyciv
	Id            int `json:"-"`
}

type Group struct{}

type StringIntList []int

func (s StringIntList) MarshalJSON() ([]byte, error) {
	if len(s) == 0 {
		return []byte(""), nil
	}
	builder := &strings.Builder{}

	for i := 0; i < len(s); i++ {
		fmt.Fprintf(builder, "%d", s[i])
		if i < len(s)-1 {
			builder.WriteString(",")
		}
	}
	return json.Marshal(builder.String())
}

type AreaLoad struct {
	Type             string        `json:"type"`
	Nodes            StringIntList `json:"nodes"`
	Members          int           `json:"members"`
	Mag              float64       `json:"mag"`
	Direction        string        `json:"direction"`
	Elevations       int           `json:"elevations"`
	Mags             int           `json:"mags"`
	ColumnDirection  StringIntList `json:"column_direction"`
	LoadedMemberAxis string        `json:"loaded_member_axis"`
	LoadGroup        string        `json:"LG"`
	Id               int           `json:"-"`
}

type SelfWeight struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Z         float64 `json:"z"`
	LoadGroup string  `json:"LG"`
	Id        int     `json:"-"`
}

type Suppression struct {
	Members          []string `json:"members"`
	Plates           []string `json:"plates"`
	Supports         []string `json:"supports"`
	Moments          []string `json:"moments"`
	DistributedLoads []string `json:"distributed_loads"`
	PointLoads       []string `json:"point_loads"`
	AreaLoads        []string `json:"area_loads"`
	Pressures        []string `json:"pressures"`
	LoadCombinations []string `json:"load_combinations"`
}

func emptySuppression() Suppression {
	return Suppression{
		Members:          []string{},
		Plates:           []string{},
		Supports:         []string{},
		Moments:          []string{},
		DistributedLoads: []string{},
		PointLoads:       []string{},
		AreaLoads:        []string{},
		Pressures:        []string{},
		LoadCombinations: []string{},
	}
}

type Suppress struct {
	Suppressions map[string]Suppression
	CurrentCase  string
}

func (s *Suppress) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	for k, v := range s.Suppressions {
		m[k] = v
	}
	m["current_case"] = s.CurrentCase
	return json.Marshal(m)
}

type Skyciv struct {
	DataVersion          int                  `json:"dataVersion"`
	Settings             Settings             `json:"settings"`
	Details              []Details            `json:"details"`
	Nodes                map[int]*Node        `json:"nodes"`
	Members              map[int]*Member      `json:"members"`
	Plates               map[int]*Plate       `json:"plates"`
	MeshedPlates         map[int]*MeshedPlate `json:"meshed_plates"`
	Sections             map[int]*Section     `json:"sections"`
	Materials            map[int]*Material    `json:"materials"`
	Supports             map[int]*Support     `json:"supports"`
	Settlements          map[int]interface{}  `json:"settlements"`
	Groups               []*Group             `json:"groups"`
	PointLoads           map[int]interface{}  `json:"point_loads"`
	Moments              map[int]interface{}  `json:"moments"`
	DistributedLoads     map[int]interface{}  `json:"distributed_loads"`
	Pressures            map[int]interface{}  `json:"pressures"`
	AreaLoads            map[int]*AreaLoad    `json:"area_loads"`
	MemberPrestressLoads map[int]interface{}  `json:"member_prestress_loads"`
	SelfWeight           map[int]*SelfWeight  `json:"self_weight"`
	LoadCombinations     Combination          `json:"load_combinations"`
	// LoadCases                map[int]interface{}  `json:"load_cases"`
	NodalMasses              map[int]interface{} `json:"nodal_masses"`
	NodalMassesConversionMap map[int]interface{} `json:"nodal_masses_conversion_map"`
	SpectralLoads            map[int]interface{} `json:"spectral_loads"`
	NotionalLoads            map[int]interface{} `json:"notional_loads"`
	Suppress                 Suppress            `json:"suppress"`
}

const (
	DataVersion       = 30
	SectionVersion    = 4
	MaterialClassWood = "wood"
)

func NewModel() *Skyciv {
	return &Skyciv{
		DataVersion: DataVersion,
		Settings: Settings{
			Units: &Units{
				Length:           "ft",
				SectionLength:    "in",
				MaterialStrength: "ksi",
				Density:          "lb/ft3",
				Force:            "kip",
				Moment:           "kip-ft",
				Pressure:         "ksf",
				Mass:             "kip",
				Translation:      "in",
				Stress:           "ksi",
			},
		},
		Details:              []Details{},
		Nodes:                make(map[int]*Node),
		Members:              make(map[int]*Member),
		Plates:               make(map[int]*Plate),
		MeshedPlates:         make(map[int]*MeshedPlate),
		Sections:             make(map[int]*Section),
		Materials:            make(map[int]*Material),
		Supports:             make(map[int]*Support),
		Settlements:          make(map[int]interface{}),
		AreaLoads:            make(map[int]*AreaLoad),
		SelfWeight:           make(map[int]*SelfWeight),
		Groups:               []*Group{nil, nil},
		PointLoads:           make(map[int]interface{}),
		Moments:              make(map[int]interface{}),
		DistributedLoads:     make(map[int]interface{}),
		Pressures:            make(map[int]interface{}),
		MemberPrestressLoads: make(map[int]interface{}),
		// LoadCombinations:         make(map[int]interface{}),
		// LoadCases:                make(map[int]interface{}),
		NodalMasses:              make(map[int]interface{}),
		NodalMassesConversionMap: make(map[int]interface{}),
		SpectralLoads:            make(map[int]interface{}),
		NotionalLoads:            make(map[int]interface{}),
		Suppress: Suppress{
			CurrentCase: "User Defined",
			Suppressions: map[string]Suppression{
				"All On":       emptySuppression(),
				"User Defined": emptySuppression(),
			},
		},
	}
}

type NodeList []*Node

func (nl NodeList) NodeIds() (ret []int) {
	for _, n := range nl {
		ret = append(ret, n.Id)
	}
	return
}

func (m *Skyciv) NewAreaLoad(nodes []*Node) (*AreaLoad, error) {
	if len(nodes) < 3 {
		return nil, fmt.Errorf("area loads must have at least 3 nodes")
	}
	nl := NodeList(nodes).NodeIds()
	al := &AreaLoad{
		Nodes:            nl,
		ColumnDirection:  nl[0:2],
		LoadedMemberAxis: "all",
		Type:             "one_way",
		Id:               len(m.AreaLoads) + 1,
	}
	m.AreaLoads[al.Id] = al

	return al, nil
}

func (m *Skyciv) NewSelfWeight() *SelfWeight {
	sw := &SelfWeight{
		Id: len(m.SelfWeight) + 1,
	}
	switch m.Settings.VerticalAxis {
	case "X":
		sw.X = -1
	case "Y":
		sw.Y = -1
	case "Z":
		sw.Z = -1
	default:
		sw.Y = -1
	}
	sw.LoadGroup = "SW1" // skyciv always assumes self weight is SW1 I think
	m.SelfWeight[sw.Id] = sw
	return sw
}

func (m *Skyciv) GetMaterial(name string) *Material {
	for _, m := range m.Materials {
		if m.Name == name {
			return m
		}
	}
	return nil
}

func (m *Skyciv) NewMaterial(name string) *Material {
	mat := &Material{
		model: m,
		Id:    len(m.Materials) + 1,
		Name:  name,
	}
	m.Materials[mat.Id] = mat
	return mat
}

func (n *Node) Colocated(x, y, z float64) bool {
	// TODO: precision of 3 assumed
	const eps2 = 0.001 * 0.001

	isClose := func(a, b float64) bool {
		d := b - a
		return d*d < eps2
	}

	return isClose(n.X, x) && isClose(n.Y, y) && isClose(n.Z, z)
}

func (m *Skyciv) FindNearestNode(x, y, z float64) (minNode *Node) {
	minDistance := math.MaxFloat64
	for _, node := range m.Nodes {
		dist := node.ToVector().Diff(Vector{x, y, z}).Length()
		if dist < minDistance {
			minDistance = dist
			minNode = node
		}
	}
	return
}

func (m *Skyciv) FindNearestMemberAndSplitAt(x, y, z float64) (*Node, error) {
	// if a node is colocated just return it
	if n := m.FindNearestNode(x, y, z); n != nil && n.Colocated(x, y, z) {
		return n, nil
	}

	// otherwise find the nearest member
	mem := m.FindNearestMember(x, y, z)
	if mem == nil {
		return nil, fmt.Errorf("could not find a nearest member to %f,%f,%f", x, y, z)
	}
	n, _, err := mem.SplitAt(x, y, z)
	if err == ErrColocated {
		return n, nil
	}
	return n, err
}

func (m *Skyciv) FindNearestMember(x, y, z float64) *Member {
	minDistance := math.MaxFloat64
	var minMember *Member
	for _, mem := range m.Members {
		t, d := mem.distanceTo(x, y, z)

		// fmt.Fprintf(os.Stderr, "t, d: %f, %f\n", t, d)

		if t < 0 || t > 1 {
			continue
		}
		if d < minDistance {
			minDistance = d
			minMember = mem
		}
	}
	// fmt.Fprintf(os.Stderr, "min: %f %d\n", minDistance, minMember.Id)
	return minMember
}

func (m *Skyciv) NewNode(x, y, z float64) *Node {
	var found *Node
	for _, n := range m.Nodes {
		if n.Colocated(x, y, z) {
			found = n
			break
		}
	}

	if found != nil {
		return found
	}

	found = &Node{X: x, Y: y, Z: z, Id: len(m.Nodes) + 1, model: m}
	m.Nodes[found.Id] = found
	return found

}
