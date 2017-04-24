package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// RefreshTokenClaims ...
type RefreshTokenClaims struct {
	tokenString string
	jwt.StandardClaims
}

// AccessTokenClaims ...
type AccessTokenClaims struct {
	tokenString string
	jwt.StandardClaims
}

const (
	accessTokenDuration  = 24     //in hours
	refreshTokenDuration = 24 * 7 //in hours
	certificateURL       = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

var (
	accessTokenSign  []byte
	refreshTokenSign []byte
	oAuthCred        *OAuth2Credentials
	maxAgePattern    = regexp.MustCompile(`\s*max-age\s*=\s*(\d+)\s*`)
	certCache        = NewCache(NoExpiration)
)

func init() {
	// rand.Seed(time.Now().UnixNano())
	var err error

	accessTokenSign, err = generateRandomSign(64)
	if err != nil {
		TraceFatalError(err)
		os.Exit(-1)
	}

	refreshTokenSign, err = generateRandomSign(64)
	if err != nil {
		TraceFatalError(err)
		os.Exit(-1)
	}

	if resourcesDir := os.Getenv("OPENSHIFT_DATA_DIR"); resourcesDir == "" {
		TraceFatalError(err)
		os.Exit(-1)
	}

	b, err := ioutil.ReadFile(filepath.Join(resourcesDir, "client_secret.json"))
	if err != nil {
		TraceFatalError(err)
		os.Exit(-1)
		return
	}

	oAuthCred, err = CredentialsFromJSON(b)
	if err != nil {
		TraceFatalError(err)
		os.Exit(-1)
		return
	}
}

func generateRandomSign(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// NewAccessToken ...
func NewAccessToken(audience, subject string) (AccessTokenClaims, error) {
	var err error
	claims := AccessTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  audience,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Duration(accessTokenDuration) * time.Hour).Unix(),
			Issuer:    "fragrances-api",
			Subject:   subject,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	claims.tokenString, err = token.SignedString(accessTokenSign)
	if err != nil {
		return claims, err
	}

	return claims, nil
}

// NewRefreshToken ...
func NewRefreshToken(audience, subject string) (RefreshTokenClaims, error) {
	var err error
	claims := RefreshTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  audience,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Duration(refreshTokenDuration) * time.Hour).Unix(),
			Issuer:    "fragrances-api",
			Subject:   subject,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	claims.tokenString, err = token.SignedString(refreshTokenSign)
	if err != nil {
		return claims, err
	}

	return claims, nil
}

// maxAge parses Cache-Control header value and extracts max-age (in seconds)
func maxAge(s string) int {
	match := maxAgePattern.FindStringSubmatch(s)
	if len(match) != 2 {
		return 0
	}
	if maxAge, err := strconv.Atoi(match[1]); err == nil {
		return maxAge
	}
	return 0
}

func certExpirationTime(h http.Header) time.Duration {
	var max int
	for _, entry := range strings.Split(h.Get("Cache-Control"), ",") {
		max = maxAge(entry)
		if max > 0 {
			break
		}
	}
	if max <= 0 {
		return 0
	}

	age, err := strconv.Atoi(h.Get("Age"))
	if err != nil {
		return 0
	}

	remainingTime := max - age
	if remainingTime <= 0 {
		return 0
	}

	return time.Duration(remainingTime) * time.Second
}

func lookupPublicKey(kid string) ([]byte, error) {
	var err error

	if cert, found := certCache.Get(kid); found {
		return []byte(cert.(string)), nil
	}

	req, err := http.NewRequest("GET", certificateURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Response not OK")
	}

	var certs map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&certs)
	if err != nil {
		return nil, err
	}

	expiration := certExpirationTime(res.Header)
	if expiration == 0 {
		return nil, errors.New("Certificate expiration time not valid")
	}
	certCache.Flush()
	for k, v := range certs {
		certCache.Set(k, v, expiration)
	}

	if cert, found := certCache.Get(kid); found {
		return []byte(cert.(string)), nil
	}

	return nil, errors.New("Certificate with key not found")
}

func parseIdTokenClaims(claims jwt.MapClaims) IdTokenClaims {
	idToken := IdTokenClaims{}

	if value, ok := claims["exp"]; ok {
		idToken.Exp = value.(float64)
	}
	if value, ok := claims["iat"]; ok {
		idToken.Iat = value.(float64)
	}
	if value, ok := claims["aud"]; ok {
		idToken.Aud = value.(string)
	}
	if value, ok := claims["iss"]; ok {
		idToken.Iss = value.(string)
	}
	if value, ok := claims["sub"]; ok {
		idToken.Sub = value.(string)
	}
	if value, ok := claims["user_id"]; ok {
		idToken.UserId = value.(string)
	}

	return idToken
}

//CheckIdToken ...
func CheckIdToken(tokenString string) (IdTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("Unexpected signing method")
		}
		if kid, ok := token.Header["kid"]; ok {
			if key, err := lookupPublicKey(kid.(string)); err == nil {
				return jwt.ParseRSAPublicKeyFromPEM(key)
			}
		}

		return nil, errors.New("Certificate not found")
	})

	if err != nil {
		return IdTokenClaims{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idToken := parseIdTokenClaims(claims)
		if idToken.Exp > float64(time.Now().Unix()) &&
			idToken.Iat <= float64(time.Now().Unix()) &&
			idToken.Aud == oAuthCred.ProjectID &&
			idToken.Iss == "https://securetoken.google.com/"+oAuthCred.ProjectID &&
			idToken.Sub != "" &&
			idToken.Sub == idToken.UserId {
			return idToken, nil
		}

	}

	return IdTokenClaims{}, jwt.NewValidationError("token is not valid", jwt.ValidationErrorClaimsInvalid)
}

//CheckAccessToken ...
func CheckAccessToken(tokenString string) (AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return accessTokenSign, nil
	})

	if err != nil {
		return AccessTokenClaims{}, err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return *claims, nil
	}

	return AccessTokenClaims{}, err
}

//CheckRefreshToken ...
func CheckRefreshToken(tokenString string) (RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return refreshTokenSign, nil
	})

	if err != nil {
		return RefreshTokenClaims{}, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return *claims, nil
	}

	return RefreshTokenClaims{}, err
}

type OAuth2Credentials struct {
	ClientID     string
	ProjectID    string
	CertURL      string
	ClientSecret string
}

func CredentialsFromJSON(jsonKey []byte) (*OAuth2Credentials, error) {
	type cred struct {
		ClientID     string `json:"client_id"`
		ProjectID    string `json:"project_id"`
		AuthURI      string `json:"auth_uri"`
		TokenURI     string `json:"token_uri"`
		CertURL      string `json:"auth_provider_x509_cert_url"`
		ClientSecret string `json:"client_secret"`
	}
	var j struct {
		Web       *cred `json:"web"`
		Installed *cred `json:"installed"`
	}
	if err := json.Unmarshal(jsonKey, &j); err != nil {
		return nil, err
	}
	var c *cred
	switch {
	case j.Web != nil:
		c = j.Web
	case j.Installed != nil:
		c = j.Installed
	default:
		return nil, errors.New("oauth2/google: no credentials found")
	}

	return &OAuth2Credentials{
		ClientID:     c.ClientID,
		ProjectID:    c.ProjectID,
		CertURL:      c.CertURL,
		ClientSecret: c.ClientSecret,
	}, nil
}
