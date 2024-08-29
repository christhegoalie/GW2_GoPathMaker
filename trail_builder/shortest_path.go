package trailbuilder

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"slices"
)

type path []point

var glBarriers map[string]pointGroup
var glPaths map[string]pointGroup

func SaveShortestTrail(mapid int, waypoints []point, pois []point, barriers map[string]pointGroup, paths map[string]pointGroup, fileName string) error {
	if glBarriers != nil || glPaths != nil {
		panic("threading unsupported")
	}
	glBarriers = barriers
	glPaths = paths
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
		for !p.optimize() {
		}

		if dist := p.Distance(); dist < min {
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
			}
		}
	}

	b, err := PointsToTrlBytes(mapid, final)
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, b, fs.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (p path) Distance() float64 {
	var out float64
	for i := 0; i < len(p)-1; i++ {
		out += p[i].Distance(p[i+1])
	}
	return out
}

func (p path) trySwap(i, j int) bool {
	//Prev to first
	orig := p[i-1].Distance(p[i])
	new := p[i-1].Distance(p[j])

	if j == i+1 {
		//Distance between swapped nodes
		orig += p[i].Distance(p[j])
		new += p[j].Distance(p[i])
	} else {
		//node to middle list
		orig += p[i].Distance(p[i+1])
		new += p[j].Distance(p[i+1])
		//middle list to second node
		orig += p[j-1].Distance(p[j])
		new += p[j-1].Distance(p[i])
	}

	//If not at end, second node to rest of list
	if j+1 < len(p) {
		orig += p[j].Distance(p[j+1])
		new += p[i].Distance(p[j+1])
	}
	if new < orig {
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
		minDist := p[index].Distance(tmp[0])
		for i := 1; i < len(tmp); i++ {
			if dist := p[index].Distance(tmp[i]); dist < minDist {
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

// returns true of no changes
func (p path) optimize() bool {
	done := true
	for i := 1; i < len(p)-1; i++ {
		for j := i + 1; j < len(p); j++ {
			if i == j {
				panic("unexpected")
			}
			if p.trySwap(i, j) {
				done = false
			}
		}
	}
	if !done {
		for i := len(p) - 1; i > 0; i-- {
			for j := i + 1; j < len(p); j++ {
				if i == j {
					panic("unexpected")
				}
				if p.trySwap(i, j) {
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
func (src point) Distance(dst point) float64 {
	var pathDistance float64
	if src.barrier(dst) {
		if path, ok := src.findPath(dst); !ok {
			return math.MaxFloat64
		} else {
			pathDistance, src = src.TakePath(path)
		}
	}

	return pathDistance + src.CalcDistance(dst)
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
		if len(b) != 2 {
			fmt.Println("Unsupported barrier")
			return false
		}
		if doIntersect(src, dst, b[0], b[1]) {
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
		if len(p) == 0 {
			continue
		}

		if !src.barrier(p[0]) && !p[len(p)-1].barrier(dst) {
			dist := src.Distance(p[0]) + p[len(p)-1].Distance(dst)
			for i := 0; i < len(p)-1; i++ {
				dist += p[i].CalcDistance(p[i+1])
			}
			if dist < minDist {
				minDist = dist
				returnPath = p
				forward = true
				found = true
			}
		}

		if !src.barrier(p[len(p)-1]) && !p[0].barrier(dst) {
			dist := src.Distance(p[len(p)-1]) + p[0].Distance(dst)
			for i := len(p) - 1; i > 0; i-- {
				dist += p[i].CalcDistance(p[i-1])
			}
			if dist < minDist {
				minDist = dist
				returnPath = p
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
