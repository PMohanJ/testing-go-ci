package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Id         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Sender     primitive.ObjectID `json:"sender" bson:"sender"`
	Content    string             `json:"content" bson:"content"`
	Chat       primitive.ObjectID `json:"chat" bson:"chat"`
	IsEdited   bool               `json:"isedited" bson:"isedited"`
	Created_at time.Time          `json:"created_at" bson:"created_at"`
	Updated_at time.Time          `json:"updated_at" bson:"updated_at"`
}
