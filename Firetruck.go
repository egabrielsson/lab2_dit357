package main

import (
	"fmt"
)

type Firetruck struct {
	row, col int
	water    int

	destRow int
	destCol int
	enRoute bool
}

var trucks []Firetruck

func createFireTruck(row, col int) int {
	truck := Firetruck{row: row, col: col, water: 0, destRow: row, destCol: col, enRoute: false}
	trucks = append(trucks, truck)
	return len(trucks) - 1
}

func extinguishFire(index int) {
	t := &trucks[index]
	if t.water > 20 {
		fmt.Println("putout the fire")
	} else {
		fmt.Println("not enough water")
	}
}

func requestWater(index int, amount int) {
	t := &trucks[index]
	t.water += amount
	fmt.Println("Truck ", index, " Is now:", t.water)
}

func moveTruck(index int, direction string) {
	t := &trucks[index]
	switch direction {
	case "north":
		t.row--
	case "south":
		t.row++
	case "west":
		t.col--
	case "east":
		t.col++
	}
}

func driveToFire(index int, fireRow, fireCol int) bool {
	t := &trucks[index]

	if !t.enRoute || t.destRow != fireRow || t.destCol != fireCol {
		t.destRow, t.destCol = fireRow, fireCol
		t.enRoute = true
	}

	if t.row == t.destRow && t.col == t.destCol {
		t.enRoute = false
		return true
	}

	if t.row < t.destRow {
		t.row++
	} else if t.row > t.destRow {
		t.row--
	} else if t.col < t.destCol {
		t.col++
	} else if t.col > t.destCol {
		t.col--
	}

	if t.row == t.destRow && t.col == t.destCol {
		t.enRoute = false
		return true
	}
	return false
}
