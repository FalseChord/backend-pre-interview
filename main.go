package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Open the file, read from file, initialize the grids.
// Solve the puzzle, Record and print the answer.
func main() {
	sourceFile, err := os.Open("./sudoku.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer sourceFile.Close()

	sourceFileScanner := bufio.NewScanner(sourceFile)

	var rawNumbereGrid [9][9]int
	var rowCount int
	var gridTitle string
	var totalSum int

	for sourceFileScanner.Scan() {
		line := sourceFileScanner.Text()

		if strings.HasPrefix(line, "Grid") {
			gridTitle = line
			rowCount = 0
			rawNumbereGrid = [9][9]int{{0}}
			continue
		}

		splitedString := strings.Split(line, "")
		var rawNumberRow [9]int
		for charIndex, splitedChar := range splitedString {
			number, err := strconv.Atoi(splitedChar)
			if err != nil {
				log.Fatal("Fail convert string to int.")
			}
			rawNumberRow[charIndex] = number
		}

		rawNumbereGrid[rowCount] = rawNumberRow
		rowCount++

		if rowCount == len(rawNumbereGrid) {
			grid := Grid{}
			grid.Init(rawNumbereGrid)
			grid = Solve(grid)

			topLeftSum := grid.Rows[0][0].Number + grid.Rows[0][1].Number + grid.Rows[0][2].Number

			fmt.Printf(" %s ", gridTitle)
			grid.Print()
			fmt.Printf("Sum of the first three numbers in the top row: %d \n\n", topLeftSum)

			totalSum += topLeftSum
		}
	}

	fmt.Printf("Sum of all Grids' first three numbers in the top row: %d \n\n", totalSum)

	if err = sourceFileScanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// Solve solves Sudoku puzzle.
// It complete the deterministic part of the puzzle at the first.
// If the puzzle still not solved, then try to guess from possible candidates.
// Loop through the check and quess cycle until all number are confirmed.
func Solve(grid Grid) Grid {
	statusStack := make([]GridStatus, 0)

	grid.Check()

	for {
		verifyResult := grid.Verify()

		// Grid complete, done.
		if verifyResult == GridStatusCompleted {
			break
		}

		// If the grid is not complete and the status is good according to the rule,
		// pick a candidate to guess, and record grid's current status for possible rollback.
		if verifyResult == GridStatusNotCompleted {
			gridToPreserve := Grid{}
			rawNumberGrid := grid.ToRawNumberGrid()
			gridToPreserve.Init(rawNumberGrid)

			rowIndex, cellIndex := grid.GetBranchCellIndex()
			remainingCandidates := grid.Rows[rowIndex][cellIndex].Candidates
			branchedCandidate := remainingCandidates[0]
			remainingCandidates = remainingCandidates[1:]
			grid.Rows[rowIndex][cellIndex].Candidates = []int{branchedCandidate}

			statusStack = append(statusStack, GridStatus{
				Grid:                gridToPreserve,
				BranchCellRowIndex:  rowIndex,
				BranchCellIndex:     cellIndex,
				RemainingCandidates: remainingCandidates,
			})
		}

		// If the guess is wrong, pop the stack until find another candidate we havn't go through.
		if verifyResult == GridStatusMalformed {
			for {
				if len(statusStack) == 0 {
					log.Fatal("Unexpect Situation: Ran Out Possible Solutions")
				}

				lastPosition := len(statusStack) - 1
				gridStatus := statusStack[lastPosition]
				statusStack[lastPosition] = GridStatus{}
				statusStack = statusStack[:lastPosition]

				if len(gridStatus.RemainingCandidates) == 0 {
					continue
				}

				grid = gridStatus.Grid
				remainingCandidates := gridStatus.RemainingCandidates
				branchedCandidate := remainingCandidates[0]
				gridStatus.RemainingCandidates = remainingCandidates[1:]

				rowIndex := gridStatus.BranchCellRowIndex
				cellIndex := gridStatus.BranchCellIndex
				grid.Rows[rowIndex][cellIndex].Candidates = []int{branchedCandidate}

				break
			}
		}

		grid.Check()
	}

	return grid
}

// GridStatus is a status record of the grid and the guessing.
type GridStatus struct {
	Grid                Grid
	BranchCellRowIndex  int
	BranchCellIndex     int
	RemainingCandidates []int
}
