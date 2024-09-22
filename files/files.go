package files

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const CategoryExtension = ".cat"

// Map Marker Extensions
const MarkerPoiExtension = ".poi"
const MarkerTrailExtension = ".trail"

// Trail Extensions
const TrailExtension = ".trl"
const AutoTrailExtension = ".atrl"     //Generates a .trl using a graph alogrithm
const CompiledTrailExtension = ".rtrl" //Generates a .trl using map/point definitions

// Root directory
const CategoriesDirectory = "categories"
const MapsDirectory = "maps"
const AssetsDirectory = "assets"
const CompiledAssetsDirectory = "compiled_assets"

// Maps info files [in map directory]
const BarriersFile = "barriers.txt"
const WaypointsFile = "waypoints.txt"
const PathsFile = "paths.txt"
const MapInfoFile = "mapinfo.txt"
const PtpPathsFile = "edges.txt"

// Export files
const OutputCategoryFile = "_markerCategories.xml"

func FilesByExtension(root string, extensions ...string) []string {
	items, _ := os.ReadDir(root)
	fileList := []string{}
	for _, item := range items {
		fullPath := fmt.Sprintf("%s/%s", root, item.Name())
		if item.IsDir() {
			fileList = append(fileList, FilesByExtension(fullPath, extensions...)...)
		}
		for _, ext := range extensions {
			if strings.HasSuffix(item.Name(), ext) {
				fileList = append(fileList, fullPath)
			}
		}
	}
	return fileList
}

func FilesWithPrefixSuffix(root string, prefix string, suffix string) []string {
	items, _ := os.ReadDir(root)
	fileList := []string{}
	for _, item := range items {
		fullPath := fmt.Sprintf("%s/%s", root, item.Name())
		if item.IsDir() {
			fileList = append(fileList, FilesByExtension(fullPath, prefix, suffix)...)
		}
		if strings.HasPrefix(item.Name(), prefix) && strings.HasSuffix(item.Name(), suffix) {
			fileList = append(fileList, fullPath)
		}
	}
	return fileList

}

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func FileChangedSince(timestamp time.Time, filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return true
	}
	return info.ModTime().After(timestamp)
}

func OldestModified(srcDir string, prefix string, suffix string) time.Time {
	var first bool = true
	files := FilesWithPrefixSuffix(srcDir, prefix, suffix)
	var out time.Time
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			panic(err)
		}
		if first || info.ModTime().Before(out) {
			first = false
			out = info.ModTime()
		}
	}
	return out
}
func RemoveWithExtension(srcDir string, prefix, suffix string) {
	files := FilesWithPrefixSuffix(srcDir, prefix, suffix)
	for _, f := range files {
		e := os.Remove(f)
		if e != nil {
			panic(e)
		}
	}
}
