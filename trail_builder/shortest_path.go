package trailbuilder

import (
	"errors"
	"fmt"
	"io/fs"
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
		p.Distance(true, true)
		for {
			opt1 := p.optimize(true)
			opt2 := p.optimize(false)
			//newDist := p.Distance(true, true)
			//log.Printf("Optimized distance %.2f to %.2f", dist, newDist)
			//dist = newDist
			if opt1 && opt2 {
				break
			}
			ct++
			if ct > 10000 {
				return errors.New("maximum optimizations exceeded")
			}
		}
		//log.Println("Optimization complete")

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
				takePath, ok := minPath[i].findPath(minPath[i+1])
				if !ok {
					panic("unexpected missing path")
				}
				final = append(final, takePath...)
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
		diffZ *= 4
	} else if diffZ < 0 {
		diffZ /= 2
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

func (src point) TakePath(path []point) (float64, point) {
	var out float64
	for i := 0; i < len(path); i++ {
		out += src.CalcDistance(path[i])
		src = path[i]
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

func (src point) findPath(dst point) ([]point, bool) {
	var found bool
	var returnPath []point
	var forward bool
	minDist := math.MaxFloat64

	for _, p := range glPaths {
		if len(p.points) == 0 {
			continue
		}

		//Don't try to recursively take the same path
		if dst == p.points[0] || dst == p.points[len(p.points)-1] {
			continue
		}

		//Check if following the path works
		if !src.barrier(p.points[0]) && !p.points[len(p.points)-1].barrier(dst) {
			var dist float64
			//Since mushroom time is mostly static, we give them a static weight
			if p.Type == GT_Mushroom {
				dist = mushroomCost
			} else {
				//Allow waypointing to reach the path, but not to go from the path end to the point
				//Don't bypass a barrier to reach the path (in this case..another path should be found)
				//Do allow searching for another path to bypass barriers after the first
				dist = src.Distance(p.points[0], true, false) + p.points[len(p.points)-1].Distance(dst, false, true)
				for i := 0; i < len(p.points)-1; i++ {
					dist += p.points[i].CalcDistance(p.points[i+1])
				}
			}
			if dist < minDist {
				minDist = dist
				returnPath = p.points
				forward = true
				found = true
			}
		}

		//Check if following the path in reverse works
		// Mushroom paths can't go in reverse
		if p.Type != GT_Mushroom && !src.barrier(p.points[len(p.points)-1]) && !p.points[0].barrier(dst) {
			//Allow waypointing to reach the path, but not to go from the path end to the point
			//Don't bypass a barrier to reach the path (in this case..another path should be found)
			//Do allow searching for another path to bypass barriers after the first
			dist := src.Distance(p.points[len(p.points)-1], true, false) + p.points[0].Distance(dst, false, true)
			for i := len(p.points) - 1; i > 0; i-- {
				dist += p.points[i].CalcDistance(p.points[i-1])
			}
			if dist < minDist {
				minDist = dist
				returnPath = p.points
				forward = false
				found = true
			}
		}
	}

	if !forward && found {
		tmp := make([]point, len(returnPath))
		copy(tmp, returnPath)
		slices.Reverse(tmp)
		return tmp, found
	}
	return returnPath, found
}
