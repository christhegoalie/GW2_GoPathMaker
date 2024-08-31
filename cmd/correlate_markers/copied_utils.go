package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
)

func trim(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(s), `"`), `"`)
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

// Read a space seperated line of key value pairs seperated by "=", and return a map
func readMap(line string, delim byte) map[string]string {
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
			if !quoted && line[i] == delim {
				needEqual = true
				out[key] = trim(tmp.String())
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
		out[key] = trim(tmp.String())
	}
	return out
}

func readPoints(filePath string) []point {
	out := []point{}
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
		out = append(out, point{X: x, Y: y, Z: z})
	}
	return out
}
