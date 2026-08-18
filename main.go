package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"C"

	"github.com/dvcrn/go-rekordbox/rekordbox"
)
import "encoding/json"

// ExtendedContent adds label name to the rekordbox content
type ExtendedContent struct {
	*rekordbox.DjmdContent
	LabelName string `json:"label_name,omitempty"`
}

type Playlist struct {
	CombinedName string                  `json:"combined_name"`
	DJMdPlaylist *rekordbox.DjmdPlaylist `json:"dj_md_playlist,omitempty"`
	DJMdContents []*ExtendedContent      `json:"dj_md_contents,omitempty"`
}

func getRecursivePlaylistName(ctx context.Context, client *rekordbox.Client, playlist *rekordbox.DjmdPlaylist, nameSoFar string) string {
	// check if has a parent
	if playlist.ParentID.String() == "root" {
		return nameSoFar
	}

	// get parent
	parent, err := client.DjmdPlaylistByID(ctx, playlist.ParentID)
	if err != nil {
		panic(err)
	}

	name := fmt.Sprintf("%s - %s", parent.Name.String(), nameSoFar)
	return getRecursivePlaylistName(ctx, client, parent, name)
}

//export getPlaylists
func getPlaylists() *C.char {
	ctx := context.Background()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	optionsFilePath := filepath.Join(homeDir, "/Library/Application Support/Pioneer/rekordboxAgent/storage/", "options.json")

	// Files and paths
	client, err := rekordbox.NewClient(optionsFilePath)
	if err != nil {
		panic(err)
	}

	defer client.Close()

	playlists, err := client.AllDjmdPlaylist(ctx)
	if err != nil {
		panic(err)
	}

	parsedPlaylists := []*Playlist{}
	for _, playlist := range playlists {
		pl := &Playlist{}

		playlistSongs, err := client.DjmdSongPlaylistByPlaylistID(ctx, playlist.ID)
		if err != nil {
			panic(err)
		}

		if len(playlistSongs) == 0 {
			continue
		}

		pl.DJMdPlaylist = playlist
		pl.CombinedName = getRecursivePlaylistName(ctx, client, playlist, playlist.Name.String())

		for _, playlistSong := range playlistSongs {
			content, err := client.DjmdContentByID(ctx, playlistSong.ContentID)
			if err != nil {
				panic(err)
			}

			// Fetch label name if LabelID exists
			labelName := ""
			if content.LabelID.Valid() {
				label, err := client.DjmdLabelByID(ctx, content.LabelID)
				if err == nil && label != nil && label.Name.Valid() {
					labelName = label.Name.String()
				}
			}

			// Create extended content with label name
			extended := &ExtendedContent{
				DjmdContent: content,
				LabelName:   labelName,
			}

			pl.DJMdContents = append(pl.DJMdContents, extended)
		}

		parsedPlaylists = append(parsedPlaylists, pl)
	}

	// marshal playlists to json
	b, err := json.Marshal(parsedPlaylists)
	if err != nil {
		panic(err)
	}

	return C.CString(string(b))
}

func main() {
}
