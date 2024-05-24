package models

import "time"

type UrlDocument struct {
	OriginalUrl  string    `bson:"original_url" redis:"original_url"`
	ShortenedUrl string    `bson:"shortened_url" redis:"shortened_url"`
	CreatedAt    time.Time `bson:"created_at" redis:"created_at"`
	ExpiredAt    time.Time `bson:"expired_at" redis:"expired_at"`
}
