package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/saeidalz13/tinyurl/api/db"
	"github.com/saeidalz13/tinyurl/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	port       string
	Collection *mongo.Collection
}

func shortenUrl(originalUrl string) string {
	hash := sha256.Sum256([]byte(originalUrl))
	return base64.URLEncoding.EncodeToString(hash[:][:7])
}

func normalizeURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		return url[len("http://"):]
	}
	if strings.HasPrefix(url, "https://") {
		return url[len("https://"):]
	}
	return url
}

func (s *Server) HandleGetUrl(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.PathValue("shortUrl")
	filter := bson.M{"shortened_url": shortUrl}

	var resultFound models.UrlDocument
	err := s.Collection.FindOne(context.Background(), filter).Decode(&resultFound)

	if err == nil {
		log.Printf("%+v\n", resultFound)
		respBytes, err := json.Marshal(models.RespOriginalUrl{OriginalUrl: resultFound.OriginalUrl})
		if err != nil {
			log.Println(err)
			http.Error(w, "response could not be prepared", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(respBytes)
		if err != nil {
			log.Println(err)
			return
		}

	} else {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleTinyUrl(w http.ResponseWriter, r *http.Request) {
	var reqUrl models.ReqUrl
	if err := json.NewDecoder(r.Body).Decode(&reqUrl); err != nil {
		log.Println(err)
		http.Error(w, "invalid type of json request", http.StatusBadRequest)
		return
	}

	shortenedUrl := ""
	normalizedUrl := normalizeURL(reqUrl.OriginalUrl)
	filter := bson.M{"original_url": normalizedUrl}
	var resultFound models.UrlDocument
	err := s.Collection.FindOne(context.Background(), filter).Decode(&resultFound)

	if err == nil {
		log.Println("shortened already in database")
		shortenedUrl = resultFound.ShortenedUrl

	} else {
		if err != mongo.ErrNoDocuments {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert into database
		doc := models.UrlDocument{
			OriginalUrl:  normalizedUrl,
			ShortenedUrl: shortenUrl(normalizedUrl),
			CreatedAt:    time.Now(),
			ExpirtedAt:   time.Now().Add(time.Hour),
		}
		result, err := s.Collection.InsertOne(context.Background(), doc)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to record the url", http.StatusInternalServerError)
			return
		}
		log.Printf("%+v\n", result)
		shortenedUrl = doc.ShortenedUrl
	}

	// preparing response
	respBytes, err := json.Marshal(models.RespShortenedUrl{ShortenedUrl: shortenedUrl})
	if err != nil {
		log.Println(err)
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
	// Load the env vars
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	mongoUri := os.Getenv("MONGO_URI")
	client, coll := db.MustConnectToDb(mongoUri)
	server := Server{
		port:       "7374",
		Collection: coll,
	}
	defer db.DisconnectClient(client)

	// start a multiplexer
	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("GET /{shortUrl}", server.HandleGetUrl)
	mux.HandleFunc("POST /shorten-url", server.HandleTinyUrl)

	// Start the server
	log.Printf("listening to port %s...\n", server.port)
	log.Fatalln(http.ListenAndServe(":"+server.port, mux))
}
