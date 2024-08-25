package maps

type POI struct {
	CategoryReference string
	XPos, YPos, ZPos  float64
	Keys              map[string]string
}
