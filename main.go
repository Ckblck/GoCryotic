package main

import (
	"fmt"
	"log"
	"time"

	handler "github.com/ckblck/gocryotic/network"
	external "github.com/ckblck/gocryotic/saving"
	local "github.com/ckblck/gocryotic/saving"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config := Config{}
	ReadConfig(&config)

	handler.DBName = config.Database.DatabaseName
	app := fiber.New()
	err := local.CreateLocalDatabase()
	external.ConnectMongo(config.Database.URI)

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

	app.Get("/api/v1/replay", handler.GetReplays)
	app.Get("/api/v1/replay/:id", handler.GetReplay)
	app.Post("/api/v1/replay", handler.AddReplay)
	app.Post("/api/v1/player", handler.AddPlayer)
	app.Delete("/api/v1/replay/:id", handler.DeleteReplay)
}

func print() {
	fmt.Println("❄️  Cryotic ")
	fmt.Println("⚠️  Do not manually rename/delete any file under the 'replays-storage' folder.")
}
