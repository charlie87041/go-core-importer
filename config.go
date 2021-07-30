package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	RedisUrl string					`json:"RedisUrl"`
	RedisChannel string				`json:"RedisChannel"`
	Locale string					`json:"Locale"`
	DefaultCenterCode  string		`json:"DefaultCenterCode"`
	DefaultEdificeCode string		`json:"DefaultEdificeCode"`
	DefaultLevelCode   string		`json:"DefaultLevelCode"`
	DefaultSpaceCode   string		`json:"DefaultSpaceCode"`
	DefaultBrand 		 string		`json:"DefaultBrand"`
	DefaultModel 		 string		`json:"DefaultModel"`
	DefaultColor 		 string		`json:"DefaultColor"`
	DefaultMaterial    string		`json:"DefaultMaterial"`
	DefaultArea		 string			`json:"DefaultArea"`
	Debug			 bool			`json:"Debug"`

}

func (ptr *Config) loadConfig()  {
	configFile,error := os.ReadFile("settings.json")

	if error != nil || ptr == nil {
		cache.debugger.debug("config file not exists. Proceeding with default values")
		ptr.defaultConfig()
		ptr.createFile()
	}else{
		json.Unmarshal(configFile, ptr)
		if ptr.RedisUrl == ""{
			cache.debugger.debug("eroor opening config file ")
			ptr.defaultConfig()
			ptr.createFile()
		}
	}

}
func (ptr *Config) defaultConfig()  {
	ptr.RedisUrl = "localhost:6379"
	ptr.RedisChannel = "database_import_data_server"
	ptr.Locale = "Spain/Madrid"
	ptr.DefaultCenterCode    = "01 "
	ptr.DefaultEdificeCode   = "01/01"
	ptr.DefaultLevelCode   	 = "01/01/01"
	ptr.DefaultSpaceCode   	 = "01/01/01/001"
	ptr.DefaultBrand 		 = "Undefined"
	ptr.DefaultModel 		 = "Undefined"
	ptr.DefaultColor 		 = "default"
	ptr.DefaultMaterial      = "default"
	ptr.DefaultArea		     = "01"
	ptr.Debug				 = true
}
func (ptr *Config) createFile()  {
	marshalledConfig, _:= json.MarshalIndent((*ptr), "", "\n \t")

	err := os.WriteFile("settings.json", marshalledConfig, 0777)
	if err != nil {
		cache.debugger.error("error writing config")
		os.Exit(1)
	}

}


