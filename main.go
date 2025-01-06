package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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
	hbSelectionContent := container.NewHBox(rgExploreType, btnStartExplore, btnCancelExploration)

	grid := initBoardContainer(len(board), len(board[0]))
	content := container.NewVBox(hbSelectionContent, grid)
	myWindow.SetContent(content)

	drawBoard(start, grid, board, nil, nil)
	explorer := explorers.NewBfsExplorer(board, start, end)

	canceled := atomic.Bool{}
	solutionDrawn := atomic.Bool{}
	canceled.Store(false)
	solutionDrawn.Store(false)
	btnStartExplore.OnTapped = func() {
		btnStartExplore.Disable()
		go func() {
			for explorer.ExploreUntilNewCellsAreFound() {
				if canceled.Load() == false {
					drawBoard(start, grid, board, explorer.Visited, explorer.ShortestPath)
					time.Sleep(10 * time.Millisecond)
				} else {
					canceled.Store(false)
					explorer.Reset()
					drawBoard(start, grid, board, explorer.Visited, explorer.ShortestPath)
					btnStartExplore.Enable()
					return
				}
			}
			drawBoard(start, grid, board, explorer.Visited, explorer.ShortestPath)
			btnStartExplore.Enable()
			solutionDrawn.Store(true)
			fmt.Println("done here", explorer.ShortestPath)
		}()
	}

	// there has to be a better way to do the below
	btnCancelExploration.OnTapped = func() {
		if solutionDrawn.CompareAndSwap(true, false) {
			explorer.Reset()
			drawBoard(start, grid, board, explorer.Visited, explorer.ShortestPath)
			return
		}
		// only swap if canceled is set to false. if it's set to true,
		// a cancellation hasn't been cleaned up yet.
		if canceled.CompareAndSwap(false, true) {
			fmt.Println("canceled change value")
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

func drawBoard(start coord.Pos, grid *fyne.Container, board [][]byte, visited map[coord.Pos]struct{}, shortestPath []coord.Pos) {
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
