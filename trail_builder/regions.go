package trailbuilder

import "fmt"

type Region struct {
	Start    Point   `json:"start"`
	End      Point   `json:"end"`
	Vertices []Point `json:"vertices"`
}
type ZoneTrail struct {
	Map     string   `json:"map"`
	File    string   `json:"file"`
	Regions []Region `json:"regions"`
}

type PointRegion struct {
	Start  *Point
	End    *Point
	Points []Point
	graph  *graph
}

func (trail ZoneTrail) PartitionPoints(input []Point) []PointRegion {
	out := make([]PointRegion, len(trail.Regions))
	for i := range out {
		out[i].Points = make([]Point, 0)
	}
pointLoop:
	for _, pt := range input {
		for i, r := range trail.Regions {
			if r.Contains(pt) {
				out[i].Points = append(out[i].Points, pt)
				continue pointLoop
			}
		}
		panic(fmt.Sprintf("Point: %+v not in a specified region", pt))
	}

	return out
}

// IsPointInPolygon checks if a point is inside a polygon using the Ray Casting algorithm
func (region Region) Contains(point Point) bool {
	n := len(region.Vertices)
	if n < 3 {
		return false
	}

	var inside bool
	for i := 0; i < n; i++ {
		current := region.Vertices[i]
		next := region.Vertices[(i+1)%n]

		if IntersectsEdge(point, current, next) {
			return !inside
		}
	}
	return inside
}

// isRayIntersectingEdge checks if a horizontal ray from the point intersects with an edge of the polygon
func IntersectsEdge(point, vertex1, vertex2 Point) bool {
	// Ensure vertex1 is below vertex2 for easier calculations
	if vertex1.Z > vertex2.Z {
		vertex1, vertex2 = vertex2, vertex1
	}

	if point.Z < vertex1.Z || point.Z > vertex2.Z || point.X > max(vertex1.X, vertex2.X) {
		return false
	}

	// Check if the point is to the left of the edge or on it
	if point.X < min(vertex1.X, vertex2.X) {
		return true
	}

	// Calculate the intersection point of the ray with the edge
	edgeSlope := (vertex2.X - vertex1.X) / (vertex2.Z - vertex1.Z)
	rayIntersectionX := vertex1.X + (point.Z-vertex1.Z)*edgeSlope

	// The ray intersects if the intersection X is greater than the point's X
	return point.X < rayIntersectionX
}
