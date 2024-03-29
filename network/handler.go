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

// GetPlayerReplays handles the GET request of getting a player's replays.
func GetPlayerReplays(c *fiber.Ctx) error {
	playerName := c.Params("name")
	statusCode, message, replays := saving.FetchPlayerReplays(playerName, DBName)
	status := "success"

	if statusCode != 200 {
		status = "error"
	}

	return c.Status(statusCode).JSON(fiber.Map{"status": status, "message": message, "data": replays})
}

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
	replays, err := saving.FetchReplays(DBName)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "An error occurred when retrieving the replays from the database.", "data": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Successfully fetched all the stored replays.", "data": replays})
}

func DownloadReplay(c *fiber.Ctx) error {
	replayID := c.Params("id")
	fileName := "REPLAY-" + replayID + "-compressed.zip"

	return c.Download(saving.FolderPath+string(os.PathSeparator)+fileName, fileName)
}

// GetReplay downloads a specific replay.
func GetReplay(c *fiber.Ctx) error {
	replayID := c.Params("id")

	var status int = 200
	var statusMsg string = "success"
	var message string = "Successfully retrieved information."
	var data interface{}

	replay, err := saving.GetReplayWithID(DBName, replayID)
	data = replay

	if err != nil {
		status = 404
		statusMsg = "error"
		message = "Replay information is not stored in the database."
		data = err.Error()
	}

	return c.Status(status).JSON(fiber.Map{"status": statusMsg, "message": message, "data": data})
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
	replayID := c.Params("id")
	replayName := "REPLAY-" + replayID + "-compressed.zip"

	success := saving.DeleteReplay(replayName)
	resultMessage1 := saving.DeleteReplayFromCollection(replayID, DBName)
	resultMessage2 := saving.DeleteReplayFromPlayersTrackers(replayID, DBName)

	data := [2]string{resultMessage1, resultMessage2}

	if success {
		return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Successfully deleted file: " + replayName + ".", "data": data})
	}

	return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Could not delete file: " + replayName + ".", "data": data})
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
