package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pankona/hashira-auth/google"
	"github.com/pankona/hashira-auth/kvstore"
	"github.com/pankona/hashira-auth/twitter"
	"github.com/pankona/hashira-auth/user"
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
	switch bucket {
	case "userIDByIDToken":
		v, ok := m.userIDByIDToken[k]
		return v, ok
	case "userIDByAccessToken":
		v, ok := m.userIDByAccessToken[k]
		return v, ok
	case "userByUserID":
		v, ok := m.userByUserID[k]
		return v, ok
	default:
		panic("unknown bucket [" + bucket + "] is specified.")
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	env := os.Getenv("GAE_ENV")
	servingBaseURL := "http://localhost:8080"
	if env != "" {
		servingBaseURL = "https://hashira-auth.appspot.com"
	}

	log.Printf("GAE_ENV: %v", env)
	log.Printf("servingBaseURL: %v", servingBaseURL)

	kvs := &memKVS{
		userIDByIDToken:     make(map[string]string),
		userByUserID:        make(map[string]user.User),
		userIDByAccessToken: make(map[string]string),
	}
	registerGoogle(kvs, servingBaseURL)
	registerTwitter(kvs, servingBaseURL)

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

func registerGoogle(kvs kvstore.KVStore, servingBaseURL string) {
	var (
		clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
		clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
	)
	g := google.New(clientID, clientSecret,
		servingBaseURL+"/auth/google/callback", kvs)
	g.Register("/auth/google/")
}

func registerTwitter(kvs kvstore.KVStore, servingBaseURL string) {
	var (
		consumerKey       = os.Getenv("TWITTER_API_TOKEN")
		consumerSecret    = os.Getenv("TWITTER_API_SECRET")
		accessToken       = os.Getenv("TWITTER_API_ACCESS_TOKEN")
		accessTokenSecret = os.Getenv("TWITTER_API_ACCESS_TOKEN_SECRET")
	)
	t := twitter.New(consumerKey, consumerSecret, accessToken, accessTokenSecret,
		servingBaseURL+"/auth/twitter/callback", kvs)
	t.Register("/auth/twitter/")
}
