package files

import (
	"encoding/xml"
	"fmt"
	"gw2_markers_gen/blish"
	"gw2_markers_gen/location"
	"gw2_markers_gen/utils"
	"log"
	"os"
	"strings"
)

func ReadPoints(filePath string) []location.Point {
	out := []location.Point{}
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}
	for i, s := range strings.Split(string(data), "\n") {
		s = utils.Trim(s)
		if s == "" {
			continue
		}
		vals := utils.ReadMap(s, ' ')
		x, y, z, e := location.GetPosition(vals)
		if e != nil {
			if i > 0 {
				log.Printf("[%s] Unknown line: ", filePath, e.Error())
			}
			continue
		}
		var allowDupe bool
		if allowDupeSt, ok := utils.MapString(vals, "AllowDuplicate"); ok {
			allowDupe = allowDupeSt == "1" || strings.EqualFold(allowDupeSt, "true") || strings.EqualFold(allowDupeSt, "yes")
		}
		out = append(out, location.Point{X: x, Y: y, Z: z, AllowDuplicate: allowDupe})
	}
	return out
}

func ReadPoiPoints(filePath string) []blish.Poi {
	out := []blish.Poi{}
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}
	lines := strings.Split(string(data), "\n")

	pair := strings.Split(lines[0], "=")
	if len(pair) != 2 {
		panic("missing category")
	}

	if !strings.EqualFold("category", pair[0]) {
		panic("invalid category")
	}

	category := pair[1]
	for i, s := range lines {
		if i == 0 {
			continue
		}
		s = utils.Trim(s)
		if s == "" {
			continue
		}
		vals := utils.ReadMap(s, ' ')
		x, y, z, e := location.GetPosition(vals)
		if e != nil {
			if i > 0 {
				log.Printf("[%s] Unknown line: ", filePath, e.Error())
			}
			continue
		}
		var tmpCat any
		var ok bool
		if tmpCat, ok = vals["category"]; !ok {
			tmpCat = category
		}
		out = append(out, blish.Poi{
			XPos: x,
			YPos: y,
			ZPos: z,
			Type: tmpCat.(string),
		})
	}
	return out
}
func ReadXMLPoints(filePath string) []blish.Poi {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}
	var pois blish.XMLPoiData
	err = xml.Unmarshal(data, &pois)
	if err != nil {
		panic(err)
	}
	return pois.Pois.Poi
}

func hasPrefix(in []byte, prefix []byte) bool {
	if len(prefix) > len(in) {
		return false
	}
	for i := range prefix {
		if prefix[i] != in[i] {
			return false
		}
	}
	return true
}
func firstIndex(in []byte, expected []byte) int {
	var i int
	for len(in[i:]) >= len(expected) {
		if hasPrefix(in[i:], expected) {
			return i
		}
		i++
	}
	return -1
}
func nextPtp(in []byte) ([]byte, []byte, bool) {
	beginBytes := []byte("Begin")
	endBytes := []byte("End")
	beginIndex := firstIndex(in, beginBytes)
	if beginIndex < 0 {
		return []byte{}, []byte{}, false
	}
	endIndex := firstIndex(in[beginIndex:], endBytes)
	if endIndex < 0 {
		return []byte{}, []byte{}, false
	}
	endIndex += beginIndex
	return in[beginIndex+len(beginBytes) : endIndex], in[endIndex+len(endBytes):], true
}
func ReadPTPPoints(filePath string) map[string]location.TypedGroup {
	out := make(map[string]location.TypedGroup)

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}

	i := 0
	for {
		i++
		key := fmt.Sprintf("%d", i)
		path := location.NewEmptyGroup(key, location.Type_Unknown)
		pathBytes, more, ok := nextPtp(data)
		if !ok {
			break
		}
		data = more

		for i, s := range strings.Split(string(pathBytes), "\n") {
			s = utils.Trim(s)
			if s == "" {
				continue
			}
			vals := utils.ReadMap(s, ' ')
			x, y, z, e := location.GetPosition(vals)
			if e != nil {
				if i > 0 {
					log.Printf("[%s] Unknown line: ", filePath, e.Error())
				}
				continue
			}

			p := location.Point{X: x, Y: y, Z: z, AllowDuplicate: false, Type: location.TypeFromMap(vals)}
			path.AddPoint(p)
		}

		if len(path.Points()) > 0 {
			out[key] = path
		}

	}
	return out
}
func ReadTypedGroup(filePath string) map[string]location.TypedGroup {
	out := make(map[string]location.TypedGroup)
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}

	lines := strings.Split(string(data), "\n")
	for _, s := range lines {
		s = utils.Trim(s)
		if s == "" {
			continue
		}
		vals := utils.ReadMap(s, ' ')
		x, y, z, e := location.GetPosition(vals)
		if e != nil {
			log.Printf("[%s] Unknown line: ", filePath, e.Error())
			continue
		}
		pt := location.Point{X: x, Y: y, Z: z, Type: location.TypeFromMap(vals)}
		if name, ok := utils.MapString(vals, "name"); ok {
			if v, ok := out[name]; ok {
				v.AddPoint(pt)
				out[name] = v
			} else {
				v := location.NewGroup(name, pt)
				out[name] = v
			}
		} else {
			log.Println("Line missing 'name' field: ")
			continue
		}
	}
	return out
}

func ReadAllPoints(path string) []blish.Poi {
	out := []blish.Poi{}
	txtFiles := FilesByExtension(path, ".poi")
	for _, f := range txtFiles {
		out = append(out, ReadPoiPoints(f)...)
	}

	xmlFiles := FilesByExtension(path, ".xml")
	for _, f := range xmlFiles {
		out = append(out, ReadXMLPoints(f)...)
	}
	return out
}
