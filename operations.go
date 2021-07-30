package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)


var lastAssetCodeInDatabase string
var lastVirtualTagCodeInDatabase string
var subscriptionOperationType int



func importAsset(assets []*ExcelAsset) ([][][]interface{}, []*GenericAsset, []*ExcelAsset) {
	container := make([][]interface{}, 0)
	tmpAssetCache := make(map[string][]interface{})
	for i, asset := range assets {

		assetData := &GenericAsset{}
		assetData.init(asset, i)
		data := make([]interface{}, 0)
		st := make([]interface{}, 0)
		st = append(st, asset)
		st = append(st, assetData)
		tmpAssetCache[assetData.temporalId] = st

		jsonOperationalState, error := json.Marshal(assetData.operational_state)
		if error != nil {
			cache.debugger.error(error)
			os.Exit(1)
		}
		data = append(data, assetData.verified, assetData.subscription_date, assetData.code, assetData.internal_item_code, assetData.acquisition_type_id, assetData.center_id, assetData.associated_surface_tag, assetData.associated_building_tag, assetData.validated, assetData.typeId, assetData.created_at, assetData.updated_at, jsonOperationalState, assetData.temporalId)
		container = append(container, data)
	}
	ids := batchInsertAsset(container)
	asset := make([]*GenericAsset, 0)
	excel := make([]*ExcelAsset, 0)
	asset_query := make([][][]interface{}, 0)
	for i := 0; i < len(ids); i += 2 {
		id := ids[i]
		code := ids[i+1]
		//updating the assets references with the actual database id
		assetTuple := tmpAssetCache[code.(string)]
		excelAsset := assetTuple[0].(*ExcelAsset)
		assetData := assetTuple[1].(*GenericAsset)
		assetData.GeneralAspect = &GeneralAspect{}
		assetData.LocalizationAspect = &LocalizationAspect{}
		assetData.LegalAspect = &LegalAspect{}
		assetData.investment = &Firstinvestment{}
		assetData.id = int64(id.(int))
		asset_query = append(asset_query, assetData.after(excelAsset))
		//aspects
		asset = append(asset, assetData)
		excel = append(excel, excelAsset)
	}
	return asset_query, asset, excel
}

func LoadNomenclators(assets *[]*ExcelAsset) {
	waitingForInsert := make(map[string][][]interface{})
	waitingForInsert["asset_models"] = make([][]interface{}, 0)
	waitingForInsert["levels"] = make([][]interface{}, 0)
	waitingForInsert["spaces"] = make([][]interface{}, 0)
	debugger.chrono("starting with nomenclator population")
	for _, asset := range *assets {
		refreshBrands(asset, &waitingForInsert)//
		refreshColors(asset, &waitingForInsert)//
		refreshMaterials(asset, &waitingForInsert)//
		refreshEdifices(asset, &waitingForInsert)
		refreshModels(asset, &waitingForInsert)//
		refreshLevels(asset, &waitingForInsert)
		refreshSpaces(asset, &waitingForInsert)
		refreshAreas(asset, &waitingForInsert)//

	}
	var wg sync.WaitGroup
	wg.Add(1)
	cache.batchInsert(&wg)
	wg.Wait()
	tables := []string{"asset_models", "brands", "edifices", "levels", "spaces", "colors", "materials"}
	cache.loadCodeablesNomenclators(tables)
	processEnqueued(&waitingForInsert)
	tables = []string{"asset_models", "levels", "spaces"}
	cache.loadCodeablesNomenclators(tables)
	refreshOperationVarsForAcquisiton()
}



func refreshColors(asset *ExcelAsset, queue *map[string][][]interface{}) {
	code, name, date := rowAttributes(asset.Color)
	if code == "" {
		(*asset).Color = config.DefaultColor
		return
	} else {
		rawCode := strings.Split(code, ";")
		code = rawCode[0]
	}
	(*asset).Color = code
	if cache.findByCode("colors", code) == -1 {
		data := []interface{}{name, code, fmt.Sprintf("Imported  color %v", code), date, date}
		cache.enqueueForInsert("colors", code, &data)
	}
}

func refreshMaterials(asset *ExcelAsset, queue *map[string][][]interface{}) {
	code, name, date := rowAttributes(asset.Material)
	if code == "" {
		(*asset).Material = config.DefaultMaterial
		return
	} else {
		rawCode := strings.Split(code, ";")
		code = rawCode[0]
	}
	(*asset).Material = code
	if cache.findByCode("materials", code) == -1 {
		data := []interface{}{name, code, fmt.Sprintf("Imported  material %v", code),  date, date}
		cache.enqueueForInsert("materials", code, &data)
	}
}

func refreshBrands(asset *ExcelAsset, queue *map[string][][]interface{}) {
	code  := asset.Brand
	 date := now()
	 name := code
	if code == "" {
		name = config.DefaultBrand
		code = "default"
		(*asset).Brand = name
	}
	if cache.findByCode("brands", name) == -1 {
		data := []interface{}{name, code, fmt.Sprintf("Imported  brand %v", asset.Brand), 1, date, date}
		cache.enqueueForInsert("brands", code, &data)
	}
}
func refreshEdifices(asset *ExcelAsset, queue *map[string][][]interface{}) {
	code := asset.BuildingCode
	date := now()
	data := make([]interface{}, 0)
	centerCode := asset.CenterCode
	if centerCode == "" {
		centerCode = config.DefaultCenterCode
	}
	centerId := cache.findByCode("centers", centerCode)
	ncode := fmt.Sprintf("%v_default", centerCode)

	if code == "" {
		if cache.findByCode("edifices", ncode) == -1 {
			data = append(data, "General", ncode, "default", centerCode, centerId, date, date)
			cache.enqueueForInsert("edifices", code, &data)
		}
	} else if ncode = fmt.Sprint(centerCode, "/", code); cache.findByCode("edifices", ncode) == -1 {
		data = append(data, code, ncode, code, centerCode, centerId, date, date)
		cache.enqueueForInsert("edifices", code, &data)
	}
	(*asset).BuildingCode = ncode
}

func refreshAreas(asset *ExcelAsset, queue *map[string][][]interface{}) {
	code := asset.AreaCode
	if code == "" {
		code = config.DefaultArea
		(*asset).AreaCode = code
	}
	areaId := cache.findByCode("areas", code)
	if areaId == -1{
		date := now()
		data := []interface{}{code, code, "Imported area" , date, date}
		cache.enqueueForInsert("areas", code, &data)
	}
}

func refreshLevels(asset *ExcelAsset, tempqueue *map[string][][]interface{}) {
	code := asset.LevelCode
	date := now()
	data := make([]interface{}, 0)
	buildingCode := asset.BuildingCode
	if buildingCode == ""{
		buildingCode = config.DefaultEdificeCode
	}
	buildingId := cache.findByCode("edifices", buildingCode)
	name := "General"
	scode:= "default"
	description := "default"
	ncode := fmt.Sprintf("%v_default", buildingCode)
	if code != ""{
		name = code
		scode = code
		ncode = fmt.Sprint(buildingCode, "/", code)
		description = fmt.Sprint("Imported level for edifice ", buildingCode)
	}
	 if cache.findByCode("levels", ncode) == -1 {
		if buildingId != -1 {
			data = append(data, name, ncode, scode, description, buildingId, date, date)
			cache.enqueueForInsert("levels", code, &data)
		} else {
			body := (*tempqueue)["levels"]
			for _, level := range body {
				if level[2] == ncode {
					(*asset).LevelCode = ncode
					return
				}
			}
			params := make([]interface{}, 0)
			params = append(params, buildingCode,  name, ncode, scode, description,  nil, date, date)
			body = append(body, params)
			(*tempqueue)["levels"] = body
		}
	}
	(*asset).LevelCode = ncode
}

func refreshSpaces(asset *ExcelAsset, tempqueue *map[string][][]interface{}) {
	code, _, date := rowAttributes(asset.SpaceCode)
	code = strings.ToUpper(code)
	data := make([]interface{}, 0)
	levelCode := asset.LevelCode
	if levelCode == ""{
		levelCode = config.DefaultLevelCode
	}
	levelId := cache.findByCode("levels", levelCode)
	name := "General"
	scode:= "default"
	description := "default"
	ncode := fmt.Sprintf("%v_default%v", levelCode, asset.id)
	//data = append(data, "General", ncode, "default", "default", levelId, true, date, date)
	if code != ""{
		name = code
		scode = code
		ncode = fmt.Sprint(levelCode, "/", code)
		description = fmt.Sprint("Imported space for level ",levelCode)
	}
	if  cache.findByCode("spaces", ncode) == -1 {
		if levelId != -1{
			data = append(data,name, ncode, scode, description, levelId,  true, date, date)
			cache.enqueueForInsert("spaces", code, &data)
		}else {
			body := (*tempqueue)["spaces"]
			for _, space := range body {
				if space[2] == ncode {
					(*asset).SpaceCode = ncode
					return
				}
			}
			params := make([]interface{}, 0)
			params = append(params, levelCode, name, ncode, scode, description, nil, true, date, date)
			body = append(body, params)
			(*tempqueue)["spaces"] = body
		}
	}
	(*asset).SpaceCode = ncode
}

func refreshOperationVarsForAcquisiton() {
	conn, error := cache.Conn.Acquire(context.Background())
	if error != nil {
		debugger.error(error.Error())
		os.Exit(2)
	}
	defer conn.Release()
	rows, err := conn.Query(context.Background(), "select id, code from operation_vars  where prefix like 'A' ")
	if err != nil {
		cache.debugger.error(err.Error())
		return
	}
	dest := make([]codeable, 0)
	var code codeable
	for rows.Next() {
		rows.Scan(&code.id, &code.code)
		dest = append(dest, code)
	}
	cache.nomenclators.Store("operation_vars", &dest)
}

func refreshModels(asset *ExcelAsset, tempqueue *map[string][][]interface{}) {
	code:= asset.Model
	name:= code
	date:= now()
	brand := asset.Brand
	if brand == ""{
		brand = config.DefaultBrand
	}
	brandId := (cache.findByCode("brands", strings.ToLower(brand))).(int)
	if code == ""{
		code = fmt.Sprint(brand, "_default")
		name = "Sin definir"
		(*asset).Model = name
	}
	if cache.findByCode("asset_models", name) == -1 {
		if brandId != -1 {
			data := []interface{}{name, code, fmt.Sprintf("%v  %v model", brand, code), brandId, date, date}
			cache.enqueueForInsert("asset_models", code, &data)
		}else {
			params := []interface{}{ brand, name, code, fmt.Sprintf("Unknown  %v model", brand), nil, date, date}
			body := (*tempqueue)["asset_models"]
			body = append(body, params)
			(*tempqueue)["asset_models"] = body
		}
	}

}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func rowAttributes(str string) (string, string, string) {
	str = strings.Split(str, "%")[0]
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	t := time.Now()
	formatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return strings.ToLower(snake), str, formatted
}

func processEnqueued(tempqueue *map[string][][]interface{}) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pendingModels(tempqueue)
	}()
	go func() {
		defer wg.Done()
		pendingLevels(tempqueue)
	}()
	wg.Wait()
	wg.Add(1)
	cache.batchInsert(&wg)
	cache.loadCodeablesNomenclators([]string{"levels"})
	pendingSpaces(tempqueue)
	wg.Add(1)
	cache.batchInsert(&wg)
	wg.Wait()
}

func pendingModels(tempqueue *map[string][][]interface{}) {
	pendingModels := (*tempqueue)["asset_models"]
	for _, model := range pendingModels {
		ref := model[0]
		data := model[1:]
		brandId := cache.findByCode("brands", ref.(string))
		data[3] = brandId
		cache.enqueueForInsert("asset_models", data[1].(string), &data)
	}
}
func pendingLevels(tempqueue *map[string][][]interface{}) {
	pendingLevels := (*tempqueue)["levels"]
	for _, level := range pendingLevels {
		ref := level[0]
		data := level[1:]
		edificeId := cache.findByCode("edifices", ref.(string))
		data[4] = edificeId
		cache.enqueueForInsert("levels", data[1].(string), &data)
	}
}
func pendingSpaces(tempqueue *map[string][][]interface{}) {
	pendingSpaces := (*tempqueue)["spaces"]
	for _, space := range pendingSpaces {
		ref := space[0]
		data := space[1:]
		levelId := cache.findByCode("levels", ref.(string))
		data[4] = levelId
		cache.enqueueForInsert("spaces", data[1].(string), &data)
	}
}

func batchInsertAsset(data [][]interface{}) []interface{} {
	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	var i int = 1
	b := func() string {
		s := ""
		for c := 1; c < 15; c++ {
			if c == 1 {
				s += fmt.Sprintf("$%v", i)
			} else {
				s += fmt.Sprintf(",$%v", i)
			}
			i += 1
		}
		return s
	}
	for _, r := range data {
		params := r
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", b()))
		valueArgs = append(valueArgs, params...)
	}

	stmt := fmt.Sprintf("INSERT INTO public.generic_assets( verified, subscription_date, code, internal_item_code,  acquisition_type_id, center_id, associated_surface_tag, associated_building_tag, validated,type_id,created_at, updated_at,  operational_state, temporal_id) VALUES %s returning id,temporal_id", strings.Join(valueStrings, ","))

	conn, error := cache.Conn.Acquire(context.Background())
	if error != nil {
		debugger.error(error.Error())
		os.Exit(2)
	}
	defer conn.Release()
	rows, error := conn.Query(context.Background(), stmt, valueArgs...)
	if error != nil {
		cache.debugger.error("error with batchInsert ", error.Error())

	}
	rsp := make([]interface{}, 0)
	for rows.Next() != false {
		var id int
		var code string
		rows.Scan(&id, &code)
		rsp = append(rsp, id)
		rsp = append(rsp, code)
	}
	return rsp
}
func now()  string{
	t := time.Now()
	formatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return formatted
}