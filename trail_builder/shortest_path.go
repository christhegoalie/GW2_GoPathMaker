package trailbuilder

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"slices"
)

type graphPath struct {
	node *graphNode
	next *graphPath
}
type edge struct {
	shortcuts []typedGroup //optional non-direct path to reach node
	cost      float64
	dest      *graphNode
}
type graphNode struct {
	location point
	required bool
	edges    []edge
}
type graph struct {
	nodes     []*graphNode
	waypoints []*graphNode
}
type path []point

var glBarriers map[string]typedGroup
var glPaths map[string]typedGroup

var glWaypoints path

// Arbitrarily high value, but we need to be able to compute the "better" of paths crossing multiple barriers
const BarrierValue = 1e7
const waypointCost = 1000
const enableWaypointing = false
const mushroomCost = 10

func SaveShortestTrail(mapid int, waypoints []point, pois []point, barriers map[string]typedGroup, paths map[string]typedGroup, baseFileName string, extension string) error {
	if glBarriers != nil || glPaths != nil {
		panic("threading unsupported")
	}
	glBarriers = barriers
	glPaths = paths
	glWaypoints = waypoints
	defer func() {
		glBarriers = nil
		glPaths = nil
	}()

	g := path(pois).toGraph()
	g.addWaypoints(waypoints)
	pathList := g.getPaths()

	for _, p := range pathList {
		for p.optimize() {
		}
	}

	final, _ := pathList.shortest()
	points := final.toPath()

	b, err := PointsToTrlBytes(mapid, points)
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s%s", baseFileName, extension)
	err = os.WriteFile(fileName, b, fs.ModePerm)
	if err != nil {
		return err
	}

	i := 0
	for _, b := range glBarriers {
		i++
		b, err := PointsToTrlBytes(mapid, b._points)
		if err == nil {
			fileName := fmt.Sprintf("%s_barrier_%d%s", baseFileName, i, extension)
			os.WriteFile(fileName, b, fs.ModePerm)
		}
	}
	return nil
}
func (l *graphPath) optimize() bool {
	return false

}

func (p path) Distance(allowWaypoints bool, bypassBarriers bool) float64 {
	var out float64
	for i := 0; i < len(p)-1; i++ {
		out += p[i].Distance(p[i+1], allowWaypoints, bypassBarriers)
	}
	return out
}

func (p path) trySwap3p(i, j int, bypasBarriers bool) bool {

	//Note: non-directed graph, so no need to compute parts of the path that don't change
	first := i
	second := j

	var delta float64

	//Remove segments
	for r := i - 1; r < j+1; r++ {
		if r+1 < len(p) {
			delta -= p[r].Distance(p[r+1], true, bypasBarriers)
		}
	}

	delta += p[i-1].Distance(p[j], true, bypasBarriers)
	if j+1 < len(p) {
		delta += p[first].Distance(p[j+1], true, bypasBarriers)
	}
	for r := j; r > i-1; r-- {
		delta += p[r].Distance(p[r-1], false, bypasBarriers) //don't allow waypoints when following paths
	}

	if delta < 0 {
		for i, j := first, second; i < j; i, j = i+1, j-1 {
			p[i], p[j] = p[j], p[i]
		}
		return true
	}
	return false
}
func (p path) trySwap2p(i, j int, bypasBarriers bool) bool {
	if len(p) < 2 || i >= len(p) || j >= len(p) {
		return false
	}

	var oldSeg path
	var newSeg path
	if j+1 < len(p) {
		oldSeg = p[i-1 : j+2]
		newSeg = make(path, len(oldSeg))
		copy(newSeg, oldSeg)
		newSeg[1] = oldSeg[len(oldSeg)-2]
		newSeg[len(newSeg)-2] = oldSeg[1]
	} else {
		oldSeg = p[i-1 : j+1]
		newSeg = make(path, len(oldSeg))
		copy(newSeg, oldSeg)
		newSeg[1] = oldSeg[len(oldSeg)-1]
		newSeg[len(newSeg)-1] = oldSeg[1]
	}

	newDist := newSeg.Distance(true, bypasBarriers)
	oldDist := oldSeg.Distance(true, bypasBarriers)
	if newDist < oldDist {
		p[i], p[j] = p[j], p[i]
		return true
	}
	return false
}

func findPathDistance(start point, path []typedGroup, dest point) float64 {
	if len(path) == 0 {
		return BarrierValue //indicates no paths
	}
	pathLen := start.CalcDistance(path[0].first()) + path[len(path)-1].last().CalcDistance(dest)
	for _, p := range path {
		pathLen += p.distance()
	}
	return pathLen
}
func (node1 *graphNode) connect(node2 *graphNode) {
	const MAX_PATH_LENGTH = 10000

	// find any possible paths to the node
	toPath, _ := node1.location.findPath(node2.location)
	fromPath, _ := node2.location.findPath(node1.location)
	toDistance := findPathDistance(node1.location, toPath, node2.location)
	fromDistance := findPathDistance(node2.location, fromPath, node1.location)
	toDirectDistance := node1.location.Distance(node2.location, false, false)
	fromDirectDistance := node2.location.Distance(node1.location, false, false)

	if toDirectDistance < toDistance {
		if toDirectDistance < MAX_PATH_LENGTH {
			node1.edges = append(node1.edges,
				edge{
					dest: node2,
					cost: toDirectDistance,
				})
		}
	} else {
		if toDistance < MAX_PATH_LENGTH {
			node1.edges = append(node1.edges,
				edge{
					dest:      node2,
					cost:      toDistance,
					shortcuts: toPath,
				})
		}
	}

	//Node 2 to node 1
	if fromDirectDistance < fromDistance {
		if fromDirectDistance < MAX_PATH_LENGTH {
			node2.edges = append(node2.edges,
				edge{
					dest: node1,
					cost: fromDirectDistance,
				})
		}
	} else {
		if fromDistance < MAX_PATH_LENGTH {
			node2.edges = append(node2.edges,
				edge{
					dest:      node1,
					cost:      fromDistance,
					shortcuts: fromPath,
				})
		}
	}
}
func (g *graph) add(pt point, required bool) *graphNode {
	node := graphNode{
		location: pt,
		edges:    make([]edge, 0),
		required: required,
	}
	for i, graphNode := range g.nodes {
		graphNode.connect(&node)
		g.nodes[i] = graphNode
	}
	g.nodes = append(g.nodes, &node)
	return &node
}

func (g graph) requiredNodes() []*graphNode {
	out := []*graphNode{}
	for _, n := range g.nodes {
		if n.required {
			out = append(out, n)
		}
	}
	return out
}

func contains(ls []*graphNode, node *graphNode) bool {
	for _, i := range ls {
		if node == i {
			return true
		}
	}
	return false
}
func remove(ls []*graphNode, node *graphNode) []*graphNode {
	for i, _node := range ls {
		if node == _node {
			ls[i] = ls[len(ls)-1]
			ls = ls[:len(ls)-1]
			return ls
		}
	}
	panic("remove item doesn't exist")
}
func (n *graphNode) closest(required []*graphNode) *graphNode {
	var currentCost float64
	var out *graphNode
	for _, edge := range n.edges {
		if !contains(required, edge.dest) {
			continue
		} else if out == nil {
			out = edge.dest
			currentCost = edge.cost
		} else {
			if edge.cost < currentCost {
				currentCost = edge.cost
				out = edge.dest
			}
		}
	}

	//Catch for valid edges to finish the path..let the otimizer fix things
	if out == nil {
		out = required[0]
	}
	return out
}

type graphPathList []*graphPath

func (p *graphNode) findEdge(dst *graphNode) (edge, error) {
	for _, edge := range p.edges {
		if edge.dest == dst {
			return edge, nil
		}
	}
	return edge{}, errors.New("expected edge")
}
func (p *graphPath) toPath() path {
	out := path{p.node.location}
	for p.next != nil {
		edge, err := p.node.findEdge(p.next.node)
		if err != nil {
			return out
		}
		for _, st := range edge.shortcuts {
			out = append(out, st._points...)
		}
		out = append(out, p.next.node.location)
		p = p.next
	}
	return out
}
func (p *graphPath) distanceTo(n *graphNode) float64 {
	for _, edge := range p.node.edges {
		if edge.dest == n {
			return edge.cost
		}
	}

	return BarrierValue
}
func (p *graphPath) EndDistance() float64 {
	var out float64
	for {
		if p == nil || p.next == nil {
			return out
		}
		out += p.distanceTo(p.next.node)
		p = p.next
	}
}
func (l graphPathList) shortest() (*graphPath, float64) {
	var out *graphPath
	var min float64
	for _, path := range l {
		dist := path.EndDistance()
		if out == nil || dist < min {
			out = path
			min = dist
		}
	}
	return out, min
}

func (g *graph) getPaths() graphPathList {
	out := make(graphPathList, len(g.waypoints))
	for i, w := range g.waypoints {
		current := &graphPath{node: w}
		out[i] = current
		required := g.requiredNodes()
		for len(required) > 0 {
			node := current.node.closest(required)
			required = remove(required, node)
			newNode := &graphPath{node: node}
			current.next = newNode
			current = newNode
		}
	}
	return out
}
func (g *graph) addWaypoints(pts []point) {
	for _, p := range pts {
		wp := g.add(p, false)
		g.waypoints = append(g.waypoints, wp)
	}
}
func (p path) toGraph() graph {
	g := graph{}
	for _, node := range p {
		g.add(node, true)
	}
	return g
}

// returns true of no changes made
func (p path) optimizeAlg(p3 bool, bypassBarriers bool) bool {
	done := true
	for i := 1; i < len(p)-1; i++ {
		for j := i + 1; j < len(p); j++ {
			if i == j {
				panic("unexpected")
			}
			if p3 {
				if p.trySwap3p(i, j, bypassBarriers) {
					done = false
				}
			} else {
				if p.trySwap2p(i, j, bypassBarriers) {
					done = false
				}
			}
		}
	}

	return done
}

func (src point) CalcDistance(dst point) float64 {
	diffX := dst.x - src.x
	diffY := dst.y - src.y
	diffZ := dst.z - src.z
	if diffZ > 0 {
		diffZ *= 1.5
	} else if diffZ < 0 {
		diffZ /= 1.5
	}
	return math.Sqrt(diffX*diffX + diffY*diffY + diffZ*diffZ)
}
func (src point) Distance(dst point, allowWaypoints bool, bypassBarriers bool) float64 {
	var pathDistance float64
	if src.barrier(dst) {
		if !bypassBarriers {
			return BarrierValue
		}
		if path, ok := src.findPath(dst); !ok {
			return BarrierValue
		} else {
			pathDistance, src = src.TakePath(path)
		}
	}

	pathDistance = pathDistance + src.CalcDistance(dst)

	//check if waypointing is faster
	if allowWaypoints && enableWaypointing {
		for _, w := range glWaypoints {
			waypointDistance := waypointCost + w.Distance(dst, false, false)
			if waypointDistance < pathDistance {
				return waypointDistance
			}
		}
	}

	return pathDistance
}

func (t typedGroup) IsMushroom() bool {
	return t.Type == GT_Mushroom
}

func (t *typedGroup) addPoint(pt point) {
	t._distance = t._distance + t._points[len(t._points)-1].CalcDistance(pt)
	t._revDistance = t._revDistance + pt.CalcDistance(t._points[len(t._points)-1])
	t._points = append(t._points, pt)
}
func (t typedGroup) last() point {
	return t._points[len(t._points)-1]
}
func (t typedGroup) first() point {
	return t._points[0]
}
func (t typedGroup) distance() float64 {
	if t.IsMushroom() {
		return mushroomCost
	}
	return t._distance
}
func (src point) TakePath(path []typedGroup) (float64, point) {
	var out float64
	for _, p := range path {
		out += src.CalcDistance(p.first())
		out += p.distance()
		src = p.last()
	}
	return out, src
}

func (src point) barrier(dst point) bool {
	for _, b := range glBarriers {
		if len(b._points) != 2 {
			fmt.Println("Unsupported barrier")
			return false
		}
		if b.Type == BT_DownOnly && dst.y < src.y {
			continue
		}
		if doIntersect(src, dst, b.first(), b.last()) {
			return true
		}
	}
	return false
}

func (src path) first() point {
	return src[0]
}
func (src path) last() point {
	return src[len(src)-1]
}

func (t typedGroup) Equals(t1 typedGroup) bool {
	return t.name == t1.name
}
func availablePaths(usedList []typedGroup) []typedGroup {
	out := make([]typedGroup, 0)
	for _, global := range glPaths {
		found := false
		for _, local := range usedList {
			if local.Equals(global) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, global)
		}
	}
	return out
}

func (t typedGroup) Reverse() typedGroup {
	rev := make([]point, len(t._points))
	copy(rev, t._points)
	slices.Reverse(rev)

	dist := t._distance
	revDist := t._revDistance

	return typedGroup{
		name:         t.name,
		reverseName:  t.reverseName,
		Type:         t.Type,
		_points:      rev,
		_distance:    revDist,
		_revDistance: dist,
	}
}

func (src point) pathTo(dst point, usedPaths []typedGroup) ([]typedGroup, bool) {
	if len(usedPaths) > 3 {
		return []typedGroup{}, false
	}
	choices := []typedGroup{}
	var addChoice = false
	start := src
	if len(usedPaths) > 0 {
		start = usedPaths[len(usedPaths)-1].last()
	}
	possiblePaths := [][]typedGroup{}
	for _, path := range availablePaths(usedPaths) {
		if !start.barrier(path.first()) {
			addChoice = true
			if !path.last().barrier(dst) {
				possiblePaths = append(possiblePaths, append(usedPaths, path))
			}
		}
		if !path.IsMushroom() && !start.barrier(path.last()) {
			addChoice = true
			if !path.first().barrier(dst) {
				possiblePaths = append(possiblePaths, append(usedPaths, path.Reverse()))
			}
		}

		if addChoice {
			choices = append(choices, path)
		}
	}

	if len(possiblePaths) == 0 {
		for _, choice := range choices {
			newPaths, ok := src.pathTo(dst, append(usedPaths, choice))
			if ok {
				possiblePaths = append(possiblePaths, newPaths)
			}
		}
	}
	if len(possiblePaths) == 0 {
		return []typedGroup{}, false
	}
	return cheapest(src, dst, possiblePaths), true
}

func calculatePathCost(src, dst point, group []typedGroup) float64 {
	total, newSrc := src.TakePath(group)
	return total + newSrc.Distance(dst, false, false)

}
func cheapest(src, dst point, groups [][]typedGroup) []typedGroup {
	min := math.MaxFloat64
	index := 0
	for i, next := range groups {
		cost := calculatePathCost(src, dst, next)
		if cost < min {
			index = i
			min = cost
		}
	}
	return groups[index]
}

func (src point) findPath(dst point) ([]typedGroup, bool) {
	return src.pathTo(dst, make([]typedGroup, 0))
}
