package maps

import (
	"fmt"
	"gw2_markers_gen/categories"
	"log"
	"os"
	"strings"
)

const infoFileName = "mapinfo.txt"
const poiExtension = ".poi"
const trailExtension = ".trail"

type Map struct {
	MapName string
	MapId   int
	POIs    []POI
	Trails  []Trail
}

var validateFileExists func(fname string) string

// Set a routine used for detecting if a file exists in our assets directory
func SetValidation(f func(fname string) string) {
	validateFileExists = f
}

// Compiles a list of all maps from source map directory
func Compile(categories []categories.Category, path string) ([]Map, []string) {
	out := []Map{}
	warns := []string{}
	items, _ := os.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			newMap, newWarns, err := compileMap(categories, fmt.Sprintf("%s/%s", path, item.Name()))
			if err != nil {
				log.Printf("Failed to load map: %s, Error: %s", item.Name(), err.Error())
				continue
			}
			warns = append(warns, newWarns...)
			out = append(out, newMap)
		}
	}
	return out, warns
}

// Helper to deal with quoted strings
func trim(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(s), `"`), `"`)
}

// Pulls Category out of the line if present
// Returns: X, X, false on line not being a valid pair
// Return: X, Warning, true on 1st line being a pair not defining a category
// Returns category, warning, true when issues are detected on the category data
// Returns category, false, true on valid configuration
func getCategory(categoryList []categories.Category, line string) (string, string, bool) {
	pair := strings.Split(line, "=")
	if len(pair) != 2 {
		return "", "", false
	}

	if !strings.EqualFold("category", pair[0]) {
		return "", fmt.Sprintf("Invalid category pair: %s", line), true
	}

	category := pair[1]

	for _, cat := range categoryList {
		if cat.MatchString(category) {
			return category, "", true
		}
	}
	return category, "category not found", true
}
