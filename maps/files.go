package maps

import (
	"errors"
	"fmt"
	"gw2_markers_gen/categories"
	"gw2_markers_gen/files"
	"log"
	"os"
	"strconv"
	"strings"
)

// read/parse a .trail file into a list of POI structures
func ReadTrails(categories []categories.Category, fileName string) ([]Trail, []string, error) {
	trails := []Trail{}
	warns := []string{}

	b, err := os.ReadFile(fileName)
	if err != nil {
		return trails, warns, err
	}

	lines := strings.Split(string(b), "\n")
	if len(lines) < 1 {
		return trails, warns, nil
	}
	i := 1
	category, warning, ok := getCategory(categories, strings.TrimSpace(lines[0]))
	if !ok {
		i = 0
		log.Printf("Warn: [%s] Category not set", fileName)
	}
	if warning != "" {
		warns = append(warns, fmt.Sprintf("Warn: [%s]: %s", fileName, warning))
	}
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		trail, newWarns, err := parseTrail(category, line)
		if err != nil {
			return trails, warns, err
		}

		trails = append(trails, trail)
		warns = append(warns, newWarns...)
	}
	return trails, warns, nil
}

// read/parse a .poi file into a list of POI structures
func ReadPOIs(categories []categories.Category, fileName string) ([]POI, []string, error) {
	pois := []POI{}
	warns := []string{}

	b, err := os.ReadFile(fileName)
	if err != nil {
		return pois, warns, err
	}

	lines := strings.Split(string(b), "\n")
	if len(lines) < 1 {
		return pois, warns, nil
	}
	i := 1
	category, warning, ok := getCategory(categories, strings.TrimSpace(lines[0]))
	if !ok {
		i = 0
		log.Printf("Warn: [%s] Category not set", fileName)
	}
	if warning != "" {
		warns = append(warns, fmt.Sprintf("Warn: [%s]: %s", fileName, warning))
	}

	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		poi, newWarns, err := parsePoi(category, line)
		if err != nil {
			return pois, warns, err
		}

		pois = append(pois, poi)
		warns = append(warns, newWarns...)
	}
	return pois, warns, nil
}

// Read a space seperated line of key value pairs seperated by "=", and return a map
func readMap(line string) map[string]string {
	out := make(map[string]string)
	needEqual := true
	quoted := false
	key := ""
	tmp := strings.Builder{}
	for i := 0; i < len(line); i++ {
		if needEqual {
			if line[i] == '=' {
				if tmp.Len() == 0 {
					continue
				}
				key = tmp.String()
				tmp.Reset()
				needEqual = false
			} else {
				tmp.WriteByte(line[i])
			}
		} else {
			if !quoted && line[i] == ' ' {
				needEqual = true
				out[key] = tmp.String()
				tmp.Reset()
				key = ""
				continue
			}

			tmp.WriteByte(line[i])
			if line[i] == '"' {
				quoted = !quoted
			}
		}
	}
	if key != "" {
		out[key] = tmp.String()
	}
	return out
}

// Return xpos, ypos, zpos, or an error from a line (read from file )if any issue is detected
func getPosition(m map[string]string) (float64, float64, float64, error) {
	var xst, yst, zst string
	var x, y, z float64
	var ok bool
	var err error

	if xst, ok = m["xpos"]; !ok {
		return x, y, z, errors.New("xpos not defined")
	}
	if yst, ok = m["ypos"]; !ok {
		return x, y, z, errors.New("ypos not defined")
	}
	if zst, ok = m["zpos"]; !ok {
		return x, y, z, errors.New("zpos not defined")
	}
	xst = trim(xst)
	yst = trim(yst)
	zst = trim(zst)
	if x, err = strconv.ParseFloat(xst, 64); err != nil {
		return x, y, z, errors.New("invalid xpos")
	}
	if y, err = strconv.ParseFloat(yst, 64); err != nil {
		return x, y, z, errors.New("invalid ypos")
	}
	if z, err = strconv.ParseFloat(zst, 64); err != nil {
		return x, y, z, errors.New("invalid zpos")
	}
	return x, y, z, nil
}

// Read the "mapinfo.txt" file from the map directory
// Returns an error if the file is not present, or does not contain a map id (resulting in no markers being generated)
func ReadMapInfo(path string) (int, string, error) {
	var id *int
	var name *string
	var fname = fmt.Sprintf("%s/%s", path, files.MapInfoFile)

	b, err := os.ReadFile(fname)
	if err != nil {
		return 0, "", err
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		pair := strings.Split(line, "=")
		if len(pair) != 2 {
			log.Printf("[%s] invalid line: %s, skipping", fname, line)
			continue
		}
		if strings.EqualFold("id", pair[0]) {
			iVal, err := strconv.ParseInt(trim(pair[1]), 10, 64)
			if err != nil {
				return 0, "", fmt.Errorf("[%s] Invalid map id: %s", fname, pair[1])
			}
			i := int(iVal)
			id = &i
		} else if strings.EqualFold("name", pair[0]) {
			name = &pair[1]
		}
	}
	if id == nil {
		return 0, "", errors.New("mapid not defined")
	}
	if name == nil {
		log.Printf("map %s name not defined, defaulting", fname)
		return *id, fmt.Sprintf("%d", *id), nil
	}
	return *id, *name, nil
}

// Walks the current Maps directory generating all POI and Trail definitions
func compileMap(categories []categories.Category, path string) (Map, []string, error) {
	id, name, err := ReadMapInfo(path)
	if err != nil {
		return Map{}, nil, err
	}
	out := Map{MapId: id, MapName: name, POIs: []POI{}, Trails: []Trail{}}
	warns := []string{}
	fileList := files.FilesByExtension(path, files.MarkerPoiExtension, files.MarkerTrailExtension)
	for _, item := range fileList {
		if strings.HasSuffix(item, files.MarkerPoiExtension) {
			newPoi, newWarns, err := ReadPOIs(categories, item)
			if err != nil {
				return out, warns, err
			}
			warns = append(warns, newWarns...)
			out.POIs = append(out.POIs, newPoi...)
		} else if strings.HasSuffix(item, files.MarkerTrailExtension) {
			newTrails, newWarns, err := ReadTrails(categories, item)
			if err != nil {
				return out, warns, err
			}
			warns = append(warns, newWarns...)
			out.Trails = append(out.Trails, newTrails...)
		}
	}

	return out, warns, nil
}
