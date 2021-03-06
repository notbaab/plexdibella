package main

import (
	"fmt"
	"github.com/jrudio/go-plex-client"
	"github.com/notbaab/plexdibella"
	"github.com/urfave/cli"
	"log"
	"os"
)

var (
	url   string
	token string
)

func test(c *cli.Context) {
	// plexdibella.GetCleanNames
	p, err := plex.New(url, token)
	if err != nil {
		log.Panicln(err)
	}

	renameMapChan := make(chan plexdibella.RenameMap, 100)
	errChan := make(chan error, 1)
	go plexdibella.StreamAllCleanNames(p, renameMapChan, errChan)
	if err != nil {
		log.Panicln(err)
	}

	for nameMap := range renameMapChan {
		if true {
			fmt.Printf("%s -> %s\n", nameMap.Src, nameMap.Dest)
		}
	}

	for err := range errChan {
		log.Panicln(err)
	}
}

func main() {
	app := cli.NewApp()

	app.Name = "plexdibella"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "url, u",
			Usage:       "Plex url or ip",
			Destination: &url,
		},
		cli.StringFlag{
			Name:        "token, tkn",
			Usage:       "abc123",
			Destination: &token,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "test",
			Usage:  "Test your connection to your Plex Media Server",
			Action: test,
		},
	}
	app.Run(os.Args)
}
