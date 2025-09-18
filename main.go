package main

import(
	"net/http"
	"net/url"
	"sync"
	"io"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"flag"
	"strconv"
)

var (
	port int
	client *http.Client
	config SiteConfig
)


func searchHandler(keyword string, w http.ResponseWriter){
	resultCh := make(chan string)
	wg := sync.WaitGroup{}

	for _, item := range config.BaseUrls {
		wg.Add(1)

		go func(item BaseUrlItem){
			defer wg.Done()
			url := fmt.Sprintf("%s?ac=detail&wd=%s", item.BaseUrl, url.QueryEscape(keyword))
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

			if resp.StatusCode != 200 {
				log.Println(item.Name, resp.Status, url)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println(item.Name, err)
				return
			}

			var cms CMSResponse
			err = json.Unmarshal(body, &cms)
			if err != nil {
				log.Println(item.Name, err)
				return 
			}

			if cms.Code != 1 || len(cms.List) == 0{
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

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		io.WriteString(w, "data: ")
		io.WriteString(w, res)
		io.WriteString(w, "\n\n")
	}
}


func main(){
	cc, err := ReadConfig()
	if err != nil {
		log.Println(err)
		return 
	}
	config = cc

	client = &http.Client{Timeout: time.Duration(config.Timeout) * time.Second,}

	http.Handle("/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			data, err := json.Marshal(config)
			if err != nil {
				log.Println(err)
				return 
			}
			io.WriteString(w, string(data))
		} else if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			var newCfg SiteConfig
			if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "bad json")
				return
			}
			config = newCfg
			io.WriteString(w, "200 OK")
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

	fmt.Println("http://127.0.0.1:" + strconv.Itoa(port))
	err = http.ListenAndServe(":" + strconv.Itoa(port), nil)
	if err != nil {
		log.Println(err)
		return 
	}
}


func init(){
	flag.IntVar(&port, "p", 22222, "listening port, between 0 and 65535")
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}