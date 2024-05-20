package models

import "time"

type UrlDocument struct {
	OriginalUrl  string    `bson:"original_url"`
	ShortenedUrl string    `bson:"shortened_url"`
	CreatedAt    time.Time `bson:"created_at"`
	ExpirtedAt   time.Time `bson:"expired_at"`
}
