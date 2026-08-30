package main

import "log"

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatalf("err open database: %v", err)
	}
	defer db.Close()
}
