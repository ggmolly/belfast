package answer

type chapterCellKey struct {
	Row    uint32
	Column uint32
}

func findMovePath(grids []chapterGrid, start chapterPos, end chapterPos) []chapterPos {
	if start == end {
		return []chapterPos{start}
	}
	walkable := make(map[chapterCellKey]bool, len(grids))
	for _, grid := range grids {
		if grid.Walkable {
			walkable[chapterCellKey{Row: grid.Row, Column: grid.Column}] = true
		}
	}
	startKey := chapterCellKey{Row: start.Row, Column: start.Column}
	endKey := chapterCellKey{Row: end.Row, Column: end.Column}
	if !walkable[startKey] || !walkable[endKey] {
		return nil
	}
	queue := []chapterCellKey{startKey}
	visited := map[chapterCellKey]bool{startKey: true}
	parent := map[chapterCellKey]chapterCellKey{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == endKey {
			break
		}
		neighbors := make([]chapterCellKey, 0, 4)
		neighbors = append(neighbors, chapterCellKey{Row: current.Row + 1, Column: current.Column})
		if current.Row > 1 {
			neighbors = append(neighbors, chapterCellKey{Row: current.Row - 1, Column: current.Column})
		}
		neighbors = append(neighbors, chapterCellKey{Row: current.Row, Column: current.Column + 1})
		if current.Column > 1 {
			neighbors = append(neighbors, chapterCellKey{Row: current.Row, Column: current.Column - 1})
		}
		for _, neighbor := range neighbors {
			if visited[neighbor] || !walkable[neighbor] {
				continue
			}
			visited[neighbor] = true
			parent[neighbor] = current
			queue = append(queue, neighbor)
		}
	}
	if !visited[endKey] {
		return nil
	}
	path := []chapterPos{}
	current := endKey
	for {
		path = append(path, chapterPos{Row: current.Row, Column: current.Column})
		if current == startKey {
			break
		}
		current = parent[current]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
