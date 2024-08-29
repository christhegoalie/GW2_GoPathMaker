package trailbuilder

import (
	"gw2_markers_gen/maps"
	"io/fs"
	"math"
	"os"
)

type path []point

func SaveShortestTrail(mapid int, waypoints []maps.POI, pois []maps.POI, fileName string) {
	var minPath path
	var min float64 = math.MaxFloat64
	for _, w := range waypoints {
		p := make(path, 1+len(pois))
		p[0] = point{w.XPos, w.YPos, w.ZPos}
		for i, poi := range pois {
			p[i+1] = point{poi.XPos, poi.YPos, poi.ZPos}
		}
		p.sort()
		for !p.optimize() {
		}

		if dist := p.Distance(); dist < min {
			minPath = p
			min = dist
		}
	}

	b, err := PointsToTrlBytes(mapid, minPath)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fileName, b, fs.ModePerm)
	if err != nil {
		panic(err)
	}
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

func (src point) Distance(dst point) float64 {
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
