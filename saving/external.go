package saving

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ckblck/gocryotic/saving/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

const replaysCollection = "cryotic_replays"
const playersCollection = "cryotic_players"

// FetchPlayerReplays returns a string array of the replays
// that a certain player has. It will return false if the retrieving failed.
func FetchPlayerReplays(playerName, databaseName string) (int, string, []model.StoredReplay) {
	type player struct {
		Nickname string   `bson:"nick"`
		Replays  []string `bson:"replays"`
	}

	playerDocument := new(player)
	coll := mongoClient.Database(databaseName).Collection(playersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"nick": playerName}
	value := coll.FindOne(ctx, filter)

	if value.Err() == mongo.ErrNoDocuments {
		return 404, "The requested player could not be found in the database.", nil
	}

	err := value.Decode(playerDocument)

	if err != nil {
		return 500, "Could not decode document. Internal server error.", nil
	}

	replaysIDs := playerDocument.Replays
	success, replays := getReplay(databaseName, replaysIDs)

	if !success {
		return 500, "Internal server error occurred while parsing documents.", nil
	}

	message := "Successfully retrieved " + strconv.Itoa(len(replays)) + " replays from player " + playerName + "."

	return 200, message, replays
}

// UploadReplay saves the replay information to Mongo DB.
// It will return true if the uploading was a success, false otherwise.
// A message will also be returned explaining what gone wrong.
// A status code will be returned.
func UploadReplay(replay *model.StoredReplay, databaseName string) (bool, string, int) {
	if replayExists(replay) {
		return false, "file is already stored in the local database.", 409
	}

	coll := mongoClient.Database(databaseName).Collection(replaysCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	value := coll.FindOne(ctx, bson.M{"identifier": replay.ReplayID})

	if value.Err() == nil {
		return false, "replay is already stored in the external database.", 409
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := coll.InsertOne(ctx, replay)

	if err != nil {
		log.Fatal(err)

		return false, "an error occurred when inserting the replay into the external database.", 500
	}

	return true, "", 201
}

// ConnectMongo attemps to stablish the initial connection.
// The collection will be created if it is the first time connecting.
func ConnectMongo(URI string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URI))

	if err != nil {
		panic(err)
	}

	mongoClient = client
}

// SavePlayer will add the specified player to the database.
func SavePlayer(player *model.RecordedPlayer, databaseName string) (bool, error) {
	type mongoPlayer struct {
		Nickname string    `bson:"nick"`
		Replays  [1]string `bson:"replays"`
	}

	coll := mongoClient.Database(databaseName).Collection(playersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	value := coll.FindOne(ctx, bson.M{"nick": player.Nickname})

	defer cancel()

	if value.Err() == nil { // Player already stored, push to its array.
		filter := bson.M{"nick": player.Nickname}
		update := bson.M{
			"$push": bson.M{"replays": player.ReplayID},
		}

		coll.UpdateOne(ctx, filter, update)

		return true, nil
	}

	storedPlayer := mongoPlayer{
		Nickname: player.Nickname,
		Replays:  [1]string{player.ReplayID},
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := coll.InsertOne(ctx, storedPlayer)

	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteReplayFromCollection will try to delete a replay from the "cryotic_replays" collection.
// A string is returning informing the result of the operation.
func DeleteReplayFromCollection(replayID, databaseName string) string {
	coll := mongoClient.Database(databaseName).Collection(replaysCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"identifier": replayID}

	result, err := coll.DeleteOne(ctx, filter)

	if err != nil {
		return "Could not remove such replayID from the replays collection. Error: " + err.Error()
	}

	if result.DeletedCount == 0 {
		return "No replay was removed from the collection, as there was not any."
	}

	return "Successfully removed replayID from the replays collection."
}

// DeleteReplayFromPlayersTrackers will try to delete the from the "cryotic_players" collection.
// It will search for all the players which participated with such replayID and remove it.
// A string is returning informing the result of the operation.
func DeleteReplayFromPlayersTrackers(replayID, databaseName string) string {
	coll := mongoClient.Database(databaseName).Collection(playersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"replays": replayID}
	update := bson.M{"$pull": filter}
	result, err := coll.UpdateMany(ctx, filter, update)

	if err != nil {
		return "Could not remove such replayID from player trackers. Error: " + err.Error()
	}

	if result.ModifiedCount == 0 {
		return "No replay was removed from player trackers, as there was not any."
	}

	return "Successfully removed " + strconv.FormatInt(result.ModifiedCount, 10) + " replays from player trackers."
}

func getReplay(databaseName string, replayIDs []string) (bool, []model.StoredReplay) {
	coll := mongoClient.Database(databaseName).Collection(replaysCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"identifier": bson.M{"$in": replayIDs}}
	cursor, err := coll.Find(ctx, filter)

	if err != nil {
		return false, nil
	}

	var replays []model.StoredReplay
	err = cursor.All(ctx, &replays)

	if err != nil {
		return false, nil
	}

	return true, replays
}

// GetReplayWithID will try to retrieve a replay from the database.
func GetReplayWithID(databaseName string, replayID string) (*model.StoredReplay, error) {
	coll := mongoClient.Database(databaseName).Collection(replaysCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"identifier": replayID}
	res := coll.FindOne(ctx, filter)

	if res.Err() != nil {
		return nil, res.Err()
	}

	replay := new(model.StoredReplay)

	if err := res.Decode(replay); err != nil {
		return replay, err
	}

	return replay, nil
}

func FetchReplays(databaseName string) ([]model.StoredReplay, error) {
	coll := mongoClient.Database(databaseName).Collection(replaysCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.D{}
	cursor, err := coll.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	var replays []model.StoredReplay

	if err := cursor.All(ctx, &replays); err != nil {
		return replays, err
	}

	return replays, nil
}

func replayExists(replay *model.StoredReplay) bool {
	path := FolderPath + string(os.PathSeparator) + "REPLAY-" + replay.ReplayID + "-compressed.zip"
	_, err := os.Stat(path)

	return os.IsExist(err)
}
