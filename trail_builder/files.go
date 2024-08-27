package trailbuilder

import (
	"fmt"
	"os"
	"strings"
)

const AssetsDir = "assets"
const CompiledAssetsDir = "compiled_assets"
const CompiledTrailExtension = ".rtrl"
const TrailExtension = ".trl"

// Recursively find all .trail and .poi files
func readFiles(path string) []string {
	items, _ := os.ReadDir(path)
	files := []string{}
	for _, item := range items {
		fullPath := fmt.Sprintf("%s/%s", path, item.Name())
		if item.IsDir() {
			files = append(files, readFiles(fullPath)...)
		} else if strings.HasSuffix(item.Name(), CompiledTrailExtension) {
			files = append(files, fullPath)
		}
	}
	return files
}
