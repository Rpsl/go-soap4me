package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Episode struct {
	show    string
	title   string
	season  string
	episode string
	torrent string
}

func downloadTorrentFile(episode Episode) string {
	content := soapDownload(episode.torrent)
	tempFile, err := ioutil.TempFile(os.TempDir(), "soap4me")

	if err != nil {
		log.Fatal(err)
	}

	if _, err := tempFile.Write([]byte(content)); err != nil {
		log.Fatal(err)
	}

	if err := tempFile.Close(); err != nil {
		log.Fatal(err)
	}

	return tempFile.Name()
}

func getEpisodePath(episode Episode) string {
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

func soapDownload(url string) string {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", USER_AGENT)

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)

	return string(contents)
}

func ParseFeed(feedUrl string) []Episode {
	feedData := soapDownload(feedUrl)

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(feedData)

	if err != nil {
		log.Panic("unable to parse feed", err)
	}

	return parseItems(feed)
}

func parseItems(items *gofeed.Feed) []Episode {
	r, _ := regexp.Compile(`(.*?) / сезон ([0-9]+) эпизод ([0-9]+) / (.*?) / (.*)`)

	episodes := []Episode{}

	for _, item := range items.Items {
		found := r.FindStringSubmatch(item.Title)

		tmp := Episode{
			show:    found[1],
			title:   found[4],
			season:  found[2],
			episode: found[3],
			torrent: item.GUID,
		}

		episodes = append(episodes, tmp)
	}

	return episodes
}
