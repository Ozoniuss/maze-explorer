package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Ozoniuss/maze-explorer/coord"
	"github.com/Ozoniuss/maze-explorer/explorers"
)

//go:embed maze.txt
var maze string

func main() {
	parts := strings.Split(maze, "\n")

	board := make([][]byte, len(parts))
	for i := range len(parts) {
		board[i] = make([]byte, len(parts))
	}

	start := coord.Pos{}
	end := coord.Pos{}
	for i := range len(parts) {
		for j := range len(parts[i]) {
			board[i][j] = parts[i][j]
			if parts[i][j] == 'S' {
				start = coord.Pos{i, j}
			}
			if parts[i][j] == 'E' {
				end = coord.Pos{i, j}
			}
		}
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("Maze Explorer")

	rgExploreType := widget.NewRadioGroup([]string{"bfs", "dfs"}, func(s string) {})
	rgExploreType.Selected = "bfs"
	rgExploreType.Horizontal = true

	btnStartExplore := widget.NewButton("Explore!", func() {})
	btnCancelExploration := widget.NewButton("Cancel", func() {})

	waitTime := binding.NewFloat()
	waitTime.Set(100)

	sldExecutionSpeed := widget.NewSliderWithData(10, 1000, waitTime)
	hbSelectionContent := container.NewGridWithColumns(6, rgExploreType, btnStartExplore, btnCancelExploration, widget.NewSeparator(), sldExecutionSpeed)

	grid := initBoardContainer(len(board), len(board[0]))
	content := container.NewVBox(hbSelectionContent, grid)
	myWindow.SetContent(content)

	drawBoard(grid, board, nil, nil)
	explorer := explorers.NewDfsExplorer(board, start, end)

	// 0 - ready for run
	// 1 - drawing
	// 2 - fully drawed, final path shows
	// 3 - canceled was hit
	// 4 - paused run by cancel
	runState := atomic.Int32{}
	runState.Store(0)

	rmu := &sync.Mutex{}

	btnStartExplore.OnTapped = func() {
		btnStartExplore.Disable()
		go func() {
			// don't want to have two simultaneous runners. Want to reset the
			// explorer state completely before any run.
			rmu.Lock()
			defer rmu.Unlock()
			runState.Store(1)
			defer explorer.Reset()
			for explorer.ExploreUntilNewCellsAreFound() {
				// run was canceled
				if runState.Load() != 3 {
					drawBoard(grid, board, explorer.GetOccupied(), explorer.ShortestPath)
					wt, err := waitTime.Get()
					if err != nil {
						panic(err)
					}
					delta := 1010 - int(wt)
					time.Sleep(time.Duration(delta) * time.Millisecond)
				} else {
					// Run was canceled, set the state to 4
					drawBoard(grid, board, explorer.GetOccupied(), explorer.ShortestPath)
					explorer.Reset()
					btnStartExplore.Enable()
					// set this only at the end, cause you can get here only
					// if you canceled, which has no effect twice.
					runState.Store(4)
					return
				}
			}
			drawBoard(grid, board, explorer.GetOccupied(), explorer.ShortestPath)
			btnStartExplore.Enable()
			// only set this to 1 if the previous state was 1. If it's not 1,
			// it means it was altered via a cancel.
			if runState.CompareAndSwap(1, 2) {
				return
			} else {
				runState.Store(0)
				explorer.Reset() // explicitly reset the explorer before drawing
				// the board
				drawBoard(grid, board, explorer.GetOccupied(), explorer.ShortestPath)
			}
		}()
	}

	// there has to be a better way to do the below
	btnCancelExploration.OnTapped = func() {
		if runState.Load() == 0 {
			return
		}
		// run in state 2, redraw required if exiting
		if runState.CompareAndSwap(1, 3) {
			fmt.Println("state 1")
			return
		}
		// run in state 1, need to stop the function
		if runState.CompareAndSwap(2, 3) {
			fmt.Println("state 2")
			// force a redraw anyway because state may be 1 even outside the
			// for loop, which no longer checks for this value.
			drawBoard(grid, board, explorer.GetOccupied(), explorer.ShortestPath)
			runState.Store(0)
			return
		}

		// state 4 means draw was left from when cancel was hit. Erase everything
		if runState.CompareAndSwap(4, 0) {
			// explorer is always reset when the function is exited.
			drawBoard(grid, board, explorer.GetOccupied(), explorer.ShortestPath)
		}
	}

	myWindow.ShowAndRun()
}

func initBoardContainer(r, c int) *fyne.Container {
	grid := container.New(layout.NewGridLayout(c))
	for range r {
		for range c {
			grid.Add(canvas.NewRectangle(color.Black))
		}
	}
	return grid
}

func drawLetter(letter string, pos coord.Pos, visited map[coord.Pos]struct{}, shortestPath []coord.Pos) *fyne.Container {
	var rect *canvas.Rectangle
	if len(shortestPath) != 0 {
		rect = canvas.NewRectangle(color.RGBA{R: 255, G: 0, B: 0, A: 128})
	} else if _, ok := visited[pos]; ok {
		rect = canvas.NewRectangle(color.RGBA{R: 0, G: 255, B: 0, A: 128})
	} else {
		rect = canvas.NewRectangle(color.Black)
	}
	rect.SetMinSize(fyne.NewSize(20, 20)) // how can I make use of this better?
	s := canvas.NewText(letter, color.White)
	c := container.NewCenter(s)
	final := container.New(layout.NewStackLayout(), rect, c)
	return final
}

var dbm = &sync.Mutex{}

func drawBoard(grid *fyne.Container, board [][]byte, visited map[coord.Pos]struct{}, shortestPath []coord.Pos) {
	dbm.Lock()
	defer dbm.Unlock()
	rectangles := grid.Objects
	size := len(board)
	for i := range len(board) {
		for j := range len(board[0]) {
			if board[i][j] == '#' {
				rectangles[i*size+j] = canvas.NewRectangle(color.White)
			} else if board[i][j] == '.' {
				if len(shortestPath) != 0 && slices.Contains(shortestPath, coord.Pos{i, j}) {
					rectangles[i*size+j] = canvas.NewRectangle(color.RGBA{R: 255, G: 0, B: 0, A: 128})
				} else if _, ok := visited[coord.Pos{i, j}]; ok {
					rectangles[i*size+j] = canvas.NewRectangle(color.RGBA{R: 0, G: 255, B: 0, A: 128})
				} else {
					rectangles[i*size+j] = canvas.NewRectangle(color.Black)
				}
			} else if board[i][j] == 'S' {
				rectangles[i*size+j] = drawLetter("S", coord.Pos{i, j}, visited, shortestPath)
			} else if board[i][j] == 'E' {
				rectangles[i*size+j] = drawLetter("E", coord.Pos{i, j}, visited, shortestPath)
			}
		}
	}
	grid.Refresh()
}
