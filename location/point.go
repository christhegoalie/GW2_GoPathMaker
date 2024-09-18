package location

import (
	"errors"
	"gw2_markers_gen/utils"
	"log"
	"math"
	"strconv"
)

type Path []Point
type PointList []Point

var GLOBAL_Barriers map[string]TypedGroup
var GLOBAL_Paths map[string]TypedGroup

type Point struct {
	X, Y, Z        float64
	AllowDuplicate bool
}

func GetPositionGeneric(m map[string]any) (float64, float64, float64, error) {
	var xst, yst, zst string
	var x, y, z float64
	var ok bool
	var err error

	if xst, ok = utils.MapString(m, "xpos"); !ok {
		return x, y, z, errors.New("xpos not defined")
	}
	if yst, ok = utils.MapString(m, "ypos"); !ok {
		return x, y, z, errors.New("ypos not defined")
	}
	if zst, ok = utils.MapString(m, "zpos"); !ok {
		return x, y, z, errors.New("zpos not defined")
	}
	xst = utils.Trim(xst)
	yst = utils.Trim(yst)
	zst = utils.Trim(zst)
	if x, err = strconv.ParseFloat(xst, 64); err != nil {
		return x, y, z, errors.New("invalid xpos")
	}
	if y, err = strconv.ParseFloat(yst, 64); err != nil {
		return x, y, z, errors.New("invalid ypos")
	}
	if z, err = strconv.ParseFloat(zst, 64); err != nil {
		return x, y, z, errors.New("invalid zpos")
	}
	return x, y, z, nil
}

func GetPosition(m map[string]any) (float64, float64, float64, error) {
	var xst, yst, zst string
	var x, y, z float64
	var ok bool
	var err error

	if xst, ok = utils.MapString(m, "xpos"); !ok {
		return x, y, z, errors.New("xpos not defined")
	}
	if yst, ok = utils.MapString(m, "ypos"); !ok {
		return x, y, z, errors.New("ypos not defined")
	}
	if zst, ok = utils.MapString(m, "zpos"); !ok {
		return x, y, z, errors.New("zpos not defined")
	}
	xst = utils.Trim(xst)
	yst = utils.Trim(yst)
	zst = utils.Trim(zst)
	if x, err = strconv.ParseFloat(xst, 64); err != nil {
		return x, y, z, errors.New("invalid xpos")
	}
	if y, err = strconv.ParseFloat(yst, 64); err != nil {
		return x, y, z, errors.New("invalid ypos")
	}
	if z, err = strconv.ParseFloat(zst, 64); err != nil {
		return x, y, z, errors.New("invalid zpos")
	}
	return x, y, z, nil
}

func (src Point) Same(point Point) bool {
	return distance(src, point) < 5
}

func distance(p1, p2 Point) float64 {
	d1, d2, d3 := p2.X-p1.X, p2.Y-p1.Y, p2.Z-p1.Z
	return math.Sqrt(d1*d1 + d2*d2 + d3*d3)
}

func (src Point) TakePath(path []TypedGroup) (float64, Point) {
	var out float64
	for _, p := range path {
		out += src.CalcDistance(p.First())
		out += p.Distance()
		src = p.Last()
	}
	return out, src
}

func (src Point) Barrier(dst Point) bool {
	for _, b := range GLOBAL_Barriers {
		if len(b._points) != 2 {
			log.Println("Unsupported barrier")
			return false
		}
		if b.Type == BT_DownOnly && dst.Y < src.Y {
			continue
		}
		if doIntersect(src, dst, b.First(), b.Last()) {
			return true
		}
	}
	return false
}

func (src Point) CalcDistance(dst Point) float64 {
	diffX := dst.X - src.X
	diffY := dst.Y - src.Y
	diffZ := dst.Z - src.Z

	diffXSq := diffX * diffX
	diffZSq := diffZ * diffZ
	planarDistance := math.Sqrt(diffXSq + diffZSq)
	climb := float64(1)
	if planarDistance > 0 {
		climb = diffY / planarDistance
	}

	// Punish our distance for steep climbs.
	// Steeper climb take far longer to traverse than regular distances
	if diffY > 0 {
		if climb > 4 { // larger than ~68 degree angle
			diffY *= 4
		}
		if climb >= 1 { //45 Degree angle
			diffY *= 2
		} else if climb >= 0.5 {
			diffY *= 1.2
		}
		//Steep drops are rewarded (griffon is too good here)
	} else if diffY < 0 {
		diffY /= 1.5
	}

	//calculate Y Squared after punishments
	diffYSq := diffY * diffY

	return math.Sqrt(diffXSq + diffYSq + diffZSq)
}
func (src Point) Distance(dst Point, allowWaypoints bool, bypassBarriers bool) float64 {
	var pathDistance float64
	if src.Barrier(dst) {
		if !bypassBarriers {
			return BarrierValue
		}
		if path, ok := src.FindPath(dst); !ok {
			return BarrierValue
		} else {
			pathDistance, src = src.TakePath(path)
		}
	}

	pathDistance = pathDistance + src.CalcDistance(dst)

	//check if waypointing is faster
	if allowWaypoints && enableWaypointing {
		for _, w := range GLOBAL_Waypoints {
			waypointDistance := waypointCost + w.Distance(dst, false, false)
			if waypointDistance < pathDistance {
				return waypointDistance
			}
		}
	}

	return pathDistance
}

func (src Point) FindPath(dst Point) ([]TypedGroup, bool) {
	return src.PathTo(dst, make([]TypedGroup, 0))
}

func (src Point) PathTo(dst Point, usedPaths []TypedGroup) ([]TypedGroup, bool) {
	if len(usedPaths) > 3 {
		return []TypedGroup{}, false
	}
	choices := []TypedGroup{}
	var addChoice = false
	start := src
	if len(usedPaths) > 0 {
		start = usedPaths[len(usedPaths)-1].Last()
	}
	possiblePaths := [][]TypedGroup{}
	for _, path := range availablePaths(usedPaths) {
		if !start.Barrier(path.First()) {
			addChoice = true
			if !path.Last().Barrier(dst) {
				possiblePaths = append(possiblePaths, append(usedPaths, path))
			}
		}
		if !path.IsOneway() && !start.Barrier(path.Last()) {
			addChoice = true
			if !path.First().Barrier(dst) {
				possiblePaths = append(possiblePaths, append(usedPaths, path.Reverse()))
			}
		}

		if addChoice {
			choices = append(choices, path)
		}
	}

	if len(possiblePaths) == 0 {
		for _, choice := range choices {
			newPaths, ok := src.PathTo(dst, append(usedPaths, choice))
			if ok {
				possiblePaths = append(possiblePaths, newPaths)
			}
		}
	}
	if len(possiblePaths) == 0 {
		return []TypedGroup{}, false
	}
	return cheapest(src, dst, possiblePaths), true
}
func availablePaths(usedList []TypedGroup) []TypedGroup {
	out := make([]TypedGroup, 0)
	for _, global := range GLOBAL_Paths {
		found := false
		for _, local := range usedList {
			if local.Equals(global) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, global)
		}
	}
	return out
}

func cheapest(src, dst Point, groups [][]TypedGroup) []TypedGroup {
	min := math.MaxFloat64
	index := 0
	for i, next := range groups {
		cost := calculatePathCost(src, dst, next)
		if cost < min {
			index = i
			min = cost
		}
	}
	return groups[index]
}

func calculatePathCost(src, dst Point, group []TypedGroup) float64 {
	total, newSrc := src.TakePath(group)
	return total + newSrc.Distance(dst, false, false)
}

func (ls PointList) Contains(point Point) bool {
	for _, p := range ls {
		if p.Same(point) {
			return true
		}
	}
	return false
}
