package boardgame

import "fmt"

const (
	Player1    Player = 1
	Player2    Player = 2
	PlayerNone Player = 0
)

type Player int

// Convert given player to an integer.
func (p Player) Int() int {
	return int(p)
}

type Game struct {
	rows         int
	cols         int
	winThreshold int
	grid         [][]int
	winner       Player
}

// NewGame - create a new game.
func NewGame(rows, cols, winThreshold int) (*Game, error) {
	if rows <= 0 || cols <= 0 || winThreshold <= 0 {
		return nil, fmt.Errorf("invalid input")
	}

	grid := make([][]int, rows)
	for i := 0; i < rows; i++ {
		grid[i] = make([]int, cols)
	}

	return &Game{
		rows:         rows,
		cols:         cols,
		winThreshold: winThreshold,
		grid:         grid,
		winner:       PlayerNone,
	}, nil
}

// Play - given player makes his move. Return true if the game is end.
func (g *Game) Play(player Player, col int) (bool, error) {
	if col < 0 || col > g.cols {
		return false, fmt.Errorf("invalid input %d from player %d\n", col, player)
	}

	// Drop the discs
	lowest := g.lowestPos(col)
	g.grid[lowest][col] = player.Int()

	// print the game board
	g.printBoard()

	// Detect whether the current player won the game.
	if g.won(player, lowest, col) {
		g.winner = player
		return true, nil
	}

	return false, nil
}

// Winner - print out winner
func (g *Game) Winner() Player {
	return g.winner
}

func (g *Game) lowestPos(col int) int {
	for i := g.rows - 1; i >= 0; i-- {
		if g.grid[i][col] == 0 {
			return i
		}
	}

	return -1
}

func (g *Game) printBoard() {
	for i := 0; i < g.rows; i++ {
		fmt.Printf("%v\n", g.grid[i])
	}
}

func (g *Game) won(player Player, currentRow, currentCol int) bool {
	winThreshold := g.winThreshold - 1

	// Check left
	nextCol := currentCol - 1
	if nextCol+1 >= winThreshold {
		count := winThreshold
		for i := nextCol; ; i-- {
			if g.grid[currentRow][i] != player.Int() {
				break
			}
			count--
			if count == 0 {
				return true
			}
		}
	}

	// Check right
	nextCol = currentCol + 1
	if g.cols-nextCol >= winThreshold {
		count := winThreshold
		for i := nextCol; ; i++ {
			if g.grid[currentRow][i] != player.Int() {
				break
			}
			count--
			if count == 0 {
				return true
			}
		}
	}

	// Check down
	nextRow := currentRow + 1
	if g.rows-nextRow >= winThreshold {
		count := winThreshold
		for i := nextRow; ; i++ {
			if g.grid[i][currentCol] != player.Int() {
				break
			}
			count--
			if count == 0 {
				return true
			}
		}
	}

	// Check left up
	nextRow = currentRow - 1
	nextCol = currentCol - 1
	if nextRow+1 >= winThreshold && nextCol+1 >= winThreshold {
		count := winThreshold
		for {
			if g.grid[nextRow][nextCol] != player.Int() {
				break
			}
			nextRow--
			nextCol--
			count--
			if count == 0 {
				return true
			}
		}
	}

	// Check left down
	nextRow = currentRow + 1
	nextCol = currentCol - 1
	if g.rows-nextRow >= winThreshold && nextCol+1 >= winThreshold {
		count := winThreshold
		for {
			if g.grid[nextRow][nextCol] != player.Int() {
				break
			}
			nextRow++
			nextCol--
			count--
			if count == 0 {
				return true
			}
		}
	}

	// Check right up
	nextRow = currentRow - 1
	nextCol = currentCol + 1
	if nextRow+1 >= winThreshold && g.cols-nextCol >= winThreshold {
		count := winThreshold
		for {
			if g.grid[nextRow][nextCol] != player.Int() {
				break
			}
			nextRow--
			nextCol++
			count--
			if count == 0 {
				return true
			}
		}
	}

	// Check right down
	nextRow = currentRow + 1
	nextCol = currentCol + 1
	if g.rows-nextRow >= winThreshold && g.cols-nextCol >= winThreshold {
		count := winThreshold
		for {
			if g.grid[nextRow][nextCol] != player.Int() {
				break
			}
			nextRow++
			nextCol++
			count--
			if count == 0 {
				return true
			}
		}
	}

	return false
}
