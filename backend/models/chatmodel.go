package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	Id            primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	IsGroupChat   bool                 `json:"isGroupChat" bson:"isGroupChat"` // should default to false
	ChatName      string               `json:"chatName" bson:"chatName"`
	Users         []primitive.ObjectID `json:"users" bson:"users"`
	LatestMessage primitive.ObjectID   `json:"latestMessage" bson:"latestMessage"`
	GroupAdmin    primitive.ObjectID   `json:"groupAdmin" bson:"groupAdmin"`
	Created_at    time.Time            `json:"created_at" bson:"created_at"`
	Updated_at    time.Time            `json:"updated_at" bson:"updated_at"`
}
