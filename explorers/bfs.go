package explorers

import (
	"slices"

	"github.com/Ozoniuss/maze-explorer/coord"
)

func NewBfsExplorer(board [][]byte, start, end coord.Pos) *BfsExplorer {
	parents := make(map[coord.Pos]coord.Pos)
	parents[start] = coord.NullPos
	return &BfsExplorer{
		Board:        board,
		Start:        start,
		End:          end,
		Q:            []bfsState{{pos: start, length: 1}},
		Visited:      make(map[coord.Pos]struct{}),
		Currlen:      0,
		Parents:      parents,
		ShortestPath: make([]coord.Pos, 0),
	}
}

type bfsState struct {
	pos    coord.Pos
	length int
}

type BfsExplorer struct {
	Board        [][]byte
	Q            []bfsState
	Start        coord.Pos
	End          coord.Pos
	Visited      map[coord.Pos]struct{}
	Parents      map[coord.Pos]coord.Pos
	Currlen      int // draw only those that exceed the current length
	ShortestPath []coord.Pos
}

func (b *BfsExplorer) buildParent() {
	curr := b.End
	arr := []coord.Pos{}
	for curr != coord.NullPos {
		arr = append(arr, curr)
		v, ok := b.Parents[curr]
		if !ok {
			panic("no parent found")
		}
		curr = v
	}
	slices.Reverse(arr)
	b.ShortestPath = arr
}

func (b *BfsExplorer) ExploreUntilNewCellsAreFound() bool {
	c := b.Currlen
	for b.Currlen == c {
		ok := b.Explore()
		if !ok {
			return false
		}
	}
	return true
}

func (b *BfsExplorer) Explore() bool {
	top := b.Q[0]
	b.Q = b.Q[1:]
	if b.Currlen <= top.length {
		b.Currlen = top.length
	}
	if b.End == top.pos {
		b.buildParent()
		return false
	}
	ns := []coord.Pos{
		{top.pos[0] - 1, top.pos[1]},
		{top.pos[0] + 1, top.pos[1]},
		{top.pos[0], top.pos[1] + 1},
		{top.pos[0], top.pos[1] - 1},
	}
	for _, n := range ns {
		if n[0] < 0 || n[0] >= len(b.Board) || n[1] < 0 || n[1] >= len(b.Board[0]) {
			continue
		}
		_, visitedAlready := b.Visited[n]
		if visitedAlready {
			continue
		}
		i, j := n[0], n[1]
		if b.Board[i][j] == '#' {
			continue
		} else {
			b.Q = append(b.Q, bfsState{
				pos:    n,
				length: top.length + 1,
			})
			// If I already set the parent it means I got there faster
			if _, ok := b.Parents[n]; !ok {
				b.Parents[n] = top.pos
			}
		}
	}
	b.Visited[top.pos] = struct{}{}
	return true
}
