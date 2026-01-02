package main

import (
	"fmt"
	"log"
	"os"

	"lsm-storage-engine/engine"
)

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  SET <key> <value>")
	fmt.Println("  GET <key>")
	fmt.Println("  DEL <key>")
	fmt.Println("  COMPACT")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	db, err := engine.New("store.log")
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "SET":
		if len(os.Args) != 4 {
			log.Fatal("SET requires <key> <value>")
		}
		if err := db.Set(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
		fmt.Println("OK")

	case "GET":
		if len(os.Args) != 3 {
			log.Fatal("GET requires <key>")
		}
		val, err := db.Get(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(val)

	case "DEL":
		if len(os.Args) != 3 {
			log.Fatal("DEL requires <key>")
		}
		if err := db.Delete(os.Args[2]); err != nil {
			log.Fatal(err)
		}
		fmt.Println("OK")

	case "COMPACT":
		if err := db.Compact(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Compaction complete")

	default:
		printUsage()
		log.Fatal("unknown command")
	}
}
