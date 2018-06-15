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
)

func download(episodes []Episode) {
	clientConfig := torrent.Config{
		DataDir:  Config.TempDir,
		NoUpload: true,
		Seed:     false,
	}

	for _, episode := range episodes {
		if _, err := os.Stat(getEpisodePath(episode)); os.IsNotExist(err) {
			client, err := torrent.NewClient(&clientConfig)

			if err != nil {
				log.Fatalf("error creating client: %s", err)
			}

			defer client.Close()

			fmt.Println(clientConfig.DataDir)
			fmt.Println(episode.show)
			files := runTorrent(client, downloadTorrentFile(episode))

			if len(files) > 1 {
				log.Panic("Our torrent have more than one file")
			}

			if client.WaitAll() {
				MoveFile(filepath.Join(Config.TempDir, files[0].Path()), getEpisodePath(episode))
			} else {
				log.Fatal("y u no complete torrents?!")
			}
		}
	}
}

func runTorrent(client *torrent.Client, torrentFile string) []*torrent.File {
	t := func() *torrent.Torrent {
		metaInfo, err := metainfo.LoadFromFile(torrentFile)

		if err != nil {
			log.Panicf("error loading torrent file %q: %s\n", torrentFile, err)
			os.Exit(1)
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
					fmt.Println(fmt.Sprintf("downloading (%s/%s)", humanize.Bytes(uint64(t.BytesCompleted())), humanize.Bytes(uint64(t.Info().TotalLength()))))
				}
				time.Sleep(time.Second * 1)
			}
		}()
	}

	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()

	return t.Files()
}
