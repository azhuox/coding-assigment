package main

import (
	"context"
	"flag"
	"log"

	"github.com/azhuox/code-interviews/koho/account"
)

func main() {
	// Parse args
	inputFile := flag.String("input_file", "", "Input file")
	flag.Parse()
	if *inputFile == "" {
		log.Fatalln("The arg 'input_file' is required")
	}

	var accountManager account.Manager
	accountManager = account.NewManager()

	log.Printf("Start processing transactions in the given input file...\n")

	err := accountManager.ProcessLoadTransactions(context.Background(), *inputFile, "./output.txt")
	if err != nil {
		log.Fatalf("Error processing transactions in the given input file: %s\n", err.Error())
	}

	log.Printf("Sucessfully process transactions in the given input file. " +
		"Please check output file for results.\n")
}
