package model

// RecordedPlayer is the struct of the
// player who was recorded in a replay.
type RecordedPlayer struct {
	Nickname string `bson:"nick" form:"nickname" validate:"required"`
	ReplayID string `bson:"replays" form:"replay_id" validate:"required"`
}
