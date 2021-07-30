package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var assetCount int64 = 0
type GenericAsset struct {
	verified bool
	validated bool
	can_counts bool
	useful_life int
	operational_state map[string]string
	subscription_date string
	created_at string
	updated_at string
	associated_surface_tag string
	associated_building_tag string
	center_id int64
	edifice_id int64
	level_id int64
	space_id int64
	typeId int64
	code string
	internal_item_code string
	acquisition_type_id int
	id int64
	temporalId string
	image string

	GeneralAspect *GeneralAspect
	LocalizationAspect *LocalizationAspect
	LegalAspect *LegalAspect
	investment *Firstinvestment

}

func (ptr *GenericAsset) init(excelAsset *ExcelAsset, tempId int)  {

	now := now()
	centerCode:= excelAsset.CenterCode
	if centerCode == ""{
		centerCode = config.DefaultCenterCode
	}
	centerId := cache.findByCode("centers", centerCode)


	acquisitionCode := excelAsset.AcquisitionTypeCode
	if acquisitionCode == ""{
		acquisitionCode = "default"
	}
	ptr.operational_state = make(map[string]string)
	acquisitionId := cache.findByCode("acquisition_types", acquisitionCode)
	typeCode := fmt.Sprintf("%v%v",excelAsset.SubfamilyCode,excelAsset.TypeCode)
	useful_life := cache.findByCode("useful_life_by_type", typeCode)
	typeId := cache.findByCode("types", typeCode)
	ptr.image = excelAsset.ImageUrl
	ptr.temporalId = fmt.Sprintf("%v",excelAsset.id)
	ptr.useful_life = useful_life.(int)
	ptr.acquisition_type_id = acquisitionId.(int)
	ptr.subscription_date = now
	ptr.center_id = realint64(centerId)
	ptr.validated = true
	ptr.internal_item_code = excelAsset.ICode
	ptr.operational_state["status"] = "active"
	ptr.operational_state["since"] = now
	ptr.created_at = now
	ptr.updated_at = now
	ptr.verified = false
	ptr.typeId = int64(typeId.(int))
	ptr.associated_building_tag = excelAsset.AssociatedBuilding
	ptr.associated_surface_tag = excelAsset.AssociatedTerrain
	ptr.makeAssetCode(lastAssetCodeInDatabase, excelAsset)


}

func (ptr *GenericAsset) after(excelAsset *ExcelAsset) [][]interface{} {
	//falta   tag
	excelAsset.id = ptr.id
	queriesForQueue := make([][]interface{},4)
	now := now()
	ptr.markTagInUse(excelAsset, now)

	ptr.GeneralAspect.init(excelAsset)
	query := getQueryForBatchInsert("general_aspects")
	gAspect := ptr.GeneralAspect
	characterisiticsJson, _ := json.Marshal(gAspect.Characteristics)
	genQ := []interface{}{query,gAspect.Description,gAspect.Denomination, gAspect.Address, gAspect.ZipCode, gAspect.SurfaceFiels, gAspect.SurfaceBuilt, gAspect.Length, characterisiticsJson,gAspect.Width,gAspect.Height,gAspect.Depth, gAspect.SerialNumber, gAspect.Matricula, gAspect.Bastidor, gAspect.OldTagCode, gAspect.Validated, gAspect.RiskLevel, gAspect.Catastral, ptr.id, gAspect.Asset_Model_Id,  now, now, gAspect.Color_Id, gAspect.Material_Id}
	queriesForQueue[0] = genQ



	ptr.LocalizationAspect.init(excelAsset)
	queryL := getQueryForBatchInsert("localization_aspects")
	locAspect := ptr.LocalizationAspect
	locQ := []interface{}{queryL,true, true, ptr.id,locAspect.Space, now, now, locAspect.Area, nil, locAspect.Str_value}
	queriesForQueue[1] = locQ
	var lender interface{}
	var owner interface{}

	ptr.LegalAspect.init(excelAsset)

	queryLeg := getQueryForBatchInsert("legal_aspects")
	legAspect := ptr.LegalAspect

	if legAspect.Proprietary != -1 {
		owner = legAspect.Proprietary
	}
	if legAspect.Assignee != -1 {
		lender = legAspect.Assignee
	}
	legQ := []interface{}{queryLeg,ptr.id, owner, lender, now, now}
	queriesForQueue[2] = legQ

	ptr.investment.init(excelAsset)
	queryInv := getQueryForBatchInsert("first_investment")
	investmnt := ptr.investment
	var stDate, endDate *string
	if investmnt.start_assigment_date != "" {
		*stDate = investmnt.start_assigment_date
	}
	if investmnt.end_assigment_date != "" {
		*endDate = investmnt.end_assigment_date
	}
	investQ := []interface{}{queryInv,stDate,endDate, true, true, ptr.id,investmnt.acquisition_type_id, investmnt.purchase_order_id, now, now, nil}
	queriesForQueue[3] = investQ

	return queriesForQueue
}

func (ptr *GenericAsset) makeAssetCode(lastCode string,  info *ExcelAsset ){
	repeat := 3
	var _lastPrefix string
	var _lastDigits int
	_lastPrefix = "AA"
	if lastCode == ""{
		_lastDigits = 0
	}else {
		_lastDigits = len(lastCode[2:])
	}
	_lastDigits += int(info.id)
	s:= fmt.Sprint(_lastDigits)
	digitLength := len(s)
	repeat = int(math.Abs(float64(4-digitLength)))
	pad := strings.Repeat("0",repeat)
	nudigits := fmt.Sprint(pad,s)
	code := fmt.Sprint("YY/","%v/",_lastPrefix,nudigits)
	code = fmt.Sprintf(code, fmt.Sprint(info.SubfamilyCode, info.TypeCode))
	ptr.code = code
}

func (ptr *GenericAsset) markTagInUse(info *ExcelAsset, now string ){
	tagType := info.TagType
	defaultCode := cache.findById("tags",1)
	if tagType[0] == 'F' && info.TagCode != ""{

		tagId := cache.findByCode("tags",info.TagCode)
		if tagId == 1 && info.TagCode != defaultCode{
			debugger.error("wrong tag found", info.TagCode)
			return
			//os.Exit(1)
		}
		data := make([]interface{},0)
		data = append(data,info.id, ptr.center_id, tagId)
		cache.enqueueForInsert("update_tag", info.TagCode, &data)
		ptr.makeTagStatus(info, now)

		tagAsset, _ := cache.nomenclators.LoadOrStore("asset_tag_ref", make([]codeable,0))
		var codeables codeable
		codeables.id = int(ptr.id)
		codeables.code = info.TagCode
		tagAsset = append(tagAsset.([]codeable),codeables)
		cache.nomenclators.Store("asset_tag_ref", tagAsset)
	}else if tagType[0] == 'V' {
		ptr.fuckingVirtualTags()
	}
}

func (ptr *GenericAsset) makeTagStatus(info *ExcelAsset, now string ){
	tagCode := info.TagCode
	tagType := info.TagType
	if tagCode == "" &&  strings.Index(tagType,"V") != 0{
		return
	}
	data := make([]interface{},0)
	tagId := cache.findByCode("tags",tagCode)
	defaultTagCode :=  cache.findById("tags",1)
	if tagId == 1 && defaultTagCode != tagCode{
		return
	}else{
		data = append(data,"inuse","App\\Models\\Tag\\Tag", tagId, now, now)
	}
	if len(data) == 5 {
		cache.enqueueForInsert("pending_tags_status", tagCode, &data)
	}
}

func(ptr *GenericAsset) fuckingVirtualTags()  {
	if lastVirtualTagCodeInDatabase == ""{
		lastVirtualTagCodeInDatabase = "0"
	}
	assetCount +=1
	numericCode,_ := strconv.ParseInt(lastVirtualTagCodeInDatabase,10,64)
	numericCode += assetCount
	textCode := fmt.Sprint(numericCode)
	if l :=len(textCode); l < 8 {
		pad := 8 - l
		textCode = fmt.Sprint(strings.Repeat("0", pad),textCode)
	}
	data := make([]interface{},0);
	now := now()
	textCode = fmt.Sprint("V", textCode)
	data = append(data, textCode, textCode,2,now, now, ptr.id, ptr.center_id)
	cache.enqueueForInsert("virtual_tags", textCode, &data)

	tagAsset, _ := cache.nomenclators.LoadOrStore("asset_tag_ref", make([]codeable,0))
	var codeables codeable
	codeables.id = int(ptr.id)
	codeables.code = textCode
	tagAsset = append(tagAsset.([]codeable),codeables)
	cache.nomenclators.Store("asset_tag_ref", tagAsset)



}
