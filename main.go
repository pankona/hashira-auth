package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pankona/hashira-auth/twitter"
)

var (
	userIDByIDToken     = make(map[string]string)
	userByUserID        = make(map[string]user)
	userIDByAccessToken = make(map[string]string)
)

type user struct {
	id   string
	name string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	registerTwitter()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func registerTwitter() {
	var (
		consumerKey       = os.Getenv("TWITTER_API_TOKEN")
		consumerSecret    = os.Getenv("TWITTER_API_SECRET")
		accessToken       = os.Getenv("TWITTER_API_ACCESS_TOKEN")
		accessTokenSecret = os.Getenv("TWITTER_API_ACCESS_TOKEN_SECRET")
	)
	t := twitter.New(consumerKey, consumerSecret, accessToken, accessTokenSecret)
	t.RegisterHandler("/auth/twitter/")
}
