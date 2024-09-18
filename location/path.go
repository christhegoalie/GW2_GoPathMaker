package location

type objectType int

// Arbitrarily high value, but we need to be able to compute the "better" of paths crossing multiple barriers
const BarrierValue = 1e7
const waypointCost = 1000
const enableWaypointing = false
const mushroomCost = 10

const (
	Type_Unknown objectType = iota
	BT_Wall
	BT_DownOnly

	GT_Mushroom
	GT_ONEWAY
)

type TypedGroup struct {
	Name         string
	ReverseName  string
	_points      PointList
	Type         objectType
	_distance    float64
	_revDistance float64
}

func (t TypedGroup) Points() []Point {
	return t._points
}
func (t TypedGroup) IsOneway() bool {
	return t.Type == GT_ONEWAY || t.IsMushroom()
}
func (t TypedGroup) IsMushroom() bool {
	return t.Type == GT_Mushroom
}

func (t *TypedGroup) AddPoint(pt Point) {
	t._distance = t._distance + t._points[len(t._points)-1].CalcDistance(pt)
	t._revDistance = t._revDistance + pt.CalcDistance(t._points[len(t._points)-1])
	t._points = append(t._points, pt)
}
func (t TypedGroup) Last() Point {
	return t._points[len(t._points)-1]
}
func (t TypedGroup) First() Point {
	return t._points[0]
}
func (t TypedGroup) Distance() float64 {
	if t.IsMushroom() {
		return mushroomCost
	}
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

func NewGroup(name string, point Point, tp objectType) TypedGroup {
	return TypedGroup{
		Name:         name,
		ReverseName:  "Reversed " + name,
		_points:      []Point{point},
		Type:         Type_Unknown,
		_distance:    0,
		_revDistance: 0,
	}
}
