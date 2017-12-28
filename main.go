package plexdibella

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/notbaab/go-plex-client"
	"log"
	"strings"
)

type RenameMap struct {
	src  string
	dest string
}

func GetCleanNamesMovies(plexInstance *plex.Plex, sectionDirectory plex.Directory) []RenameMap {
	// first get the section meta data
	section, pErr := plexInstance.GetSection(sectionDirectory.Key)
	spew.Dump(section)

	if pErr != nil {
		log.Println(pErr)
		log.Println(section)
	}
	renamedMap := []RenameMap{}

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
				renamedMap = append(renamedMap, RenameMap{part.File, fullFileName})
			}
		}
	}

	return renamedMap
}
