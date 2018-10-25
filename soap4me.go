package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Episode struct {
	show    string
	title   string
	season  string
	episode string
	torrent string
	path    string
}

// todo escape file names
func GetEpisodePath(episode Episode) string {
	return fmt.Sprintf(
		"%s/%s/Season %s/s%se%s %s.mp4",
		Config.ShowsDir,
		episode.show,
		episode.season,
		episode.season,
		episode.episode,
		episode.title,
	)
}

// отдельный метод, т.к. необходимо устанавливать user-agent на каждый запрос,
// в противном случае soap4me будет возвращать 403 ошибку
func doRequest(url string) string {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", UserAgent)
	response, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	contents, _ := ioutil.ReadAll(response.Body)
	return string(contents)
}
