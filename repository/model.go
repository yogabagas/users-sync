package repository

import "time"

type LogData struct {
	NIK         string `bson:"nik"`
	Status      int    `bson:"status"`
	Description string `bson:"description"`
}

type UserData struct {
	NIK         string    `bson:"nik"`
	Name        string    `bson:"name"`
	Role        string    `bson:"role"`
	Directorate string    `bson:"directorate"`
	Status      int       `bson:"status"`
	Description string    `bson:"description"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}
