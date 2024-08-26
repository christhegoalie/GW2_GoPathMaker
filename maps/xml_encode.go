package maps

import (
	"fmt"
	"os"
	"strings"
)

func Save(maps []Map, path string) error {
	for _, m := range maps {
		fname := fmt.Sprintf("%s/map%d.xml", path, m.MapId)
		f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString(`<?xml version="1.0" encoding="utf-8"?><overlaydata><pois>`)
		for _, p := range m.POIs {
			f.WriteString(encodePoi(m.MapId, p))
		}
		for _, t := range m.Trails {
			f.WriteString(encodeTrail(m.MapId, t))
		}
		f.WriteString(`</pois></overlaydata>`)
	}
	return nil
}

func encodePoi(mapid int, p POI) string {
	txt := strings.Builder{}
	txt.WriteString(fmt.Sprintf(`<poi type="%s" xpos="%.6f" ypos="%.6f" zpos="%.6f" mapid="%d"`, p.CategoryReference, p.XPos, p.YPos, p.ZPos, mapid))
	for key, val := range p.Keys {
		txt.WriteString(fmt.Sprintf(" %s=%s", key, val))
	}
	txt.WriteString("/>")
	return txt.String()
}
func encodeTrail(mapid int, t Trail) string {
	txt := strings.Builder{}
	txt.WriteString(fmt.Sprintf(`<trail type="%s" trailData="%s" mapid="%d"`, t.CategoryReference, t.TrailDataFile, mapid))
	for key, val := range t.Keys {
		txt.WriteString(fmt.Sprintf(" %s=%s", key, val))
	}
	txt.WriteString("/>")
	return txt.String()
}
