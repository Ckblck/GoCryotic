package network

import (
	"os"
	"reflect"
	"regexp"

	"github.com/ckblck/gocryotic/saving"
	"github.com/ckblck/gocryotic/saving/model"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

var zipFilePattern = regexp.MustCompile("REPLAY-([\\w\\d]*)-compressed\\.zip")

// DBName is the name of the Mongo Database.
var DBName string

// AddPlayer adds a player to the mongo database.
// It is normally used when a replay starts, and it needs
// to store the recorded players.
func AddPlayer(c *fiber.Ctx) error {
	player := new(model.RecordedPlayer)

	if err := c.BodyParser(player); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "An error occurred while parsing the body.", "data": err.Error()})
	}

	if !validateStruct(player, c) {
		return nil
	}

	success, error := saving.SavePlayer(player, DBName)

	if success {
		return c.Status(201).JSON(fiber.Map{"status": "success", "message": "Successfully stored player into the database.", "data": player})
	}

	return c.Status(400).JSON(fiber.Map{"status": "error", "message": "An error occurred while trying to store the player into the database.", "data": error.Error()})
}

// GetReplays lists all the replays stored in the local database.
func GetReplays(c *fiber.Ctx) error {
	replays := saving.RetrieveReplays()

	if replays == nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "An error occurred when retrieving the replays from the local database.", "data": nil})
	}

	return c.Status(201).JSON(fiber.Map{"status": "success", "message": "Successfully fetched all the stored replays.", "data": replays})
}

// GetReplay downloads a specific replay.
func GetReplay(c *fiber.Ctx) error {
	fileName := "REPLAY-" + c.Params("id") + "-compressed.zip"

	return c.Download(saving.FolderPath+string(os.PathSeparator)+fileName, fileName)
}

// AddReplay adds a zipped replay to the local database.
func AddReplay(c *fiber.Ctx) error {
	zippedReplay, err := c.FormFile("file")

	if zippedReplay == nil || err != nil {
		return c.Status(422).JSON(fiber.Map{"status": "error", "message": "Body form-data with key 'file' expected but not received.", "data": err.Error()})
	}

	storedReplay := new(model.StoredReplay)

	if err := c.BodyParser(storedReplay); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "An error occurred while parsing the body.", "data": err.Error()})
	}

	if !validateStruct(storedReplay, c) {
		return nil
	}

	zipName := zippedReplay.Filename
	var matchesRegex bool = zipFilePattern.MatchString(zipName)

	if !matchesRegex {
		return c.Status(401).JSON(fiber.Map{"status": "error", "message": "File name: " + zipName + " does not match regex: " + zipFilePattern.String(), "data": nil})
	}

	success, message, statusCode := saving.UploadReplay(storedReplay, DBName)

	if !success {
		return c.Status(statusCode).JSON(fiber.Map{"status": "error", "message": "Could not save file: " + message, "data": nil})
	}

	replayID := zipFilePattern.FindStringSubmatch(zipName)[1]
	err = c.SaveFile(zippedReplay, saving.FolderPath+string(os.PathSeparator)+zipName)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Could not save file: " + zipName + ". (internal error).", "data": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"status": "success", "message": "Successfully saved file: " + zipName + ".", "data": replayID})
}

// DeleteReplay tries to delete a zipped file from the database.
func DeleteReplay(c *fiber.Ctx) error {
	replayName := c.Params("id")
	replayName = "REPLAY-" + replayName + "-compressed.zip"

	// TODO Delete from player and croytic databases.

	success := saving.DeleteReplay(replayName)

	if success {
		return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Successfully deleted file: " + replayName + ".", "data": replayName})
	}

	return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Could not delete file: " + replayName + ".", "data": replayName})
}

func validateStruct(s interface{}, c *fiber.Ctx) bool {
	validate := validator.New()
	err := validate.Struct(s)

	if err != nil {
		errors := len(err.(validator.ValidationErrors))
		properties := make([]string, errors)

		for i, err := range err.(validator.ValidationErrors) {
			field, _ := reflect.TypeOf(s).Elem().FieldByName(err.Field())
			property := "Key '" + field.Tag.Get("form") + "' of type '" + err.Type().Name() + "' required but not received."
			properties[i] = property
		}

		c.Status(422).JSON(fiber.Map{"status": "error", "message": "Incorrect form-data received.", "data": properties})

		return false
	}

	return true
}
