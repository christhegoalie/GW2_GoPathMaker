package trailbuilder

import (
	"errors"
	"fmt"
	"gw2_markers_gen/files"
	"gw2_markers_gen/maps"
	"io/fs"
	"log"
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

		var mapName string
		var fileName string
		var ok bool
		m := readMap(string(b), '\n')
		if mapName, ok = m["map"]; !ok {
			log.Printf("Missing map name: %s", f)
			continue
		} else if fileName, ok = m["file"]; !ok {
			log.Printf("File name not specified: %s", f)
			continue
		}
		mapName = trim(mapName)
		fileName = trim(fileName)
		mapPath := fmt.Sprintf("%s%s", mapsPath, mapName)
		if !fileExists(mapPath) {
			log.Printf("Src POI File %s not found: %s", fileName, f)
			continue
		}
		mapId, _, err := maps.ReadMapInfo(mapPath)
		if err != nil {
			return err
		}

		barrierFile := fmt.Sprintf("%s/%s", mapPath, files.BarriersFile)
		waypointsFile := fmt.Sprintf("%s/%s", mapPath, files.WaypointsFile)
		pathsFile := fmt.Sprintf("%s/%s", mapPath, files.PathsFile)
		poiFile := fmt.Sprintf("%s/%s", mapPath, fileName)

		barriers := readTypedGroup(barrierFile)
		waypoints := readPoints(waypointsFile)
		paths := readTypedGroup(pathsFile)
		pois := readPoints(poiFile)

		if checkCompileTime {
			lastCompile := dstFileInfo.ModTime()
			if !forceRecompile {
				if !fileChangedSince(lastCompile, dstPath) &&
					!fileChangedSince(lastCompile, barrierFile) &&
					!fileChangedSince(lastCompile, waypointsFile) &&
					!fileChangedSince(lastCompile, pathsFile) &&
					!fileChangedSince(lastCompile, poiFile) {
					continue
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
