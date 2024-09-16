package maps

import (
	"errors"
	"fmt"
	"strings"
)

// Convert a line of trail information into a trail object
func parseTrail(category string, line string) (Trail, []string, error) {
	warns := []string{}
	var traildata string
	var ok bool
	m := readMap(line)
	if traildata, ok = m["trailData"]; !ok {
		return Trail{}, warns, errors.New("traildata not defined")
	}
	delete(m, "trailData")
	traildata = trim(traildata)
	if validateFileExists != nil {
		if warn := validateFileExists(traildata); warn != "" {
			warns = append(warns, warn)
		}
	}
	if cat, ok := m["category"]; ok {
		category = trim(cat)
		delete(m, "category")
	}

	return Trail{
		CategoryReference: category,
		TrailDataFile:     traildata,
		Keys:              m,
	}, warns, nil
}

// Convert a line of poi information into a POI object
func parsePoi(category string, line string) (POI, []string, error) {
	warns := []string{}
	m := readMap(line)
	x, y, z, err := getPosition(m)
	if err != nil {
		return POI{}, warns, fmt.Errorf("error in line: %s, error: %s", line, err.Error())
	}
	delete(m, "xpos")
	delete(m, "ypos")
	delete(m, "zpos")
	if cat, ok := m["category"]; ok {
		category = trim(cat)
		delete(m, "category")
	}
	var allowDupe bool
	if allowDupeSt, ok := m["AllowDuplicate"]; ok {
		allowDupe = allowDupeSt == "1" || strings.EqualFold(allowDupeSt, "true") || strings.EqualFold(allowDupeSt, "yes")
		delete(m, "AllowDuplicate")
	}

	return POI{
		CategoryReference: category,
		XPos:              x,
		YPos:              y,
		ZPos:              z,
		AllowDuplicate:    allowDupe,
		Keys:              m,
	}, warns, nil
}
