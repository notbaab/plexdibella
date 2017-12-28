package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/notbaab/go-plex-client"
	"github.com/notbaab/plexdibella"
	"log"
)

func main() {
	Plex, err := plex.New("url", "token")
	fmt.Println(Plex.URL)
	id, err := Plex.GetMachineID()

	if err != nil {
		log.Println(err)
	}

	fmt.Println(id)
	sections, err := Plex.GetLibraries()
	spew.Dump(sections)

	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "movie" {
			renameMap := plexdibella.GetCleanNamesMovies(Plex, section)
			fmt.Printf("%+v\n", renameMap)
		}
	}
}
