package main

import (
	"log"
	"socialv3/internal/db"
	"socialv3/internal/store"
)

func main() {
	addr := "postgres://social_user:social_password@localhost:5432/social_db?sslmode=disable"
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	strg := store.NewStorage(conn)
	_ = db.Seed(conn, strg)
}
