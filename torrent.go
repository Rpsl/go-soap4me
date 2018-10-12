package main

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/dustin/go-humanize"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// todo
// В перспективе можно вынести в интерфейс и сделать разные backend для скачивания
// aria2c, api для transmission. Можно даже из .torrent файла грепать ссылку на http сида и так скачивать
func DownloadEpisodes(episodes []Episode) {
	cleanUp()
	runtime.GOMAXPROCS(Config.Threads)

	var wg sync.WaitGroup

	wg.Add(Config.Threads)
	var threads = 0

	// тут можно качать до трех файла одновременно
	for _, episode := range episodes {
		if _, err := os.Stat(episode.path); os.IsNotExist(err) {
			threads++
			go doEpisode(episode, &wg)
		}

		if threads == Config.Threads {
			wg.Wait()
			threads = 0
		}
	}
}

func doEpisode(episode Episode, wg *sync.WaitGroup) {
	defer wg.Done()

	storageProvider := storage.NewMapPieceCompletion()

	defer storageProvider.Close()

	clientConfig := torrent.ClientConfig{
		DataDir:  Config.TempDir,
		NoUpload: true,
		Seed:     false,
		NoDefaultPortForwarding: true,
		DefaultStorage:          storage.NewMMapWithCompletion(Config.TempDir, storageProvider),
	}

	client, err := torrent.NewClient(&clientConfig)

	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}

	defer client.Close()

	log.Printf("Start downloading: %s - %s", episode.show, episode.title)

	tor := runTorrent(client, downloadTorrentFile(episode))

	if len(tor.Files()) > 1 {
		client.Close()
		log.Fatal("Our torrent have more than one file")
		return
	}

	if client.WaitAll() {
		moveFile(filepath.Join(Config.TempDir, tor.Files()[0].Path()), episode.path)
		tor.Drop()
		client.Close()
	}
}

func runTorrent(client *torrent.Client, torrentFile string) *torrent.Torrent {
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
					if t.BytesCompleted() == t.Info().TotalLength() {
						continue
					}
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

	return t
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

func moveFile(fromPath string, toPath string) {
	err := os.MkdirAll(filepath.Dir(toPath), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Rename(fromPath, toPath)
	if err != nil {
		log.Fatal(err)
	}
}

// Удаление старых временных файлов, если они есть
func cleanUp() {
	files, err := filepath.Glob(filepath.Join(os.TempDir(), "soap4me*"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		err = os.Remove(f)
		if err != nil {
			log.Println(fmt.Printf("Can't remove old torrent file :: %s", f))
		}
	}
}
