package plexdibella

import (
	"errors"
	"fmt"
	"github.com/jrudio/go-plex-client"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type RenameMap struct {
	Src  string
	Dest string
}

type streamFunc func(*plex.Plex, plex.Directory, chan RenameMap)

func RenameMediaLibraryFiles(p *plex.Plex) error {
	renameMap, err := GetAllCleanNames(p)

	if err != nil {
		return err
	}

	for _, renamePair := range renameMap {
		dir := filepath.Dir(renamePair.Dest)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("file does not exist")
			os.MkdirAll(dir, os.ModePerm)
		}

		err = os.Rename(renamePair.Src, renamePair.Dest)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func streamCleanNames(p *plex.Plex, renameMapChan chan RenameMap, section plex.Directory, cleanFunc streamFunc) {
	subRenameMapChan := make(chan RenameMap, 100)
	go cleanFunc(p, section, subRenameMapChan)
	for renameMap := range subRenameMapChan {
		renameMapChan <- renameMap
	}
}

func StreamAllCleanNames(p *plex.Plex, renameMapChan chan RenameMap, errChan chan error) {
	sections, err := p.GetLibraries()
	defer close(renameMapChan)
	defer close(errChan)
	if err != nil {
		errChan <- err
		return
	}

	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "show" {
			streamCleanNames(p, renameMapChan, section, StreamCleanNamesTv)
		} else if section.Type == "movie" {
			streamCleanNames(p, renameMapChan, section, StreamCleanNamesMovies)
		}
	}
}

// wraps the clean tv and movies call into a single call
func GetAllCleanNames(p *plex.Plex) ([]RenameMap, error) {
	renamedMap := []RenameMap{}
	sections, err := p.GetLibraries()
	if err != nil {
		return renamedMap, err
	}

	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "show" {
			tvMap, err := GetCleanNamesTv(p, section)
			if err != nil {
				log.Println(err)
				return renamedMap, err
			}
			renamedMap = append(renamedMap, tvMap...)
		} else if section.Type == "movie" {
			movieMap, err := GetCleanNamesMovies(p, section)
			if err != nil {
				log.Println(err)
				return renamedMap, err
			}
			renamedMap = append(renamedMap, movieMap...)
		}
	}

	return renamedMap, nil
}

func getFirstMediaPart(metadata plex.Metadata) (plex.Part, error) {
	if len(metadata.Media) == 0 || len(metadata.Media[0].Part) == 0 {
		return plex.Part{}, errors.New("No media file found in container")

	}

	return metadata.Media[0].Part[0], nil
}

func matchPrefix(paths []string, file string) (string, error) {
	for _, path := range paths {
		if strings.Contains(file, path) {
			return path, nil
		}
	}

	return "", errors.New("File not in path")
}

func getLibraryLocations(directory plex.Directory) []string {
	locations := directory.Location

	var libraryLocations []string
	for _, location := range locations {
		path := location.Path
		libraryLocations = append(libraryLocations, path)
	}

	return libraryLocations
}

func cleanEpisodeName(prefixPath, title string, episode plex.Metadata) string {
	return fmt.Sprintf("%s/%s/Season %d/S%dE%d %s.%s", prefixPath, title, episode.ParentIndex, episode.ParentIndex, episode.Index, episode.Title, episode.Media[0].Part[0].Container)
}

func CrawlToEpisode(p *plex.Plex, libraryKey string, episodeChan chan plex.Metadata) {
	defer close(episodeChan)

	section, pErr := p.GetLibraryContent(libraryKey, "")
	if pErr != nil {
		log.Println(pErr)
		return
	}

	topLevelTVLibraries := section.MediaContainer.Metadata
	for _, tvShow := range topLevelTVLibraries {
		seasons, pErr := p.GetMetadataChildren(tvShow.RatingKey)
		if pErr != nil {
			log.Println(pErr)
			continue
		}
		for _, season := range seasons.MediaContainer.Metadata {
			episodesSection, pErr := p.GetEpisodes(season.RatingKey)
			if pErr != nil {
				log.Println(pErr)
				continue
			}

			for _, episode := range episodesSection.MediaContainer.Metadata {
				episodeChan <- episode
			}
		}
	}
}

func StreamCleanNamesTv(p *plex.Plex, sectionDirectory plex.Directory, renameMapChan chan RenameMap) {
	episodeChan := make(chan plex.Metadata, 100)
	tvPaths := getLibraryLocations(sectionDirectory)
	defer close(renameMapChan)

	go CrawlToEpisode(p, sectionDirectory.Key, episodeChan)
	for episode := range episodeChan {
		title := episode.GrandparentTitle
		part, err := getFirstMediaPart(episode)
		file := part.File

		libraryPrefix, err := matchPrefix(tvPaths, file)

		if err != nil {
			log.Println(err)
			continue
		}

		cleanName := cleanEpisodeName(libraryPrefix, title, episode)
		if cleanName == file {
			continue
		}
		renameMapChan <- RenameMap{Src: file, Dest: cleanName}
	}
}

func GetCleanNamesTv(p *plex.Plex, sectionDirectory plex.Directory) ([]RenameMap, error) {
	renameMapChan := make(chan RenameMap, 100)
	renamedMap := []RenameMap{}

	go StreamCleanNamesTv(p, sectionDirectory, renameMapChan)
	for rename := range renameMapChan {
		renamedMap = append(renamedMap, rename)
	}

	return renamedMap, nil
}

func CrawlMovies(p *plex.Plex, libraryKey string, movieChan chan plex.Metadata) {
	defer close(movieChan)

	// TODO: Get library content has all the metadata we want, use it plus a filter
	section, pErr := p.GetLibraryContent(libraryKey, "")
	if pErr != nil {
		log.Println(pErr)
		return
	}

	for _, movie := range section.MediaContainer.Metadata {
		movieChan <- movie
	}
}

func StreamCleanNamesMovies(p *plex.Plex, sectionDirectory plex.Directory, renameMapChan chan RenameMap) {
	movieChan := make(chan plex.Metadata, 100)
	moviePaths := getLibraryLocations(sectionDirectory)
	defer close(renameMapChan)

	go CrawlMovies(p, sectionDirectory.Key, movieChan)
	for movie := range movieChan {
		part, err := getFirstMediaPart(movie)
		file := part.File
		libraryPrefix, err := matchPrefix(moviePaths, file)

		if err != nil {
			log.Println(err)
			continue
		}

		cleanName := fmt.Sprintf("%s/%s.%s", libraryPrefix, movie.Title, part.Container)
		if cleanName == file {
			continue // already clean
		}
		renameMapChan <- RenameMap{Src: file, Dest: cleanName}
	}
}

func GetCleanNamesMovies(p *plex.Plex, sectionDirectory plex.Directory) ([]RenameMap, error) {
	renameMapChan := make(chan RenameMap, 100)
	renamedMap := []RenameMap{}

	go StreamCleanNamesMovies(p, sectionDirectory, renameMapChan)
	for rename := range renameMapChan {
		renamedMap = append(renamedMap, rename)
	}

	return renamedMap, nil
}
