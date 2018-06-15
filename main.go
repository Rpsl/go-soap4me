package main

import (
	"fmt"
	"github.com/crgimenes/goconfig"
	_ "github.com/crgimenes/goconfig/yaml"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36"

type ConfigStruct struct {
	FeedUrl  string `yaml:"feed_url" cfg:"feed_url"`
	TempDir  string `yaml:"tmp_dir" cfg:"tmp_dir"`
	ShowsDir string `yaml:"shows_dir" cfg:"shows_dir"`
}

var Config = ConfigStruct{}

func init() {
	goconfig.File = "config.yaml"
	err := goconfig.Parse(&Config)

	if err != nil {
		log.Panic(err)
		return
	}

	createDir(Config.TempDir)
	createDir(Config.ShowsDir)
}

func main() {
	CleanUp()

	if offline() {
		fmt.Println("No network connection. Cannot get RSS feed")
		return
	}

	episodes := ParseFeed(Config.FeedUrl)

	download(episodes)

}

func MoveFile(fromPath string, toPath string) {
	err := os.MkdirAll(filepath.Dir(toPath), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Rename(fromPath, toPath)
	if err != nil {
		panic(err)
	}
}

func CleanUp() {
	files, err := filepath.Glob(filepath.Join(os.TempDir(), "soap4me*"))
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		err = os.Remove(f)
		if err != nil {
			log.Println(fmt.Printf("Can't remove old torrent file :: %s", f))
		}
	}
}

func offline() bool {
	//Chances are if Google's down, the internet is down.
	_, err := http.Get("https://google.com/")

	return err != nil
}

func createDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)

		if err != nil {
			log.Panicf("Can't create dir %s %s", path, err)
		}
	}
}
