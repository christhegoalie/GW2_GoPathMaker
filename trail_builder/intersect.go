package trailbuilder

import (
	"math"
)

// Function to find orientation of ordered triplet (p, q, r).
// 0 -> p, q, and r are collinear
// 1 -> Clockwise
// 2 -> Counterclockwise
func orientation(p, q, r Point) int {
	val := (q.Z-p.Z)*(r.X-q.X) - (q.X-p.X)*(r.Z-q.Z)
	if val == 0 {
		return 0 // Collinear
	}
	if val > 0 {
		return 1 // Clockwise
	}
	return 2 // Counterclockwise
}

// Function to check if point q lies on segment pr
func onSegment(p, q, r Point) bool {
	if q.X <= math.Max(p.X, r.X) && q.X >= math.Min(p.X, r.X) &&
		q.Z <= math.Max(p.Z, r.Z) && q.Z >= math.Min(p.Z, r.Z) {
		return true
	}
	return false
}

// Function to check if two line segments intersect
func doIntersect(p1, q1, p2, q2 Point) bool {
	// Find the four orientations needed for general and special cases
	o1 := orientation(p1, q1, p2)
	o2 := orientation(p1, q1, q2)
	o3 := orientation(p2, q2, p1)
	o4 := orientation(p2, q2, q1)

	// General case
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Special cases
	// p1, q1 and p2 are collinear and p2 lies on segment p1q1
	if o1 == 0 && onSegment(p1, p2, q1) {
		return true
	}

	// p1, q1 and p2 are collinear and q2 lies on segment p1q1
	if o2 == 0 && onSegment(p1, q2, q1) {
		return true
	}

	// p2, q2 and p1 are collinear and p1 lies on segment p2q2
	if o3 == 0 && onSegment(p2, p1, q2) {
		return true
	}

	// p2, q2 and q1 are collinear and q1 lies on segment p2q2
	if o4 == 0 && onSegment(p2, q1, q2) {
		return true
	}

	// If none of the cases
	return false
}
