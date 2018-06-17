package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36"

type Episode struct {
	show    string
	title   string
	season  string
	episode string
	torrent string
	path    string
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

func ParseFeed(feedUrl string) []Episode {
	feedData := doRequest(feedUrl)

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(feedData)

	if err != nil {
		log.Fatal("unable to parse feed", err)
	}

	return parseItems(feed)
}

func parseItems(items *gofeed.Feed) []Episode {
	var episodes []Episode

	r, _ := regexp.Compile(`(.*?) / сезон ([0-9]+) эпизод ([0-9]+) / (.*?) / (.*)`)

	for _, item := range items.Items {
		match := r.FindStringSubmatch(item.Title)

		if match == nil {
			continue
		}

		ep := Episode{
			show:    match[1],
			title:   match[4],
			season:  match[2],
			episode: match[3],
			torrent: item.GUID,
		}
		ep.path = GetEpisodePath(ep)

		episodes = append(episodes, ep)
	}

	return episodes
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
