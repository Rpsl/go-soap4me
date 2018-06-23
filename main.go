package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/crgimenes/goconfig"
	_ "github.com/crgimenes/goconfig/yaml"
)

type ConfigStruct struct {
	FeedUrl  string `yaml:"feed_url" cfg:"feed_url"`
	TempDir  string `yaml:"tmp_dir" cfg:"tmp_dir"`
	ShowsDir string `yaml:"shows_dir" cfg:"shows_dir"`
	Debug    bool   `yaml:"debug" cfg:"debug" cfgDefault:"false"`
	Threads  int    `yaml:"threads" cfg:"threads" cfgDefault:"1"`
}

var Config = ConfigStruct{}

func init() {
	goconfig.File = "config.yaml"
	err := goconfig.Parse(&Config)

	if err != nil {
		log.Fatal(err)
		return
	}

	Config.TempDir, _ = filepath.Abs(Config.TempDir)
	Config.ShowsDir, _ = filepath.Abs(Config.ShowsDir)

	log.Println("Temporary dir is:", Config.TempDir)
	log.Println("Shows dir is:", Config.ShowsDir)

	createDir(Config.TempDir)
	createDir(Config.ShowsDir)
}

// по хорошему это можно завернуть в loop и пусть демон всегда живет
// но надо ли?
func main() {
	episodes := ParseFeed(Config.FeedUrl)

	DownloadEpisodes(episodes)

}

func createDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)

		if err != nil {
			log.Fatalf("Can't create dir %s %s", path, err)
		}
	}
}
