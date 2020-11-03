package model

// StoredReplay is the struct of the
// replay information that will be saved in the Mongo database.
type StoredReplay struct {
	WorldName     string `bson:"world_name" form:"world_name" validate:"required"`
	ReplayID      string `bson:"identifier" form:"identifier" validate:"required"`
	FullWorldName string `bson:"full_world_name" form:"full_world_name" validate:"required"`
	EpochDate     uint64 `bson:"epoch" form:"epoch" validate:"required"`
	MaxTicks      uint64 `bson:"ticks" form:"ticks" validate:"required"`
}
