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

const collectionName = "cryotic_collection"

// UploadReplay saves the replay information to Mongo DB.
// It will return true if the uploading was a success, false otherwise.
// A message will be also returned explaining what gone wrong.
func UploadReplay(replay *model.StoredReplay, databaseName string) (bool, string) {
	if replayExists(replay) {
		return false, "file is already stored in the local database."
	}

	coll := mongoClient.Database(databaseName).Collection(collectionName)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	value := coll.FindOne(ctx2, bson.M{"identifier": replay.ReplayID})

	if value.Err() == nil { // Replay already added.
		return false, "replay is already stored in the external database."
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func replayExists(replay *model.StoredReplay) bool {
	path := FolderPath + string(os.PathSeparator) + "REPLAY-" + replay.ReplayID + "-compressed.zip"
	_, err := os.Stat(path)

	return os.IsExist(err)
}
