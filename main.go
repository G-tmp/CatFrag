package main

import(
	"net/http"
	"sync"
	"io"
	"encoding/json"
	"fmt"
	"log"
	"time"
)


func searchHandler(keyword string, w http.ResponseWriter){
	config, err := ReadConfig()
	if err != nil {
		log.Println(err)
		return 
	}
	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	resultCh := make(chan string)
	wg := sync.WaitGroup{}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for _, item := range config.BaseUrls {
		wg.Add(1)

		go func(item BaseUrlItem){
			defer wg.Done()
			url := fmt.Sprintf("%s?ac=detail&wd=%s", item.BaseUrl, keyword)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Println(item.Name, err)
				return 
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
			resp, err := client.Do(req)
			if err != nil {
				log.Println(item.Name ,err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println(item.Name, err)		
				return
			}

			cms := CMSResponse{}
			err = json.Unmarshal(body, &cms)
			if err != nil {
				log.Println(item.Name, err)
				return 
			}

			if cms.Code != 1 {
				return
			}

			re := ResultData{
				item.Name, cms.List,
			}
			data, err := json.Marshal(re)
			if err != nil {
				log.Println(err)
				return 
			}

			resultCh <- string(data)
			// resultCh <- item.Name + string(body)
		}(item)
	}

	for res := range resultCh {
		io.WriteString(w, "data: ")
		io.WriteString(w, res)
		io.WriteString(w, "\n\n")
        flusher.Flush()
	}
}


func main(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.Handle("/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			rawConfig, err := ReadRawConfig()
			if err != nil {
				log.Println(err)
				return 
			}
			io.WriteString(w, string(rawConfig))
		} else if r.Method == http.MethodPost {

		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			keyword := r.URL.Query().Get("keyword")
			if keyword == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"error": "keyword is required"}`)
				return
			}

			searchHandler(keyword, w)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("http://127.0.0.1:22222")
	err := http.ListenAndServe(":22222", nil)
	if err != nil {
		log.Println(err)
		return 
	}
}
