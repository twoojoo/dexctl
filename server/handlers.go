package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/twoojoo/dexctl/utils"
	"golang.org/x/oauth2"
)

type AuthorizationResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type UserClaimsJWT struct {
	AtHash        string `json:"at_hash"`
	Aud           string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Iat           int    `json:"iat"`
	Iss           string `json:"iss"`
	Name          string `json:"name"`
	Sub           string `json:"sub"`
}

type ApplicationHanlder struct {
	state           string
	provider        *oidc.Provider
	oauth2Config    oauth2.Config
	idTokenVerifier *oidc.IDTokenVerifier
	userinfo        bool
}

func (a ApplicationHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/login":
		a.handleLogin(w, r)
	case "/callback":
		a.handleCallback(w, r)
	case "/favicon.ico":
		w.Write([]byte{})
	default:
		log.Println("called unknown URL:", r.URL.String())
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte{})
		log.Fatal(http.StatusNotFound)
	}
}

func (a ApplicationHanlder) handleExample(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you're the user with id"))
}

func (a ApplicationHanlder) handleLogin(w http.ResponseWriter, r *http.Request) {
	providerURL := a.oauth2Config.AuthCodeURL(a.state)
	http.Redirect(w, r, providerURL, 307)
}

func (a ApplicationHanlder) handleCallback(w http.ResponseWriter, r *http.Request) {
	var err error
	var token *oauth2.Token

	ctx := context.Background()

	switch r.Method {
	case http.MethodGet: //Oauth2 flow
		if errMsg := r.FormValue("error"); errMsg != "" {
			w.Write(errorPage(errMsg + ": " + r.FormValue("error_description")))
			go exitWithDelay(1)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			w.Write(errorPage(fmt.Sprintf("no code in request: %q", r.Form)))
			go exitWithDelay(1)
			return
		}

		if state := r.FormValue("state"); state != a.state {
			w.Write(errorPage("state mismatch"))
			go exitWithDelay(1)
			return
		}

		token, err = a.oauth2Config.Exchange(ctx, code)
	case http.MethodPost: // Form request from frontend to refresh a token.
		refresh := r.FormValue("refresh_token")
		if refresh == "" {
			w.Write(errorPage(fmt.Sprintf("no refresh_token in request: %q", r.Form)))
			go exitWithDelay(1)
			return
		}

		t := &oauth2.Token{
			RefreshToken: refresh,
			Expiry:       time.Now().Add(-time.Hour),
		}

		token, err = a.oauth2Config.TokenSource(ctx, t).Token()
	default:
		http.Error(w, fmt.Sprintf("method not implemented: %s", r.Method), http.StatusBadRequest)
		go exitWithDelay(1)
		return
	}

	if a.userinfo {
		tokenSource := a.oauth2Config.TokenSource(ctx, token)
		userInfo, err := a.provider.UserInfo(ctx, tokenSource)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch user info: %e", err), http.StatusInternalServerError)
			go exitWithDelay(1)
			return
		}

		w.Write(successPage)

		p, err := utils.PrettifyJSON(userInfo)
		if err != nil {
			w.Write(errorPage(fmt.Sprintf("failed to prettify userinfo: %v", err)))
			go exitWithDelay(1)
			return
		}

		fmt.Println(p)

		go exitWithDelay(0)
		return
	}

	if err != nil {
		w.Write(errorPage(fmt.Sprintf("failed to get token: %v", err)))
		go exitWithDelay(1)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		w.Write(errorPage("no id_token in token response"))
		go exitWithDelay(1)
		return
	}

	if _, err := a.idTokenVerifier.Verify(r.Context(), rawIDToken); err != nil {
		w.Write(errorPage(fmt.Sprintf("failed to verify ID token: %v", err)))
		go exitWithDelay(1)
		return
	}

	w.Write(successPage)

	p, err := utils.PrettifyJSON(token)
	fmt.Println("id_token:", rawIDToken)
	if err != nil {
		w.Write(errorPage(fmt.Sprintf("failed to prettify token: %v", err)))
		go exitWithDelay(1)
		return
	}

	fmt.Println(p)

	go exitWithDelay(0)
	return
}

func exitWithDelay(status int) {
	time.Sleep(300 * time.Millisecond)
	os.Exit(status)
}
