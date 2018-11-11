package twitter

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
)

type KVStore interface {
	Store(bucket, k string, v interface{})
	Load(bucket, k string) (interface{}, bool)
}

type Twitter struct {
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
	credential        *oauth.Credentials
	client            *anaconda.TwitterApi
	kvstore           KVStore
}

func New(consumerKey, consumerSecret,
	accessToken, accessTokenSecret string,
	kvstore KVStore) *Twitter {
	if consumerKey == "" || consumerSecret == "" ||
		accessToken == "" || accessTokenSecret == "" {
		panic("not enough parameter")
	}

	t := &Twitter{
		consumerKey:       consumerKey,
		consumerSecret:    consumerSecret,
		accessToken:       accessToken,
		accessTokenSecret: accessTokenSecret,
		kvstore:           kvstore,
	}
	t.client = anaconda.NewTwitterApiWithCredentials(
		accessToken, accessTokenSecret,
		consumerKey, consumerSecret)

	return t
}

func (t *Twitter) Register(pattern string) {
	http.Handle(pattern, http.StripPrefix(pattern, t))
}

func (t *Twitter) handleRequestToken(w http.ResponseWriter, r *http.Request) {
	url, tmpCred, err := t.client.AuthorizationURL("http://localhost:8080/auth/twitter/callback")
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	// TODO: must expire
	// TODO: support multi session
	t.credential = tmpCred
	http.Redirect(w, r, url, http.StatusFound)
}

func (t *Twitter) handleAccessToken(w http.ResponseWriter, r *http.Request) {
	c, _, err := t.client.GetCredentials(t.credential, r.URL.Query().Get("oauth_verifier"))
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	cli := anaconda.NewTwitterApiWithCredentials(c.Token, c.Secret, t.consumerKey, t.consumerSecret)

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

func (t *Twitter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "callback":
		t.handleAccessToken(w, r)
	default:
		t.handleRequestToken(w, r)
	}
}
