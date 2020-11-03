package saving

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ckblck/gocryotic/saving/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

const replaysCollection = "cryotic_replays"
const playersCollection = "cryotic_players"

// UploadReplay saves the replay information to Mongo DB.
// It will return true if the uploading was a success, false otherwise.
// A message will be also returned explaining what gone wrong.
func UploadReplay(replay *model.StoredReplay, databaseName string) (bool, string) {
	if replayExists(replay) {
		return false, "file is already stored in the local database."
	}

	coll := mongoClient.Database(databaseName).Collection(replaysCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	value := coll.FindOne(ctx, bson.M{"identifier": replay.ReplayID})

	if value.Err() == nil { // Replay already added.
		return false, "replay is already stored in the external database."
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := coll.InsertOne(ctx, replay)

	if err != nil {
		log.Fatal(err)

		return false, "an error occurred when inserting the replay into the external database."
	}

	return true, ""
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
		update := bson.M{"$push": bson.M{"replays": player.ReplayID}}

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

func replayExists(replay *model.StoredReplay) bool {
	path := FolderPath + string(os.PathSeparator) + "REPLAY-" + replay.ReplayID + "-compressed.zip"
	_, err := os.Stat(path)

	return os.IsExist(err)
}
