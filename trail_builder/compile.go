package trailbuilder

import (
	"errors"
	"fmt"
	"gw2_markers_gen/files"
	"gw2_markers_gen/location"
	"gw2_markers_gen/maps"
	"gw2_markers_gen/utils"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
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
		filePrefix := strings.TrimSuffix(strings.TrimPrefix(f, filesPath), files.AutoTrailExtension)
		tmp := dstRoot + filePrefix
		filePrefix = path.Base(filePrefix)
		baseDstPath := path.Dir(tmp)
		templateOutputFileName := fmt.Sprintf("%s/%s", baseDstPath, filePrefix)

		oldestTime := files.OldestModified(baseDstPath, filePrefix, files.TrailExtension)
		checkCompileTime := oldestTime != time.Time{}

		b, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Error compiling resource: %s, Error: %s", f, err.Error())
			continue
		}

		var ok bool
		var fileLs []string
		var mapName string

		m := utils.ReadMap(string(b), '\n')
		if mapName, ok = utils.MapString(m, "map"); !ok {
			log.Printf("Missing map name: %s", f)
			continue
		} else if fileLs, ok = utils.MapStringArray(m, "file"); !ok {
			log.Printf("File name not specified: %s", f)
			continue
		}
		mapName = utils.Trim(mapName)
		mapPath := fmt.Sprintf("%s%s", mapsPath, mapName)
		mapId, _, err := maps.ReadMapInfo(mapPath)
		if err != nil {
			return err
		}

		barrierFile := fmt.Sprintf("%s/%s", mapPath, files.BarriersFile)
		waypointsFile := fmt.Sprintf("%s/%s", mapPath, files.WaypointsFile)
		pathsFile := fmt.Sprintf("%s/%s", mapPath, files.PathsFile)
		ptpPathsFile := fmt.Sprintf("%s/%s", mapPath, files.PtpPathsFile)

		barriers := files.ReadTypedGroup(barrierFile)
		waypoints := files.ReadPoints(waypointsFile)
		paths := files.ReadTypedGroup(pathsFile)
		ptpPaths := files.ReadPTPPoints(ptpPathsFile)
		var pois []location.Point = []location.Point{}
		for _, f := range fileLs {
			poiFile := fmt.Sprintf("%s/%s", mapPath, f)
			pois = append(pois, files.ReadPoints(poiFile)...)
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
			lastCompile := oldestTime
			if !forceRecompile {
				changed := false
				if files.FileChangedSince(lastCompile, barrierFile) ||
					files.FileChangedSince(lastCompile, waypointsFile) ||
					files.FileChangedSince(lastCompile, pathsFile) ||
					files.FileChangedSince(lastCompile, ptpPathsFile) {
					changed = true
				}
				for _, f := range fileLs {
					poiFile := fmt.Sprintf("%s/%s", mapPath, f)
					if files.FileChangedSince(lastCompile, poiFile) {
						changed = true
						break
					}
				}
				if !changed {
					continue
				}
			}
		}

		files.RemoveWithExtension(baseDstPath, filePrefix, files.TrailExtension)
		os.MkdirAll(dstRoot, fs.ModePerm)
		err = SaveShortestTrail(mapId, waypoints, pois, barriers, paths, ptpPaths, templateOutputFileName, files.TrailExtension)
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

func checkForDuplicates(pts []location.Point) error {
	var err error
	for i := 0; i < len(pts); i++ {
		for j := i + 1; j < len(pts); j++ {
			if pts[i].Same(pts[j]) {
				if !pts[i].AllowDuplicate || !pts[j].AllowDuplicate {
					err = fmt.Errorf("duplicate point i:%d, (%+v), j:%d, (%+v)", i, pts[i], j, pts[j])
					log.Println(err.Error())
				}
			}
		}
	}
	return err
}
