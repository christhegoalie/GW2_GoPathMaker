package location

import (
	"gw2_markers_gen/utils"
	"strings"
)

type ObjectType int

// Arbitrarily high value, but we need to be able to compute the "better" of paths crossing multiple barriers
const BarrierValue = 1e7
const waypointCost = 5000
const mushroomCost = 10
const leylineScale = 0.4
const updraftScale = 0.2

const (
	Type_Unknown ObjectType = iota
	BT_Wall
	BT_DownOnly

	GT_Leyline
	GT_Mushroom
	GT_ONEWAY
	GT_Updraft
	GT_Waypoint
)

type TypedGroup struct {
	Name         string
	_points      PointList
	_distance    float64
	_revDistance float64
}

func (t TypedGroup) Points() []Point {
	return t._points
}
func (t *TypedGroup) IsOneway() bool {
	for _, p := range t._points {
		if p.Type.IsOneway() {
			return true
		}
	}
	return false
}
func (t ObjectType) IsBarrier() bool {
	return t == BT_DownOnly || t == BT_Wall
}
func (t ObjectType) IsOneway() bool {
	return t == GT_ONEWAY || t.IsMushroom() || t.IsLeyline() || t.IsUpdraft() || t.IsWaypoint()
}
func (t ObjectType) IsMushroom() bool {
	return t == GT_Mushroom
}
func (t ObjectType) IsLeyline() bool {
	return t == GT_Leyline
}
func (t ObjectType) IsUpdraft() bool {
	return t == GT_Updraft
}
func (t ObjectType) IsWaypoint() bool {
	return t == GT_Waypoint
}

func (t *TypedGroup) AddPoint(pt Point) {
	if len(t._points) > 0 {
		t._distance = t._distance + t._points[len(t._points)-1].CalcDistance(pt)
		t._revDistance = t._revDistance + pt.CalcDistance(t._points[len(t._points)-1])
	}
	t._points = append(t._points, pt)
}
func (t TypedGroup) Last() Point {
	return t._points[len(t._points)-1]
}
func (t TypedGroup) First() Point {
	return t._points[0]
}
func (t TypedGroup) Distance() float64 {
	return t._distance
}

func (src Path) First() Point {
	return src[0]
}
func (src Path) Last() Point {
	return src[len(src)-1]
}

func (t TypedGroup) Equals(t1 TypedGroup) bool {
	return t.Name == t1.Name
}

func NewEmptyGroup(name string, tp ObjectType) TypedGroup {
	return TypedGroup{
		Name:         name,
		_points:      []Point{},
		_distance:    0,
		_revDistance: 0,
	}
}
func NewGroup(name string, point Point) TypedGroup {
	return TypedGroup{
		Name:         name,
		_points:      []Point{point},
		_distance:    0,
		_revDistance: 0,
	}
}

func TypeFromMap(vals map[string]any) ObjectType {
	if typeString, ok := utils.MapString(vals, "type"); ok {
		switch strings.ToLower(typeString) {
		case "downonly":
			return BT_DownOnly
		case "wall":
			return BT_Wall
		case "mushroom":
			return GT_Mushroom
		case "oneway":
			return GT_ONEWAY
		case "leyline":
			return GT_Leyline
		case "updraft":
			return GT_Updraft
		case "waypoint":
			return GT_Waypoint
		}
	}
	return Type_Unknown
}
