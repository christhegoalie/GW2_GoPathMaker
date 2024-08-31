package files

import (
	"fmt"
	"os"
	"strings"
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
