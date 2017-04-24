package main

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

var (
	NameFields = map[string]LangField{
		"ru": LangField{
			BrandsName:         "name",
			ComponentsName:     "name_ru",
			CountriesName:      "name_ru",
			GenderName:         "name_ru",
			GroupsName:         "name_ru",
			NotesName:          "name_ru",
			PerfumInfo:         "name",
			SeasonsName:        "name_ru",
			ShopsName:          "name_ru",
			TsodName:           "name_ru",
			TypesName:          "name_ru",
			PerfumsDescription: "description_ru",
		},
		"en": LangField{
			BrandsName:         "name",
			ComponentsName:     "name_en",
			CountriesName:      "name_en",
			GenderName:         "name_en",
			GroupsName:         "name_en",
			NotesName:          "name_en",
			PerfumInfo:         "name",
			SeasonsName:        "name_en",
			ShopsName:          "name_en",
			TsodName:           "name_en",
			TypesName:          "name_en",
			PerfumsDescription: "description_en",
		},
		"default": LangField{
			BrandsName:         "name",
			ComponentsName:     "name_ru",
			CountriesName:      "name_ru",
			GenderName:         "name_ru",
			GroupsName:         "name_ru",
			NotesName:          "name_ru",
			PerfumInfo:         "name",
			SeasonsName:        "name_ru",
			ShopsName:          "name_ru",
			TsodName:           "name_ru",
			TypesName:          "name_ru",
			PerfumsDescription: "description_ru",
		},
	}

	dbmap           *gorp.DbMap
	regex           *regexp.Regexp
	PfumsCountCache = map[string]PfumsCountCacheItem{
		"brands": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT brands.id AS id, brands.uuid AS uid FROM brands",
			getCountDbQuery: "SELECT COUNT(*) FROM parfum_info WHERE brand_id=$1",
			count:           make(map[string]int64),
		},
		"components": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT components.id AS id, components.uuid AS uid FROM components",
			getCountDbQuery: "SELECT COUNT(DISTINCT parfum_info_id) FROM parfums WHERE component_id=$1",
			count:           make(map[string]int64),
		},
		"countries": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT countries.id AS id, countries.uuid AS uid FROM countries",
			getCountDbQuery: "SELECT COUNT(*) FROM parfum_info WHERE country_id=$1",
			count:           make(map[string]int64),
		},
		"genders": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT gender.id AS id, gender.uuid AS uid FROM gender",
			getCountDbQuery: "SELECT COUNT(*) FROM parfum_info WHERE gender_id=$1",
			count:           make(map[string]int64),
		},
		"groups": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT groups.id AS id, groups.uuid AS uid FROM groups",
			getCountDbQuery: "SELECT COUNT(*) FROM parfum_info WHERE group_id=$1",
			count:           make(map[string]int64),
		},
		"notes": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT notes.id AS id, notes.uuid AS uid FROM notes",
			getCountDbQuery: "SELECT COUNT(DISTINCT parfum_info_id) FROM parfums WHERE note_id=$1",
			count:           make(map[string]int64),
		},
		"seasons": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT seasons.id AS id, seasons.uuid AS uid FROM seasons",
			getCountDbQuery: "SELECT COUNT(*) FROM parfum_info WHERE season_id=$1",
			count:           make(map[string]int64),
		},
		"timesOfDay": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT times_of_day.id AS id, times_of_day.uuid AS uid FROM times_of_day",
			getCountDbQuery: "SELECT COUNT(*) FROM parfum_info WHERE tsod_id=$1",
			count:           make(map[string]int64),
		},
		"types": PfumsCountCacheItem{
			getItemsDbQuery: "SELECT types.id AS id, types.uuid AS uid FROM types",
			getCountDbQuery: "SELECT COUNT(*) FROM parfum_info WHERE type_id=$1",
			count:           make(map[string]int64),
		},
	}
)

type PfumsCountCacheItem struct {
	getItemsDbQuery string
	getCountDbQuery string
	mutex           sync.RWMutex
	count           map[string]int64
}

type LangField struct {
	BrandsName         string
	ComponentsName     string
	CountriesName      string
	GenderName         string
	GroupsName         string
	NotesName          string
	PerfumInfo         string
	SeasonsName        string
	ShopsName          string
	TsodName           string
	TypesName          string
	PerfumsDescription string
}

type ConditionParams struct {
	AuxConditionString   string
	WhereConditionString string
	AndConditionString   string
}

type ConditionSelectTemplateParams struct {
	ConditionTableField string
	ConditionTableName  string
	ConditionUuid       string
}

type CountTemplateParams struct {
	DistinctTableField string
}

type BaseQueryTemplateParams struct {
	Order         string
	Offset        string
	Limit         string
	FromTableName string
}

type QueryTemplateParams struct {
	LangField
	ConditionParams
	ConditionSelectTemplateParams
	CountTemplateParams
	BaseQueryTemplateParams
}

type SearchQueryTemplateParams struct {
	LangField
	BaseQueryTemplateParams
	InfoUid      string
	Name         string
	YearFrom     string
	YearTo       string
	DescUid      string
	Desc         string
	BrandUid     string
	Brand        string
	GenderUid    string
	Gender       string
	GroupUid     string
	Group        string
	CountryUid   string
	Country      string
	SeasonUid    string
	Season       string
	TsodUid      string
	Tsod         string
	TypeUid      string
	Type         string
	PerfumUid    string
	NoteUid      string
	Note         string
	ComponentUid string
	Component    string
}

func NewSearchQueryTemplateParams() *SearchQueryTemplateParams {
	return &SearchQueryTemplateParams{}
}

func (sp *SearchQueryTemplateParams) ParseBaseParams(params *BaseParams) error {
	sp.BaseQueryTemplateParams.Order = "name"

	if params.Lang.Valid {
		sp.LangField = getNameFields(params.Lang.String)
	} else {
		sp.LangField = getNameFields("default")
	}

	if params.Offset.Valid {
		sp.BaseQueryTemplateParams.Offset = strconv.FormatInt(params.Offset.Int64, 10)
	} else {
		sp.BaseQueryTemplateParams.Offset = strconv.FormatInt(DEFAULT_OFFSET, 10)
	}

	if params.Limit.Valid {
		sp.BaseQueryTemplateParams.Limit = strconv.FormatInt(params.Limit.Int64, 10)
	} else {
		sp.BaseQueryTemplateParams.Limit = strconv.FormatInt(DEFAULT_LIMIT, 10)
	}

	return nil
}

func (sp *SearchQueryTemplateParams) ParseSearchParams(params *SearchParams) error {
	sp.ParseBaseParams(&params.Base)

	if params.InfoUid.Valid && len(params.InfoUid.String) > 0 {
		sp.InfoUid = addUidToQuery(params.InfoUid.String, "parfum_info.uuid")
	}

	if params.Name.Valid && len(params.Name.String) > 0 {
		sp.Name = addSubstringToQuery(params.Name.String, "parfum_info."+sp.LangField.PerfumInfo, params.CompareMode, params.CaseSensitive)
	}

	if params.YearFrom.Valid && len(params.YearFrom.Int64) > 0 {
		sp.YearFrom = addIntToQueryConditionGE(params.YearFrom.Int64, "parfum_info.year")
	}

	if params.YearTo.Valid && len(params.YearTo.Int64) > 0 {
		sp.YearTo = addIntToQueryConditionLE(params.YearTo.Int64, "parfum_info.year")
	}

	if params.DescUid.Valid && len(params.DescUid.String) > 0 {
		sp.DescUid = addUidToQuery(params.DescUid.String, "descriptions.uuid")
	}

	if params.Desc.Valid && len(params.Desc.String) > 0 {
		sp.Desc = addSubstringToQuery(params.Desc.String, "descriptions."+sp.LangField.PerfumsDescription, params.CompareMode, params.CaseSensitive)
	}

	if params.BrandUid.Valid && len(params.BrandUid.String) > 0 {
		sp.BrandUid = addUidToQuery(params.BrandUid.String, "brands.uuid")
	}

	if params.Brand.Valid && len(params.Brand.String) > 0 {
		sp.Brand = addSubstringToQuery(params.Brand.String, "brands."+sp.LangField.BrandsName, params.CompareMode, params.CaseSensitive)
	}

	if params.GenderUid.Valid && len(params.GenderUid.String) > 0 {
		sp.GenderUid = addUidToQuery(params.GenderUid.String, "gender.uuid")
	}

	if params.Gender.Valid && len(params.Gender.String) > 0 {
		sp.Gender = addSubstringToQuery(params.Gender.String, "gender."+sp.LangField.GenderName, params.CompareMode, params.CaseSensitive)
	}

	if params.GroupUid.Valid && len(params.GroupUid.String) > 0 {
		sp.GroupUid = addUidToQuery(params.GroupUid.String, "groups.uuid")
	}

	if params.Group.Valid && len(params.Group.String) > 0 {
		sp.Group = addSubstringToQuery(params.Group.String, "groups."+sp.LangField.GroupsName, params.CompareMode, params.CaseSensitive)
	}

	if params.CountryUid.Valid && len(params.CountryUid.String) > 0 {
		sp.CountryUid = addUidToQuery(params.CountryUid.String, "countries.uuid")
	}

	if params.Country.Valid && len(params.Country.String) > 0 {
		sp.Country = addSubstringToQuery(params.Country.String, "countries."+sp.LangField.CountriesName, params.CompareMode, params.CaseSensitive)
	}

	if params.SeasonUid.Valid && len(params.SeasonUid.String) > 0 {
		sp.SeasonUid = addUidToQuery(params.SeasonUid.String, "seasons.uuid")
	}

	if params.Season.Valid && len(params.Season.String) > 0 {
		sp.Season = addSubstringToQuery(params.Season.String, "seasons."+sp.LangField.SeasonsName, params.CompareMode, params.CaseSensitive)
	}

	if params.TsodUid.Valid && len(params.TsodUid.String) > 0 {
		sp.TsodUid = addUidToQuery(params.TsodUid.String, "times_of_day.uuid")
	}

	if params.Tsod.Valid && len(params.Tsod.String) > 0 {
		sp.Tsod = addSubstringToQuery(params.Tsod.String, "times_of_day."+sp.LangField.TsodName, params.CompareMode, params.CaseSensitive)
	}

	if params.TypeUid.Valid && len(params.TypeUid.String) > 0 {
		sp.TypeUid = addUidToQuery(params.TypeUid.String, "types.uuid")
	}

	if params.Type.Valid && len(params.Type.String) > 0 {
		sp.Type = addSubstringToQuery(params.Type.String, "types."+sp.LangField.TypesName, params.CompareMode, params.CaseSensitive)
	}

	if params.PerfumUid.Valid && len(params.PerfumUid.String) > 0 {
		sp.PerfumUid = addUidToQuery(params.PerfumUid.String, "parfums.uuid")
	}

	if params.NoteUid.Valid && len(params.NoteUid.String) > 0 {
		sp.NoteUid = addUidToQuery(params.NoteUid.String, "notes.uuid")
	}

	if params.Note.Valid && len(params.Note.String) > 0 {
		sp.Note = addSubstringToQuery(params.Note.String, "notes."+sp.LangField.NotesName, params.CompareMode, params.CaseSensitive)
	}

	if params.ComponentUid.Valid && len(params.ComponentUid.String) > 0 {
		sp.ComponentUid = addUidToQuery(params.ComponentUid.String, "components.uuid")
	}

	if params.Component.Valid && len(params.Component.String) > 0 {
		sp.Component = addSubstringToQuery(params.Component.String, "components."+sp.LangField.ComponentsName, params.CompareMode, params.CaseSensitive)
	}

	return nil
}

func addUidToQuery(slice []string, fieldId string) string {
	ret := ""
	if num := len(slice); num > 0 {
		for i, value := range slice {
			if normalized := regex.FindString(value); normalized != "" {
				if i > 0 {
					ret += " OR "
				}
				ret += fieldId + "='" + normalized + "'"
			}
		}
	}

	return ret
}

func addSubstringToQuery(slice []string, fieldId string, cm, cs NullString) string {
	ret := ""
	if num := len(slice); num > 0 {
		for i, value := range slice {
			if normalized := regex.FindString(value); normalized != "" {
				if i > 0 {
					ret += " OR "
				}

				if !cs.Valid || cs.String != "y" {
					ret += `LOWER`
				}

				ret += `(` + fieldId + `) LIKE `

				if !cs.Valid || cs.String != "y" {
					ret += `LOWER`
				}

				if cm.String == "st" { //strict
					ret += `('` + normalized + `')`
				} else if cm.String == "bw" { //begin with
					ret += `('` + normalized + `%')`
				} else if cm.String == "ew" { //end with
					ret += `('%` + normalized + `')`
				} else { //at any position in the string
					ret += `('%` + normalized + `%')`
				}
			}
		}
	}

	return ret
}

func addIntToQueryConditionGE(slice []int64, fieldId string) string {
	ret := ""
	if num := len(slice); num > 0 {
		for i, value := range slice {
			if i > 0 {
				ret += " OR "
			}
			ret += fieldId + ">=" + strconv.FormatInt(value, 10)
		}
	}

	return ret
}

func addIntToQueryConditionLE(slice []int64, fieldId string) string {
	ret := ""
	if num := len(slice); num > 0 {
		for i, value := range slice {
			if i > 0 {
				ret += " OR "
			}
			ret += fieldId + "<=" + strconv.FormatInt(value, 10)
		}
	}

	return ret
}

// DBObjecter ...
type DBObjecter interface {
	GetRecords(params *BaseParams) ([]interface{}, error)
	GetPerfumInfoWithUid(params *BaseParams, uuids []string) ([]interface{}, error)
	GetCount() (int64, error)
	GetPerfumsCount(uuids []string) (int64, error)
}

// UserDB ...
type UserDB struct {
	UserId       string `db:"user_id"`
	AccessToken  string `db:"access_token"`
	RefreshToken string `db:"refresh_token"`
	ExpiresAt    int64  `db:"expires_at"`
	CreatedAt    int64  `db:"created_at"`
	UpdatedAt    int64  `db:"updated_at"`
}

// ImagesDB ...
type ImageDB struct {
	Id            int64          `db:"id"`
	SmallImgFname sql.NullString `db:"small_img_filename"`
	SmallImgPath  sql.NullString `db:"small_img_path"`
	SmallImgLink  sql.NullString `db:"small_img_link"`
	SmallImgEtag  sql.NullString `db:"small_img_etag"`
	LargeImgFname sql.NullString `db:"large_img_filename"`
	LargeImgPath  sql.NullString `db:"large_img_path"`
	LargeImgLink  sql.NullString `db:"large_img_link"`
	LargeImgEtag  sql.NullString `db:"large_img_etag"`
	ImgUuid       sql.NullString `db:"uuid"`
	Note          sql.NullString `db:"note"`
}

var templateFile = "query_templates.tmpl"
var tmpl *template.Template
var whereIsUsed = false

func init() {
	_ = godotenv.Load("openshift.env")
	dir := os.Getenv("OPENSHIFT_REPO_DIR")
	if dir == "" {
		TraceFatal("Variable OPENSHIFT_REPO_DIR is not defined")
		os.Exit(-1)
	}

	templateFile = filepath.Join(dir, templateFile)
	tmpl = template.Must(template.New("").Funcs(
		template.FuncMap{
			"SetWhereIsUsed": func(newValue bool) bool {
				whereIsUsed = newValue
				return whereIsUsed
			},
			"GetWhereIsUsed": func() bool {
				return whereIsUsed
			},
		},
	).ParseFiles(templateFile))

	regex = regexp.MustCompile(`(([\p{L}|\p{Nd}]+\s*[-|&]?\s*)+[\p{L}|\p{Nd}]+[']?([\p{L}|\p{Nd}]+\s*[-|&]?\s*)+[\p{L}|\p{Nd}]*)|([\p{L}|\p{Nd}]{1,2})`)
}

// InitDb ...
func InitDb() *gorp.DbMap {
	db, err := sql.Open("postgres", os.Getenv("OPENSHIFT_POSTGRESQL_DB_URL")+"/"+os.Getenv("FRAGRANCES_DB_NAME")+"?sslmode=disable")
	if err != nil {
		TracePrintError(err)
		os.Exit(-1)
	}
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(UserDB{}, "users").SetKeys(false, "UserId")
	dbmap.AddTableWithName(BrandV1{}, "brands").SetKeys(false, "Id")
	dbmap.AddTableWithName(ImageDB{}, "images").SetKeys(false, "Id")
	dbmap.AddTableWithName(PerfumInfoV1{}, "parfum_info").SetKeys(false, "Id")
	dbmap.AddTableWithName(ComponentV1{}, "components").SetKeys(false, "Id")
	dbmap.AddTableWithName(CountryV1{}, "countries").SetKeys(false, "Id")
	dbmap.AddTableWithName(GenderV1{}, "gender").SetKeys(false, "Id")
	dbmap.AddTableWithName(GroupV1{}, "groups").SetKeys(false, "Id")
	dbmap.AddTableWithName(SeasonV1{}, "seasons").SetKeys(false, "Id")
	dbmap.AddTableWithName(TimeOfDayV1{}, "times_of_day").SetKeys(false, "Id")
	dbmap.AddTableWithName(TypeV1{}, "types").SetKeys(false, "Id")
	dbmap.AddTableWithName(PerfumCompositionDBRecordV1{}, "parfums").SetKeys(false, "PerfumId")
	// dbmap.TraceOn("[gorp]", log.New(os.Stdout, "fga:", log.Lmicroseconds))

	go func() {
		c := time.Tick(time.Duration(4) * time.Hour)
		for _, pcci := range PfumsCountCache {
			CachePerfumsCount(&pcci)
		}

		for _ = range c {
			for _, pcci := range PfumsCountCache {
				CachePerfumsCount(&pcci)
				time.Sleep(time.Duration(10) * time.Second)
			}
		}
	}()

	return dbmap
}

func GetDbMap() *gorp.DbMap {
	return dbmap
}

// GetUserByUserId ...
func GetUserByUserId(UserId string) (user *UserDB, err error) {
	if UserId == "" {
		TracePrint("user == nil")
		return nil, errors.New("bad arg")
	}
	obj, err := dbmap.Get(UserDB{}, UserId)
	if err != nil {
		TracePrintError(err)
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	user = obj.(*UserDB)
	return user, nil
}

// GetUserByAccessToken ...
func GetUserByAccessToken(tok string) (*UserDB, error) {
	if tok == "" {
		return nil, errors.New("bad arg")
	}
	var user UserDB
	if err := dbmap.SelectOne(&user, "SELECT * FROM users WHERE access_token=$1", tok); err != nil {
		TracePrintError(err)
		return nil, err
	}
	return &user, nil
}

// GetUserByRefreshToken ...
func GetUserByRefreshToken(tok string) (*UserDB, error) {
	if tok == "" {
		return nil, errors.New("bad arg")
	}
	var user UserDB
	if err := dbmap.SelectOne(&user, "SELECT * FROM users WHERE refresh_token=$1", tok); err != nil {
		TracePrintError(err)
		return nil, err
	}
	return &user, nil
}

// UserInsert ...
func UserInsert(userId, accessToken, refreshToken string, expiresAt int64) (*UserDB, error) {
	if accessToken == "" || userId == "" {
		return nil, errors.New("bad arg")
	}

	now := time.Now()
	newUser := &UserDB{
		UserId:       userId,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    now.Unix(),
		UpdatedAt:    now.Unix(),
	}
	if err := dbmap.Insert(newUser); err != nil {
		TracePrintError(err)
		return nil, err
	}
	return newUser, nil
}

// Update ...
func (u *UserDB) Update(accessToken, refreshToken string, expiresAt int64) (bool, error) {
	if u.UserId == "" || accessToken == "" || refreshToken == "" {
		return false, errors.New("bad arg")
	}

	u.AccessToken = accessToken
	u.RefreshToken = refreshToken
	u.ExpiresAt = expiresAt
	u.UpdatedAt = time.Now().Unix()
	count, err := dbmap.Update(u)
	if err != nil {
		TracePrintError(err)
		return false, err
	} else if count == 0 {
		return false, nil
	}
	return true, nil
}

// Delete ...
func (u *UserDB) Delete() (bool, error) {
	if u.UserId == "" {
		return false, errors.New("bad arg")
	}
	count, err := dbmap.Delete(u)
	if err != nil {
		TracePrintError(err)
		return false, err
	} else if count == 0 {
		return false, nil
	}
	return true, nil
}

func addIdsToQuery(ids []string, fieldId string) string {
	ret := ""
	if idsLen := len(ids); idsLen > 0 {
		for i, id := range ids {
			ret += fieldId + "='" + id + "'"
			if i < (idsLen - 1) {
				ret += " OR "
			}
		}
	}

	return ret
}

func addParamsToQuery(base string, params *BaseParams, condition string) string {
	if condition != "" {
		base += " WHERE (" + condition + ")"
	}
	base += " ORDER BY name ASC"

	if params.Offset.Valid {
		base += " OFFSET " + strconv.FormatInt(params.Offset.Int64, 10)
	} else {
		base += " OFFSET " + strconv.FormatInt(DEFAULT_OFFSET, 10)
	}

	if params.Limit.Valid {
		base += " LIMIT " + strconv.FormatInt(params.Limit.Int64, 10)
	} else {
		base += " LIMIT " + strconv.FormatInt(DEFAULT_LIMIT, 10)
	}

	return base
}

func getNameFields(lang string) LangField {
	if lang == "" || lang == "default" {
		return NameFields["default"]
	}
	if names, ok := NameFields[lang]; ok {
		return names
	}
	return NameFields["default"]
}

func setDbQueryBaseParams(params *BaseParams, dbParams *QueryTemplateParams) error {
	dbParams.Order = "name"

	if params.Lang.Valid {
		dbParams.LangField = getNameFields(params.Lang.String)
	} else {
		dbParams.LangField = getNameFields("default")
	}

	if params.Offset.Valid {
		dbParams.Offset = strconv.FormatInt(params.Offset.Int64, 10)
	} else {
		dbParams.Offset = strconv.FormatInt(DEFAULT_OFFSET, 10)
	}

	if params.Limit.Valid {
		dbParams.Limit = strconv.FormatInt(params.Limit.Int64, 10)
	} else {
		dbParams.Limit = strconv.FormatInt(DEFAULT_LIMIT, 10)
	}

	return nil
}

// GetImageById ...
func GetImageById(id int64) (*ImageDB, error) {
	image := ImageDB{}
	if err := dbmap.SelectOne(&image, "SELECT * FROM images WHERE id=$1", id); err != nil {
		TracePrintError(err)
		return nil, err
	}
	return &image, nil
}

// GetImageByUuid ...
func GetImageByUuid(uuid string) (*ImageDB, error) {
	image := ImageDB{}
	if err := dbmap.SelectOne(&image, "SELECT * FROM images WHERE uuid=$1", uuid); err != nil {
		TracePrintError(err)
		return nil, err
	}
	return &image, nil
}

func CachePerfumsCount(cacheItem *PfumsCountCacheItem) error {
	type DbItem struct {
		Id  string `db:"id"`
		Uid string `db:"uid"`
	}

	var items []DbItem
	if _, err := dbmap.Select(&items, cacheItem.getItemsDbQuery); err != nil {
		return err
	}

	if len(items) > 0 {
		temp := make(map[string]int64)
		for _, b := range items {
			count, err := dbmap.SelectInt(cacheItem.getCountDbQuery, b.Id)
			if err != nil {
				return err
			}
			temp[b.Uid] = count
		}
		if len(temp) > 0 {
			cacheItem.mutex.Lock()
			for k, v := range temp {
				cacheItem.count[k] = v
			}
			cacheItem.mutex.Unlock()
		}
	}

	return nil
}

func GetPerfumsCount(table, uid string) (int64, bool) {
	cacheItem, found := PfumsCountCache[table]
	if !found {
		return 0, false
	}

	cacheItem.mutex.RLock()
	value, found := cacheItem.count[uid]
	cacheItem.mutex.RUnlock()

	return value, found
}
