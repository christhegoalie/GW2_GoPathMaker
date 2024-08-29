package trailbuilder

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const MapsDir = "maps"
const AssetsDir = "assets"
const CompiledAssetsDir = "compiled_assets"
const CompiledTrailExtension = ".rtrl"
const CompiledAutoTrailExtension = ".atrl"
const TrailExtension = ".trl"

// Recursively find all .trail and .poi files
func readFiles(path string, extension string) []string {
	items, _ := os.ReadDir(path)
	files := []string{}
	for _, item := range items {
		fullPath := fmt.Sprintf("%s/%s", path, item.Name())
		if item.IsDir() {
			files = append(files, readFiles(fullPath, extension)...)
		} else if strings.HasSuffix(item.Name(), extension) {
			files = append(files, fullPath)
		}
	}
	return files
}

func readPoints(fpath string, fname string) []point {
	fname = fmt.Sprintf("%s/%s", fpath, fname)
	out := []point{}
	data, err := os.ReadFile(fname)
	if err != nil {
		log.Println(err)
	}
	for i, s := range strings.Split(string(data), "\n") {
		s = trim(s)
		if s == "" {
			continue
		}
		vals := readMap(s, ' ')
		x, y, z, e := getPosition(vals)
		if e != nil {
			if i > 0 {
				log.Println("Unknown line: ", e.Error())
			}
			continue
		}
		out = append(out, point{x: x, y: y, z: z})
	}
	return out
}
func readPointGroups(fpath string, fname string) map[string]pointGroup {
	fname = fmt.Sprintf("%s/%s", fpath, fname)
	out := make(map[string]pointGroup)
	data, err := os.ReadFile(fname)
	if err != nil {
		log.Println(err)
	}

	lines := strings.Split(string(data), "\n")
	for _, s := range lines {
		s = trim(s)
		if s == "" {
			continue
		}
		vals := readMap(s, ' ')
		x, y, z, e := getPosition(vals)
		if e != nil {
			log.Println("Unknown line: ", e.Error())
			continue
		}
		pt := point{x: x, y: y, z: z}

		if name, ok := vals["name"]; ok {
			if v, ok := out[name]; ok {
				v = append(v, pt)
				out[name] = v
			} else {
				out[name] = []point{pt}
			}
		} else {
			log.Println("Line missing 'name' field: ")
			continue
		}
	}
	return out
}

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
