package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	clientID            = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	clientSecret        = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
	userIDByIDToken     = make(map[string]string)
	userByUserID        = make(map[string]user)
	userIDByAccessToken = make(map[string]string)
)

const servingURL = "https://hashira-auth.appspot.com"

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

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  servingURL + "/auth/google/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	state := "foobar" // Don't do this in production.

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = json.MarshalIndent(resp, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// check if the user already exists
		uid, ok := userIDByIDToken[idToken.Subject]
		if ok {
			token := uuid.NewV4()
			userIDByAccessToken[token.String()] = uid
			cookie := &http.Cookie{
				Name:   "Authorization",
				Value:  token.String(),
				Domain: servingURL,
				Path:   "/",
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// create new user
		var (
			userID = uuid.NewV4()
			token  = uuid.NewV4()
		)
		username, err := fetchPhraseFromMashimashi()
		if err != nil {
			// TODO: error handling
		}

		userIDByIDToken[idToken.Subject] = userID.String()
		userByUserID[userID.String()] = user{
			id:   userID.String(),
			name: username,
		}
		userIDByAccessToken[token.String()] = userID.String()

		cookie := &http.Cookie{
			Name:   "Authorization",
			Value:  token.String(),
			Domain: servingURL,
			Path:   "/",
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func fetchPhraseFromMashimashi() (string, error) {
	resp, err := http.Get("https://strongest-mashimashi.appspot.com/api/v1/phrase")
	if err != nil {
		return "", err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
