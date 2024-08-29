package trailbuilder

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"slices"
)

type path []point

var glBarriers map[string]typedGroup
var glPaths map[string]typedGroup

var glWaypoints path

// Arbitrarily high value, but we need to be able to compute the "better" of paths crossing multiple barriers
const BarrierValue = 1e7
const waypointCost = 1000
const enableWaypointing = false
const mushroomCost = 10

func SaveShortestTrail(mapid int, waypoints []point, pois []point, barriers map[string]typedGroup, paths map[string]typedGroup, baseFileName string, extension string) error {
	finals := [][]point{}
	if glBarriers != nil || glPaths != nil {
		panic("threading unsupported")
	}
	glBarriers = barriers
	glPaths = paths
	glWaypoints = waypoints
	defer func() {
		glBarriers = nil
		glPaths = nil
	}()

	var minPath path
	var min float64 = math.MaxFloat64
	for _, w := range waypoints {
		p := make(path, 1+len(pois))
		p[0] = w
		for i, poi := range pois {
			p[i+1] = poi
		}
		p.sort()
		ct := 0
		dist := p.Distance(true, true)
		log.Printf("Begin Optimization: %d", mapid)
		for {
			opt1 := p.optimize(true)
			opt2 := p.optimize(false)
			newDist := p.Distance(true, true)
			log.Printf("Optimized distance %.2f to %.2f", dist, newDist)
			dist = newDist
			if opt1 && opt2 {
				break
			}
			ct++
			if ct > 10000 {
				return errors.New("maximum optimizations exceeded")
			}
		}
		log.Printf("Optimization Complete: %d", mapid)

		if dist := p.Distance(true, true); dist < min {
			minPath = p
			min = dist
		}
	}

	final := make([]point, 0)
	for i := 0; i < len(minPath); i++ {
		final = append(final, minPath[i])
		if i+1 < len(minPath) {
			if minPath[i].barrier(minPath[i+1]) {
				takenPaths, ok := minPath[i].findPath(minPath[i+1])
				if !ok {
					panic("unexpected missing path")
				}
				for _, p := range takenPaths {
					final = append(final, p.points...)
				}
			} else if i+1 < len(minPath) {
				dist := minPath[i].Distance(minPath[i+1], false, false)

				if enableWaypointing {
					var waypoint *point = nil
					for i, w := range waypoints {
						wpDistance := waypointCost + w.Distance(minPath[i+1], false, true)
						if wpDistance < dist {
							waypoint = &w
							dist = wpDistance
						}
					}
					if waypoint != nil {
						finals = append(finals, final)
						final = []point{}
					}
				}
			}
		}
	}
	finals = append(finals, final)

	if len(finals) == 1 {
		b, err := PointsToTrlBytes(mapid, finals[0])
		if err != nil {
			return err
		}
		fileName := fmt.Sprintf("%s%s", baseFileName, extension)
		err = os.WriteFile(fileName, b, fs.ModePerm)
		if err != nil {
			return err
		}
	} else {
		for i, ls := range finals {
			b, err := PointsToTrlBytes(mapid, ls)
			if err != nil {
				return err
			}
			fileName := fmt.Sprintf("%s_%d%s", baseFileName, i+1, extension)
			err = os.WriteFile(fileName, b, fs.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	i := 0
	for _, b := range glBarriers {
		i++
		b, err := PointsToTrlBytes(mapid, b.points)
		if err == nil {
			fileName := fmt.Sprintf("%s_barrier_%d%s", baseFileName, i, extension)
			os.WriteFile(fileName, b, fs.ModePerm)
		}
	}
	return nil
}

func (p path) Distance(allowWaypoints bool, bypassBarriers bool) float64 {
	var out float64
	for i := 0; i < len(p)-1; i++ {
		out += p[i].Distance(p[i+1], allowWaypoints, bypassBarriers)
	}
	return out
}

func (p path) trySwap3p(i, j int) bool {

	//Note: non-directed graph, so no need to compute parts of the path that don't change
	first := i
	second := j

	var delta float64

	//Remove segments
	for r := i - 1; r < j+1; r++ {
		if r+1 < len(p) {
			delta -= p[r].Distance(p[r+1], true, true)
		}
	}

	delta += p[i-1].Distance(p[j], true, true)
	if j+1 < len(p) {
		delta += p[first].Distance(p[j+1], true, true)
	}
	for r := j; r > i-1; r-- {
		delta += p[r].Distance(p[r-1], false, true) //don't allow waypoints when following paths
	}

	if delta < 0 {
		for i, j := first, second; i < j; i, j = i+1, j-1 {
			p[i], p[j] = p[j], p[i]
		}
		return true
	}
	return false
}
func (p path) trySwap2p(i, j int) bool {
	if len(p) < 2 || i >= len(p) || j >= len(p) {
		return false
	}

	var oldSeg path
	var newSeg path
	if j+1 < len(p) {
		oldSeg = p[i-1 : j+2]
		newSeg = make(path, len(oldSeg))
		copy(newSeg, oldSeg)
		newSeg[1] = oldSeg[len(oldSeg)-2]
		newSeg[len(newSeg)-2] = oldSeg[1]
	} else {
		oldSeg = p[i-1 : j+1]
		newSeg = make(path, len(oldSeg))
		copy(newSeg, oldSeg)
		newSeg[1] = oldSeg[len(oldSeg)-1]
		newSeg[len(newSeg)-1] = oldSeg[1]
	}

	newDist := newSeg.Distance(true, true)
	oldDist := oldSeg.Distance(true, true)
	if newDist < oldDist {
		p[i], p[j] = p[j], p[i]
		return true
	}
	return false
}

func (p path) sort() {
	tmp := make([]point, len(p)-1)
	copy(tmp, p[1:])
	index := 0
	for len(tmp) > 0 {
		minIndex := 0
		minDist := p[index].Distance(tmp[0], true, true)
		for i := 1; i < len(tmp); i++ {
			if dist := p[index].Distance(tmp[i], true, true); dist < minDist {
				minDist = dist
				minIndex = i
			}
		}

		p[index+1] = tmp[minIndex]
		index++
		tmp[minIndex] = tmp[len(tmp)-1]
		tmp = tmp[:len(tmp)-1]
	}
}

// returns true of no changes made
func (p path) optimize(p3 bool) bool {
	done := true
	for i := 1; i < len(p)-1; i++ {
		for j := i + 1; j < len(p); j++ {
			if i == j {
				panic("unexpected")
			}
			if p3 {
				if p.trySwap3p(i, j) {
					done = false
				}
			} else {
				if p.trySwap2p(i, j) {
					done = false
				}
			}
		}
	}

	return done
}

func (src point) CalcDistance(dst point) float64 {
	diffX := dst.x - src.x
	diffY := dst.y - src.y
	diffZ := dst.z - src.z
	if diffZ > 0 {
		diffZ *= 1.5
	} else if diffZ < 0 {
		diffZ /= 1.5
	}
	return math.Sqrt(diffX*diffX + diffY*diffY + diffZ*diffZ)
}
func (src point) Distance(dst point, allowWaypoints bool, bypassBarriers bool) float64 {
	var pathDistance float64
	if src.barrier(dst) {
		if !bypassBarriers {
			return BarrierValue
		}
		if path, ok := src.findPath(dst); !ok {
			return BarrierValue
		} else {
			pathDistance, src = src.TakePath(path)
		}
	}

	pathDistance = pathDistance + src.CalcDistance(dst)

	//check if waypointing is faster
	if allowWaypoints && enableWaypointing {
		for _, w := range glWaypoints {
			waypointDistance := waypointCost + w.Distance(dst, false, false)
			if waypointDistance < pathDistance {
				return waypointDistance
			}
		}
	}

	return pathDistance
}

func (t typedGroup) IsMushroom() bool {
	return t.Type == GT_Mushroom
}
func (src point) TakePath(path []typedGroup) (float64, point) {
	var out float64
	for _, p := range path {
		if p.IsMushroom() {
			out += mushroomCost
			src = p.points.last()
		} else {
			for _, item := range p.points {
				out += src.CalcDistance(item)
				src = item
			}
		}
	}
	return out, src
}

func (src point) barrier(dst point) bool {
	for _, b := range glBarriers {
		if len(b.points) != 2 {
			fmt.Println("Unsupported barrier")
			return false
		}
		if b.Type == BT_DownOnly && dst.y < src.y {
			continue
		}
		if doIntersect(src, dst, b.points[0], b.points[1]) {
			return true
		}
	}
	return false
}

func (src path) first() point {
	return src[0]
}
func (src path) last() point {
	return src[len(src)-1]
}

func (t typedGroup) Equals(t1 typedGroup) bool {
	return t.name == t1.name
}
func availablePaths(usedList []typedGroup) []typedGroup {
	out := make([]typedGroup, 0)
	for _, global := range glPaths {
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

func (t typedGroup) Reverse() typedGroup {
	rev := make([]point, len(t.points))
	copy(rev, t.points)
	slices.Reverse(rev)
	return typedGroup{
		name:        t.name,
		reverseName: t.reverseName,
		Type:        t.Type,
		points:      rev,
	}
}

func (src point) pathTo(dst point, usedPaths []typedGroup) ([]typedGroup, bool) {
	choices := []typedGroup{}
	var addChoice = false
	start := src
	if len(usedPaths) > 0 {
		start = usedPaths[len(usedPaths)-1].points.last()
	}
	possiblePaths := [][]typedGroup{}
	for _, path := range availablePaths(usedPaths) {
		if !start.barrier(path.points.first()) {
			addChoice = true
			if !path.points.last().barrier(dst) {
				possiblePaths = append(possiblePaths, append(usedPaths, path))
			}
		}
		if !path.IsMushroom() && !start.barrier(path.points.last()) {
			addChoice = true
			if !path.points.first().barrier(dst) {
				possiblePaths = append(possiblePaths, append(usedPaths, path.Reverse()))
			}
		}

		if addChoice {
			choices = append(choices, path)
		}
	}

	if len(possiblePaths) == 0 {
		for _, choice := range choices {
			newPaths, ok := src.pathTo(dst, append(usedPaths, choice))
			if ok {
				possiblePaths = append(possiblePaths, newPaths)
			}
		}
	}
	if len(possiblePaths) == 0 {
		return []typedGroup{}, false
	}
	return cheapest(src, dst, possiblePaths), true
}

func calculatePathCost(src, dst point, group []typedGroup) float64 {
	total, newSrc := src.TakePath(group)
	return total + newSrc.Distance(dst, false, false)

}
func cheapest(src, dst point, groups [][]typedGroup) []typedGroup {
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

func (src point) findPath(dst point) ([]typedGroup, bool) {
	return src.pathTo(dst, make([]typedGroup, 0))
}
