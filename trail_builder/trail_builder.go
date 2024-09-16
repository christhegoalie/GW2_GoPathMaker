package trailbuilder

import (
	"encoding/binary"
	"errors"
	"fmt"
	"gw2_markers_gen/maps"
	"log"
	"math"
	"strconv"
	"strings"
)

type objectType int

const (
	Type_Unknown objectType = iota
	BT_Wall
	BT_DownOnly

	GT_Mushroom
	GT_ONEWAY
)

type typedGroup struct {
	name         string
	reverseName  string
	_points      []Point
	Type         objectType
	_distance    float64
	_revDistance float64
}
type Point struct {
	X, Y, Z        float64
	AllowDuplicate bool
}

func trim(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(s), `"`), `"`)
}

func setVal(txt string, v *float64) error {
	flt, err := strconv.ParseFloat(trim(txt), 64)
	if err != nil {
		return err
	}
	*v = flt
	return nil
}
func lineToTriple(line string) (Point, error) {
	set := []bool{false, false, false}
	out := Point{}
	ls := strings.Split(line, " ")
	for _, line := range ls {
		pair := strings.Split(line, "=")
		if len(pair) != 2 {
			continue
		}
		if pair[0] == "xpos" {
			if err := setVal(pair[1], &out.X); err != nil {
				return out, err
			}
			set[0] = true
		} else if pair[0] == "ypos" {
			if err := setVal(pair[1], &out.Y); err != nil {
				return out, err
			}
			set[1] = true
		} else if pair[0] == "zpos" {
			if err := setVal(pair[1], &out.Z); err != nil {
				return out, err
			}
			set[2] = true
		}
	}

	if !set[0] {
		return out, errors.New("no xpos found")
	} else if !set[1] {
		return out, errors.New("no ypos found")
	} else if !set[2] {
		return out, errors.New("no zpos found")
	}
	return out, nil
}

func LinesToTRLBytes(lines []string) ([]byte, error) {
	skipped := 0
	if len(lines) == 0 {
		return []byte{}, errors.New("invalid file, no mapid")
	}

	var mapId int64
	var ok bool
	var err error
	var mapVal any
	var mapIdStr string

	m := readMap(strings.TrimSpace(lines[0]), ' ')
	if mapVal, ok = m["mapid"]; ok {
		if mapIdStr, ok = mapVal.(string); !ok {
			return []byte{}, errors.New("dupplicate mapid fields")
		}
		mapId, err = strconv.ParseInt(trim(mapIdStr), 10, 32)
		if err != nil {
			return []byte{}, err
		}
	}

	out := make([]byte, 8+12*(len(lines)-1))
	offset := 8
	binary.LittleEndian.PutUint32(out[4:], uint32(mapId))
	for i := 1; i < len(lines); i++ {
		pt, err := lineToTriple(strings.TrimSpace(lines[i]))
		if err != nil {
			skipped += 12
			log.Printf("error on line %d: %s", i, err.Error())
			continue
		}
		binary.LittleEndian.PutUint32(out[offset:], math.Float32bits(float32(pt.X)))
		binary.LittleEndian.PutUint32(out[offset+4:], math.Float32bits(float32(pt.Y)))
		binary.LittleEndian.PutUint32(out[offset+8:], math.Float32bits(float32(pt.Z)))
		offset += 12
	}
	if skipped != 0 {
		return out[:len(out)-skipped], nil
	}
	return out, nil
}

func PointsToTrlBytes(mapId int, points []Point) ([]byte, error) {
	out := make([]byte, 8+12*(len(points)))
	offset := 8
	binary.LittleEndian.PutUint32(out[4:], uint32(mapId))
	for _, p := range points {
		binary.LittleEndian.PutUint32(out[offset:], math.Float32bits(float32(p.X)))
		binary.LittleEndian.PutUint32(out[offset+4:], math.Float32bits(float32(p.Y)))
		binary.LittleEndian.PutUint32(out[offset+8:], math.Float32bits(float32(p.Z)))
		offset += 12
	}
	return out, nil
}

func TRLBytesToLines(bytes []byte) ([]string, error) {
	out := make([]string, 0)
	if len(bytes) < 8 {
		return out, errors.New("mapid header not found")
	}
	if (len(bytes)-8)%12 != 0 {
		return out, errors.New("invalid tlr file")
	}
	mapid := binary.LittleEndian.Uint32(bytes[4:])
	out = append(out, fmt.Sprintf("mapid=%d", mapid))
	for i := 8; i < len(bytes); i += 12 {
		p := Point{
			X: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i:]))),
			Y: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i+4:]))),
			Z: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i+8:]))),
		}
		out = append(out, fmt.Sprintf(`xpos="%.6f" ypos="%.6f" zpos="%.6f"`, p.X, p.Y, p.Z))
	}

	return out, nil
}
func TRLBytesToPOIs(category string, bytes []byte) (int, []maps.POI, error) {
	out := make([]maps.POI, 0)
	if len(bytes) < 8 {
		return 0, out, errors.New("mapid header not found")
	}
	if (len(bytes)-8)%12 != 0 {
		return 0, out, errors.New("invalid tlr file")
	}
	mapid := binary.LittleEndian.Uint32(bytes[4:])
	for i := 8; i < len(bytes); i += 12 {
		p := Point{
			X: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i:]))),
			Y: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i+4:]))),
			Z: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i+8:]))),
		}
		out = append(out, maps.POI{CategoryReference: category, XPos: p.X, YPos: p.Y, ZPos: p.Z})
	}

	return int(mapid), out, nil
}

func addUpdateKey(m map[string]any, key string, val string) {
	if old, ok := m[key]; ok {
		var arr []string
		switch v := old.(type) {
		case []string:
			arr = append(v, trim(val))
		case string:
			arr = []string{v, trim(val)}
		default:
			panic("invalid data type")
		}
		m[key] = arr
	} else {
		m[key] = trim(val)
	}
}

// Read a space seperated line of key value pairs seperated by "=", and return a map
func readMap(line string, delim byte) map[string]any {
	out := make(map[string]any)
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
			if !quoted && line[i] == delim {
				needEqual = true
				addUpdateKey(out, key, tmp.String())
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
		addUpdateKey(out, key, tmp.String())
	}
	return out
}

func mapString(src map[string]any, key string) (string, bool) {
	if val, ok := src[key]; ok {
		if v, ok := val.(string); ok {
			return v, true
		}
	}
	return "", false
}
func mapStringArray(src map[string]any, key string) ([]string, bool) {
	if val, ok := src[key]; ok {
		switch v := val.(type) {
		case string:
			return []string{v}, true
		case []string:
			return v, true
		}
	}
	return []string{}, false
}
