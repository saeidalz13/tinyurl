package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type ReqUrl struct {
	OriginalUrl string `json:"original_url"`
}

type RespUrl struct {
	ShortenedUrl string `json:"shortened_url"`
}

func HandleTinyUrl(w http.ResponseWriter, r *http.Request) {
	var reqUrl ReqUrl
	if err := json.NewDecoder(r.Response.Body).Decode(&reqUrl); err != nil {
		log.Println(err)
		http.Error(w, "invalid type of json request", http.StatusBadRequest)
		return
	}


	respBytes, err := json.Marshal(RespUrl{ShortenedUrl: reqUrl.OriginalUrl})	
	if err != nil {
		http.Error(w, "response could not be prepared", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respBytes)
	if err != nil {
		log.Println(err)
		return
	}
}

func main() {
	// start a multiplexer
	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("GET /", HandleTinyUrl)

	log.Fatalln(http.ListenAndServe(":7374", mux))
}