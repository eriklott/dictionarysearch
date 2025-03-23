package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/eriklott/dictionarysearch/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	var (
		connectionString string
		src              string
	)
	flag.StringVar(&connectionString, "db", "", "postgresql connection string")
	flag.StringVar(&src, "src", "", "The source dictionary textfile path")
	flag.Parse()

	// Open database connection
	db, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		log.Fatalf("failed pgxpool.New: %s", err)
	}
	defer db.Close()

	// Open the file
	file, err := os.Open(src)
	if err != nil {
		log.Fatalf("failed to open source file: %s", err)
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Loop through the file and read each line
	for scanner.Scan() {
		line := scanner.Text() // Get the line as a string

		// Remove comments from line
		commentIndex := strings.Index(line, "#")
		if commentIndex != -1 {
			line = line[:commentIndex-1]
		}

		// Split the word off the front of the line of text
		word, pronunciation, found := strings.Cut(line, " ")
		if !found {
			log.Fatal("failed to find cut string")
		}

		// Get pronunciation number
		pronunciationNum := 1
		if len(word) > 3 {
			last3 := word[len(word)-3:]
			if strings.HasPrefix(last3, "(") {
				word = word[:len(word)-3]
				pronunciationNum, err = strconv.Atoi(string(last3[1]))
				if err != nil {
					log.Fatal("failed to get pronunciation number")
				}
			}
		}

		// Insert word into database
		err := sqlc.New().InsertWord(ctx, db, word)
		if err != nil {
			log.Fatalf("failed sqlc.InsertWord: %s", err)
		}
		// log.Printf("inserted word: %s", word)

		// Split the pronunciation into symbols
		symbols := strings.Split(pronunciation, " ")
		for i, symbol := range symbols {

			// Insert word symbol into database
			insertWordSymbolParams := sqlc.InsertWordSymbolParams{
				Word:             word,
				PronunciationNum: int16(pronunciationNum),
				Position:         int16(i + 1),
				Symbol:           symbol,
			}
			err = sqlc.New().InsertWordSymbol(ctx, db, insertWordSymbolParams)
			if err != nil {
				log.Fatalf("failed sqlc.InsertWordSymbol: %s", err)
			}

			// log.Printf("inserted word_symbol: %+v", insertWordSymbolParams)
		}
	}

	// Check for errors during the scan
	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file: %s", err)
	}
}
