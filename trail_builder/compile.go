package trailbuilder

import (
	"errors"
	"fmt"
	"gw2_markers_gen/maps"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func compilePaths(srcPath string) error {
	filesPath := fmt.Sprintf("%s/%s/", srcPath, CompiledAssetsDir)
	dstRoot := fmt.Sprintf("%s/%s/", srcPath, AssetsDir)
	files := readFiles(srcPath, CompiledTrailExtension)

	for _, f := range files {
		dstPath := dstRoot + strings.TrimPrefix(f, filesPath)
		dstPath = strings.TrimSuffix(dstPath, CompiledTrailExtension) + TrailExtension

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
	filesPath := fmt.Sprintf("%s/%s/", srcPath, CompiledAssetsDir)
	dstRoot := fmt.Sprintf("%s/%s/", srcPath, AssetsDir)
	files := readFiles(srcPath, CompiledAutoTrailExtension)
	mapsPath := fmt.Sprintf("%s/%s/", srcPath, MapsDir)

	for _, f := range files {
		dstPath := dstRoot + strings.TrimPrefix(f, filesPath)
		dstPath = strings.TrimSuffix(dstPath, CompiledAutoTrailExtension) + TrailExtension

		//srcInfo, err := os.Stat(srcPath)
		//if err != nil {
		//	return err
		//}
		//dstInfo, err := os.Stat(dstPath)
		//Skip recompiling the resource if no changes have been made
		//if err == nil && dstInfo.ModTime().After(srcInfo.ModTime()) {
		//	continue
		//}

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
		barriers := readPointGroups(mapPath, "barriers.txt")
		waypoints := readPoints(mapPath, "waypoints.txt")
		paths := readPointGroups(mapPath, "paths.txt")
		pois := readPoints(mapPath, fileName)

		os.MkdirAll(filepath.Dir(dstPath), fs.ModePerm)
		err = SaveShortestTrail(mapId, waypoints, pois, barriers, paths, dstPath)
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
