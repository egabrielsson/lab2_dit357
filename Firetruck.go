package main

import (
	"fmt"
)

type Firetruck struct {
	row   int
	col   int
	water int
}

var trucks []Firetruck

func createFireTruck(row, cal int) {
	truck := Firetruck{row: row, col: cal, water: 0}
	trucks = append(trucks, truck)

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
		t.row++

	case "south":
		t.row--

	case "West":
		t.col++

	case "East":
		t.col--
	}
}
