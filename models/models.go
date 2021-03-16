package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Kot struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	MyID          int32              `json:"id" bson:"id"`
	Url           string             `json:"url" bson:"url"`
	CompressedUrl string             `json:"compressed_url" bson:"compressed_url"`
}
