package trailbuilder

import (
	"fmt"
	"gw2_markers_gen/location"
	"io/fs"
	"log"
	"os"
	"sync"
)

func SaveShortestTrail(
	mapid int,
	waypoints []location.Point,
	pois []location.Point,
	barriers map[string]location.TypedGroup,
	paths map[string]location.TypedGroup,
	ptpPaths map[string]location.TypedGroup,
	baseFileName string,
	extension string) error {

	location.SetGlobals(barriers, paths, waypoints, ptpPaths)
	defer location.ResetGlobals()

	g := location.Path(pois).ToGraph()
	g.AddWaypoints(waypoints)
	pathList := g.GetPaths()

	wg := sync.WaitGroup{}
	log.Printf("Optimizing Map: %d", mapid)
	for i := range pathList {
		index := i
		p := pathList[i]
		wg.Add(1)
		go func() {
			log.Printf("[%d] Starting distance: %.2f", index+1, p.EndDistance())
			defer wg.Add(-1)
			for p.Optimize(false) {
			}
			log.Printf("[%d] Final map distance: %.2f", index+1, p.EndDistance())
		}()
	}
	wg.Wait()

	final, _ := pathList.Shortest()
	outputPaths := final.ToPath()

	for i, points := range outputPaths {
		b, err := PointsToTrlBytes(mapid, points)
		if err != nil {
			return err
		}

		fileName := fmt.Sprintf("%s_%d%s", baseFileName, i+1, extension)
		log.Printf("Generating file: %s", fileName)
		err = os.WriteFile(fileName, b, fs.ModePerm)
		if err != nil {
			return err
		}
	}

	/*
		i := 0
		for _, b := range glBarriers {
			i++
			b, err := PointsToTrlBytes(mapid, b._points)
			if err == nil {
				fileName := fmt.Sprintf("%s_barrier_%d%s", baseFileName, i, extension)
				os.WriteFile(fileName, b, fs.ModePerm)
			}
		}
	*/
	return nil
}

/*
func SaveShortestTrailWithZones(mapid int, waypoints []location.Point, srcPoints []location.Point, zoneTrail ZoneTrail, barriers map[string]location.TypedGroup, paths map[string]location.TypedGroup, ptpPaths map[string]location.TypedGroup, baseFileName string, extension string) error {
	location.SetGlobals(barriers, paths, waypoints, ptpPaths)
	defer location.ResetGlobals()

	regions := zoneTrail.PartitionPoints(srcPoints)
	for i, r := range regions {
		g := location.Path(r.Points).ToGraph()
		if r.Start == nil {
			g.AddWaypoints(waypoints)
		} else {
			g.AddWaypoints([]location.Point{*r.Start})
		}
		if r.End != nil {
			g.SetEndpoint(*r.End)
		}
		r.graph = &g
		regions[i] = r
	}

	outputList := []*location.GraphPath{}
	log.Printf("Optimizing Map: %d", mapid)
	for _, region := range regions {
		wg := sync.WaitGroup{}
		pathList := region.graph.GetPaths()
		for i := range pathList {
			index := i
			p := pathList[i]
			wg.Add(1)
			go func() {
				log.Printf("[%d] Starting distance: %.2f", index+1, p.EndDistance())
				defer wg.Add(-1)
				for p.Optimize(p.BindEnd) {
				}
				log.Printf("[%d] Final map distance: %.2f", index+1, p.EndDistance())
			}()
		}
		wg.Wait()
		shortest, _ := pathList.Shortest()
		outputList = append(outputList, shortest)
	}

	points := make([][]location.Point, 0)
	current := []location.Point{}
	for i, path := range outputList {
		pts := path.ToPath()
		//Only draw the first point (waypoint/start) if this is the first entry
		if len(current) != 0 {
			pts = pts[1:]
		}
		//If an endpoint binding is set, strip it. (these are strictly for controlling the approximate destination)
		if path.BindEnd {
			pts = pts[:len(pts)-1]
		}
		current = append(current, pts...)

		//If the next segment has no starting point, this path ends (and you're expected to waypoint)
		if regions[i].Start != nil {
			points = append(points, current)
			current = []location.Point{}
		}
	}
	if len(current) > 0 {
		points = append(points, current)
	}

	if len(points) == 1 {
		b, err := PointsToTrlBytes(mapid, points[0])
		if err != nil {
			return err
		}
		fileName := fmt.Sprintf("%s%s", baseFileName, extension)
		err = os.WriteFile(fileName, b, fs.ModePerm)
		if err != nil {
			return err
		}
	} else {
		for index, ls := range points {
			b, err := PointsToTrlBytes(mapid, ls)
			if err != nil {
				return err
			}
			fileName := fmt.Sprintf("%s_%d%s", baseFileName, index+1, extension)
			err = os.WriteFile(fileName, b, fs.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	i := 0
	for _, b := range location.GLOBAL_Barriers {
		i++
		b, err := PointsToTrlBytes(mapid, b.Points())
		if err == nil {
			fileName := fmt.Sprintf("%s_barrier_%d%s", baseFileName, i, extension)
			os.WriteFile(fileName, b, fs.ModePerm)
		}
	}
	return nil
}
*/
