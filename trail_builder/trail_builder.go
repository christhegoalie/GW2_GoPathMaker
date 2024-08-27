package trailbuilder

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type point struct {
	x, y, z float64
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
func lineToTriple(line string) (point, error) {
	set := []bool{false, false, false}
	out := point{}
	ls := strings.Split(line, " ")
	for _, line := range ls {
		pair := strings.Split(line, "=")
		if len(pair) != 2 {
			continue
		}
		if pair[0] == "xpos" {
			if err := setVal(pair[1], &out.x); err != nil {
				return out, err
			}
			set[0] = true
		} else if pair[0] == "ypos" {
			if err := setVal(pair[1], &out.y); err != nil {
				return out, err
			}
			set[1] = true
		} else if pair[0] == "zpos" {
			if err := setVal(pair[1], &out.z); err != nil {
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

	if len(lines) == 0 {
		return []byte{}, errors.New("invalid file, no mapid")
	}
	mapId, err := strconv.ParseInt(lines[0], 10, 32)
	if err != nil {
		return []byte{}, err
	}

	out := make([]byte, 8+12*(len(lines)-1))
	offset := 8
	binary.LittleEndian.PutUint32(out[4:], uint32(mapId))
	for i := 1; i < len(lines); i++ {
		pt, err := lineToTriple(lines[i])
		if err != nil {
			return out, fmt.Errorf("error on line %d: %w", i, err)
		}
		binary.LittleEndian.PutUint32(out[offset:], math.Float32bits(float32(pt.x)))
		binary.LittleEndian.PutUint32(out[offset+4:], math.Float32bits(float32(pt.y)))
		binary.LittleEndian.PutUint32(out[offset+8:], math.Float32bits(float32(pt.z)))
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
	out = append(out, fmt.Sprintf("%d", mapid))
	for i := 8; i < len(bytes); i += 12 {
		p := point{
			x: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i:]))),
			y: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i+4:]))),
			z: float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i+8:]))),
		}
		out = append(out, fmt.Sprintf(`xpos="%.6f" ypos="%.6f" zpos="%.6f"`, p.x, p.y, p.z))
	}

	return out, nil
}
