package main

import (
	"net/http"
)

//Route ...
type Route struct {
	Name     string
	Method   string
	Pattern  string
	Endpoint http.HandlerFunc
}

// Routes ...
type Routes []Route

// validateByGoogle. Path prefix: /users
var validateByGoogle = Routes{
	Route{
		"GetUser",
		"GET",
		"/{userId}",
		GetUserEndpoint,
	},
	Route{
		"CreateUser",
		"POST",
		"/{userId}",
		CreateUserEndpoint,
	},
	Route{
		"UpdateUser",
		"PUT",
		"/{userId}",
		UpdateUserEndpoint,
	},
	Route{
		"DeleteUser",
		"DELETE",
		"/{userId}",
		DeleteUserEndpoint,
	},
	Route{
		"Logout",
		"GET",
		"/{userId}/logout",
		LogoutEndpoint,
	},
	Route{
		"GetUserFavorites",
		"GET",
		"/{userId}/favorites",
		GetUserFavoritesEndpoint,
	},
	Route{
		"CreateUserFavorites",
		"POST",
		"/{userId}/favorites",
		CreateUserFavoritesEndpoint,
	},
	Route{
		"UpdateUserFavorites",
		"PUT",
		"/{userId}/favorites",
		UpdateUserFavoritesEndpoint,
	},
	Route{
		"DeleteUserFavorites",
		"DELETE",
		"/{userId}/favorites",
		DeleteUserFavoritesEndpoint,
	},
}

var validateByDatabase = Routes{
	Route{
		"GetBrands",
		"GET",
		"/brands",
		GetBrandsEndpoint,
	},
	Route{
		"GetBrand",
		"GET",
		"/brands/{brandId}",
		GetBrandEndpoint,
	},
	Route{
		"GetPerfumsByBrand",
		"GET",
		"/brands/{brandId}/perfums",
		GetBrandPerfumsEndpoint,
	},
	Route{
		"GetComponents",
		"GET",
		"/components",
		GetComponentsEndpoint,
	},
	Route{
		"GetComponent",
		"GET",
		"/components/{componentId}",
		GetComponentEndpoint,
	},
	Route{
		"GetPerfumsByComponent",
		"GET",
		"/components/{componentId}/perfums",
		GetComponentPerfumsEndpoint,
	},
	Route{
		"GetCountries",
		"GET",
		"/countries",
		GetCountriesEndpoint,
	},
	Route{
		"GetCountry",
		"GET",
		"/countries/{countryId}",
		GetCountryEndpoint,
	},
	Route{
		"GetPerfumsByCountry",
		"GET",
		"/countries/{countryId}/perfums",
		GetCountryPerfumsEndpoint,
	},
	Route{
		"GetGenders",
		"GET",
		"/gender",
		GetGendersEndpoint,
	},
	Route{
		"GetGender",
		"GET",
		"/gender/{genderId}",
		GetGenderEndpoint,
	},
	Route{
		"GetPerfumsByGender",
		"GET",
		"/gender/{genderId}/perfums",
		GetGenderPerfumsEndpoint,
	},
	Route{
		"GetGroups",
		"GET",
		"/groups",
		GetGroupsEndpoint,
	},
	Route{
		"GetGroup",
		"GET",
		"/groups/{groupId}",
		GetGroupEndpoint,
	},
	Route{
		"GetPerfumsByGroup",
		"GET",
		"/groups/{groupId}/perfums",
		GetGroupPerfumsEndpoint,
	},
	Route{
		"GetNotes",
		"GET",
		"/notes",
		GetNotesEndpoint,
	},
	Route{
		"GetNote",
		"GET",
		"/notes/{noteId}",
		GetNoteEndpoint,
	},
	Route{
		"GetPerfumsByNote",
		"GET",
		"/notes/{noteId}/perfums",
		GetNotePerfumsEndpoint,
	},
	Route{
		"GetSeasons",
		"GET",
		"/seasons",
		GetSeasonsEndpoint,
	},
	Route{
		"GetSeason",
		"GET",
		"/seasons/{seasonId}",
		GetSeasonEndpoint,
	},
	Route{
		"GetPerfumsBySeason",
		"GET",
		"/seasons/{seasonId}/perfums",
		GetSeasonPerfumsEndpoint,
	},
	Route{
		"GetTimesOfDay",
		"GET",
		"/timeofday",
		GetTimesOfDayEndpoint,
	},
	Route{
		"GetTimeOfDay",
		"GET",
		"/timeofday/{tsodId}",
		GetTimeOfDayEndpoint,
	},
	Route{
		"GetPerfumsByTimeOfDay",
		"GET",
		"/timeofday/{tsodId}/perfums",
		GetTimeOfDayPerfumsEndpoint,
	},
	Route{
		"GetTypes",
		"GET",
		"/types",
		GetTypesEndpoint,
	},
	Route{
		"GetType",
		"GET",
		"/types/{typeId}",
		GetTypeEndpoint,
	},
	Route{
		"GetPerfumsByType",
		"GET",
		"/types/{typeId}/perfums",
		GetTypePerfumsEndpoint,
	},
	Route{
		"GetPerfums",
		"GET",
		"/perfums",
		GetPerfumsEndpoint,
	},
	Route{
		"GetPerfumDetails",
		"GET",
		"/perfums/find",
		GetPerfumFindEndpoint,
	},
	Route{
		"GetPerfumDetails",
		"GET",
		"/perfums/{perfumId}",
		GetPerfumDetailedInfoEndpoint,
	},
	Route{
		"GetImagesSmall",
		"GET",
		"/images/{imageUuid}/small",
		GetSmallImageEndpoint,
	},
	Route{
		"GetImagesLarge",
		"GET",
		"/images/{imageUuid}/large",
		GetLargeImageEndpoint,
	},
	Route{
		"Test",
		"GET",
		"/test",
		TestEndpoint,
	},
}
