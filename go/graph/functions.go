package graph

import (
	"container/heap"
)

const (
	unseen = iota
	seen
	queued = 1 << 30 // assume that no weight will be > 2^30
)

// O(V + E). It does not matter to traverse back
// on a bidirectional edge, because any vertex dfs is
// recursing on is marked as visited and won't be visited
// again anyway.
func (g *Graph) dfs(node *node, finishList *[]Node) {
	node.state = seen
	for _, edge := range node.edges {
		if edge.end.state == unseen {
			edge.end.parent = node
			g.dfs(edge.end, finishList)
		}
	}
	*finishList = append(*finishList, node.container)
}

func (g *Graph) dfsReversedEdges(node *node, finishList *[]Node) {
	node.state = seen
	for _, edge := range node.reversedEdges {
		if edge.end.state == unseen {
			edge.end.parent = node
			g.dfsReversedEdges(edge.end, finishList)
		}
	}
	*finishList = append(*finishList, node.container)
}

// Topologically sorts a directed acyclic graph.
// If the graph is cyclic, the sort order will change
// based on which node the sort starts on. O(V+E) complexity.
func (g *Graph) TopologicalSort() []Node {
	sorted := make([]Node, 0, len(g.nodes))
	// sort preorder (first jacket, then shirt)
	for _, node := range g.nodes {
		if node.state == unseen {
			g.dfs(node, &sorted)
		}
	}
	// now make post order for correct sort (jacket follows shirt). O(V)
	length := len(sorted)
	for i := 0; i < length/2; i++ {
		sorted[i], sorted[length-i-1] = sorted[length-i-1], sorted[i]
	}
	return sorted
}

// Returns reversed copy of the directed graph g. O(V+E) complexity.
// This function can be used to copy an undirected graph.
func (g *Graph) Reverse() *Graph {
	reversed := New(Directed)
	if g.kind == Undirected {
		reversed = New(Undirected)
	}
	// O(V)
	for _ = range g.nodes {
		reversed.MakeNode()
	}
	// O(V + E)
	for _, node := range g.nodes {
		for _, edge := range node.edges {
			reversed.MakeEdge(reversed.nodes[edge.end.index].container, reversed.nodes[node.index].container)
		}
	}
	return reversed
}

// Returns a slice of strongly connected nodes on a directed graph.
// If passed an undirected graph, returns nil.
// The returned components have reversed, nonexclusive edges.
// For example, if this is passed the graph
//     a->b, c
//     b->a, c
//     c
// will return components
//     [[c->a, b], [b->a], [a->b]]
// where -> represents the edges that the node contains.
// O(V+E) complexity.
func (g *Graph) StronglyConnectedComponents() [][]Node {
	if g.kind == Undirected {
		return nil
	}
	components := make([][]Node, 0)
	finishOrder := g.TopologicalSort()
	for i := range finishOrder {
		finishOrder[i].node.state = unseen
	}
	for _, sink := range finishOrder {
		if g.nodes[sink.node.index].state == unseen {
			component := make([]Node, 0)
			g.dfsReversedEdges(g.nodes[sink.node.index], &component)
			components = append(components, component)
		}
	}
	return components
}

// This function will sort the edges vertices and return a the edges
// in sorted order. An edge is represented by a length 2 array
// containing the start node and the end node.
//func Sort(g *Graph) []Edge { // O(V+E)
//	var edges []Edge
//	if g.kind == Directed {
//		edges = make([]Edge, 0, len(g.edges))
//	} else {
//		edges = make([]Edge, 0, len(g.edges)/2)
//	}
//	for _, node := range g.nodes { // O(V)
//		for _, edge := range g.edges[node] { // O(E)
//			if g.kind == Directed || edge.Start.node != node {
//				edges = append(edges, edge)
//			}
//		}
//	}
//	sort.Sort(edgeSlice(edges)) // O(E)
//	return edges
//}

// This function will return the edges corresponding to the
// minimum spanning tree in the graph based off of the edge's weight values.
func (g *Graph) MinimumSpanningTree() []Edge {
	// create priority queue for vertices
	nodesBase := nodeSlice(make([]*node, 0))
	nodes := &nodesBase
	heap.Init(nodes)
	for _, node := range g.nodes[1:] {
		node.state = 1<<30 - 1 // bit 31 is for queue testing
		node.state |= queued
		heap.Push(nodes, node)
	}
	g.nodes[0].state = 0
	heap.Push(nodes, g.nodes[0]) // now add root element

	for nodes.Len() > 0 {
		min := heap.Pop(nodes).(*node)
		for _, edge := range min.edges {
			v := edge.end // get the other side of the edge
			if nodes.QueueContains(v) && edge.weight < (v.state & ^queued) {
				v.parent = min
				heap.Remove(nodes, v.data)     // remove it
				v.state = edge.weight | queued // update state, add queue bit
				heap.Push(nodes, v)            // add it back in, WORD, O(lg n) update key time!
			}
		}
	}

	mst := make([]Edge, 0)
	for i := range g.nodes {
		if g.nodes[i].parent != nil {
			mst = append(mst, Edge{Weight: g.edgeBetweenWeight(g.nodes[i], g.nodes[i].parent), Start: g.nodes[i].container,
				End: g.nodes[i].parent.container, Kind: g.kind})
		}
	}

	return mst
}

// only called when the graph is guaranteed to have an edge
// between the two nodes
func (g *Graph) edgeBetweenWeight(v, u *node) int {
	for _, edge := range u.edges {
		// one of the two is always u
		if edge.end == v {
			return edge.weight
		}
	}
	return 0
}

type nodeSlice []*node

func (n *nodeSlice) Push(x interface{}) {
	p := x.(*node)
	p.data = len(*n) // index into heap
	*n = append(*n, x.(*node))
}
func (n *nodeSlice) Pop() interface{} {
	rNode := (*n)[len(*n)-1]
	rNode.state &= ^queued
	*n = (*n)[0 : len(*n)-1]
	return rNode
}
func (n nodeSlice) Len() int {
	return len(n)
}
func (n nodeSlice) Less(i, j int) bool {
	return n[i].state & ^queued < n[j].state & ^queued
}
func (n nodeSlice) Swap(i, j int) {
	n[j].data, n[i].data = n[i].data, n[j].data // swap data containing indices
	n[j], n[i] = n[i], n[j]
}
func (n nodeSlice) QueueContains(node *node) bool { // extend heap interface
	return node.state&queued > 0
}
