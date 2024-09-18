package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gw2_markers_gen/files"
	"gw2_markers_gen/location"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Correlations struct {
	list []Correlation
}
type Correlation struct {
	category string
	pois     []location.Point
	entries  []location.PointList
}

type EntryPoint struct {
	Name         string
	Location     location.Point
	PossiblePois []location.Point
}

type Summary struct {
	Data              EntryPoints
	MinReferences     int
	MaxReferences     int
	AverageReferences float64
	ExpectedPoints    int
}
type EntryPoints []EntryPoint

// Takes a directory with correlation files
// Every group should contain a .poi file EX: warclaw.poi
// the "group" should contain files with the prefix "category_", and the ".txt" extension EX: warclaw_1.txt
// Each correlation file should contain a subset of the locations from the poi file
func main() {
	srcDirectory := *flag.String("s", "correlations", "Correlation Directory")
	flag.Parse()

	correlations := getPOICorrelations(srcDirectory)
	for _, c := range correlations.list {
		log.Printf("Generating correlation info for: %s", c.category)
		//Build a list of all unique points
		//find the file with the highest point count (best approximation of the actual point count)
		totalPoints := 0
		index := -1
		for i, entry := range c.entries {
			if len(entry) > totalPoints {
				index = i
				totalPoints = len(entry)
			}
		}
		if index == -1 {
			log.Printf("No data found for: %s", c.category)
			continue
		}

		list := make(EntryPoints, 0)
		for i, p := range c.entries[index] {
			list = append(list, EntryPoint{Name: fmt.Sprintf("Point_%d", i), Location: p, PossiblePois: copyPoints(c.pois)})
		}

		//Calculate possible pois for all points
		for pointIndex, point := range list {
			//Any instance containing the same point, excludes all other pois on that map
			for i, c := range c.entries {
				//Only perform computation if the point is already on the map
				if i != index && !c.Contains(point.Location) {
					continue
				}
				for _, p := range c {
					if !p.Same(point.Location) {
						point.removePoi(p)
						list[pointIndex] = point
					}
				}
			}
		}

		var min, max, ct int = math.MaxInt, 0, 0
		var avg float64
		for _, e := range list {
			poiCount := len(e.PossiblePois)
			if poiCount < min {
				min = poiCount
			}
			if poiCount > max {
				max = poiCount
			}
			avg += float64(poiCount)
			ct++
		}
		if ct > 0 {
			avg = avg / float64(ct)
		}
		writeResults(srcDirectory, c.category, Summary{ExpectedPoints: len(list), MinReferences: min, MaxReferences: max, AverageReferences: avg, Data: list})
	}
}

func writeResults(path string, category string, summary Summary) {
	log.Printf("%s Summary. Points: %d, Min: %d, Max: %d, Avg: %.1f", category, summary.ExpectedPoints, summary.MinReferences, summary.MaxReferences, summary.AverageReferences)
	b, _ := json.MarshalIndent(summary, "", "\t")
	os.WriteFile(fmt.Sprintf("%s/%s.txt", path, category), b, os.ModePerm)
}
func copyPoints(pts []location.Point) []location.Point {
	out := make([]location.Point, len(pts))
	copy(out, pts)
	return out
}
func (p *EntryPoint) removePoi(poi location.Point) {
	for i, item := range p.PossiblePois {
		if item.Same(poi) {
			p.PossiblePois[i] = p.PossiblePois[len(p.PossiblePois)-1]
			p.PossiblePois = p.PossiblePois[:len(p.PossiblePois)-1]
			return
		}
	}
}
func getPOICorrelations(pathName string) Correlations {
	out := Correlations{}
	fileList := files.FilesByExtension(pathName, files.MarkerPoiExtension)
	for _, f := range fileList {
		category := strings.TrimSuffix(filepath.Base(f), files.MarkerPoiExtension)
		points := files.ReadPoints(f)
		out.list = append(out.list, Correlation{pois: points, category: category, entries: findEntries(pathName, category)})
	}
	return out
}
func findEntries(pathName string, category string) []location.PointList {
	entries := []location.PointList{}
	fList := files.FilesByExtension(fmt.Sprintf("%s/%s", pathName, category), category, ".txt")
	for _, f := range fList {
		entries = append(entries, files.ReadPoints(f))
	}

	return entries
}
