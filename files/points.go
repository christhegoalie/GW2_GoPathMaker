package files

import (
	"encoding/xml"
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
				log.Println("Unknown line: ", e.Error())
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
				log.Println("Unknown line: ", e.Error())
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
			log.Println("Unknown line: ", e.Error())
			continue
		}
		pt := location.Point{X: x, Y: y, Z: z}
		if name, ok := utils.MapString(vals, "name"); ok {
			if v, ok := out[name]; ok {
				v.AddPoint(pt)
				out[name] = v
			} else {
				v := location.NewGroup(name, pt, location.Type_Unknown)
				if typeString, ok := utils.MapString(vals, "type"); ok {
					switch strings.ToLower(typeString) {
					case "downonly":
						v.Type = location.BT_DownOnly
					case "wall":
						v.Type = location.BT_Wall
					case "mushroom":
						v.Type = location.GT_Mushroom
					case "oneway":
						v.Type = location.GT_ONEWAY
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
