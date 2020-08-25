package main

import (
	"fmt"
)

// Grid represents the status of the whole game in a specific time.
type Grid struct {
	Rows    [9]CellGroup
	Columns [9]CellGroup
	Regions [9]CellGroup
}

// Init initial the grid with a 9x9 sequence of numbers.
func (g *Grid) Init(rawNumberes [9][9]int) {
	for rowIndex, row := range rawNumberes {
		for columnIndex, rawNumber := range row {
			regionIndex := columnIndex/3 + (rowIndex/3)*3
			regionPosition := columnIndex%3 + (rowIndex%3)*3

			cell := Cell{
				Row:        &g.Rows[rowIndex],
				Column:     &g.Columns[columnIndex],
				Region:     &g.Regions[regionIndex],
				Candidates: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			}

			if rawNumber != 0 {
				cell.Candidates = []int{rawNumber}
			}
			g.Rows[rowIndex][columnIndex] = &cell
			g.Columns[columnIndex][rowIndex] = &cell
			g.Regions[regionIndex][regionPosition] = &cell
		}
	}
}

// Check checks the grid and assign number to cells in a deterministic way, it performs 2 checks.
// 1. If there's any cell has only one candidate.
// 2. If there's a candidate owns by exactly one cell in a CellGroup.
// If a cell matches any criteria above, then assign the candidate to the cell.
func (g *Grid) Check() {
	for {
		var modified bool

		for _, cellGroup := range g.Rows {
			for _, cell := range cellGroup {
				if len(cell.Candidates) == 1 && !cell.Confirmed {
					cell.Mark(cell.Candidates[0])
					modified = true
				}
			}
		}

		for _, cellGroup := range g.getAllCellGroups() {
			if uniqCandidate := cellGroup.FindUniqueCandidate(); uniqCandidate != 0 {
				for _, cell := range cellGroup {
					if !cell.Confirmed {
						for _, candidate := range cell.Candidates {
							if candidate == uniqCandidate {
								cell.Mark(uniqCandidate)
								modified = true
							}
						}
					}
				}
			}
		}

		if !modified {
			break
		}
	}
}

// Print shows the arrangement of each cell and their number in the grid.
func (g *Grid) Print() {
	for _, r := range g.Rows {
		fmt.Println("")
		for _, c := range r {
			fmt.Print(c.Number)
		}
	}
	fmt.Println("")
}

// PrintStatus shows the currnet status and candidates for each cell in the grid.
func (g *Grid) PrintStatus() {
	for _, r := range g.Rows {
		fmt.Println("")
		for _, c := range r {
			fmt.Println(c)
		}
	}
}

func (g *Grid) getAllCellGroups() []CellGroup {
	cellGroups := make([]CellGroup, 0)
	cellGroups = append(cellGroups, g.Rows[:]...)
	cellGroups = append(cellGroups, g.Columns[:]...)
	cellGroups = append(cellGroups, g.Regions[:]...)
	return cellGroups
}

// Verify checks whether the grid is complete, not complete or malformed.
func (g *Grid) Verify() string {
	isFullfilled := true
	for _, row := range g.Rows {
		for _, cell := range row {
			if !cell.Confirmed {
				isFullfilled = false
				if len(cell.Candidates) == 0 {
					return GridStatusMalformed
				}
			}
		}
	}
	if !isFullfilled {
		return GridStatusNotCompleted
	}
	for _, cellGroup := range g.getAllCellGroups() {
		if !cellGroup.CheckIfComplete() {
			return GridStatusMalformed
		}
	}
	return GridStatusCompleted
}

// GetBranchCellIndex determined which cell to start to traversal remaining possible solutions.
// The logic is pick the cell with least candidates left for less quessing and calculation time.
func (g *Grid) GetBranchCellIndex() (rowIndex int, cellIndex int) {
	var candidateLength int
	for rindex, row := range g.Rows {
		for cindex, cell := range row {
			if !cell.Confirmed {
				if len(cell.Candidates) == MinimumBranchableCandidates {
					return rindex, cindex
				}

				if len(cell.Candidates) == 0 || candidateLength > len(cell.Candidates) {
					rowIndex = rindex
					cellIndex = cindex
				}
			}
		}
	}
	return rowIndex, cellIndex
}

// ToRawNumberGrid transform the grid to 9x9 sequential numbers, for grid duplication.
func (g *Grid) ToRawNumberGrid() [9][9]int {
	var rawNumbereGrid [9][9]int
	for rowIndex, row := range g.Rows {
		for cellIndex, cell := range row {
			rawNumbereGrid[rowIndex][cellIndex] = cell.Number
		}
	}
	return rawNumbereGrid
}

// CellGroup is abstraction of a set of numbers in a Sudoku game.
// The number of cells in the same CellGroup should be disdinct.
type CellGroup [9]*Cell

// TrimCandidate eliminates candidates from cells which do not confirm it's number in the CellGroup.
func (c *CellGroup) TrimCandidate(trimedItem int) {
	for _, cell := range c {
		if !cell.Confirmed {
			for index, candidate := range cell.Candidates {
				if candidate == trimedItem {
					lastPosition := len(cell.Candidates) - 1
					cell.Candidates[index] = cell.Candidates[lastPosition]
					cell.Candidates[lastPosition] = 0
					cell.Candidates = cell.Candidates[:lastPosition]
				}
			}
		}
	}
}

// FindUniqueCandidate finds if there is any candidate owns by only one cell in the CellGroup.
func (c *CellGroup) FindUniqueCandidate() int {
	candidateCount := make(map[int]int)
	for _, cell := range c {
		if !cell.Confirmed {
			for _, cand := range cell.Candidates {
				candidateCount[cand]++
			}
		}
	}
	for candidate, count := range candidateCount {
		if count == 1 {
			return candidate
		}
	}
	return 0
}

// CheckIfComplete checks if the number of cells in the CellGroup follows the rule of Sudoku,
// which means all numbers should br distinct.
func (c *CellGroup) CheckIfComplete() bool {
	checkSet := [10]bool{false}
	for _, cell := range c {
		checkSet[cell.Number] = true
	}

	if checkSet[0] {
		return false
	}

	for i := 1; i <= 9; i++ {
		if !checkSet[i] {
			return false
		}
	}

	return true
}

// Cell represent a square which contains a single number in Sudoku.
type Cell struct {
	Confirmed  bool
	Number     int
	Candidates []int
	Row        *CellGroup
	Column     *CellGroup
	Region     *CellGroup
}

// Mark marks the status of the cell to confirmed(the answer is comfirmed).
// Mark also triggers candidate elimination to the CellGroups it belongs.
func (c *Cell) Mark(answer int) {
	c.Confirmed = true
	c.Number = answer
	c.Row.TrimCandidate(answer)
	c.Column.TrimCandidate(answer)
	c.Region.TrimCandidate(answer)
}
