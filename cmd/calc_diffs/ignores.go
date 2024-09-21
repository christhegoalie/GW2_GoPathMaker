package main

import (
	"gw2_markers_gen/blish"
)

func init() {
	ignore = func() blish.PoiList {
		return blish.PoiList{
			// Other
			{XPos: 362.8005, YPos: 133.9013, ZPos: 400.1024},
			{XPos: 615.403, YPos: 318.6896, ZPos: -130.094},
			{XPos: 641.5231, YPos: 376.7664, ZPos: -534.1235},

			// Mine
			{XPos: 364.7122, YPos: 133.2208, ZPos: 408.3469},
			{XPos: 607.4794, YPos: 318.1981, ZPos: -126.1629},
			{XPos: 644.8059, YPos: 375.4574, ZPos: -530.1046},
		}
	}
}
