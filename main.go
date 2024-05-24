package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/saeidalz13/tinyurl/api/db"
	"github.com/saeidalz13/tinyurl/api/frontend"
	"github.com/saeidalz13/tinyurl/api/models"
	"github.com/saeidalz13/tinyurl/internal/url"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	port        string
	mux         *http.ServeMux
	Collection  *mongo.Collection
	RedisClient *redis.Client
}

func (s *Server) searchRedis(ctx context.Context, key string) (models.UrlDocument, error) {
	var resultFound models.UrlDocument
	foundShortUrl, err := s.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return resultFound, err
	}
	if err := json.Unmarshal(foundShortUrl, &resultFound); err != nil {
		return resultFound, err
	}
	return resultFound, nil
}

func (s *Server) HandleGetUrl(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	shortUrl := r.PathValue("shortUrl")
	log.Println("Short URL Request:", shortUrl)
	var originalUrl string

	resultFound, err := s.searchRedis(ctx, shortUrl)
	if err == nil {
		originalUrl = resultFound.OriginalUrl

	} else {
		log.Println("Redis error:", err)

		var resultFoundMongo models.UrlDocument
		filter := bson.M{"shortened_url": shortUrl}
		err = s.Collection.FindOne(context.Background(), filter).Decode(&resultFoundMongo)
		if err == nil {
			originalUrl = resultFoundMongo.OriginalUrl

		} else {
			log.Println(err)
			http.Error(w, "Invalid URL", http.StatusNotFound)
			return
		}
	}

	originalUrl = "https://www." + originalUrl 
	log.Println("Original URL Found:", originalUrl)
	http.Redirect(w, r, originalUrl, http.StatusSeeOther)
}

func (s *Server) HandleShortenUrl(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	originalUrl := r.FormValue("link")
	if originalUrl == "" {
		log.Println("nothing received from the form input")
	} else {
		log.Println("original url from form:", originalUrl)
	}

	// Look into the request and normalize the original URL
	// var reqUrl models.ReqUrl
	// if err := json.NewDecoder(r.Body).Decode(&reqUrl); err != nil {
	// 	log.Println(err)
	// 	http.Error(w, "invalid type of json request", http.StatusBadRequest)
	// 	return
	// }
	normalizedUrl := url.NormalizeURL(originalUrl)
	log.Println("incoming request URL:", normalizedUrl)

	// Declare necessary vars
	var shortenedUrl string
	var resultFound models.UrlDocument

	resultFound, err := s.searchRedis(ctx, normalizedUrl)
	if err == nil {
		shortenedUrl = resultFound.ShortenedUrl

	} else {
		log.Printf("Redis error: %s\n", err)

		// Search MongoDB to check for available short url
		filter := bson.M{"original_url": normalizedUrl}
		err = s.Collection.FindOne(ctx, filter).Decode(&resultFound)
		if err == nil {
			log.Println("shortened already in database")
			shortenedUrl = resultFound.ShortenedUrl

		} else {
			if err != mongo.ErrNoDocuments {
				// TODO: Needs more detailed error handling
				log.Println(err)
			}

			// Prepare shortened URL and object to store
			doc := models.UrlDocument{
				OriginalUrl:  normalizedUrl,
				ShortenedUrl: url.ShortenURL(normalizedUrl),
				CreatedAt:    time.Now(),
				ExpiredAt:    time.Now().Add(time.Hour),
			}
			log.Println("Shortened URL:", doc.ShortenedUrl)

			// Add the object to cache
			if err := s.RedisClient.HSet(ctx, doc.ShortenedUrl, doc).Err(); err != nil {
				// TODO: Needs more detailed error handling
				log.Println(err)
			}

			// Add object to MongoDB
			_, err := s.Collection.InsertOne(ctx, doc)
			if err != nil {
				// TODO: Needs more detailed error handling
				log.Println(err)
				// http.Error(w, "failed to record the url", http.StatusInternalServerError)
				// return
			}
			shortenedUrl = doc.ShortenedUrl
		}

	}

	// preparing response
	respBytes, err := json.Marshal(models.RespShortenedUrl{ShortenedUrl: "http://localhost:7374/"+shortenedUrl})
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

func (s *Server) HandleHome(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Saeid"
	}
	frontend.Hello(name).Render(context.Background(), w)
}

func main() {
	// Load the env vars
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	mongoUri := os.Getenv("MONGO_URI")
	redisUri := os.Getenv("REDIS_URI")
	port := os.Getenv("PORT")

	// Starting databases
	rdb := db.MustConnectToRedis(redisUri)
	client, coll := db.MustConnectToDb(mongoUri)
	defer db.DisconnectClient(client)
	defer rdb.Close()

	// Initialize server
	server := Server{
		port:        port,
		mux:         http.NewServeMux(),
		Collection:  coll,
		RedisClient: rdb,
	}

	// Handlers
	server.mux.HandleFunc("GET /", server.HandleHome)
	server.mux.HandleFunc("GET /{shortUrl}", server.HandleGetUrl)
	server.mux.HandleFunc("POST /shorten-url", server.HandleShortenUrl)

	// Start the server
	log.Printf("listening to port %s...\n", server.port)
	log.Fatalln(http.ListenAndServe(":"+server.port, server.mux))
}
