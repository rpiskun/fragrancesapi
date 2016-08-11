package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/context"
	"github.com/unrolled/render"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// getAccessToken ...
func getAccessToken(r *http.Request) (string, error) {
	tok := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(tok), "bearer ") {
		return "", errors.New("Get access token")
	}
	return tok[7:], nil
}

// ValidateTokenByGoogle ...
func ValidateTokenByGoogle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	jsonRender := render.New()
	tok, err := getAccessToken(r)
	if err != nil {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	validateUrl := "https://www.googleapis.com/oauth2/v3/tokeninfo?access_token=" + tok
	req, err := http.NewRequest("GET", validateUrl, nil)
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if resp.StatusCode != http.StatusOK {
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := resp.Body.Close(); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}
	var validTokenData ValidatedTokenData
	if err := json.Unmarshal(body, &validTokenData); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
	context.Set(r, "accessToken", tok)
	context.Set(r, "validatedUserId", validTokenData.Sub)
	context.Set(r, "expiresOn", validTokenData.Exp)
	next(w, r)
}

// ValidateTokenByDatabase ...
func ValidateTokenByDatabase(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	jsonRender := render.New()
	tok, err := getAccessToken(r)
	if err != nil {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}
	user, err := GetUserByAccessToken(tok)
	if err != nil {
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}
	if user.ExpiresOn < time.Now().Unix() {
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}
	context.Set(r, "user", user)
	next(w, r)
}
