package network

import (
	"os"
	"reflect"
	"regexp"

	external "github.com/ckblck/gocryotic/saving"
	local "github.com/ckblck/gocryotic/saving"
	"github.com/ckblck/gocryotic/saving/model"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

var zipFilePattern = regexp.MustCompile("REPLAY-([\\w\\d]*)-compressed\\.zip")

// DBName is the name of the Mongo Database.
var DBName string

type response struct {
	Status     int      `json:"status"`
	Message    string   `json:"message,omitempty"`
	Properties []string `json:"properties,omitempty"`
}

// GetReplays lists all the replays stored in the local database.
func GetReplays(c *fiber.Ctx) error {
	replays := local.RetrieveReplays()

	response := response{Status: 201, Properties: replays}
	c.SendStatus(201)

	return c.JSON(response)
}

// GetReplay downloads a specific replay.
func GetReplay(c *fiber.Ctx) error {
	fileName := "REPLAY-" + c.Params("id") + "-compressed.zip"

	return c.Download(local.FolderPath+string(os.PathSeparator)+fileName, fileName)
}

// AddReplay adds a zipped replay to the local database.
func AddReplay(c *fiber.Ctx) error {
	zippedReplay, err := c.FormFile("file")

	if zippedReplay == nil || err != nil {
		response := response{Status: 400, Message: "Body form-data with key 'file' expected but not received."}
		sendResponse(c, &response, 400)

		return err
	}

	storedReplay := new(model.StoredReplay)

	if err := c.BodyParser(storedReplay); err != nil {
		response := response{Status: 400, Message: "An error occurred while parsing the body."}
		sendResponse(c, &response, 400)

		return err
	}

	if !validateStruct(storedReplay, c) {
		return nil
	}

	zipName := zippedReplay.Filename
	var matchesRegex bool = zipFilePattern.MatchString(zipName)

	if !matchesRegex {
		response := response{Status: 400, Message: "File name: " + zipName + " does not match regex: " + zipFilePattern.String()}
		sendResponse(c, &response, 400)

		return nil
	}

	success, message := external.UploadReplay(storedReplay, DBName)

	if !success {
		response := response{Status: 400, Message: "Could not save file: " + message}
		sendResponse(c, &response, 400)

		return nil
	}

	replayID := zipFilePattern.FindStringSubmatch(zipName)[1]
	err2 := c.SaveFile(zippedReplay, local.FolderPath+string(os.PathSeparator)+zipName)

	if err2 != nil {
		response := response{Status: 400, Message: "Could not save file: " + zipName + ". (internal error)", Properties: []string{replayID}}
		sendResponse(c, &response, 400)

		return err2
	}

	response := response{Status: 201, Message: "Successfully saved file: " + zipName + ".", Properties: []string{replayID}}
	sendResponse(c, &response, 201)

	return nil
}

// DeleteReplay tries to delete a zipped file from the database.
func DeleteReplay(c *fiber.Ctx) error {
	replayName := c.Params("id")
	replayName = "REPLAY-" + replayName + "-compressed.zip"

	success := local.DeleteReplay(replayName)

	if success {
		response := response{Status: 201, Message: "Successfully deleted file: " + replayName + "."}

		return c.JSON(response)
	}

	response := response{Status: 400, Message: "Could not delete file: " + replayName + "."}

	return c.JSON(response)
}

func sendResponse(c *fiber.Ctx, r *response, statusCode int) {
	c.SendStatus(statusCode)
	c.JSON(r)
}

func validateStruct(s *model.StoredReplay, c *fiber.Ctx) bool {
	validate := validator.New()
	err := validate.Struct(s)

	if err != nil {
		errors := len(err.(validator.ValidationErrors))
		properties := make([]string, errors)

		for i, err := range err.(validator.ValidationErrors) {
			field, _ := reflect.TypeOf(s).Elem().FieldByName(err.Field())
			property := "Body key '" + field.Tag.Get("form") + "' of type '" + err.Type().Name() + "' required but not received."
			properties[i] = property
		}

		response := response{Status: 400, Message: "Incorrect body received.", Properties: properties}
		sendResponse(c, &response, 400)

		return false
	}

	return true
}
