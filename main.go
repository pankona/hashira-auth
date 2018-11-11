package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"

	"github.com/pankona/hashira-auth/google"
	"github.com/pankona/hashira-auth/twitter"
)

type memKVS struct {
	userIDByIDToken     map[string]string
	userIDByAccessToken map[string]string
	userByUserID        map[string]user.User
}

func (m *memKVS) Store(bucket, k string, v interface{}) {
	switch bucket {
	case "userIDByIDToken":
		m.userIDByIDToken[k] = v.(string)
	case "userIDByAccessToken":
		m.userIDByAccessToken[k] = v.(string)
	case "userByUserID":
		m.userByUserID[k] = v.(user.User)
	default:
		panic("unknown bucket [" + bucket + "] is specified.")
	}
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

	kvs := &memKVS{
		userIDByIDToken:     make(map[string]string),
		userByUserID:        make(map[string]user.User),
		userIDByAccessToken: make(map[string]string),
	}
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

		userID, ok := kvs.userIDByAccessToken[a.Value]
		if !ok {
			msg := fmt.Sprintf("User with id [%s] not found...", a.Value)
			w.Write([]byte(msg))
			return
		}

		u := kvs.userByUserID[userID]
		msg := fmt.Sprintf("Hello, %s!", u.Name)
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
