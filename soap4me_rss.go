package main

import (
	"log"
	"regexp"

	"github.com/mmcdole/gofeed"
)

const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36"

type Rss struct {
	FeedUrl string
}

func (rss Rss) ParseFeed() []Episode {
	feedData := doRequest(rss.FeedUrl)

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(feedData)

	if err != nil {
		log.Fatal("unable to parse feed", err)
	}

	return rss.parseItems(feed)
}

func (rss Rss) parseItems(items *gofeed.Feed) []Episode {
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
