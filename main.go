package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/notbaab/go-plex-client"
	"log"
	"strings"
)

func GetCleanNamesMovies(plexInstance *plex.Plex, sectionDirectory plex.Directory) {
	// first get the section meta data
	section, pErr := plexInstance.GetSection(sectionDirectory.Key)
	spew.Dump(section)

	if pErr != nil {
		log.Println(pErr)
		log.Println(section)
	}

	// Walk over metadata container
	for _, metadata := range section.MediaContainer.Metadata {
		// get all the media associated with the movie
		// I'm honestly not sure how to differentiate between media and parts
		for _, media := range metadata.Media {
			// Walk over all the parts.
			for _, part := range media.Part {
				newFileName := metadata.Title + "." + part.Container
				libraryLocation := ""

				// Walk over locations to see where it's located
				for _, location := range sectionDirectory.Location {
					if strings.HasPrefix(part.File, location.Path) {
						libraryLocation = location.Path
						break
					}
				}

				// got the matching library location
				fullFileName := libraryLocation + "/" + newFileName
				fmt.Printf("Changing file %s to %s\n", part.File, fullFileName)
			}
		}
	}
}

func main() {
	Plex, err := plex.New("url", "Magic Token")
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
			GetCleanNamesMovies(Plex, section)
		}
	}
}
