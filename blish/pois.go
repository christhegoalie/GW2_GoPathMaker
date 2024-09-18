package blish

import (
	"encoding/xml"
	"gw2_markers_gen/location"
)

type XMLPoiData struct {
	XMLName xml.Name `xml:"overlaydata"`
	Pois    Pois     `xml:"pois"`
}

type Pois struct {
	Poi []Poi `xml:"poi"`
}

type Poi struct {
	Type     string  `xml:"type,attr"`
	MapID    int     `xml:"mapid,attr"`
	XPos     float64 `xml:"xpos,attr"`
	YPos     float64 `xml:"ypos,attr"`
	ZPos     float64 `xml:"zpos,attr"`
	GUID     string  `xml:"guid,attr"`
	Behavior int     `xml:"behavior,attr"`
}

type PoiList []Poi

func (p Poi) Point() location.Point {
	return location.Point{
		X: p.XPos,
		Y: p.YPos,
		Z: p.ZPos,
	}
}

func (ls PoiList) Contains(point location.Point) bool {
	for _, p := range ls {
		if p.Point().Same(point) {
			return true
		}
	}
	return false
}
