package explorers

import (
	"slices"

	"github.com/Ozoniuss/maze-explorer/coord"
)

type DfsExplorer struct {
	Board        [][]byte
	S            []coord.Pos
	Start        coord.Pos
	End          coord.Pos
	Visited      map[coord.Pos]struct{}
	Currlen      int // draw only those that exceed the current length
	ShortestPath []coord.Pos
	beginning    bool
}

func NewDfsExplorer(board [][]byte, start, end coord.Pos) *DfsExplorer {
	return &DfsExplorer{
		Board:        board,
		S:            make([]coord.Pos, 0),
		Start:        start,
		End:          end,
		Visited:      make(map[coord.Pos]struct{}),
		ShortestPath: make([]coord.Pos, 0),
		beginning:    true,
	}
}

// TODO: this can heavily be optimized
func (d *DfsExplorer) GetOccupied() map[coord.Pos]struct{} {
	m := make(map[coord.Pos]struct{}, len(d.S))
	for _, c := range d.S {
		m[c] = struct{}{}
	}
	return m
}

func (d *DfsExplorer) GetPath() []coord.Pos {
	return d.ShortestPath
}

func (d *DfsExplorer) Reset() {
	d.S = make([]coord.Pos, 0)
	d.Visited = make(map[coord.Pos]struct{})
	d.ShortestPath = make([]coord.Pos, 0)
	d.beginning = true
}

func (d *DfsExplorer) Advance() bool {
	return d.Explore()
}

func (d *DfsExplorer) Explore() bool {

	if d.beginning {
		d.beginning = false
		d.S = append(d.S, d.Start)
		return true
	}
	top := d.S[len(d.S)-1]
	d.Visited[top] = struct{}{}

	if top == d.End {
		d.ShortestPath = slices.Clone(d.S)
		return false
	}

	ns := []coord.Pos{
		{top[0] - 1, top[1]},
		{top[0] + 1, top[1]},
		{top[0], top[1] - 1},
		{top[0], top[1] + 1},
	}
	shouldRemove := true
	for _, n := range ns {
		if n[0] < 0 || n[0] >= len(d.Board) || n[1] < 0 || n[1] >= len(d.Board[0]) {
			continue
		}
		_, visitedAlready := d.Visited[n]
		if visitedAlready {
			continue
		}
		i, j := n[0], n[1]
		if d.Board[i][j] == '#' {
			continue
		} else {
			shouldRemove = false
			d.S = append(d.S, n)
			break
		}
	}
	if shouldRemove {
		d.S = d.S[:len(d.S)-1]
	}
	return true
}
