package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"os"
	"sort"
	"strconv"
	"sync"
)
//fileDir = "C:\\xamppe\\first_upload.xlsx"
const (
	sheet ="Hoja1"
	step = 1000
)

type ExcelAsset struct {
	ZipCode 				int16
	AssociatedTerrain 		string
	AssociatedBuilding 		string
	Lenght 					float32
	SurfaceSoil 			float32
	SurfaceBuilt 			float32
	id 						int64
	Address					string
	AcquisitionTypeCode 	string
	BorrowStart 			string
	BorrowEnd 				string
	SubfamilyCode 			string
	TypeCode 				string
	ImageUrl 				string
	TagCode 				string
	TagType 				string
	ICode 					string
	Denomination 			string
	Catastral 				string
	Characteristics 		string
	Measures 				string
	Brand 					string
	Model 					string
	Serial 					string
	Matricula 				string
	Bastidor 				string
	Description 			string
	CenterCode 				string
	BuildingCode 			string
	LevelCode 				string
	SpaceCode 				string
	CostCenterCode 			string
	AreaCode 				string
	OwnerEntity 			string
	LendEntity 				string
	Color 					string
	Material 				string

}

func (ptr *ExcelAsset) init(row []string)  {
	if len(row) == 0{
		return
	}
	if len(row) < 38 {
		pad := 38 -len(row)
		for i := 0; i < pad; i++ {
			row = append(row, "")
		}
	}
	ids,_ := strconv.ParseInt(row[0],10,64)

	ptr.id =ids
	ptr.AcquisitionTypeCode = row[1]
	ptr.BorrowStart = row[2]
	ptr.BorrowEnd = row[3]
	ptr.SubfamilyCode = row[4]
	ptr.TypeCode = row[5]
	ptr.ImageUrl = row[6]
	ptr.TagCode = row[7]
	ptr.TagType = row[8]
	ptr.ICode = row[9]
	ptr.Denomination = row[10]
	ptr.Catastral = row[11]
	ptr.Address = row[12]
	i64ZipCode, _ := strconv.ParseInt(row[13], 10, 16)
	ptr.ZipCode =int16(i64ZipCode)
	f64SurfaceSoil,_ := strconv.ParseFloat(row[14], 32)
	ptr.SurfaceSoil = float32(f64SurfaceSoil)
	f64SurfaceBuilt, _ := strconv.ParseFloat(row[15], 32)
	ptr.SurfaceBuilt = float32(f64SurfaceBuilt)
	ptr.AssociatedTerrain = row[16]
	ptr.AssociatedBuilding = row[17]
	f64Lenght,_ := strconv.ParseFloat(row[18], 32)
	ptr.Lenght = float32(f64Lenght)
	ptr.Characteristics = row[19]
	ptr.Measures = row[20]
	ptr.Brand = row[21]
	ptr.Model = row[22]
	ptr.Serial = row[23]
	ptr.Matricula = row[24]
	ptr.Bastidor = row[25]
	ptr.Description = row [26]
	ptr.CenterCode = row[27]
	ptr.BuildingCode = row[28]
	ptr.LevelCode = row[29]
	ptr.SpaceCode = row[30]
	ptr.CostCenterCode = row[31]
	ptr.AreaCode = row[32]
	ptr.OwnerEntity = row[33]
	ptr.LendEntity = row[34]
	ptr.Color = row[36]
	ptr.Material = row[37]

}

func sheetCunk( rows [][]string, assets *[]*ExcelAsset){
	ln := len(rows)

	step:= 4
	rest :=ln % step
	for i := 0; i < (ln - rest); i+=step {
		iIndex := i

		wg := sync.WaitGroup{}
		channel := make(chan *ExcelAsset, step)
		wg.Add(step)
		for j := 0; j < step; j++ {
			jIndex := j
			rowIndex := iIndex + jIndex
			row := rows[rowIndex]
			go doMake(channel,row, &wg)
		}
		wg.Wait()
		for i := 0; i < step; i++ {
			excelAsset:= <- channel
			if excelAsset == nil || excelAsset.id == 0{
				continue
			}
			*assets = append(*assets, excelAsset)
		}
		close(channel)
	}
	if rest >0 {
		for i := ln - rest; i < ln; i++ {
			iIndex := i
			row := rows[iIndex]
			if len(row) == 0 {
				continue
			}
			if row[0] == "" {
				continue
			}
			asset := &ExcelAsset{}
			asset.init(row)
			*assets = append(*assets,asset)
		}
	}
}

func doMake(ch chan *ExcelAsset, row []string, wg *sync.WaitGroup){
	defer wg.Done()
		if len(row) == 0{
			return
		}
		asset := &ExcelAsset{}
		asset.init(row)
		ch <- asset
}


func loadData(fileDir string)  []*ExcelAsset{
	excel, error := excelize.OpenFile(fileDir)
	if error != nil {
		debugger.error(error.Error())
		os.Exit(1)
	}
	debugger.chrono("Finished opening excel at ")
	rows, err := excel.GetRows(sheet)
	//skipping excel headers
	rows = rows[1:]
	if err != nil {
		debugger.error(err.Error())
		os.Exit(1)
	}
	debugger.chrono("Finished loading excel data rows at ")
	assets := make([]*ExcelAsset,0)

	sheetCunk(rows, &assets)
	fmt.Println("assets", len(assets))
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].id < assets[j].id
	})
	return assets
}


