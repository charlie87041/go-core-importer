package main

import (
	"fmt"
	"strings"
)

type LocalizationAspect struct {
	Validated bool
	Risk int8
	Active bool
	GenericAssetId int64
	id int64
	Center int64
	Edifice int64
	Level int64
	PatrimonialOperation int64
	Space int64
	CostCenter int
	Area int64
	Title string
	Str_value string
	Created_At string
	Updated_At string
}

func (ptr *LocalizationAspect) init(excelAsset *ExcelAsset)  {
	ptr.Validated = true
	ptr.Active = true
	now := now()

	levelId := cache.findByCode("levels", excelAsset.LevelCode)
	EdificeId := cache.findByCode("edifices", excelAsset.BuildingCode)
	AreaId := cache.findByCode("areas", excelAsset.AreaCode)
	centerId := cache.findByCode("centers", excelAsset.CenterCode)
	costCenterId := cache.findByCode("cost_centers", excelAsset.CostCenterCode)

	ptr.Center = realint64(centerId)
	ptr.CostCenter = costCenterId.(int)
	ptr.Edifice = realint64(EdificeId)
	ptr.Level = realint64(levelId)
	ptr.Area = realint64(AreaId)
	spaceId := cache.findByCode("spaces", excelAsset.SpaceCode)
	strVal := strings.Replace(excelAsset.SpaceCode,"default","", -1)
	ptr.Space = realint64(spaceId)
	ptr.Str_value = strVal
	ptr.Created_At = now
	ptr.Updated_At = now
	ptr.GenericAssetId = excelAsset.id
	ptr.makeTitle(excelAsset)
}

func (ptr *LocalizationAspect) after(asset *GenericAsset) []string {
	return make([]string,0)
}

func (ptr *LocalizationAspect) makeTitle(info *ExcelAsset)  {
	t := (cache.fetch("centers", info.CenterCode)).(codeable)
	title := t.code.([]interface{})[1]
	if !strings.Contains(info.BuildingCode, "default") {
		edifice := (cache.fetch("edifices", info.BuildingCode)).(codeable)
		edificeTitle := edifice.code.([]interface{})[1]
		title = fmt.Sprint(title,"/",edificeTitle)
	}
	if !strings.Contains(info.LevelCode, "default") {
		level := (cache.fetch("levels", info.LevelCode)).(codeable)
		levelTitle := level.code.([]interface{})[1]
		title = fmt.Sprint(title,"/",levelTitle)
	}
	if !strings.Contains(info.SpaceCode, "default") {
		space := (cache.fetch("spaces", info.SpaceCode)).(codeable)
		spaceTitle := space.code.([]interface{})[1]
		title = fmt.Sprint(title,"/",spaceTitle)
	}
	area := (cache.fetch("areas", info.AreaCode)).(codeable)
	areaTitle := area.code.([]interface{})[1]
	title = fmt.Sprint(title, fmt.Sprintf("(%v)",areaTitle))
	ptr.Title = title.(string)
}
