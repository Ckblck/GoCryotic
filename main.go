package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ckblck/gocryotic/network"
	"github.com/ckblck/gocryotic/saving"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config := Config{}
	ReadConfig(&config)

	network.DBName = config.Database.DatabaseName
	app := fiber.New()
	err := saving.CreateLocalDatabase()
	saving.ConnectMongo(config.Database.URI)

	if err != nil {
		panic(err)
	}

	Routes(app)
	time.AfterFunc(500*time.Millisecond, print)

	log.Fatal(app.Listen(config.Server.Host + ":" + config.Server.Port))
}

// Routes creates the REST-API routes.
func Routes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("👋")
	})

	app.Get("/api/v1/replay", network.GetReplays)
	app.Get("/api/v1/replay/:id", network.GetReplay)
	app.Post("/api/v1/replay", network.AddReplay)
	app.Post("/api/v1/player", network.AddPlayer)
	app.Get("/api/v1/player/:name", network.GetPlayerReplays)
	app.Delete("/api/v1/replay/:id", network.DeleteReplay)
}

func print() {
	fmt.Println("❄️  Cryotic ")
	fmt.Println("⚠️  Do not manually rename/delete any file under the 'replays-storage' folder.")
}
