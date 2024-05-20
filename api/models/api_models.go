package models

type ReqUrl struct {
	OriginalUrl string `json:"original_url"`
}

type RespShortenedUrl struct {
	ShortenedUrl string `json:"shortened_url"`
}

type RespOriginalUrl struct {
	OriginalUrl string `json:"original_url"`
}
