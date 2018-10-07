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

	renameMap, err := plexdibella.GetAllCleanNames(p)
	if err != nil {
		log.Panicln(err)
	}

	fmt.Printf("%d files to rename\n", len(renameMap))
	for _, nameMap := range renameMap {
		if false {
			fmt.Printf("%s -> %s\n", nameMap.Src, nameMap.Dest)
		}
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
