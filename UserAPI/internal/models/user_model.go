package models

type User struct {
	UserId		string	`json:"userId" bson:"userId"`
	Email		string	`json:"email" bson:"email"`
	Password	string	`json:"password" bson:"password"`
	Name		string	`json:"name" bson:"name"`
}