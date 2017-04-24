package main

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/unrolled/render"
	"fmt"
	"net/http"
	"sort"
)

type MakeObjParams struct {
	Base       BaseParams
	Id         string
	Total      int64
	PerfumsNum NullInt64
	DbQuery    QueryTemplateParams
}

type Objecter interface {
	MakeObj(pParams interface{}) (Objecter, error)
	MakeExtraObj(params *MakeObjParams, uuids []string) (Objecter, error)
	Count(pParams interface{}) (int64, error)
	ExtraCount(uuids []string) (int64, error)
	Json(w http.ResponseWriter, status int) error
}

// Link ...
type LinkV1 struct {
	Href   string `db:"-" json:"href"`
	Rel    string `db:"-" json:"rel"`
	Method string `db:"-" json:"method"`
}

// PerfumInfoV1 ...
type PerfumInfoV1 struct {
	Id              string         `db:"info_id" json:"-"`
	Uuid            string         `db:"info_uuid" json:"id"`
	Name            string         `db:"name" json:"name"`
	DescriptionUuid string         `db:"description_uuid" json:"description_id"`
	Description     string         `db:"description" json:"description"`
	Year            int64          `db:"info_year" json:"year"`
	BrandUuid       string         `db:"brand_uuid" json:"brand_id"`
	BrandName       string         `db:"brand_name" json:"brand_name"`
	GenderUuid      string         `db:"gender_uuid" json:"gender_id"`
	GenderName      string         `db:"gender_name" json:"gender_name"`
	GroupUuid       string         `db:"group_uuid" json:"group_id"`
	GroupName       string         `db:"group_name" json:"group_name"`
	CountryUuid     string         `db:"country_uuid" json:"country_id"`
	CountryName     string         `db:"country_name" json:"country_name"`
	SeasonUuid      string         `db:"season_uuid" json:"season_id"`
	SeasonName      string         `db:"season_name" json:"season_name"`
	TsodUuid        string         `db:"tsod_uuid" json:"tsod_id"`
	TsodName        string         `db:"tsod_name" json:"tsod_name"`
	TypeUuid        string         `db:"type_uuid" json:"type_id"`
	TypeName        string         `db:"type_name" json:"type_name"`
	ImgUuid         sql.NullString `db:"img_uuid" json:"-"`
	StarsUuid       sql.NullString `db:"stars_uuid" json:"stars_id"`
	ShopUuid        sql.NullString `db:"shop_uuid" json:"shop_id"`
	Links           []LinkV1       `db:"-" json:"links"`
	SmallImgUrl     string         `db:"-" json:"small_img_url"`
	LargeImgUrl     string         `db:"-" json:"large_img_url"`
}

type PerfumsInfoV1 struct {
	ObjList []PerfumInfoV1 `db:"-" json:"perfums_info_list"`
	Total   int64          `db:"-" json:"total"`
	Offset  int64          `db:"-" json:"offset"`
	Amount  int64          `db:"-" json:"amount"`
}

func NewPerfumsInfoFactory(version string) Objecter {
	switch version {
	case "v1":
		return &PerfumsInfoV1{ObjList: make([]PerfumInfoV1, 0)}
	}

	return nil
}

func (obj *PerfumsInfoV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}
	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	fmt.Println("================================= Query begin =======================================")
	fmt.Println(query.String())
	fmt.Println("================================= Query end =========================================")

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/perfum/" + obj.ObjList[i].Uuid,
				Rel:    "PerfumInfo",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImgUuid.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImgUuid.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImgUuid.String + "/large"
		}
	}

	return obj, nil
}

func (obj *PerfumsInfoV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if len(uids) > 0 {
		params.DbQuery.WhereConditionString = addIdsToQuery(uids, "parfum_info.uuid")
		params.Base.Ids.Valid = false
	}

	composition := NewPerfumsCompositionFactory(params.Base.Version)
	return composition.MakeObj(params)
}

func (obj *PerfumsInfoV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *PerfumsInfoV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.WhereConditionString = addIdsToQuery(uids, "parfum_info.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *PerfumsInfoV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type PerfumCompositionDBRecordV1 struct {
	PerfumId       string `db:"perfum_id"`
	PerfumUuid     string `db:"perfum_uuid"`
	NoteUuid       string `db:"note_uuid"`
	NoteName       string `db:"note_name"`
	ComponentUuid  string `db:"component_uuid"`
	ComponentName  string `db:"component_name"`
	PerfumInfoUuid string `db:"info_uuid"`
}

type ComponentItemV1 struct {
	Id    string   `json:"component_id"`
	Name  string   `json:"component_name"`
	Links []LinkV1 `json:"links"`
}

func NewComponentItemV1(id, name string) *ComponentItemV1 {
	return &ComponentItemV1{Id: id, Name: name}
}

type ByComponentName []ComponentItemV1

func (c ByComponentName) Len() int {
	return len(c)
}

func (c ByComponentName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c ByComponentName) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

type NoteItemV1 struct {
	Id             string            `json:"note_id"`
	Name           string            `json:"note_name"`
	Components     []ComponentItemV1 `json:"components"`
	Links          []LinkV1          `json:"links"`
	ComponentCount int64             `json:"component_count"`
}

func NewNoteItemV1(id, name string) *NoteItemV1 {
	return &NoteItemV1{Id: id, Name: name, Components: []ComponentItemV1{}}
}

func (note *NoteItemV1) AddComponentItem(componentToAdd *ComponentItemV1) *NoteItemV1 {
	if componentToAdd == nil {
		return note
	}

	componentToAdd.Links = append(componentToAdd.Links,
		LinkV1{
			Href:   baseUrl + "/component/" + componentToAdd.Id,
			Rel:    "ComponentInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/component/" + componentToAdd.Id + "/perfums",
			Rel:    "ComponentPerfums",
			Method: "GET",
		})
	note.Components = append(note.Components, *componentToAdd)

	return note
}

type ByNoteName []NoteItemV1

func (n ByNoteName) Len() int {
	return len(n)
}

func (n ByNoteName) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n ByNoteName) Less(i, j int) bool {
	return n[i].Name < n[j].Name
}

type PerfumCompositionV1 struct {
	PerfumInfoV1
	Notes           []NoteItemV1 `json:"notes"`
	TotalComponents int64        `json:"total_components"`
}

func (obj *PerfumCompositionV1) AddNoteItem(noteToAdd *NoteItemV1) *PerfumCompositionV1 {
	if noteToAdd == nil {
		return obj
	}

	noteToAdd.Links = append(noteToAdd.Links,
		LinkV1{
			Href:   baseUrl + "/note/" + noteToAdd.Id,
			Rel:    "NoteInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/note/" + noteToAdd.Id + "/perfums",
			Rel:    "NotePerfums",
			Method: "GET",
		})

	obj.Notes = append(obj.Notes, *noteToAdd)

	return obj
}

func (obj *PerfumCompositionV1) AddPerfumInfoItem(info *PerfumInfoV1) *PerfumCompositionV1 {
	if info == nil {
		return obj
	}

	obj.PerfumInfoV1 = *info
	obj.PerfumInfoV1.Links = []LinkV1{
		LinkV1{
			Href:   baseUrl + "/brand/" + info.BrandUuid,
			Rel:    "BrandInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/brand/" + info.BrandUuid + "/perfums",
			Rel:    "BrandPerfums",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/country/" + info.CountryUuid,
			Rel:    "CountryInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/country/" + info.CountryUuid + "/perfums",
			Rel:    "CountryPerfums",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/gender/" + info.GenderUuid,
			Rel:    "GenderInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/gender/" + info.GenderUuid + "/perfums",
			Rel:    "GenderPerfums",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/group/" + info.GroupUuid,
			Rel:    "GroupInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/group/" + info.GroupUuid + "/perfums",
			Rel:    "GroupPerfums",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/season/" + info.SeasonUuid,
			Rel:    "SeasonInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/season/" + info.SeasonUuid + "/perfums",
			Rel:    "SeasonPerfums",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/timeofday/" + info.TsodUuid,
			Rel:    "TimeofdayInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/timeofday/" + info.TsodUuid + "/perfums",
			Rel:    "TimeofdayPerfums",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/type/" + info.TypeUuid,
			Rel:    "TypeInfo",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/type/" + info.TypeUuid + "/perfums",
			Rel:    "TypePerfums",
			Method: "GET",
		},
		LinkV1{
			Href:   baseUrl + "/perfum/" + info.Uuid,
			Rel:    "PerfumInfo",
			Method: "GET",
		},
	}

	if info.ImgUuid.Valid {
		obj.SmallImgUrl = baseUrl + "/image/" + info.ImgUuid.String + "/small"
		obj.LargeImgUrl = baseUrl + "/image/" + info.ImgUuid.String + "/large"
	}

	return obj
}

type PerfumsCompositionV1 struct {
	ObjList []PerfumCompositionV1 `db:"-" json:"perfums_composition"`
	Total   int64                 `db:"-" json:"total"`
	Offset  int64                 `db:"-" json:"offset"`
	Amount  int64                 `db:"-" json:"amount"`
}

func NewPerfumCompositionV1() *PerfumCompositionV1 {
	return &PerfumCompositionV1{
		Notes: []NoteItemV1{},
		PerfumInfoV1: PerfumInfoV1{
			Links: []LinkV1{},
		},
	}
}

func NewPerfumsCompositionFactory(version string) Objecter {
	switch version {
	case "v1":
		return &PerfumsCompositionV1{
			ObjList: []PerfumCompositionV1{},
		}
	}

	return nil
}

func (obj *PerfumsCompositionV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	query := bytes.NewBufferString("")

	perfumInfos := PerfumsInfoV1{}
	perfumInfos.MakeObj(params)

	perfumInfoMap := make(map[string]*PerfumInfoV1)
	for i := range perfumInfos.ObjList {
		perfumInfoMap[perfumInfos.ObjList[i].Uuid] = &perfumInfos.ObjList[i]
	}

	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_perfums_on_perfum_info_uuid", &params.DbQuery); err != nil {
		return nil, err
	}

	var records []PerfumCompositionDBRecordV1
	if _, err = dbmap.Select(&records, query.String()); err != nil {
		return nil, err
	}

	type NoteList struct {
		Name       string
		Components map[string]string
	}

	type Perfum struct {
		Notes      map[string]NoteList
		PerfumInfo PerfumInfoV1
	}

	perfums := make(map[string]Perfum)

	for _, record := range records {
		if perfum, found := perfums[record.PerfumInfoUuid]; !found {
			perfums[record.PerfumInfoUuid] = Perfum{
				Notes: map[string]NoteList{
					record.NoteUuid: NoteList{
						Name: record.NoteName,
						Components: map[string]string{
							record.ComponentUuid: record.ComponentName,
						},
					},
				},
			}
		} else {
			if note, found := perfum.Notes[record.NoteUuid]; !found {
				perfums[record.PerfumInfoUuid].Notes[record.NoteUuid] = NoteList{
					Name: record.NoteName,
					Components: map[string]string{
						record.ComponentUuid: record.ComponentName,
					},
				}
			} else {
				note.Components[record.ComponentUuid] = record.ComponentName
			}
		}
		if perfumInfo, found := perfumInfoMap[record.PerfumInfoUuid]; found {
			info := perfums[record.PerfumInfoUuid]
			info.PerfumInfo = *perfumInfo
			perfums[record.PerfumInfoUuid] = info
		}
	}

	obj.Amount = int64(len(perfums))
	obj.Total = params.Total
	if params.Base.Offset.Valid {
		obj.Offset = params.Base.Offset.Int64
	} else {
		obj.Offset = 0
	}

	for _, perfum := range perfums {
		pCompos := NewPerfumCompositionV1()
		pCompos.AddPerfumInfoItem(&perfum.PerfumInfo)
		// pCompos.PerfumInfoV1 = perfum.PerfumInfo
		for noteId, note := range perfum.Notes {
			newNote := NewNoteItemV1(noteId, note.Name)
			for compId, compName := range note.Components {
				newComp := NewComponentItemV1(compId, compName)
				newNote.AddComponentItem(newComp)
			}
			sort.Sort(ByComponentName(newNote.Components))
			newNote.ComponentCount = int64(len(newNote.Components))
			pCompos.TotalComponents += newNote.ComponentCount
			pCompos.AddNoteItem(newNote)
		}
		sort.Sort(ByNoteName(pCompos.Notes))
		obj.ObjList = append(obj.ObjList, *pCompos)
	}

	return obj, nil
}

func (obj *PerfumsCompositionV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	return obj, nil
}

func (obj *PerfumsCompositionV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfums"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *PerfumsCompositionV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfums"
	dbQuery.ConditionTableField = "parfum_info_id"
	dbQuery.ConditionTableName = "parfum_info"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "parfum_info.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *PerfumsCompositionV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

// Brand ...
type BrandV1 struct {
	Id           string         `db:"id" json:"-"`
	Uuid         string         `db:"brand_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

// Brands ...
type BrandsV1 struct {
	ObjList []BrandV1 `db:"-" json:"brands_list"`
	Total   int64     `db:"-" json:"total"`
	Offset  int64     `db:"-" json:"offset"`
	Amount  int64     `db:"-" json:"amount"`
}

func NewBrandsFactory(version string) Objecter {
	switch version {
	case "v1":
		return &BrandsV1{ObjList: make([]BrandV1, 0)}
	}

	return nil
}

func (obj *BrandsV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "brands.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_brands", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("brands", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/brand/" + obj.ObjList[i].Uuid,
				Rel:    "BrandInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/brand/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "BrandPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *BrandsV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.ConditionTableField = "brand_id"
	params.DbQuery.ConditionTableName = "brands"
	params.DbQuery.ConditionUuid = addIdsToQuery(uids, "brands.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.WhereConditionString = query.String()
	query.Reset()

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *BrandsV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "brands"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *BrandsV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.ConditionTableField = "brand_id"
	dbQuery.ConditionTableName = "brands"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "brands.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *BrandsV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

// ComponentDB ...
type ComponentV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"component_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

// Components ...
type ComponentsV1 struct {
	ObjList []ComponentV1 `db:"-" json:"components"`
	Total   int64         `db:"-" json:"total"`
	Offset  int64         `db:"-" json:"offset"`
	Amount  int64         `db:"-" json:"amount"`
}

func NewComponentsFactory(version string) Objecter {
	switch version {
	case "v1":
		return &ComponentsV1{ObjList: make([]ComponentV1, 0)}
	}

	return nil
}

func (obj *ComponentsV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "components.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_components", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("components", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/component/" + obj.ObjList[i].Uuid,
				Rel:    "ComponentInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/component/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "ComponentPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *ComponentsV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.WhereConditionString = addIdsToQuery(uids, "components.uuid")
	query := bytes.NewBufferString("")
	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}
	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}
	if err := tmpl.ExecuteTemplate(query, "condition_innerjoin_component_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.AuxConditionString = query.String()
	params.DbQuery.WhereConditionString = ""
	params.DbQuery.AndConditionString = ""

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *ComponentsV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "components"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *ComponentsV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfums"
	dbQuery.ConditionTableField = "component_id"
	dbQuery.ConditionTableName = "components"
	dbQuery.DistinctTableField = "parfum_info_id"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "components.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *ComponentsV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type CountryV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"country_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

// Countries ...
type CountriesV1 struct {
	ObjList []CountryV1 `db:"-" json:"countries_list"`
	Total   int64       `db:"-" json:"total"`
	Offset  int64       `db:"-" json:"offset"`
	Amount  int64       `db:"-" json:"amount"`
}

func NewCountriesFactory(version string) Objecter {
	switch version {
	case "v1":
		return &CountriesV1{ObjList: make([]CountryV1, 0)}
	}

	return nil
}

func (obj *CountriesV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "countries.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_countries", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("countries", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/country/" + obj.ObjList[i].Uuid,
				Rel:    "CountryInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/country/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "CountryPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *CountriesV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.ConditionTableField = "country_id"
	params.DbQuery.ConditionTableName = "countries"
	params.DbQuery.ConditionUuid = addIdsToQuery(uids, "countries.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.WhereConditionString = query.String()
	query.Reset()

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *CountriesV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "countries"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *CountriesV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.ConditionTableField = "country_id"
	dbQuery.ConditionTableName = "countries"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "countries.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *CountriesV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type GenderV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"gender_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

// GendersV1 ...
type GendersV1 struct {
	ObjList []GenderV1 `db:"-" json:"gender_list"`
	Total   int64      `db:"-" json:"total"`
	Offset  int64      `db:"-" json:"offset"`
	Amount  int64      `db:"-" json:"amount"`
}

func NewGendersFactory(version string) Objecter {
	switch version {
	case "v1":
		return &GendersV1{ObjList: make([]GenderV1, 0)}
	}

	return nil
}

func (obj *GendersV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "gender.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_gender", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("genders", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/gender/" + obj.ObjList[i].Uuid,
				Rel:    "GenderInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/gender/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "GenderPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *GendersV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.ConditionTableField = "gender_id"
	params.DbQuery.ConditionTableName = "gender"
	params.DbQuery.ConditionUuid = addIdsToQuery(uids, "gender.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.WhereConditionString = query.String()
	query.Reset()

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *GendersV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "gender"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *GendersV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.ConditionTableField = "gender_id"
	dbQuery.ConditionTableName = "gender"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "gender.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *GendersV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type GroupV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"group_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

// GroupsV1 ...
type GroupsV1 struct {
	ObjList []GroupV1 `db:"-" json:"groups_list"`
	Total   int64     `db:"-" json:"total"`
	Offset  int64     `db:"-" json:"offset"`
	Amount  int64     `db:"-" json:"amount"`
}

func NewGroupsFactory(version string) Objecter {
	switch version {
	case "v1":
		return &GroupsV1{ObjList: make([]GroupV1, 0)}
	}

	return nil
}

func (obj *GroupsV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "groups.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_groups", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("groups", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/group/" + obj.ObjList[i].Uuid,
				Rel:    "GroupInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/group/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "GroupPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *GroupsV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.ConditionTableField = "group_id"
	params.DbQuery.ConditionTableName = "groups"
	params.DbQuery.ConditionUuid = addIdsToQuery(uids, "groups.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.WhereConditionString = query.String()
	query.Reset()

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *GroupsV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "groups"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *GroupsV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.ConditionTableField = "group_id"
	dbQuery.ConditionTableName = "groups"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "groups.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *GroupsV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type NoteV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"note_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

type NotesV1 struct {
	ObjList []NoteV1 `db:"-" json:"notes_list"`
	Total   int64    `db:"-" json:"total"`
	Offset  int64    `db:"-" json:"offset"`
	Amount  int64    `db:"-" json:"amount"`
}

func NewNotesFactory(version string) Objecter {
	switch version {
	case "v1":
		return &NotesV1{ObjList: make([]NoteV1, 0)}
	}

	return nil
}

func (obj *NotesV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "notes.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_notes", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("notes", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/note/" + obj.ObjList[i].Uuid,
				Rel:    "NoteInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/note/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "NotePerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *NotesV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.WhereConditionString = addIdsToQuery(uids, "notes.uuid")
	query := bytes.NewBufferString("")
	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "notes.uuid")
	}
	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}
	if err := tmpl.ExecuteTemplate(query, "condition_innerjoin_note_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.AuxConditionString = query.String()
	params.DbQuery.WhereConditionString = ""
	params.DbQuery.AndConditionString = ""
	query.Reset()

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *NotesV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "notes"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *NotesV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfums"
	dbQuery.ConditionTableField = "note_id"
	dbQuery.ConditionTableName = "notes"
	dbQuery.DistinctTableField = "parfum_info_id"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "notes.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *NotesV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type SeasonV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"season_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

type SeasonsV1 struct {
	ObjList []SeasonV1 `db:"-" json:"seasons_list"`
	Total   int64      `db:"-" json:"total"`
	Offset  int64      `db:"-" json:"offset"`
	Amount  int64      `db:"-" json:"amount"`
}

func NewSeasonsFactory(version string) Objecter {
	switch version {
	case "v1":
		return &SeasonsV1{ObjList: make([]SeasonV1, 0)}
	}

	return nil
}

func (obj *SeasonsV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "seasons.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_seasons", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("seasons", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/season/" + obj.ObjList[i].Uuid,
				Rel:    "SeasonInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/season/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "SeasonPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *SeasonsV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.ConditionTableField = "season_id"
	params.DbQuery.ConditionTableName = "seasons"
	params.DbQuery.ConditionUuid = addIdsToQuery(uids, "seasons.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.WhereConditionString = query.String()
	query.Reset()

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *SeasonsV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "seasons"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *SeasonsV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.ConditionTableField = "season_id"
	dbQuery.ConditionTableName = "seasons"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "seasons.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *SeasonsV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type TimeOfDayV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"tsod_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

type TimesOfDayV1 struct {
	ObjList []TimeOfDayV1 `db:"-" json:"timeofday_list"`
	Total   int64         `db:"-" json:"total"`
	Offset  int64         `db:"-" json:"offset"`
	Amount  int64         `db:"-" json:"amount"`
}

func NewTimesOfDayFactory(version string) Objecter {
	switch version {
	case "v1":
		return &TimesOfDayV1{ObjList: make([]TimeOfDayV1, 0)}
	}

	return nil
}

func (obj *TimesOfDayV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "times_of_day.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_timeofday", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("timesOfDay", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/timeofday/" + obj.ObjList[i].Uuid,
				Rel:    "TimeofdayInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/timeofday/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "TimeofdayPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *TimesOfDayV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.ConditionTableField = "tsod_id"
	params.DbQuery.ConditionTableName = "times_of_day"
	params.DbQuery.ConditionUuid = addIdsToQuery(uids, "times_of_day.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.WhereConditionString = query.String()
	query.Reset()

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *TimesOfDayV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "times_of_day"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *TimesOfDayV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.ConditionTableField = "tsod_id"
	dbQuery.ConditionTableName = "times_of_day"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "times_of_day.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *TimesOfDayV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type TypeV1 struct {
	Id           int64          `db:"id" json:"-"`
	Uuid         string         `db:"type_uuid" json:"id"`
	Name         string         `db:"name" json:"name"`
	ImageId      sql.NullString `db:"img_uuid" json:"-"`
	PerfumsCount int64          `db:"-" json:"perfums_count"`
	Links        []LinkV1       `db:"-" json:"links"`
	SmallImgUrl  string         `db:"-" json:"small_img_url"`
	LargeImgUrl  string         `db:"-" json:"large_img_url"`
}

type TypesV1 struct {
	ObjList []TypeV1 `db:"-" json:"types_list"`
	Total   int64    `db:"-" json:"total"`
	Offset  int64    `db:"-" json:"offset"`
	Amount  int64    `db:"-" json:"amount"`
}

func NewTypesFactory(version string) Objecter {
	switch version {
	case "v1":
		return &TypesV1{ObjList: make([]TypeV1, 0)}
	}

	return nil
}

func (obj *TypesV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*MakeObjParams)

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.WhereConditionString = addIdsToQuery(params.Base.Ids.String, "types.uuid")
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_types", &params.DbQuery); err != nil {
		return nil, err
	}

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		// obj.ObjList[i].PerfumsCount = params.PerfumsNum
		perfumsCount, _ := GetPerfumsCount("types", obj.ObjList[i].Uuid)
		obj.ObjList[i].PerfumsCount = perfumsCount
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/type/" + obj.ObjList[i].Uuid,
				Rel:    "TypeInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/type/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "TypePerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *TypesV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	if params == nil || len(uids) == 0 {
		return nil, errors.New("invalid args")
	}

	params.DbQuery.ConditionTableField = "type_id"
	params.DbQuery.ConditionTableName = "types"
	params.DbQuery.ConditionUuid = addIdsToQuery(uids, "types.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &params.DbQuery); err != nil {
		return nil, err
	}
	params.DbQuery.WhereConditionString = query.String()
	query.Reset()

	if err := setDbQueryBaseParams(&params.Base, &params.DbQuery); err != nil {
		return nil, err
	}

	if params.Base.Ids.Valid {
		params.DbQuery.AndConditionString = addIdsToQuery(params.Base.Ids.String, "parfum_info.uuid")
	}

	if err := tmpl.ExecuteTemplate(query, "perfum_info_base", params.DbQuery); err != nil {
		return nil, err
	}

	pinfos := NewPerfumsInfoFactory(params.Base.Version)
	return pinfos.MakeObj(params)
}

func (obj *TypesV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "types"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *TypesV1) ExtraCount(uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, errors.New("invalid args")
	}

	dbQuery := QueryTemplateParams{}
	dbQuery.FromTableName = "parfum_info"
	dbQuery.ConditionTableField = "type_id"
	dbQuery.ConditionTableName = "types"
	dbQuery.ConditionUuid = addIdsToQuery(uids, "types.uuid")
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "condition_select_id_eq_uuid", &dbQuery); err != nil {
		return 0, err
	}
	dbQuery.WhereConditionString = query.String()
	query.Reset()
	if err := tmpl.ExecuteTemplate(query, "select_count", &dbQuery); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *TypesV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

type PerfumsSearchResultV1 struct {
	Links  []LinkV1 `json:"links"`
	Total  int64    `json:"total"`
	Offset int64    `json:"offset"`
	Amount int64    `json:"amount"`
}

func NewPerfumsSearchResultFactory(version string) Objecter {
	switch version {
	case "v1":
		return &PerfumsSearchResultV1{Links: make([]LinkV1, 0)}
	}

	return nil
}

func (obj *PerfumsSearchResultV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return nil, err
	}
	search.Order = "perfum_info.info_uuid"
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "perfum_search", &search); err != nil {
		return nil, err
	}

	var results []string
	if _, err = dbmap.Select(&results, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(results))

	for _, result := range results {
		obj.Links = append(obj.Links,
			LinkV1{
				Href:   baseUrl + "/perfum/" + result,
				Rel:    "PerfumInfo",
				Method: "GET",
			},
		)
	}

	return obj, nil
}

func (obj *PerfumsSearchResultV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	return obj, nil
}

func (obj *PerfumsSearchResultV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return 0, err
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "perfum_search_count", &search); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *PerfumsSearchResultV1) ExtraCount(uids []string) (int64, error) {
	return 0, nil
}

func (obj *PerfumsSearchResultV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

// UserReq
type UserReq struct {
	UserId string `json:"user_id"`
}

// LoginReq ...
type LoginReq struct {
	AuthcodeString string `json:"auth_code" binding:"required"`
}

// LoginResp represents an authenticated response.
type LoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"user_id"`
}

// UserResp ...
type UserResp struct {
	UserId    string   `json:"user_id"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	Links     []LinkV1 `json:"links"`
}

// IdTokenClaims ...
type IdTokenClaims struct {
	Iss    string  `json:"iss"`
	Sub    string  `json:"sub"`
	Aud    string  `json:"aud"`
	Iat    float64 `json:"iat"`
	Exp    float64 `json:"exp"`
	UserId string  `json:"user_id"`
}

//BrandsSearchResultV1
type BrandsSearchResultV1 struct {
	ObjList []BrandV1 `db:"-" json:"brands_list"`
	Total   int64     `json:"total"`
	Offset  int64     `json:"offset"`
	Amount  int64     `json:"amount"`
}

func NewBrandsSearchResultFactory(version string) Objecter {
	switch version {
	case "v1":
		return &BrandsSearchResultV1{ObjList: make([]BrandV1, 0)}
	}

	return nil
}

func (obj *BrandsSearchResultV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return nil, err
	}
	if search.BrandUid == "" && search.Brand == "" {
		// return empty object
		return obj, nil
	}
	search.Order = "brands." + search.BrandsName
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "brands_search", &search); err != nil {
		return nil, err
	}

	fmt.Println("================================= Query begin =======================================")
	fmt.Println(query.String())
	fmt.Println("================================= Query end =========================================")

	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		obj.ObjList[i].PerfumsCount, _ = GetPerfumsCount("brands", obj.ObjList[i].Uuid)
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/brand/" + obj.ObjList[i].Uuid,
				Rel:    "BrandInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/brand/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "BrandPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *BrandsSearchResultV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	return obj, nil
}

func (obj *BrandsSearchResultV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return 0, err
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "brands_search_count", &search); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *BrandsSearchResultV1) ExtraCount(uids []string) (int64, error) {
	return 0, nil
}

func (obj *BrandsSearchResultV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

//ComponentsSearchResultV1
type ComponentsSearchResultV1 struct {
	ObjList []ComponentV1 `db:"-" json:"components"`
	Total   int64         `db:"-" json:"total"`
	Offset  int64         `db:"-" json:"offset"`
	Amount  int64         `db:"-" json:"amount"`
}

func NewComponentsSearchResultFactory(version string) Objecter {
	switch version {
	case "v1":
		return &ComponentsSearchResultV1{ObjList: make([]ComponentV1, 0)}
	}

	return nil
}

func (obj *ComponentsSearchResultV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return nil, err
	}
	if search.ComponentUid == "" && search.Component == "" {
		// return empty object
		return obj, nil
	}
	search.Order = "components." + search.ComponentsName
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "components_search", &search); err != nil {
		return nil, err
	}
	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		obj.ObjList[i].PerfumsCount, _ = GetPerfumsCount("components", obj.ObjList[i].Uuid)
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/component/" + obj.ObjList[i].Uuid,
				Rel:    "ComponentInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/component/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "ComponentPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *ComponentsSearchResultV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	return obj, nil
}

func (obj *ComponentsSearchResultV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return 0, err
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "components_search_count", &search); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *ComponentsSearchResultV1) ExtraCount(uids []string) (int64, error) {
	return 0, nil
}

func (obj *ComponentsSearchResultV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

// CountriesSearchResultV1
type CountriesSearchResultV1 struct {
	ObjList []CountryV1 `db:"-" json:"countries_list"`
	Total   int64       `db:"-" json:"total"`
	Offset  int64       `db:"-" json:"offset"`
	Amount  int64       `db:"-" json:"amount"`
}

func NewCountriesSearchResultFactory(version string) Objecter {
	switch version {
	case "v1":
		return &CountriesSearchResultV1{ObjList: make([]CountryV1, 0)}
	}

	return nil
}

func (obj *CountriesSearchResultV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return nil, err
	}
	if search.CountryUid == "" && search.Country == "" {
		// return empty object
		return obj, nil
	}
	search.Order = "countries." + search.CountriesName
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "countries_search", &search); err != nil {
		return nil, err
	}
	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		obj.ObjList[i].PerfumsCount, _ = GetPerfumsCount("countries", obj.ObjList[i].Uuid)
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/country/" + obj.ObjList[i].Uuid,
				Rel:    "CountryInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/country/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "CountryPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *CountriesSearchResultV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	return obj, nil
}

func (obj *CountriesSearchResultV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return 0, err
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "countries_search_count", &search); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *CountriesSearchResultV1) ExtraCount(uids []string) (int64, error) {
	return 0, nil
}

func (obj *CountriesSearchResultV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}

// GroupsSearchResultV1
type GroupsSearchResultV1 struct {
	ObjList []GroupV1 `db:"-" json:"groups_list"`
	Total   int64     `db:"-" json:"total"`
	Offset  int64     `db:"-" json:"offset"`
	Amount  int64     `db:"-" json:"amount"`
}

func NewGroupsSearchResultFactory(version string) Objecter {
	switch version {
	case "v1":
		return &GroupsSearchResultV1{ObjList: make([]GroupV1, 0)}
	}

	return nil
}

func (obj *GroupsSearchResultV1) MakeObj(pParams interface{}) (Objecter, error) {
	if pParams == nil {
		return nil, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return nil, err
	}
	if search.GroupUid == "" && search.Group == "" {
		// return empty object
		return obj, nil
	}
	search.Order = "groups." + search.GroupsName
	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "groups_search", &search); err != nil {
		return nil, err
	}
	if _, err = dbmap.Select(&obj.ObjList, query.String()); err != nil {
		return nil, err
	}

	obj.Total = params.Total
	obj.Offset = params.Base.Offset.Int64
	obj.Amount = int64(len(obj.ObjList))

	for i := 0; i < len(obj.ObjList); i++ {
		obj.ObjList[i].PerfumsCount, _ = GetPerfumsCount("groups", obj.ObjList[i].Uuid)
		obj.ObjList[i].Links = []LinkV1{
			LinkV1{
				Href:   baseUrl + "/group/" + obj.ObjList[i].Uuid,
				Rel:    "GroupInfo",
				Method: "GET",
			},
			LinkV1{
				Href:   baseUrl + "/group/" + obj.ObjList[i].Uuid + "/perfums",
				Rel:    "GroupPerfums",
				Method: "GET",
			},
		}

		if obj.ObjList[i].ImageId.Valid {
			obj.ObjList[i].SmallImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/small"
			obj.ObjList[i].LargeImgUrl = baseUrl + "/image/" + obj.ObjList[i].ImageId.String + "/large"
		}
	}

	return obj, nil
}

func (obj *GroupsSearchResultV1) MakeExtraObj(params *MakeObjParams, uids []string) (Objecter, error) {
	return obj, nil
}

func (obj *GroupsSearchResultV1) Count(pParams interface{}) (int64, error) {
	if pParams == nil {
		return 0, errors.New("invalid args")
	}

	params := pParams.(*SearchParams)

	search := NewSearchQueryTemplateParams()
	if err := search.ParseSearchParams(params); err != nil {
		return 0, err
	}

	query := bytes.NewBufferString("")
	if err := tmpl.ExecuteTemplate(query, "groups_search_count", &search); err != nil {
		return 0, err
	}

	count, err := dbmap.SelectInt(query.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (obj *GroupsSearchResultV1) ExtraCount(uids []string) (int64, error) {
	return 0, nil
}

func (obj *GroupsSearchResultV1) Json(w http.ResponseWriter, status int) error {
	render := render.New()
	return render.JSON(w, status, obj)
}
