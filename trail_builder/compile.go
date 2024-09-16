package trailbuilder

import (
	"errors"
	"fmt"
	"gw2_markers_gen/files"
	"gw2_markers_gen/maps"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var forceRecompile bool = false

func compilePaths(srcPath string) error {
	filesPath := fmt.Sprintf("%s/%s/", srcPath, files.CompiledAssetsDirectory)
	dstRoot := fmt.Sprintf("%s/%s/", srcPath, files.AssetsDirectory)
	fileList := files.FilesByExtension(srcPath, files.CompiledTrailExtension)

	for _, f := range fileList {
		dstPath := dstRoot + strings.TrimPrefix(f, filesPath)
		dstPath = strings.TrimSuffix(dstPath, files.CompiledTrailExtension) + files.TrailExtension

		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			return err
		}
		dstInfo, err := os.Stat(dstPath)
		//Skip recompiling the resource if no changes have been made
		if err == nil && dstInfo.ModTime().After(srcInfo.ModTime()) {
			continue
		}

		b, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Error compiling resource: %s, Error: %s", f, err.Error())
			continue
		}
		lines := strings.Split(string(b), "\n")
		fileData, err := LinesToTRLBytes(lines)
		if err != nil {
			log.Printf("Error compiling resource: %s, Error: %s", f, err.Error())
			continue
		}

		os.MkdirAll(filepath.Dir(dstPath), fs.ModePerm)
		err = os.WriteFile(dstPath, fileData, fs.ModePerm)
		if err != nil {
			log.Printf("Error saving compiled resource: %s, Error: %s", f, err.Error())
			continue
		}
	}
	return nil
}

func compileAutoPaths(srcPath string) error {
	filesPath := fmt.Sprintf("%s/%s/", srcPath, files.CompiledAssetsDirectory)
	dstRoot := fmt.Sprintf("%s/%s/", srcPath, files.AssetsDirectory)
	fileList := files.FilesByExtension(srcPath, files.AutoTrailExtension)
	mapsPath := fmt.Sprintf("%s/%s/", srcPath, files.MapsDirectory)

	for _, f := range fileList {
		dstPath := dstRoot + strings.TrimPrefix(f, filesPath)
		baseDstPath := strings.TrimSuffix(dstPath, files.AutoTrailExtension)
		dstPath = baseDstPath + files.TrailExtension

		dstFileInfo, err := os.Stat(dstPath)
		checkCompileTime := err == nil

		b, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Error compiling resource: %s, Error: %s", f, err.Error())
			continue
		}

		var ok bool
		var fileLs []string
		var mapName string

		m := readMap(string(b), '\n')
		if mapName, ok = mapString(m, "map"); !ok {
			log.Printf("Missing map name: %s", f)
			continue
		} else if fileLs, ok = mapStringArray(m, "file"); !ok {
			log.Printf("File name not specified: %s", f)
			continue
		}
		mapName = trim(mapName)
		mapPath := fmt.Sprintf("%s%s", mapsPath, mapName)
		mapId, _, err := maps.ReadMapInfo(mapPath)
		if err != nil {
			return err
		}

		barrierFile := fmt.Sprintf("%s/%s", mapPath, files.BarriersFile)
		waypointsFile := fmt.Sprintf("%s/%s", mapPath, files.WaypointsFile)
		pathsFile := fmt.Sprintf("%s/%s", mapPath, files.PathsFile)

		barriers := readTypedGroup(barrierFile)
		waypoints := readPoints(waypointsFile)
		paths := readTypedGroup(pathsFile)
		var pois []Point = []Point{}
		for _, f := range fileLs {
			poiFile := fmt.Sprintf("%s/%s", mapPath, f)
			pois = append(pois, readPoints(poiFile)...)
		}
		if len(pois) == 0 {
			log.Printf("No POIs found for: %s", mapName)
			continue
		}
		if err := checkForDuplicates(pois); err != nil {
			log.Printf("Path generation failed [%s], error: %s", mapName, err.Error())
			continue
		}

		if checkCompileTime {
			lastCompile := dstFileInfo.ModTime()
			if !forceRecompile {
				if fileChangedSince(lastCompile, dstPath) ||
					fileChangedSince(lastCompile, barrierFile) ||
					fileChangedSince(lastCompile, waypointsFile) ||
					fileChangedSince(lastCompile, pathsFile) {
					continue
				}
				for _, f := range fileLs {
					poiFile := fmt.Sprintf("%s/%s", mapPath, f)
					if fileChangedSince(lastCompile, poiFile) {
						continue
					}
				}
			}
		}

		os.MkdirAll(filepath.Dir(dstPath), fs.ModePerm)
		err = SaveShortestTrail(mapId, waypoints, pois, barriers, paths, baseDstPath, files.TrailExtension)
		if err != nil {
			log.Printf("Error saving compiled resource: %s, Error: %s", f, err.Error())
			continue
		}
	}
	return nil
}
func CompileResources(srcPath string) error {
	err1 := compilePaths(srcPath)
	if err1 != nil {
		log.Printf("Failed to compile paths: %s", err1.Error())
	}
	err2 := compileAutoPaths(srcPath)
	if err2 != nil {
		log.Printf("Failed to compile auto paths: %s", err2.Error())
	}
	if err1 != nil {
		return err1
	}
	return err2
}

func fileExists(fname string) bool {
	if _, err := os.Stat(fname); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func distance(p1, p2 Point) float64 {
	d1, d2, d3 := p2.X-p1.X, p2.Y-p1.Y, p2.Z-p1.Z
	return math.Sqrt(d1*d1 + d2*d2 + d3*d3)
}

func (src Point) same(point Point) bool {
	return distance(src, point) < 4
}
func checkForDuplicates(pts []Point) error {
	var err error
	for i := 0; i < len(pts); i++ {
		for j := i + 1; j < len(pts); j++ {
			if pts[i].same(pts[j]) {
				if !pts[i].AllowDuplicate || !pts[j].AllowDuplicate {
					err = fmt.Errorf("duplicate point i:%d, (%+v), j:%d, (%+v)", i, pts[i], j, pts[j])
					log.Println(err.Error())
				}
			}
		}
	}
	return err
}
