package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName string             `bson:"first_name,omitempty" json:"first_name,omitempty"`
	LastName  string             `bson:"last_name,omitempty" json:"last_name,omitempty"`
	Email     string             `bson:"email,omitempty" json:"email,omitempty"`
	Password  string             `bson:"password,omitempty" json:"password,omitempty"`
	Tokens    []string           `bson:"tokens,omitempty" json:"tokens,omitempty"`
}
