package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var (
	resourcesDir = ""
)

func init() {
	if resourcesDir = os.Getenv("FRAGRANCES_API_RESOURCES_BASE_DIR"); resourcesDir == "" {
		log.Fatalln("Variable FRAGRANCES_API_RESOURCES_BASE_DIR is not defined")
		os.Exit(-1)
	}
}

// GetUserEndpoint ...
func GetUserEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	userId := vars["userId"]
	validatedUserId := context.Get(r, "validatedUserId").(string)
	if userId != validatedUserId {
		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
		return
	}
	user, err := GetUserByGplusId(userId)
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if user == nil {
		w.Header().Set("Cache-Control", "no-cache")
		jsonRender.JSON(w, http.StatusNotFound, &UserResp{
			UserId:    userId,
			CreatedAt: "",
			UpdatedAt: "",
			Links: []LinkV1{
				LinkV1{
					Href:   "/api/v1/users/" + userId,
					Rel:    "create",
					Method: "POST",
				},
			},
		})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	jsonRender.JSON(w, http.StatusOK, &UserResp{
		UserId:    userId,
		CreatedAt: time.Unix(user.CreatedAt, 0).String(),
		UpdatedAt: time.Unix(user.UpdatedAt, 0).String(),
		Links: []LinkV1{
			LinkV1{
				Href:   "/api/v1/users/" + userId,
				Rel:    "get",
				Method: "GET",
			},
			LinkV1{
				Href:   "/api/v1/users/" + userId,
				Rel:    "update",
				Method: "PUT",
			},
		},
	})
}

// CreateUserEndpoint ...
func CreateUserEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	userId := vars["userId"]
	validatedUserId := context.Get(r, "validatedUserId").(string)
	if userId != validatedUserId {
		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
		return
	}

	createdUser, err := UserInsert(userId, context.Get(r, "accessToken").(string), context.Get(r, "expiresOn").(string))
	if err != nil || createdUser == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Location", "http://"+r.Host+"/api/v1/users/"+userId)
	jsonRender.JSON(w, http.StatusCreated, &UserResp{
		UserId:    userId,
		CreatedAt: time.Unix(createdUser.CreatedAt, 0).String(),
		UpdatedAt: time.Unix(createdUser.UpdatedAt, 0).String(),
		Links: []LinkV1{
			LinkV1{
				Href:   "/api/v1/users/" + userId,
				Rel:    "get",
				Method: "GET",
			},
			LinkV1{
				Href:   "/api/v1/users/" + userId,
				Rel:    "update",
				Method: "PUT",
			},
		},
	})
}

// UpdateUserEndpoint ...
func UpdateUserEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	userId := vars["userId"]
	validatedUserId := context.Get(r, "validatedUserId").(string)
	if userId != validatedUserId {
		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
		return
	}
	user, err := GetUserByGplusId(userId)
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if user == nil {
		jsonRender.JSON(w, http.StatusNotFound, &UserResp{
			UserId:    userId,
			CreatedAt: "",
			UpdatedAt: "",
			Links: []LinkV1{
				LinkV1{
					Href:   "/api/v1/users/" + userId,
					Rel:    "create",
					Method: "POST",
				},
			},
		})
		return
	}
	updated, err := user.Update(context.Get(r, "accessToken").(string), context.Get(r, "expiresOn").(string))
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if !updated {
		w.Header().Set("Cache-Control", "no-cache")
		jsonRender.JSON(w, http.StatusNotFound, &UserResp{
			UserId:    userId,
			CreatedAt: "",
			UpdatedAt: "",
			Links: []LinkV1{
				LinkV1{
					Href:   "/api/v1/users/" + userId,
					Rel:    "create",
					Method: "POST",
				},
			},
		})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	jsonRender.JSON(w, http.StatusOK, &UserResp{
		UserId:    userId,
		CreatedAt: time.Unix(user.CreatedAt, 0).String(),
		UpdatedAt: time.Unix(user.UpdatedAt, 0).String(),
		Links: []LinkV1{
			LinkV1{
				Href:   "/api/v1/users/" + userId,
				Rel:    "get",
				Method: "GET",
			},
			LinkV1{
				Href:   "/api/v1/users/" + userId,
				Rel:    "update",
				Method: "PUT",
			},
		},
	})
}

//DeleteUserEndpoint ...
func DeleteUserEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	validatedUserId := context.Get(r, "validatedUserId").(string)
	userId := vars["userId"]
	if userId != validatedUserId {
		jsonRender.JSON(w, http.StatusForbidden, map[string]string{"status": "forbidden"})
		return
	}
	user, err := GetUserByGplusId(userId)
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if user == nil {
		jsonRender.JSON(w, http.StatusNotFound, &UserResp{
			UserId:    userId,
			CreatedAt: "",
			UpdatedAt: "",
			Links: []LinkV1{
				LinkV1{
					Href:   "/api/v1/users/" + userId,
					Rel:    "create",
					Method: "POST",
				},
			},
		})
		return
	}
	deleted, err := user.Delete()
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	} else if !deleted {
		w.Header().Set("Cache-Control", "no-cache")
		jsonRender.JSON(w, http.StatusNotFound, &UserResp{
			UserId:    userId,
			CreatedAt: "",
			UpdatedAt: "",
			Links: []LinkV1{
				LinkV1{
					Href:   "/api/v1/users/" + userId,
					Rel:    "create",
					Method: "POST",
				},
			},
		})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	jsonRender.JSON(w, http.StatusOK, &UserResp{
		UserId:    userId,
		CreatedAt: "",
		UpdatedAt: "",
		Links: []LinkV1{
			LinkV1{
				Href:   "/api/v1/users/" + userId,
				Rel:    "create",
				Method: "POST",
			},
		},
	})
}

// LogoutEndpoint ...
func LogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	tok := context.Get(r, "accessToken").(string)
	url := "https://accounts.google.com/o/oauth2/revoke?token=" + tok
	resp, err := http.Get(url)
	if err != nil {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}
	defer resp.Body.Close()
	jsonRender.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

// GetSmallImageEndpoint ...
func GetSmallImageEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["imageUuid"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}
	imageDb, err := GetImageByUuid(uid)
	if err != nil {
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	} else if imageDb == nil {
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}

	if !imageDb.SmallImgPath.Valid {
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}
	fp := filepath.Join(
		resourcesDir,
		imageDb.SmallImgPath.String,
		imageDb.SmallImgFname.String)
	fp = filepath.ToSlash(fp)
	http.ServeFile(w, r, fp)
}

// GetLargeImageEndpoint ...
func GetLargeImageEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	vars := mux.Vars(r)
	uid, ok := vars["imageUuid"]
	if !ok {
		jsonRender.JSON(w, http.StatusBadRequest, map[string]string{"status": "bad request"})
		return
	}
	imageDb, err := GetImageByUuid(uid)
	if err != nil {
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	} else if imageDb == nil {
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}

	if !imageDb.LargeImgFname.Valid {
		jsonRender.JSON(w, http.StatusNotFound, map[string]string{"status": "not found"})
		return
	}
	fp := path.Join(
		resourcesDir,
		imageDb.LargeImgPath.String,
		imageDb.LargeImgFname.String)
	fp = filepath.ToSlash(fp)
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := pinfos.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if _, err := obj.MakeObj(&MakeObjParams{Base: *params, Total: count}); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
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
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := composition.Json(w, http.StatusOK); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// GetPerfumFindEndpoint ...
func GetPerfumFindEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	params := NewSearchParams()
	params.Parse(r)
	obj := NewSearchResultFactory(params.Base.Version)
	if obj == nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	count, err := obj.Count(params)
	if err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	params.Total = count
	if _, err := obj.MakeObj(params); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}

	if err := obj.Json(w, http.StatusOK); err != nil {
		jsonRender.JSON(w, http.StatusInternalServerError, map[string]string{"status": "internal server error"})
		return
	}
}

// TempEndpoint ...
func TempEndpoint(w http.ResponseWriter, r *http.Request) {

}

func TestEndpoint(w http.ResponseWriter, r *http.Request) {
	jsonRender := render.New()
	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.WhereConditionString = " < WHERE > "
	dbQuery.AndConditionString = " < AND > "
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", &dbQuery); err != nil {
		return
	}

	fmt.Println("TestEndpoint")

	jsonRender.JSON(w, http.StatusNotFound, map[string]string{"QUERY": query.String()})
}
