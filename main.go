package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pankona/hashira-auth/google"
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

type memKVS struct {}

func (m *memKVS) Store(bucket, k string, v interface{}) {
	panic("implement me")
}

func (m *memKVS) Load(bucket, k string) (interface{}, bool) {
	panic("implement me")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	kvs := &memKVS{}
	registerGoogle(kvs)
	registerTwitter(kvs)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		a, err := r.Cookie("Authorization")
		if err != nil {
			msg := fmt.Sprintf("No Authorization info found...")
			fmt.Fprintf(w, "Cookies: %v\n", r.Cookies())
			fmt.Fprintf(w, "%s\n", msg)
			return
		}

		userID, ok := userIDByAccessToken[a.Value]
		if !ok {
			msg := fmt.Sprintf("User with id [%s] not found...", a.Value)
			w.Write([]byte(msg))
			return
		}

		user := userByUserID[userID]
		msg := fmt.Sprintf("Hello, %s!", user.name)
		w.Write([]byte(msg))
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func registerGoogle(kvs google.KVStore) {
	var (
		clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
		clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
	)
	g := google.New(clientID, clientSecret, kvs)
	g.Register("/auth/google")
}

func registerTwitter(kvs twitter.KVStore) {
	var (
		consumerKey       = os.Getenv("TWITTER_API_TOKEN")
		consumerSecret    = os.Getenv("TWITTER_API_SECRET")
		accessToken       = os.Getenv("TWITTER_API_ACCESS_TOKEN")
		accessTokenSecret = os.Getenv("TWITTER_API_ACCESS_TOKEN_SECRET")
	)
	t := twitter.New(consumerKey, consumerSecret, accessToken, accessTokenSecret, kvs)
	t.Register("/auth/twitter/")
}
