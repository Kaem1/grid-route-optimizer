package graph

// MultiSourceBFS assigns every vertex reachable from sources to whichever
// source discovers it first, using a single FIFO queue seeded with the
// sources in order (algorithm spec section 4: "kolejka FIFO inicjalizowana
// startami w kolejności indeksów robotów; sąsiedzi rozpatrywani w kolejności
// rosnącej"). Neighbors are already visited in ascending (Y,X) order via
// Neighbors, so the result is fully deterministic. Vertices unreachable
// from every source are simply absent from the returned maps.
//
// owner maps a vertex to the index (into sources) of the source that
// discovered it; dist maps it to its graph distance from that source.
// Duplicate entries in sources keep the earliest index.
func MultiSourceBFS(g *Graph, sources []VertexID) (owner map[VertexID]int, dist map[VertexID]int) {
	owner = make(map[VertexID]int, len(sources))
	dist = make(map[VertexID]int, len(sources))

	queue := make([]VertexID, 0, len(sources))
	for i, s := range sources {
		if _, seen := owner[s]; seen {
			continue
		}
		owner[s] = i
		dist[s] = 0
		queue = append(queue, s)
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, n := range g.Neighbors(current) {
			if _, seen := owner[n]; seen {
				continue
			}
			owner[n] = owner[current]
			dist[n] = dist[current] + 1
			queue = append(queue, n)
		}
	}

	return owner, dist
}
