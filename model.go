package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

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
}

type Settings struct {
	Units                                Units  `json:"units"`
	Precision                            string `json:"precision"`
	PrecisionValues                      int    `json:"precision_values"`
	EvaluationPoints                     int    `json:"evaluation_points"`
	VerticalAxis                         string `json:"vertical_axis"`
	MemberOffsetsAxis                    string `json:"member_offsets_axis"`
	ProjectionSystem                     string `json:"projection_system"`
	SolverTimeout                        int    `json:"solver_timeout"`
	AccurateBucklingShape                bool   `json:"accurate_buckling_shape"`
	BucklingJohnson                      bool   `json:"buckling_johnson"`
	NonLinearTolerance                   string `json:"non_linear_tolerance"`
	NonLinearTheory                      string `json:"small"`
	AutoStabilizeModel                   bool   `json:"auto_stabilize_model"`
	OnlySolveUserDefinedLoadCombinations bool   `json:"only_solve_user_defined_load_combinations"`
	IncludeRigidLinksForRealAreaLoads    bool   `json:"include_rigid_links_for_area_loads"`
}

type Details struct{}

type Node struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type Member struct {
	Type          string  `json:"type"`
	CableLength   *int    `json:"cable_length"`
	NodeA         int     `json:"node_A"`
	NodeB         int     `json:"node_B"`
	SectionId     int     `json:"section_id"`
	RotationAngle int     `json:"rotation_angle"`
	FixityA       string  `json:"fixity_A"`
	FixityB       string  `json:"fixity_B"`
	OffsetAx      float32 `json:"offset_Ax,string"`
	OffsetAy      float32 `json:"offset_Ay,string"`
	OffsetAz      float32 `json:"offset_Az,string"`
	OffsetBx      float32 `json:"offset_Bx,string"`
	OffsetBy      float32 `json:"offset_By,string"`
	OffsetBz      float32 `json:"offset_Bz,string"`
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
	X float32
	Y float32
}

func (c *Coord) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float32{c.X, c.Y})
}

type Locat struct {
	Coords      []Coord
	Placeholder string
	DimensionId string
	Dimension   float32
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
	Value float32 `json:"value"`
	Locat Locat   `json:"locat"`
}

type Dimensions struct {
	H Dimension `json:"h"`
	B Dimension `json:"b"`
}

type Operations struct {
	Rotation    float32 `json:"rotation"`
	Translation Coord   `json:"translation"`
	MirrorZ     bool    `json:"mirror_z"`
	MirrorY     bool    `json:"mirror_y"`
}

type Material struct {
	Name               string   `json:"name"`
	ElasticityModulus  float32  `json:"elasticity_modulus"`
	Density            float32  `json:"density"`
	PoissonsRatio      float32  `json:"poissons_ratio"`
	YieldStrength      *float32 `json:"yield_strength"`
	UltimateStrength   float32  `json:"ultimate_strength"`
	Class              string   `json:"class"`
	ElasticityModulusX *float32 `json:"elasticity_modulus_x"`
	ElasticityModulusY *float32 `json:"elasticity_modulus_y"`
	ShearModulusXY     *float32 `json:"shear_modulus_xy"`
	ShearModulusXZ     *float32 `json:"shear_modulus_xz"`
	ShearModulusYZ     *float32 `json:"shear_modulus_yz"`
	Id                 int      `json:"id"`
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
}

type SectionAux struct {
	Composite       bool `json:"composite"`
	Qz              float32
	Qy              float32
	CentroidPoint   []float32 `json:"centroid_point"`
	CentroidLength  []float32 `json:"centroid_length"`
	Depth           float32   `json:"depth"`
	Width           float32   `json:"width"`
	Alpha           float32   `json:"alpha"`
	Zy              float32
	Zz              float32
	Polygons        []Polygon `json:"polygons"`
	WarpingConstant float32   `json:"warping_constant"`
	ShearAreaZ      float32   `json:"shear_area_z"`
	ShearAreaY      float32   `json:"shear_area_y"`
	TorsionRadius   float32   `json:"torsion_radius"`
	NonPrismatic    *int      `json:"non_prismatic"`
}

type Section struct {
	Version    int     `json:"version"`
	Name       string  `json:"name"`
	Area       float32 `json:"area"`
	Iz         float32
	Iy         float32
	MaterialId int        `json:"material_id"`
	Aux        SectionAux `json:"aux"`
	J          float32    `json:"J"`
}

type Support struct {
	DirectionCode string  `json:"direction_code"`
	Tx            float32 `json:"tx"`
	Ty            float32 `json:"ty"`
	Tz            float32 `json:"tz"`
	Rx            float32 `json:"rx"`
	Ry            float32 `json:"ry"`
	Rz            float32 `json:"rz"`
	Node          int     `json:"node"`
	RestraintCode string  `json:"restraint_code"`
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
	return []byte(builder.String()), nil
}

type AreaLoad struct {
	Type             string        `json:"type"`
	Nodes            StringIntList `json:"nodes"`
	Members          int           `json:"members"`
	Mag              float32       `json:"mag"`
	Direction        string        `json:"direction"`
	Elevations       int           `json:"elevations"`
	Mags             int           `json:"mags"`
	ColumnDirection  StringIntList `json:"column_direction"`
	LoadedMemberAxis string        `json:"loaded_member_axis"`
	LoadGroup        string        `json:"LG"`
}

type SelfWeight struct {
	X         float32 `json:"x"`
	Y         float32 `json:"y"`
	Z         float32 `json:"z"`
	LoadGroup string  `json:"LG"`
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

type Supress struct {
	Suppressions map[string]Suppression
	CurrentCase  string
}

func (s *Supress) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	for k, v := range s.Suppressions {
		m[k] = v
	}
	m["current_case"] = s.CurrentCase
	return json.Marshal(m)
}

type Skyciv struct {
	DataVersion              int                 `json:"dataVersion"`
	Settings                 Settings            `json:"settings"`
	Details                  []Details           `json:"details"`
	Nodes                    map[int]Node        `json:"nodes"`
	Members                  map[int]Member      `json:"members"`
	Plates                   map[int]Plate       `json:"plates"`
	MeshedPlates             map[int]MeshedPlate `json:"meshed_plates"`
	Sections                 map[int]Section     `json:"sections"`
	Materials                map[int]Material    `json:"materials"`
	Supports                 map[int]Support     `json:"supports"`
	Settlements              map[int]interface{} `json:"settlements"`
	Groups                   []*Group            `json:"groups"`
	PointLoads               []interface{}       `json:"point_loads"`
	Moments                  []interface{}       `json:"moments"`
	DistributedLoads         []interface{}       `json:"distributed_loads"`
	Pressures                []interface{}       `json:"pressures"`
	AreaLoads                map[int]AreaLoad    `json:"area_loads"`
	MemberPrestressLoads     map[int]interface{} `json:"member_prestress_laods"`
	SelfWeight               map[int]SelfWeight  `json:"self_weight"`
	LoadCombinations         map[int]interface{} `json:"load_combinations"`
	LoadCases                map[int]interface{} `json:"load_cases"`
	NodalMasses              map[int]interface{} `json:"nodal_masses"`
	NodalMassesConversionMap map[int]interface{} `json:"nodal_masses_conversion_map"`
	SpectralLoads            map[int]interface{} `json:"spectral_loads"`
	NotionalLoads            map[int]interface{} `json:"notional_loads"`
	Suppress                 Supress             `json:"suppress"`
}
