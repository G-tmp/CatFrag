package main

import (
	"os"
	"encoding/json"
)

// Read config.json data

type SiteConfig struct {
	SiteName              string        `json:"site_name"`
	Timeout               int           `json:"timeout"`
	BaseUrls              []BaseUrlItem `json:"base_urls"`
	BgImgPc				string			`json:"background_image_pc"`
}

type BaseUrlItem struct {
	Name 				string			`json:"name"`
	BaseUrl				string			`json:"base_url"`
}


func ReadRawConfig() ([]byte, error){
	return os.ReadFile("config.json")
}


func ReadConfig() (SiteConfig, error) {
	config, err := ReadRawConfig()
	if err != nil {
		return SiteConfig{}, err
	}
	
	var siteConfig SiteConfig

	err = json.Unmarshal(config, &siteConfig)
	return siteConfig, err
}