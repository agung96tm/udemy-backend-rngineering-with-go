package main

import (
	"log"
	"socialv2/internal/db"
	"socialv2/internal/env"
	"socialv2/internal/store"
)

func main() {
	env.InitEnv()

	addr := env.GetString("DB_ADDR", "postgres://user:password@host:port/db?sslmode=disable")
	conn, err := db.NewDB(addr, 3, 3, "15m")
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	strg := store.NewStorage(conn)
	_ = db.Seed(strg, conn)
}
