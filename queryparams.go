package main

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	DEFAULT_OFFSET = 0
	DEFAULT_LIMIT  = 10
)

var (
	lang = map[string]string{
		"default": "ru",
		"ru":      "ru",
	}
	supportedVersions = map[string]string{
		"default": "v1",
		"v1":      "v1",
	}
)

type NullInt64 struct {
	Int64 int64 `json:"int64"`
	Valid bool  `json:"valid"`
}

type NullString struct {
	String string `json:"string"`
	Valid  bool   `json:"valid"`
}

type NullSliceInt64 struct {
	Int64 []int64 `json:"[]int64"`
	Valid bool    `json:"valid"`
}

func (ns *NullSliceInt64) append(param string) {
	if param == "" {
		return
	}

	params := strings.Split(param, ",")
	for _, param := range params {
		if intValue, err := strconv.ParseInt(param, 0, 64); err == nil {
			ns.Int64 = append(ns.Int64, intValue)
			ns.Valid = true
		}
	}
}

type NullSliceString struct {
	String []string `json:"[]string"`
	Valid  bool     `json:"valid"`
}

func (ns *NullSliceString) append(param string) {
	if param == "" {
		return
	}

	params := strings.Split(param, ",")
	if len(params) > 0 {
		ns.String = append(ns.String, params...)
		ns.Valid = true
	}
}

// QueryParams ...
type BaseParams struct {
	Limit   NullInt64
	Offset  NullInt64
	Ids     NullSliceString
	Lang    NullString
	Order   NullString
	Version string
}

func getApiVersion(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 2 {
		if version, ok := supportedVersions[strings.ToLower(parts[2])]; ok {
			return version
		}
	}

	return supportedVersions["default"]
}

func NewBaseParams(table string) *BaseParams {
	return &BaseParams{}
}

func (params *BaseParams) Parse(r *http.Request) *BaseParams {
	var err error

	params.Version = getApiVersion(r.URL.Path)
	query := r.URL.Query()
	if param := query.Get("limit"); param != "" {
		if params.Limit.Int64, err = strconv.ParseInt(param, 0, 64); err == nil {
			params.Limit.Valid = true
		}
	}

	if param := query.Get("offset"); param != "" {
		if params.Offset.Int64, err = strconv.ParseInt(param, 0, 64); err == nil {
			params.Offset.Valid = true
		}
	}

	if l, ok := lang[query.Get("lang")]; ok {
		params.Lang.String = l
		params.Lang.Valid = true
	} else {
		params.Lang.String = lang["default"]
		params.Lang.Valid = true
	}

	params.Ids.append(query.Get("id"))

	return params
}

type SearchParams struct {
	Base         BaseParams
	InfoUid      NullSliceString
	Name         NullSliceString
	YearFrom     NullSliceInt64
	YearTo       NullSliceInt64
	DescUid      NullSliceString
	Desc         NullSliceString
	BrandUid     NullSliceString
	Brand        NullSliceString
	GenderUid    NullSliceString
	Gender       NullSliceString
	GroupUid     NullSliceString
	Group        NullSliceString
	CountryUid   NullSliceString
	Country      NullSliceString
	SeasonUid    NullSliceString
	Season       NullSliceString
	TsodUid      NullSliceString
	Tsod         NullSliceString
	TypeUid      NullSliceString
	Type         NullSliceString
	PerfumUid    NullSliceString
	NoteUid      NullSliceString
	Note         NullSliceString
	ComponentUid NullSliceString
	Component    NullSliceString
	Total        int64
}

func NewSearchParams() *SearchParams {
	return &SearchParams{}
}

func (sp *SearchParams) Parse(r *http.Request) *SearchParams {
	sp.Base.Parse(r)
	query := r.URL.Query()

	sp.InfoUid.append(query.Get("info_id"))
	sp.Name.append(query.Get("name"))
	sp.YearFrom.append(query.Get("year_fr"))
	sp.YearTo.append(query.Get("year_to"))
	sp.DescUid.append(query.Get("desc_id"))
	sp.Desc.append(query.Get("desc"))
	sp.BrandUid.append(query.Get("brand_id"))
	sp.Brand.append(query.Get("brand"))
	sp.GenderUid.append(query.Get("gender_id"))
	sp.Gender.append(query.Get("gender"))
	sp.GroupUid.append(query.Get("group_id"))
	sp.Group.append(query.Get("group"))
	sp.CountryUid.append(query.Get("country_id"))
	sp.Country.append(query.Get("country"))
	sp.SeasonUid.append(query.Get("season_id"))
	sp.Season.append(query.Get("season"))
	sp.TsodUid.append(query.Get("tsod_id"))
	sp.Tsod.append(query.Get("tsod"))
	sp.TypeUid.append(query.Get("type_id"))
	sp.Type.append(query.Get("type"))
	sp.PerfumUid.append(query.Get("perfum_id"))
	sp.NoteUid.append(query.Get("note_id"))
	sp.Note.append(query.Get("note"))
	sp.ComponentUid.append(query.Get("component_id"))
	sp.Component.append(query.Get("component"))

	return sp
}
