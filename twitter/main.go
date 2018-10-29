package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
)

var (
	rootDir           string
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
	credential        *oauth.Credentials
	tc                *anaconda.TwitterApi
)

func RequestTokenHandler(w http.ResponseWriter, r *http.Request) {
	url, tmpCred, err := tc.AuthorizationURL("http://localhost:8080/auth/twitter/callback")
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}
	credential = tmpCred
	http.Redirect(w, r, url, http.StatusFound)
}

func AccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	c, _, err := tc.GetCredentials(credential, r.URL.Query().Get("oauth_verifier"))
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	cli := anaconda.NewTwitterApiWithCredentials(c.Token, c.Secret, consumerKey, consumerSecret)

	v := url.Values{}
	v.Set("include_entities", "true")
	v.Set("skip_status", "true")
	u, err := cli.GetSelf(v)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	// TODO: use u.IdStr to identify user
	fmt.Fprintf(w, "%v", u.IdStr)
}

func main() {
	consumerKey = os.Getenv("TWITTER_API_TOKEN")
	consumerSecret = os.Getenv("TWITTER_API_SECRET")
	accessToken = os.Getenv("TWITTER_API_ACCESS_TOKEN")
	accessTokenSecret = os.Getenv("TWITTER_API_ACCESS_TOKEN_SECRET")
	if consumerKey == "" || consumerSecret == "" ||
		accessToken == "" || accessTokenSecret == "" {
		panic("not enough parameter")
	}

	tc = anaconda.NewTwitterApiWithCredentials(
		accessToken, accessTokenSecret,
		consumerKey, consumerSecret)

	http.HandleFunc("/auth/twitter", RequestTokenHandler)
	http.HandleFunc("/auth/twitter/callback", AccessTokenHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
