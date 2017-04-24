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

// Unauthorized. Path prefix: /
// Handle without middleware
var publicRoutes = Routes{
	Route{
		"Login",
		"POST",
		"/login",
		LoginEndpoint,
	},
	Route{
		"Token",
		"POST",
		"/token",
		TokenEndpoint,
	},
}

var privateRoutes = Routes{
	Route{
		"GetUser",
		"GET",
		"/user/{userId}",
		GetUserEndpoint,
	},
	Route{
		"DeleteUser",
		"DELETE",
		"/user/{userId}",
		DeleteUserEndpoint,
	},
	Route{
		"Logout",
		"PUT",
		"/user/{userId}/logout",
		LogoutEndpoint,
	},
	// Route{
	// 	"RefreshToken",
	// 	"GET",
	// 	"/users/{userId}/refresh",
	// 	RefreshTokenEndpoint,
	// },
	Route{
		"GetUserFavorites",
		"GET",
		"/user/{userId}/favorites",
		GetUserFavoritesEndpoint,
	},
	Route{
		"CreateUserFavorites",
		"POST",
		"/user/{userId}/favorites",
		CreateUserFavoritesEndpoint,
	},
	Route{
		"UpdateUserFavorites",
		"PUT",
		"/user/{userId}/favorites",
		UpdateUserFavoritesEndpoint,
	},
	Route{
		"DeleteUserFavorites",
		"DELETE",
		"/user/{userId}/favorites",
		DeleteUserFavoritesEndpoint,
	},
	Route{
		"GetBrands",
		"GET",
		"/brands",
		GetBrandsEndpoint,
	},
	Route{
		"GetBrandsFind",
		"GET",
		"/brands/find",
		GetBrandsFindEndpoint,
	},
	Route{
		"GetBrand",
		"GET",
		"/brand/{brandId}",
		GetBrandEndpoint,
	},
	Route{
		"GetPerfumsByBrand",
		"GET",
		"/brand/{brandId}/perfums",
		GetBrandPerfumsEndpoint,
	},
	Route{
		"GetComponents",
		"GET",
		"/components",
		GetComponentsEndpoint,
	},
	Route{
		"GetComponentsFind",
		"GET",
		"/components/find",
		GetComponentsFindEndpoint,
	},
	Route{
		"GetComponent",
		"GET",
		"/component/{componentId}",
		GetComponentEndpoint,
	},
	Route{
		"GetPerfumsByComponent",
		"GET",
		"/component/{componentId}/perfums",
		GetComponentPerfumsEndpoint,
	},
	Route{
		"GetCountries",
		"GET",
		"/countries",
		GetCountriesEndpoint,
	},
	Route{
		"GetCountriesFind",
		"GET",
		"/countries/find",
		GetCountriesFindEndpoint,
	},
	Route{
		"GetCountry",
		"GET",
		"/country/{countryId}",
		GetCountryEndpoint,
	},
	Route{
		"GetPerfumsByCountry",
		"GET",
		"/country/{countryId}/perfums",
		GetCountryPerfumsEndpoint,
	},
	Route{
		"GetGenders",
		"GET",
		"/genders",
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
		"GetGroupsFind",
		"GET",
		"/groups/find",
		GetGroupsFindEndpoint,
	},
	Route{
		"GetGroup",
		"GET",
		"/group/{groupId}",
		GetGroupEndpoint,
	},
	Route{
		"GetPerfumsByGroup",
		"GET",
		"/group/{groupId}/perfums",
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
		"/note/{noteId}",
		GetNoteEndpoint,
	},
	Route{
		"GetPerfumsByNote",
		"GET",
		"/note/{noteId}/perfums",
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
		"/season/{seasonId}",
		GetSeasonEndpoint,
	},
	Route{
		"GetPerfumsBySeason",
		"GET",
		"/season/{seasonId}/perfums",
		GetSeasonPerfumsEndpoint,
	},
	Route{
		"GetTimesOfDay",
		"GET",
		"/timesofday",
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
		"/type/{typeId}",
		GetTypeEndpoint,
	},
	Route{
		"GetPerfumsByType",
		"GET",
		"/type/{typeId}/perfums",
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
		GetPerfumsFindEndpoint,
	},
	Route{
		"GetPerfumDetails",
		"GET",
		"/perfum/{perfumId}",
		GetPerfumDetailedInfoEndpoint,
	},
	Route{
		"GetImagesSmall",
		"GET",
		"/image/{imageId}/small",
		GetSmallImageEndpoint,
	},
	Route{
		"GetImagesLarge",
		"GET",
		"/image/{imageId}/large",
		GetLargeImageEndpoint,
	},
}
