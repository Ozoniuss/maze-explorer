package explorers

import "github.com/Ozoniuss/maze-explorer/coord"

type Explorer interface {
	// Fields that should be marked as occupied.
	GetOccupied() map[coord.Pos]struct{}
	// Complete path from start to finish.
	GetPath() []coord.Pos
	// Advance the explorer to the next drawable state.
	Advance() bool
	// Reset the explorer to its state when just created.
	Reset()
}
