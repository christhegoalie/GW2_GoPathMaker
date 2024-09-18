package maps

import (
	"errors"
	"fmt"
	"gw2_markers_gen/location"
	"gw2_markers_gen/utils"
	"strings"
)

// Convert a line of trail information into a trail object
func parseTrail(category string, line string) (Trail, []string, error) {
	warns := []string{}
	var traildata string
	var ok bool
	m := utils.ReadMap(line, ' ')
	if traildata, ok = utils.MapString(m, "trailData"); !ok {
		return Trail{}, warns, errors.New("traildata not defined")
	}
	delete(m, "trailData")
	traildata = utils.Trim(traildata)
	if validateFileExists != nil {
		if warn := validateFileExists(traildata); warn != "" {
			warns = append(warns, warn)
		}
	}
	if cat, ok := utils.MapString(m, "category"); ok {
		category = utils.Trim(cat)
		delete(m, "category")
	}

	return Trail{
		CategoryReference: category,
		TrailDataFile:     traildata,
		Keys:              utils.ToStringMap(m),
	}, warns, nil
}

// Convert a line of poi information into a POI object
func parsePoi(category string, line string) (POI, []string, error) {
	warns := []string{}
	m := utils.ReadMap(line, ' ')
	x, y, z, err := location.GetPosition(m)
	if err != nil {
		return POI{}, warns, fmt.Errorf("error in line: %s, error: %s", line, err.Error())
	}
	delete(m, "xpos")
	delete(m, "ypos")
	delete(m, "zpos")
	if cat, ok := utils.MapString(m, "category"); ok {
		category = utils.Trim(cat)
		delete(m, "category")
	}
	var allowDupe bool
	if allowDupeSt, ok := utils.MapString(m, "AllowDuplicate"); ok {
		allowDupeSt = utils.Trim(allowDupeSt)
		allowDupe = allowDupeSt == "1" || strings.EqualFold(allowDupeSt, "true") || strings.EqualFold(allowDupeSt, "yes")
		delete(m, "AllowDuplicate")
	}

	return POI{
		CategoryReference: category,
		XPos:              x,
		YPos:              y,
		ZPos:              z,
		AllowDuplicate:    allowDupe,
		Keys:              utils.ToStringMap(m),
	}, warns, nil
}
