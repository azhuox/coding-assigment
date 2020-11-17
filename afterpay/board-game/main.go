package main

import (
	"bufio"
	"github.com/azhuox/coding-assigment/afterpay/board-game/boardgame"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	game, _ := boardgame.NewGame(6, 7, 4)
	stdinReader := bufio.NewReader(os.Stdin)

	var player boardgame.Player
	count := 0
	for {
		if count&0x1 == 0 {
			player = boardgame.Player1
		} else {
			player = boardgame.Player2
		}
		count++

		log.Printf("Input your move, player %d:\n", player.Int())
		cmdString, err := stdinReader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input from stdin: %s\n", err.Error())
		}

		col, err := strconv.Atoi(strings.Trim(cmdString, "\n"))
		if err != nil {
			log.Fatalf("Invalid input from player %d\n: %s", player, err.Error())
		}

		won, _ := game.Play(player, col)
		if won {
			log.Printf("player %d won the game", game.Winner())
			break
		}

	}

}
