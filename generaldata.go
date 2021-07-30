package main

import (
	"fmt"
	"strconv"
	"strings"
)

type GeneralAspect struct {
	Validated			bool
	RiskLevel			int8
	ZipCode				string
	Width				float32
	Height				float32
	Depth				float32
	SurfaceFiels		float32
	SurfaceBuilt		float32
	Length				float32
	id 					int64
	Generic_Asset_Id    int64
	Asset_Model_Id    	int64
	Brand_Id	    	int64
	Associated_Surface 	int64
	Associated_Build 	int64
	Color_Id		 	int64
	Material_Id			int64
	TechStatus_Id	 	int64
	Description 		string
	Denomination 		string
	SerialNumber		string
	OldTagCode			string
	Created_At			string
	Updated_At			string
	Address				string
	Characteristics		string
	Matricula			string
	Bastidor			string
	Catastral			string

}

func (ptr *GeneralAspect) init(excelAsset *ExcelAsset){
	modelId := cache.findByCode("asset_models", excelAsset.Model)
	colorId := cache.findByCode("colors", excelAsset.Color)
	materiaId := cache.findByCode("materials", excelAsset.Material)
	var width, height, depth float32
	if excelAsset.Measures != "" {
		width, height, depth = extractmeasures(excelAsset)
	}
	now := now()
	ptr.Validated = true
	ptr.Description = excelAsset.Description
	ptr.Associated_Build, _ = strconv.ParseInt(excelAsset.AssociatedBuilding,10,64)
	ptr.Associated_Surface,_ = strconv.ParseInt(excelAsset.AssociatedTerrain,10,64)

	ptr.Asset_Model_Id = realint64(modelId)
	ptr.Depth = depth
	ptr.Width = width
	ptr.Height = height
	ptr.Color_Id = realint64(colorId)
	ptr.Material_Id = realint64(materiaId)
	ptr.Denomination = excelAsset.Denomination
	ptr.Address = excelAsset.Address
	ptr.ZipCode = fmt.Sprint(excelAsset.ZipCode)
	ptr.SurfaceFiels = excelAsset.SurfaceSoil
	ptr.SurfaceBuilt = excelAsset.SurfaceBuilt
	ptr.Length = excelAsset.Lenght
	ptr.Characteristics = excelAsset.Characteristics
	ptr.SerialNumber = excelAsset.Serial
	ptr.Matricula = excelAsset.Matricula
	ptr.Bastidor = excelAsset.Bastidor
	ptr.RiskLevel = 0
	ptr.Catastral = excelAsset.Catastral
	ptr.Generic_Asset_Id = excelAsset.id
	ptr.Created_At = now
	ptr.Updated_At = now
}

func (ptr *GeneralAspect) after(asset *GenericAsset)  []string{

	return make([]string,0)
}

func extractmeasures(excelAsset *ExcelAsset)  (float32, float32, float32){
	var width, height, depth float32
	if excelAsset.Measures != "" {
		s := strings.ToLower(excelAsset.Measures)
		spl := strings.Split(s, "x")
		if len(spl) == 3 {
			w,_ := strconv.ParseFloat(spl[0],32)
			h,_ := strconv.ParseFloat(spl[1],32)
			d,_ := strconv.ParseFloat(spl[2],32)
			width = float32(w)
			height = float32(h)
			depth = float32(d)
		}
		if len(spl) == 2 {
			w,_ := strconv.ParseFloat(spl[0],32)
			h,_ := strconv.ParseFloat(spl[1],32)
			width = float32(w)
			height = float32(h)
		}
		if len(spl) == 1 {
			w,_ := strconv.ParseFloat(spl[0],32)
			width = float32(w)
		}
	}
	return width, height, depth
}