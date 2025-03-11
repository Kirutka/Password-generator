package main

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800
	screenHeight = 600
	cellSize     = 10
	cols         = screenWidth / cellSize
	rows         = screenHeight / cellSize
)

type Game struct {
	grid      [][]bool
	nextGrid  [][]bool
	lastUpdate time.Time
	paused    bool
}

func NewGame() *Game {
	g := &Game{
		grid:     make([][]bool, rows),
		nextGrid: make([][]bool, rows),
		paused:   false,
	}

	for i := range g.grid {
		g.grid[i] = make([]bool, cols)
		g.nextGrid[i] = make([]bool, cols)
		for j := range g.grid[i] {
			g.grid[i][j] = rand.Intn(10) == 0 // 10% chance to be alive
		}
	}

	return g
}

func (g *Game) Update() error {
	// Toggle pause with SPACE
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if time.Since(g.lastUpdate) > 250*time.Millisecond {
			g.paused = !g.paused
			g.lastUpdate = time.Now()
		}
	}

	if g.paused {
		// Add cells with left mouse button
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			col := x / cellSize
			row := y / cellSize
			if row >= 0 && row < rows && col >= 0 && col < cols {
				g.grid[row][col] = true
			}
		}
		// Remove cells with right mouse button
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			x, y := ebiten.CursorPosition()
			col := x / cellSize
			row := y / cellSize
			if row >= 0 && row < rows && col >= 0 && col < cols {
				g.grid[row][col] = false
			}
		}
		return nil
	}

	// Update every 100ms
	if time.Since(g.lastUpdate) < 100*time.Millisecond {
		return nil
	}
	g.lastUpdate = time.Now()

	// Calculate next generation
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			neighbors := g.countNeighbors(i, j)
			if g.grid[i][j] {
				g.nextGrid[i][j] = neighbors == 2 || neighbors == 3
			} else {
				g.nextGrid[i][j] = neighbors == 3
			}
		}
	}

	// Swap grids
	g.grid, g.nextGrid = g.nextGrid, g.grid

	return nil
}

func (g *Game) countNeighbors(row, col int) int {
	count := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}
			r := (row + i + rows) % rows
			c := (col + j + cols) % cols
			if g.grid[r][c] {
				count++
			}
		}
	}
	return count
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Dark background
	screen.Fill(color.RGBA{30, 30, 30, 255})

	// Draw cells
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if g.grid[i][j] {
				drawCell(screen, j, i, cellSize, color.RGBA{0, 255, 0, 255})
			}
		}
	}

	// Draw grid
	drawGrid(screen, cellSize, color.RGBA{60, 60, 60, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func drawCell(screen *ebiten.Image, x, y, size int, clr color.Color) {
	cell := ebiten.NewImage(size-1, size-1)
	cell.Fill(clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x*size)+0.5, float64(y*size)+0.5)
	screen.DrawImage(cell, op)
}

func drawGrid(screen *ebiten.Image, cellSize int, clr color.Color) {
	// Vertical lines
	for x := 0; x <= screenWidth; x += cellSize {
		line := ebiten.NewImage(1, screenHeight)
		line.Fill(clr)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), 0)
		screen.DrawImage(line, op)
	}

	// Horizontal lines
	for y := 0; y <= screenHeight; y += cellSize {
		line := ebiten.NewImage(screenWidth, 1)
		line.Fill(clr)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(y))
		screen.DrawImage(line, op)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Game of Life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}