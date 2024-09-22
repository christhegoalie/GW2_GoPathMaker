package location

func SetGlobals(barriers map[string]TypedGroup, paths map[string]TypedGroup, waypoints []Point, ptpPaths map[string]TypedGroup) {
	if GLOBAL_Barriers != nil || GLOBAL_Paths != nil || GLOBAL_Waypoints != nil {
		panic("threading unsupported")
	}
	GLOBAL_Barriers = barriers
	GLOBAL_Paths = paths
	GLOBAL_Waypoints = waypoints
	GLOBAL_PtpPaths = ptpPaths
}

func ResetGlobals() {
	GLOBAL_Barriers = nil
	GLOBAL_Paths = nil
	GLOBAL_Waypoints = nil
	GLOBAL_PtpPaths = nil
}
