# Playlist Exporter
[![License](https://img.shields.io/github/license/tomdewildt/playlist-exporter)](https://github.com/tomdewildt/playlist-exporter/blob/master/LICENSE)

A simple tool to move files referenced in `*.m3u8` playlists into a directory local to those `*.m3u8` playlists.

# How To Run

Prerequisites:
* go version ```1.25.3``` or later

### Production

1. Run `go mod download` to download dependencies.
2. Run the following command to build the binary:
     * For MacOS: `GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Name=playlist-exporter -X main.Version=<version>" -o playlist-exporter ./cmd/playlist-exporter`
     * For Windows: `GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Name=playlist-exporter -X main.Version=<version>" -o playlist-exporter.exe ./cmd/playlist-exporter`

# References

[Go Docs](https://go.dev/doc/)

[M3U8 Docs](https://docs.fileformat.com/audio/m3u/)
