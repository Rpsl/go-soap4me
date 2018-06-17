package main

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"path/filepath"
	"time"
	"io/ioutil"
	"runtime"
	"sync"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/missinggo/filecache"
)

func DownloadEpisodes(episodes []Episode) {
	runtime.GOMAXPROCS(2)

	var wg sync.WaitGroup

	// todo нужно перенести в конфиг
	wg.Add(2)
	var threads = 0

	// тут можно качать до трех файла одновременно
	// а кол-во загрузок вынести в конфиг
	for _, episode := range episodes {
		if _, err := os.Stat(episode.path); os.IsNotExist(err) {
			threads++
			go doEpisode(episode, &wg)
		}

		// todo кол-во тредов нужно уменьшать
		if threads == 2 {
			wg.Wait()
		}
	}
}

func doEpisode(episode Episode, wg *sync.WaitGroup)  {
	defer wg.Done()

	fileCache, err := filecache.NewCache(Config.TempDir)
	if err != nil {
		return
	}
	fileCache.SetCapacity(10 << 30)
	storageProvider := fileCache.AsResourceProvider()


	clientConfig := torrent.Config{
		DataDir:  Config.TempDir,
		NoUpload: true,
		Seed:     false,
		NoDefaultPortForwarding:true,
		DefaultStorage: storage.NewResourcePieces(storageProvider),
	}

	client, err := torrent.NewClient(&clientConfig)

	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}

	defer client.Close()

	log.Printf("Start downloading: %s - %s", episode.show, episode.title)

	files := runTorrent(client, downloadTorrentFile(episode))

	if len(files) > 1 {
		client.Close()
		log.Fatal("Our torrent have more than one file")
		return
	}

	if client.WaitAll() {
		MoveFile(filepath.Join(Config.TempDir, files[0].Path()), episode.path)
		client.Close()
	}
}

func runTorrent(client *torrent.Client, torrentFile string) []*torrent.File {
	t := func() *torrent.Torrent {
		metaInfo, err := metainfo.LoadFromFile(torrentFile)

		if err != nil {
			log.Fatalf("error loading torrent file %q: %s\n", torrentFile, err)
		}

		t, err := client.AddTorrent(metaInfo)

		if err != nil {
			log.Fatal(err)
		}
		return t
	}()

	if Config.Debug {
		go func() {
			for {
				select {
				case <-t.GotInfo():
					fmt.Printf(
						"%s: %s/%s\n",
						t.Files()[0].DisplayPath(),
						humanize.Bytes(uint64(t.BytesCompleted())),
						humanize.Bytes(uint64(t.Info().TotalLength())),
					)
				}
				time.Sleep(time.Second * 5)
			}
		}()
	}

	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()

	return t.Files()
}


func downloadTorrentFile(episode Episode) string {
	content := doRequest(episode.torrent)
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
