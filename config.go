package main

import (
	"os"
	"encoding/json"
)

// Read config.json data

type SiteConfig struct {
	SiteName              string        `json:"site_name"`
	Timeout               int           `json:"timeout"`
	BgImgPc				  string		`json:"background_image_pc"`
	BgImgPhone			  string		`json:"background_image_phone"`
	BaseUrls              []BaseUrlItem `json:"base_urls"`
}

type BaseUrlItem struct {
	Name 				string			`json:"name"`
	BaseUrl				string			`json:"base_url"`
	FilterTid			[]int			`json:"filter_type_id"`
}


func readRawConfig() ([]byte, error){
	return os.ReadFile("config.json")
}


func ReadConfig() (SiteConfig, error) {
	config, err := readRawConfig()
	if err != nil {
		return SiteConfig{}, err
	}
	
	var siteConfig SiteConfig

	err = json.Unmarshal(config, &siteConfig)
	return siteConfig, err
}