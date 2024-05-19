package main

import (
	"log/slog"
	"time"
	"twitch_bot_v3/database"
	"twitch_bot_v3/server"
)

func main() {
	slog.Info("Hi! I'm AbevBot v3 :)")
	defer cleanup()

	// Start
	database.Init()
	server.Init()

	// "Main loop"
	for {
		time.Sleep(time.Second)
	}
}

func cleanup() {
	slog.Info("Closing the bot")
	database.Close()
}
