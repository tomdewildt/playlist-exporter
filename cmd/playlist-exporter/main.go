package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jacoblockett/sanitizefilename"

	"github.com/tomdewildt/playlist-exporter/internal/io"
)

// Name of the binary
var Name string

// Version of the binary
var Version string

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Move files referenced in '*.m3u8' playlists into a directory local to those '*.m3u8' playlists.

Usage: %s [source] [target]

Arguments:
  source: The directory containing the '*.m3u8' playlists.
  target: The directory local to the '*.m3u8' playlists where the audio files should be moved.`, Name)
	}
	flag.Parse()

	// Extract the arguments
	if len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(1)
	}
	source := filepath.Clean(strings.TrimSpace(flag.Arg(0)))
	target := filepath.Clean(strings.TrimSpace(flag.Arg(1)))

	// Validate arguments
	if info, err := os.Stat(source); os.IsNotExist(err) || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s: %s: directory does not exist\n", Name, source)
		os.Exit(1)
	}
	if info, err := os.Stat(target); os.IsNotExist(err) || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s: %s: directory does not exist\n", Name, target)
		os.Exit(1)
	}
	if source == target {
		fmt.Fprintf(os.Stderr, "%s: source and target directories are the same\n", Name)
		os.Exit(1)
	}

	// Collect playlists
	fmt.Printf("Finding playlists in: '%s'\n\n", source)
	playlists, err := findPlaylists(source, target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", Name, err)
		os.Exit(1)
	}

	numPlaylists := len(playlists)
	numFilesCopied := 0
	numFilesExisted := 0
	numFilesDeleted := 0

	// Loop through playlists
	for _, playlist := range playlists {
		fmt.Printf("Processing playlist: '%s'\n", playlist)

		// Ensure directory exists
		name := filepath.Clean(strings.Replace(strings.Replace(playlist, ".m3u8", "", 1), source, "", 1))
		dir := filepath.Join(target, name)
		if err := io.EnsureDirExists(dir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", Name, err)
			os.Exit(1)
		}

		// Load current files
		currentFiles, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", Name, err)
			os.Exit(1)
		}
		newFiles := []string{}

		// Load current playlist
		currentPlaylist, err := loadPlaylist(playlist)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", Name, err)
			os.Exit(1)
		}
		newPlaylist := []string{}

		// Loop through playlist
		trackIndex := 1
		for i := 0; i < len(currentPlaylist); i++ {
			line := currentPlaylist[i]

			// Handle metadata lines
			if strings.HasPrefix(line, "#EXTM3U") || strings.TrimSpace(line) == "" {
				newPlaylist = append(newPlaylist, line)
				continue
			}

			// Handle track lines
			if strings.HasPrefix(line, "#EXTINF") {
				currentPath := filepath.Clean(strings.TrimSpace(currentPlaylist[i+1]))
				currentExt := filepath.Ext(currentPath)

				// Parse metadata
				artist, track, err := parseMetadata(line)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: metadata parsing error: %v\n", Name, err)
					os.Exit(1)
				}

				// Create new filename
				newFilename := fmt.Sprintf("%03d %s - %s%s", trackIndex, track, artist, currentExt)
				newFilename = sanitizefilename.Sanitize(newFilename)
				newPath := filepath.Clean(strings.TrimSpace(filepath.Join(dir, newFilename)))

				// Add file to playlist
				newPlaylist = append(newPlaylist, line)
				newPlaylist = append(newPlaylist, newPath)
				newFiles = append(newFiles, newPath)

				// Increment track and line index
				trackIndex++
				i++ // Skip the next line as it is the file path

				// Check if file exists
				if _, err := os.Stat(newPath); err == nil {
					numFilesExisted++
					continue
				}

				// Copy file
				numFilesCopied++
				if err := io.CopyFile(currentPath, newPath); err != nil {
					fmt.Fprintf(os.Stderr, "%s: failed to copy file '%s' to '%s': %v\n", Name, currentPath, newPath, err)
					os.Exit(1)
				}
			}
		}

		// Remove files no longer present
		for _, file := range currentFiles {
			if file.IsDir() {
				continue
			}
			path := filepath.Join(dir, file.Name())
			if !slices.Contains[[]string, string](newFiles, path) {
				numFilesDeleted++
				if err := os.Remove(path); err != nil {
					fmt.Fprintf(os.Stderr, "%s: failed to delete file '%s': %v\n", Name, path, err)
					os.Exit(1)
				}
			}
		}

		// Write new playlist
		if err := writePlaylist(playlist, newPlaylist); err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to write playlist '%s': %v\n", Name, playlist, err)
			os.Exit(1)
		}
	}

	// Print results
	fmt.Print("\nProcessing complete:\n")
	fmt.Printf("%d playlists processed\n", numPlaylists)
	fmt.Printf("%d files copied\n", numFilesCopied)
	fmt.Printf("%d files already existed\n", numFilesExisted)
	fmt.Printf("%d files deleted\n", numFilesDeleted)
}

func parseMetadata(metadata string) (artist string, track string, err error) {
	index := strings.Index(metadata, ",")
	if index == -1 {
		err = fmt.Errorf("invalid metadata format: no comma found")
		return
	}
	detail := metadata[index+1:]
	artistTrack := strings.SplitN(detail, " - ", 2)
	if len(artistTrack) != 2 {
		err = fmt.Errorf("invalid artist-track format")
		return
	}
	artist = strings.TrimSpace(artistTrack[0])
	track = strings.TrimSpace(artistTrack[1])
	return
}

func findPlaylists(source string, target string) ([]string, error) {
	playlists := []string{}
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		// Handle errors
		if err != nil {
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			return err
		}

		// Skip target
		if info.IsDir() && path == target {
			return filepath.SkipDir
		}

		// Skip hidden
		if io.IsHidden(info, path) {
			if info.IsDir() {
				return filepath.SkipDir // Skip hidden directories altogether
			}
			return nil // Skip hidden files
		}

		// Collect playlists
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".m3u8") {
			playlists = append(playlists, path)
		}

		return nil
	})
	return playlists, err
}

func loadPlaylist(playlist string) ([]string, error) {
	data, err := os.ReadFile(playlist)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func writePlaylist(playlist string, data []string) error {
	return os.WriteFile(playlist, []byte(strings.Join(data, "\n")), 0644)
}
