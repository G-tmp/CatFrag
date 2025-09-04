package main

import(
	"encoding/json"
	"strings"
)

// Parse api return json CMS data

type Episode struct {
	EpName		string `json:"ep_name"`
	EpURL		string `json:"ep_url"`
}

type CMSItem struct {
	VodName		string		`json:"vod_name"`
	VodPic		string		`json:"vod_pic"`
	Vods		[]Episode 	`json:"vods"`
}

type CMSResponse struct{
	Code	int			`json:"code"`
	List	[]CMSItem 	`json:"list"`
}

func (c *CMSItem)UnmarshalJSON(data []byte) error{
	// Create a shadow struct with vod_play_url as string
	type Alias CMSItem
	aux := struct{
		VodsURL	string	`json:"vod_play_url"`
		*Alias		// Embedded field
	}{
		Alias: (*Alias)(c),
	}
	
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	var eps []Episode
	if aux.VodsURL != "" {
		playSource1 := strings.Split(aux.VodsURL, "$$$")[0]
		parts := strings.Split(playSource1, "#")
		for _, p := range parts {
			kv := strings.SplitN(p, "$", 2)
			if len(kv) == 2 {
				eps = append(eps, Episode{
					EpName: kv[0],
					EpURL: kv[1],
				})
			}
		}
	}
	c.Vods = eps

	return nil
}


type ResultData struct {
	Name 		string			`json:"name"`
	Result 			[]CMSItem 		`json:"result"`
}