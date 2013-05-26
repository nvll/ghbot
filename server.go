package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Actor_t struct {
	Id          int64
	Login       string
	Gravatar_Id string
	Url         string
	Avatar_Url  string
}

type Repo_t struct {
	Id   int64
	Name string
	Url  string
}

type Author_t struct {
	Email string
	Name  string
}

type Commit_t struct {
	Sha      string
	Author   Author_t
	Message  string
	Distinct bool
	Url      string
}

type Payload_t struct {
	Push_Id       int64
	Size          int
	Distinct_Size int
	Ref           string
	Head          string
	Before        string
	Commits       []Commit_t
}

type GithubEvents_t struct {
	Id         string
	Type       string
	Actor      Actor_t
	Repo       Repo_t
	Payload    Payload_t
	Public     bool
	Created_At string
}

func sendHTTPRequest(method, url string) ([]byte, error) {
	var err error
	var request *http.Request
	var resp *http.Response
	var json_blob []byte
	client := &http.Client{}

	if request, err = http.NewRequest(method, url, nil); err != nil {
		return nil, fmt.Errorf("request creation failed %s\n", err.Error())
	}

	request.Header.Add("Accept", "application/json")
	if resp, err = client.Do(request); err != nil {
		return nil, fmt.Errorf("request failed %s\n", err.Error())
	}

	if json_blob, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("couldn't parse response %s\n", err.Error())
	}

	resp.Body.Close()
	return json_blob, nil
}

var requestIds struct {
	sync.RWMutex
	ids map[string]int
}

func startServer(parent chan bool, cfg Config, msg chan string) {
	var err error
	var json_blob []byte
	var events []GithubEvents_t
	var event_time, last_poll_time time.Time

	last_poll_time = time.Now()

	for {
		for u := 0; u < len(cfg.Users); u++ {
			request_url := fmt.Sprintf("https://api.github.com/users/%s/events?client_id=%s&client_secret=%s", cfg.Users[u], cfg.ClientId, cfg.ClientSecret)

			json_blob, err = sendHTTPRequest("GET", request_url)
			if err := json.Unmarshal(json_blob, &events); err != nil {
				log.Printf("%v\n", string(json_blob))
				log.Fatal("Couldn't parse config json: ", err.Error())
			}

			for i := 0; i < len(events); i++ {
				switch events[i].Type {
				case "PushEvent":
					if event_time, err = time.Parse(time.RFC3339, events[i].Created_At); err != nil {
						log.Printf("Couldn't parse time, ignoring event\n")
						continue
					}

					if event_time.After(last_poll_time) {
						for c := 0; c < len(events[i].Payload.Commits); c++ {
							s := fmt.Sprintf("%s pushed to [%s] \"%s\"\n", events[i].Actor.Login, events[i].Repo.Name, strings.Split(events[i].Payload.Commits[c].Message, "\n")[0])
							log.Print(s)
							msg <- s
						}
					}
				}
			}
		}

		last_poll_time = time.Now()
		time.Sleep(5 * time.Minute)
	}
}
