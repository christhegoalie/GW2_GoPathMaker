package trailbuilder

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func readPoints(filePath string) []Point {
	out := []Point{}
	data, err := os.ReadFile(filePath)
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
		var allowDupe bool
		if allowDupeSt, ok := mapString(vals, "AllowDuplicate"); ok {
			allowDupe = allowDupeSt == "1" || strings.EqualFold(allowDupeSt, "true") || strings.EqualFold(allowDupeSt, "yes")
		}
		out = append(out, Point{X: x, Y: y, Z: z, AllowDuplicate: allowDupe})
	}
	return out
}

func readTypedGroup(filePath string) map[string]typedGroup {
	out := make(map[string]typedGroup)
	data, err := os.ReadFile(filePath)
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
		pt := Point{X: x, Y: y, Z: z}
		if name, ok := mapString(vals, "name"); ok {
			if v, ok := out[name]; ok {
				v.addPoint(pt)
				out[name] = v
			} else {
				v := typedGroup{
					name:         name,
					reverseName:  "Reversed " + name,
					_points:      []Point{pt},
					Type:         Type_Unknown,
					_distance:    0,
					_revDistance: 0,
				}
				if typeString, ok := mapString(vals, "type"); ok {
					switch strings.ToLower(typeString) {
					case "downonly":
						v.Type = BT_DownOnly
					case "wall":
						v.Type = BT_Wall
					case "mushroom":
						v.Type = GT_Mushroom
					case "oneway":
						v.Type = GT_ONEWAY
					}
				}
				out[name] = v
			}
		} else {
			log.Println("Line missing 'name' field: ")
			continue
		}
	}
	return out
}

func getPosition(m map[string]any) (float64, float64, float64, error) {
	var xst, yst, zst string
	var x, y, z float64
	var ok bool
	var err error

	if xst, ok = mapString(m, "xpos"); !ok {
		return x, y, z, errors.New("xpos not defined")
	}
	if yst, ok = mapString(m, "ypos"); !ok {
		return x, y, z, errors.New("ypos not defined")
	}
	if zst, ok = mapString(m, "zpos"); !ok {
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

func fileChangedSince(timestamp time.Time, filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return true
	}
	return info.ModTime().After(timestamp)
}
