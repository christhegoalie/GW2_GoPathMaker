package location

import (
	"errors"
	"slices"
)

type algorithm int

const (
	ALG_2p = iota
	ALG_4p
)

var GLOBAL_Waypoints Path

type GraphPath struct {
	BindEnd bool
	node    *graphNode
	next    *GraphPath
}
type edge struct {
	shortcuts []TypedGroup //optional non-direct path to reach node
	cost      float64
	dest      *graphNode
}
type graphNode struct {
	location Point
	required bool
	edges    []edge
}
type Graph struct {
	useEndpoint bool
	nodes       []*graphNode
	waypoints   []*graphNode
}
type GraphPathTraversal struct {
	path   *GraphPath
	length int
}

func (path GraphPathTraversal) reverse() GraphPathTraversal {
	var out *GraphPath
	var tmp *GraphPath
	ls := []*graphNode{}
	for a := path.path; a != nil; a = a.next {
		ls = append(ls, a.node)
	}
	for i := len(ls) - 1; i >= 0; i-- {
		if i != 0 && !edgeExists(ls[i], ls[i-1]) {
			return GraphPathTraversal{}
		}
		if out == nil {
			out = &GraphPath{node: ls[i]}
			tmp = out
		} else {
			tmp.next = &GraphPath{node: ls[i]}
			tmp = tmp.next
		}
	}
	return GraphPathTraversal{
		path:   out,
		length: path.length,
	}
}

func edgeExists(node1 *graphNode, node2 *graphNode) bool {
	for _, e := range node1.edges {
		if node2 == e.dest {
			return true
		}
	}
	return false
}

func (path *GraphPath) trySwapNext(target *GraphPath, alg algorithm) bool {
	if path == nil || path.next == nil || target == nil {
		return false
	}

	start := path
	node1 := path.next
	middle := path.next.next
	node2 := target
	end := target.next

	var p1, p2 *GraphPath
	if alg == ALG_2p {
		if middle != target {
		} else {
			fp := fullPath(middle, node2)
			rev := fp.reverse()
			p1 = newPath(start, node1, &fp, node2, end)
			p2 = newPath(start, node1, &rev, node2, end)
		}
	} else if alg == ALG_4p {
		if middle == target {
			p1 = newPath(start, node1, node2, end)
			p2 = newPath(start, node2, node1, end)
		} else {
			fp := fullPath(middle, node2)
			rev := fp.reverse()
			if rev.length == 0 {
				return false
			}
			p1 = newPath(start, node1, &fp, node2, end)
			p2 = newPath(start, node2, &rev, node1, end)
			if p2 == nil || p1 == nil {
				return false
			}
		}
	}

	if p1Dist, p2Dist := p1.EndDistance(), p2.EndDistance(); p2Dist < p1Dist {
		if end != nil && end.next != nil {
			tmp := p2
			for tmp.next != nil {
				tmp = tmp.next
			}
			tmp.next = end.next
		}
		path.next = p2.next
		return true
	}
	return false
}
func (p *GraphPath) hasDuplicates() bool {
	check := p
	for check != nil {
		comp := check.next
		for comp != nil {
			if comp.node.location == check.node.location {
				return true
			}
			comp = comp.next
		}
		check = check.next
	}
	return false
}
func (path *GraphPath) length() int {
	len := 0
	for path != nil {
		len++
		path = path.next
	}
	return len
}
func (path *GraphPath) Optimize(bindEnd bool) bool {
	if path == nil || path.next == nil {
		return false
	}

	src := path
	for {
		target := src.next.next
		for {
			if target == nil {
				break
			}
			if bindEnd && target.next == nil {
				break
			}
			if found := src.trySwapNext(target, ALG_4p); found {
				return found
			}
			target = target.next
		}
		src = src.next
		if src.next.next == nil {
			break
		}
	}

	src = path
	for {
		target := src.next.next
		for {
			if target == nil {
				break
			}
			if bindEnd && target.next == nil {
				break
			}
			if found := src.trySwapNext(target, ALG_2p); found {
				return true
			}
			target = target.next
		}
		src = src.next
		if src.next.next == nil {
			break
		}
	}

	return false
}

func fullPath(start *GraphPath, end *GraphPath) GraphPathTraversal {
	out := GraphPath{node: start.node}
	var tmp *GraphPath = &out
	length := 1
	if start.next != nil {
		for a := start.next; a != nil; a = a.next {
			if a == end {
				break
			}
			tmp.next = &GraphPath{node: a.node}
			tmp = tmp.next
			length++
		}
	}
	return GraphPathTraversal{
		path:   &out,
		length: length,
	}
}

func newPath(args ...any) *GraphPath {
	root := GraphPath{}
	tmp := &root
	for _, i := range args {
		switch v := i.(type) {
		case *GraphPath:
			if v != nil {
				tmp.next = &GraphPath{node: v.node}
				tmp = tmp.next
			}
		case *GraphPathTraversal:
			if v == nil {
				return nil
			}
			for a := v.path; a != nil; a = a.next {
				tmp.next = &GraphPath{node: a.node}
				tmp = tmp.next
			}
		default:
			panic("invalid type")
		}
	}
	if root.next == nil {
		panic("no path")
	}
	return root.next
}

type graphPathList []*GraphPath

func (p *graphNode) findEdge(dst *graphNode) (edge, error) {
	for _, edge := range p.edges {
		if edge.dest == dst {
			return edge, nil
		}
	}
	return edge{}, errors.New("expected edge")
}
func (p *GraphPath) ToPath() Path {
	out := Path{p.node.location}
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
func (p *GraphPath) distanceTo(n *graphNode) float64 {
	for _, edge := range p.node.edges {
		if edge.dest == n {
			return edge.cost
		}
	}

	return BarrierValue
}
func (p *GraphPath) EndDistance() float64 {
	var out float64
	for {
		if p == nil || p.next == nil {
			return out
		}
		out += p.distanceTo(p.next.node)
		p = p.next
	}
}
func (l graphPathList) Shortest() (*GraphPath, float64) {
	var out *GraphPath
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

func (g *Graph) GetPaths() graphPathList {
	out := make(graphPathList, len(g.waypoints))
	for i, w := range g.waypoints {
		current := &GraphPath{BindEnd: g.useEndpoint, node: w}
		out[i] = current
		required := g.requiredNodes()
		for len(required) > 0 {
			node := current.node.closest(required)
			required = remove(required, node)
			newNode := &GraphPath{BindEnd: g.useEndpoint, node: node}
			current.next = newNode
			current = newNode
		}
	}
	return out
}
func (g *Graph) AddWaypoints(pts []Point) {
	for _, p := range pts {
		wp := g.add(p, false, false)
		g.waypoints = append(g.waypoints, wp)
	}
}
func (g *Graph) SetEndpoint(pt Point) {
	g.useEndpoint = true
	g.add(pt, true, true)
}
func (p Path) ToGraph() Graph {
	g := Graph{}
	for _, node := range p {
		g.add(node, true, false)
	}
	return g
}

// returns true of no changes made
func (p Path) optimizeAlg(p3 bool, bypassBarriers bool) bool {
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

func (t TypedGroup) Reverse() TypedGroup {
	rev := make([]Point, len(t._points))
	copy(rev, t._points)
	slices.Reverse(rev)

	dist := t._distance
	revDist := t._revDistance

	return TypedGroup{
		Name:         t.Name,
		ReverseName:  t.ReverseName,
		Type:         t.Type,
		_points:      rev,
		_distance:    revDist,
		_revDistance: dist,
	}
}

func (g Graph) requiredNodes() []*graphNode {
	out := []*graphNode{}
	for _, n := range g.nodes {
		if n.required {
			out = append(out, n)
		}
	}
	return out
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

func (node1 *graphNode) connect(node2 *graphNode) {
	const MAX_PATH_LENGTH = 10000

	// find any possible paths to the node
	toPath, _ := node1.location.FindPath(node2.location)
	fromPath, _ := node2.location.FindPath(node1.location)
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
func (g *Graph) add(pt Point, required bool, endNode bool) *graphNode {
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
	if !endNode && g.useEndpoint && len(g.nodes) > 1 {
		tmp := g.nodes[len(g.nodes)-1]
		g.nodes[len(g.nodes)-1] = g.nodes[len(g.nodes)-2]
		g.nodes[len(g.nodes)-2] = tmp
	}
	return &node
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

func (p Path) Distance(allowWaypoints bool, bypassBarriers bool) float64 {
	var out float64
	for i := 0; i < len(p)-1; i++ {
		out += p[i].Distance(p[i+1], allowWaypoints, bypassBarriers)
	}
	return out
}

func (p Path) trySwap3p(i, j int, bypasBarriers bool) bool {

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
func (p Path) trySwap2p(i, j int, bypasBarriers bool) bool {
	if len(p) < 2 || i >= len(p) || j >= len(p) {
		return false
	}

	var oldSeg Path
	var newSeg Path
	if j+1 < len(p) {
		oldSeg = p[i-1 : j+2]
		newSeg = make(Path, len(oldSeg))
		copy(newSeg, oldSeg)
		newSeg[1] = oldSeg[len(oldSeg)-2]
		newSeg[len(newSeg)-2] = oldSeg[1]
	} else {
		oldSeg = p[i-1 : j+1]
		newSeg = make(Path, len(oldSeg))
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

func findPathDistance(start Point, path []TypedGroup, dest Point) float64 {
	if len(path) == 0 {
		return BarrierValue //indicates no paths
	}
	pathLen := start.CalcDistance(path[0].First()) + path[len(path)-1].Last().CalcDistance(dest)
	for index, p := range path {
		//Add the distance from the previous path
		if index > 0 {
			pathLen += path[index-1].Last().CalcDistance(p.First())
		}
		//Add the distance of each path
		pathLen += p.Distance()
	}
	return pathLen
}
