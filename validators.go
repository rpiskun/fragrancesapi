package main

import (
	// "encoding/json"
	"errors"
	// "fmt"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
	// "io"
	// "io/ioutil"

	"net/http"
	"strings"
	"time"
)

// getAccessToken ...
func getAccessToken(r *http.Request) (string, error) {
	tok := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(tok), "bearer ") {
		return "", errors.New("Token type is not Bearer")
	}
	return tok[7:], nil
}

// ValidateAccessToken ...
func ValidateAccessToken(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	jsonRender := render.New()
	tok, err := getAccessToken(r)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	accessTokenClaims, err := CheckAccessToken(tok)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	user, err := GetUserByAccessToken(tok)
	if err != nil {
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	if user.ExpiresAt < time.Now().Unix() {
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	// fmt.Println("Access token:", tok)
	// fmt.Println("URL:", r.URL.RequestURI())

	context.Set(r, "user", user)
	context.Set(r, "token", accessTokenClaims)
	next(w, r)
}
