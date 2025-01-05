package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"slices"
	"strings"
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
	fmt.Println(parts[0], parts[1])

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

	fmt.Println(start, end)
	myApp := app.New()
	myWindow := myApp.NewWindow("Maze Explorer")
	myWindow.Resize(fyne.NewSquareSize(512))

	rgExploreType := widget.NewRadioGroup([]string{"bfs", "dfs"}, func(s string) {})
	rgExploreType.Selected = "bfs"
	rgExploreType.Horizontal = true

	var explorer *explorers.BfsExplorer
	btnStartExplore := widget.NewButton("Explore!", func() {

	})

	hbSelectionContent := container.NewHBox(rgExploreType, btnStartExplore)

	grid := initBoardContainer(len(board), len(board[0]))
	content := container.NewVBox(hbSelectionContent, grid)
	myWindow.SetContent(content)
	explorer = explorers.NewBfsExplorer(board, start, end)
	clen := explorer.Currlen
	go func() {
		for explorer.Explore() {
			if clen < explorer.Currlen {
				clen = explorer.Currlen
				drawBoard(grid, board, explorer.Visited, explorer.ShortestPath)
				time.Sleep(10 * time.Millisecond)
			}
		}
		fmt.Println("sp", explorer.ShortestPath)
		drawBoard(grid, board, explorer.Visited, explorer.ShortestPath)
		grid.Refresh()

	}()
	myWindow.ShowAndRun()
}

func initBoardContainer(r, c int) *fyne.Container {
	grid := container.New(layout.NewGridLayout(c))
	for range r {
		for range c {
			grid.Add(canvas.NewRectangle(color.Black))
			fmt.Println("ce")
		}
	}
	fmt.Println("leno", len(grid.Objects))
	return grid
}

func drawBoard(grid *fyne.Container, board [][]byte, visited map[coord.Pos]struct{}, shortestPath []coord.Pos) {
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
				rectangles[i*size+j] = canvas.NewText("S", color.White)
			} else if board[i][j] == 'E' {
				rectangles[i*size+j] = canvas.NewText("E", color.White)
			}
		}
	}
	grid.Refresh()
}
