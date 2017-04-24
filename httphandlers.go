package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
	"fmt"
)

var (
	resourcesDir = ""
)

func init() {
	if resourcesDir = os.Getenv("OPENSHIFT_DATA_DIR"); resourcesDir == "" {
		TraceFatal("Variable OPENSHIFT_DATA_DIR is not defined")
		os.Exit(-1)
	}
}

// LoginEndpoint ...
func LoginEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()

	if err := r.ParseForm(); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	idToken, ok := r.Form["id_token"]
	if !ok {
		TracePrintError(errors.New("id_token is not exist in request form"))
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	idTokenClaims, err := CheckIdToken(idToken[0])
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "unauthorized"})
		return
	}

	accessToken, err := NewAccessToken(idTokenClaims.Aud, idTokenClaims.Sub)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	fmt.Println("Access token:", accessToken.tokenString)

	refreshToken, err := NewRefreshToken(idTokenClaims.Aud, idTokenClaims.Sub)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	user, err := GetUserByUserId(idTokenClaims.Sub)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	if user != nil {
		// update existing user
		updated, err := user.Update(accessToken.tokenString, refreshToken.tokenString, accessToken.ExpiresAt)
		if err != nil {
			TracePrintError(err)
			jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
			return
		} else if !updated {
			TracePrint("user not updated")
			jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
			return
		}
	} else {
		// create new user
		createdUser, err := UserInsert(idTokenClaims.Sub, accessToken.tokenString, refreshToken.tokenString, accessToken.ExpiresAt)
		if err != nil || createdUser == nil {
			TracePrint("new user not created")
			jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
			return
		}
		w.Header().Set("Location", baseUrl+"/users/"+createdUser.UserId)
	}

	jsonRender.JSON(w, http.StatusOK, &LoginResp{
		AccessToken:  accessToken.tokenString,
		RefreshToken: refreshToken.tokenString,
		UserId:       idTokenClaims.Sub,
	})
}

// TokenEndpoint
func TokenEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()

	if err := r.ParseForm(); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	grantType, ok := r.Form["grant_type"]
	if !ok {
		TracePrint("grant_type is not exist in request form")
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	if grantType[0] != "refresh_token" {
		TracePrint("grant_type is not supported")
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	clientId, ok := r.Form["client_id"]
	if !ok {
		TracePrint("client_id is not exist in request form")
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	if clientId[0] != oAuthCred.ClientID {
		TracePrint("clientId[0] != oAuthCred.ClientID")
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	token, ok := r.Form["refresh_token"]
	if !ok {
		TracePrint("refresh_token is not exist in request form")
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	tokenClaims, err := CheckRefreshToken(token[0])
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	user, err := GetUserByRefreshToken(token[0])
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}
	if user == nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	if user.UserId != tokenClaims.Subject {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusUnauthorized, map[string]string{"status": "unauthorized"})
		return
	}

	accessToken, err := NewAccessToken(tokenClaims.Audience, tokenClaims.Subject)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
	refreshToken, err := NewRefreshToken(tokenClaims.Audience, tokenClaims.Subject)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	updated, err := user.Update(accessToken.tokenString, refreshToken.tokenString, accessToken.ExpiresAt)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if !updated {
		TracePrint("user not updated")
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	jsonRender.JSON(w, http.StatusOK, &LoginResp{
		AccessToken:  accessToken.tokenString,
		RefreshToken: refreshToken.tokenString,
		UserId:       tokenClaims.Subject,
	})
}

// LogoutEndpoint
func LogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	userId := vars["userId"]
	user := context.Get(r, "user").(*UserDB)
	if user == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if userId != user.UserId {
		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
		return
	}

	updated, err := user.Update(user.AccessToken, user.RefreshToken, time.Now().Unix())
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if !updated {
		TracePrint("user not updated")
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	jsonRender.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RefreshTokenEndpoint
// func RefreshTokenEndpoint(w http.ResponseWriter, r *http.Request) {
// 	jsonRender := render.New()
// 	vars := mux.Vars(r)
// 	userId := vars["userId"]
// 	token := context.Get(r, "token").(AccessTokenClaims)
// 	user := context.Get(r, "user").(*UserDB)
// 	if user == nil {
// 		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
// 		return
// 	}

// 	if userId != user.UserId {
// 		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
// 		return
// 	}

// 	accessToken, err := NewAccessToken(token.Audience, token.Subject)
// 	if err != nil {
// 		log.Println("ERROR RefreshTokenEndpoint: NewAccessToken >>", err)
// 		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
// 		return
// 	}

// 	w.Header().Set("Cache-Control", "no-cache")
// 	updated, err := user.Update(accessToken.tokenString, user.RefreshToken, accessToken.ExpiresAt)
// 	if err != nil {
// 		log.Println("ERROR RefreshTokenEndpoint: user.Update >>", err)
// 		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
// 		return
// 	} else if !updated {
// 		log.Println("ERROR RefreshTokenEndpoint: user.Update >>", errors.New("user not updated"))
// 		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
// 		return
// 	}

// 	jsonRender.JSON(w, http.StatusOK, map[string]string{
// 		"access_token": accessToken.tokenString,
// 	})
// }

// GetUserEndpoint ...
func GetUserEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	userId := vars["userId"]
	user := context.Get(r, "user").(*UserDB)
	if user == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if userId != user.UserId {
		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	jsonRender.JSON(w, http.StatusOK, &UserResp{
		UserId:    user.UserId,
		CreatedAt: strconv.FormatInt(user.CreatedAt, 10),
		UpdatedAt: strconv.FormatInt(user.UpdatedAt, 10),
		Links: []LinkV1{
			LinkV1{
				Href:   baseUrl + "/user/" + userId + "/logout",
				Rel:    "Logout",
				Method: "PUT",
			},
			LinkV1{
				Href:   baseUrl + "/user/" + userId + "/refresh",
				Rel:    "RefreshToken",
				Method: "GET",
			},
		},
	})
}

//DeleteUserEndpoint ...
func DeleteUserEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	user := context.Get(r, "user").(*UserDB)
	userId := vars["userId"]

	if user == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if userId != user.UserId {
		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
		return
	}
	deleted, err := user.Delete()
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if !deleted {
		TracePrint("user is not deleted")
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	jsonRender.JSON(w, http.StatusOK, &UserResp{
		UserId:    userId,
		CreatedAt: "",
		UpdatedAt: "",
		Links: []LinkV1{
			LinkV1{
				Href:   baseUrl + "/login",
				Rel:    "Login",
				Method: "POST",
			},
		},
	})
}

// GetUserFavoritesEndpoint ...
func GetUserFavoritesEndpoint(w http.ResponseWriter, r *http.Request) {

}

// CreateUserFavoritesEndpoint ...
func CreateUserFavoritesEndpoint(w http.ResponseWriter, r *http.Request) {

}

// UpdateUserFavoritesEndpoint ...
func UpdateUserFavoritesEndpoint(w http.ResponseWriter, r *http.Request) {

}

// DeleteUserFavoritesEndpoint ...
func DeleteUserFavoritesEndpoint(w http.ResponseWriter, r *http.Request) {

}

func getFileHash(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

// GetSmallImageEndpoint ...
func GetSmallImageEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["imageId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}
	imageDb, err := GetImageByUuid(uid)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	} else if imageDb == nil {
		TracePrint("Image not found")
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}

	if !imageDb.SmallImgPath.Valid {
		TracePrint("Image path is not valid")
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}
	fp := filepath.Join(
		resourcesDir,
		imageDb.SmallImgPath.String,
		imageDb.SmallImgFname.String)

	// hash, err := getFileHash(fp)
	// if err != nil {
	// 	TracePrintError(err)
	// 	jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
	// 	return
	// }
	// if etag := r.Header.Get("If-None-Match"); etag != "" {
	// 	if hash == etag {
	// 		jsonRender.JSON(w, http.StatusNotModified, map[string]string{"status": "not modified"})
	// 		return
	// 	}
	// }
	// w.Header().Set("Etag", hash)
	w.Header().Set("Cache-Control", "max-age=2629000")
	http.ServeFile(w, r, fp)
}

// GetLargeImageEndpoint ...
func GetLargeImageEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["imageId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}
	imageDb, err := GetImageByUuid(uid)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	} else if imageDb == nil {
		TracePrint("Image not found")
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}

	if !imageDb.LargeImgFname.Valid {
		TracePrint("Image path is not valid")
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}
	fp := path.Join(
		resourcesDir,
		imageDb.LargeImgPath.String,
		imageDb.LargeImgFname.String)
	// hash, err := getFileHash(fp)
	// if err != nil {
	// 	TracePrintError(err)
	// 	jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
	// 	return
	// }
	// if etag := r.Header.Get("If-None-Match"); etag != "" {
	// 	if hash == etag {
	// 		jsonRender.JSON(w, http.StatusNotModified, map[string]string{"status": "not modified"})
	// 		return
	// 	}
	// }
	// w.Header().Set("Etag", hash)
	w.Header().Set("Cache-Control", "max-age=2629000")
	http.ServeFile(w, r, fp)
}

// GetBrandsEndpoint ...
func GetBrandsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("brands")
	params.Parse(r)

	obj := NewBrandsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetBrandEndpoint ...
func GetBrandEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["brandId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("brands")
	params.Parse(r)

	obj := NewBrandsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetBrandPerfums ...
func GetBrandPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["brandId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("brands")
	params.Parse(r)

	obj := NewBrandsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

func GetComponentsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("components")
	params.Parse(r)

	obj := NewComponentsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

func GetComponentEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["componentId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("components")
	params.Parse(r)

	obj := NewComponentsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

func GetComponentPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["componentId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("components")
	params.Parse(r)

	obj := NewComponentsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetCountriesEndpoint ...
func GetCountriesEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("countries")
	params.Parse(r)

	obj := NewCountriesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetCountryEndpoint ...
func GetCountryEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["countryId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("countries")
	params.Parse(r)

	obj := NewCountriesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetCountryPerfumsEndpoint ...
func GetCountryPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["countryId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("countries")
	params.Parse(r)

	obj := NewCountriesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetGendersEndpoint ...
func GetGendersEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("gender")
	params.Parse(r)

	obj := NewGendersFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetCountryEndpoint ...
func GetGenderEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["genderId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("gender")
	params.Parse(r)

	obj := NewGendersFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetGenderPerfumsEndpoint ...
func GetGenderPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["genderId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("gender")
	params.Parse(r)

	obj := NewGendersFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetGroupsEndpoint ...
func GetGroupsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("groups")
	params.Parse(r)

	obj := NewGroupsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetGroupEndpoint ...
func GetGroupEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["groupId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("groups")
	params.Parse(r)

	obj := NewGroupsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetGroupPerfumsEndpoint ...
func GetGroupPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["groupId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("groups")
	params.Parse(r)

	obj := NewGroupsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetNotesEndpoint ...
func GetNotesEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("notes")
	params.Parse(r)

	obj := NewNotesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetNoteEndpoint ...
func GetNoteEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["noteId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("notes")
	params.Parse(r)

	obj := NewNotesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetNotePerfumsEndpoint ...
func GetNotePerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["noteId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("notes")
	params.Parse(r)

	obj := NewNotesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetSeasonsEndpoint ...
func GetSeasonsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("seasons")
	params.Parse(r)

	obj := NewSeasonsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetSeasonEndpoint ...
func GetSeasonEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["seasonId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("seasons")
	params.Parse(r)

	obj := NewSeasonsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetSeasonPerfumsEndpoint ...
func GetSeasonPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["seasonId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("seasons")
	params.Parse(r)

	obj := NewSeasonsFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetTimesOfDayEndpoint ...
func GetTimesOfDayEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("tsod")
	params.Parse(r)

	obj := NewTimesOfDayFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetTimeOfDayEndpoint ...
func GetTimeOfDayEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["tsodId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("tsod")
	params.Parse(r)

	obj := NewTimesOfDayFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetTimeOfDayPerfumsEndpoint ...
func GetTimeOfDayPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["tsodId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("tsod")
	params.Parse(r)

	obj := NewTimesOfDayFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetTypesEndpoint ...
func GetTypesEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("types")
	params.Parse(r)

	obj := NewTypesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetTypeEndpoint ...
func GetTypeEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["typeId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("types")
	params.Parse(r)

	obj := NewTypesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	if _, err := obj.MakeObj(
		&MakeObjParams{
			Base:       *params,
			Total:      1,
			PerfumsNum: NullInt64{Valid: true, Int64: count},
		},
	); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetTypePerfumsEndpoint ...
func GetTypePerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["typeId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("types")
	params.Parse(r)

	obj := NewTypesFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	pinfos, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetPerfumsEndpoint ...
func GetPerfumsEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewBaseParams("perfums")
	params.Parse(r)

	obj := NewPerfumsInfoFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(&params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetPerfumDetailedInfoEndpoint ...
func GetPerfumDetailedInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["perfumId"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}

	params := NewBaseParams("perfums")
	params.Parse(r)

	obj := NewPerfumsInfoFactory(params.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.ExtraCount([]string{uid})
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if count == 0 {
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}

	params.Ids.String = []string{uid}
	params.Ids.Valid = true

	composition, err := obj.MakeExtraObj(
		&MakeObjParams{
			Base:  *params,
			Total: count,
			Id:    uid,
		},
		[]string{uid},
	)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := composition.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetPerfumFindEndpoint ...
func GetPerfumsFindEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewSearchParams()
	params.Parse(r)
	obj := NewPerfumsSearchResultFactory(params.Base.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Total = count
	if _, err := obj.MakeObj(params); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetBrandsFindEndpoint ...
func GetBrandsFindEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewSearchParams()
	params.Parse(r)
	obj := NewBrandsSearchResultFactory(params.Base.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Total = count
	if _, err := obj.MakeObj(params); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetComponentsFindEndpoint ...
func GetComponentsFindEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewSearchParams()
	params.Parse(r)
	obj := NewComponentsSearchResultFactory(params.Base.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Total = count
	if _, err := obj.MakeObj(params); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetCountriesFindEndpoint ...
func GetCountriesFindEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewSearchParams()
	params.Parse(r)
	obj := NewCountriesSearchResultFactory(params.Base.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Total = count
	if _, err := obj.MakeObj(params); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetGroupsFindEndpoint ...
func GetGroupsFindEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewSearchParams()
	params.Parse(r)
	obj := NewGroupsSearchResultFactory(params.Base.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(params)
	if err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Total = count
	if _, err := obj.MakeObj(params); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		TracePrintError(err)
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}
