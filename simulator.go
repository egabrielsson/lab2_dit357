package main

import (
	"fmt"
)

// cell states - empty - fire - extinguished

type Grid[T any] struct {
	ug *grid[T]
	rg Rect
}

type grid[T any] struct {
	Width int
	Cells []T
}

type Point struct {
	X, Y int
}

type Rect struct {
	Min, Max Point
}

func NewGrid[T any](w, h int) Grid[T] {
	if w < 0 || h < 0 {
		panic(fmt.Sprintf("negative dimensions: NewGrid(%d,%d)", w, h))
	}
	gd := Grid[T]{}
	gd.ug = &grid[T]{}
	gd.rg.Max = Point{w, h}
	gd.ug.Width = w
	gd.ug.Cells = make([]T, w*h)
	return gd
}
