package fat

import (
	"fmt"
	"github.com/mitchellh/go-fs"
)

// Directory implements fs.Directory and is used to interface with
// a directory on a FAT filesystem.
type Directory struct {
	dirCluster *DirectoryCluster
	fat        *FAT
}

func (d *Directory) Entries() []fs.DirectoryEntry {
	for i, entry := range d.dirCluster.entries {
		if entry.deleted {
			continue
		}

		if entry.longName != "" {
			fmt.Printf("%d: %s (LONG)\n", i, entry.longName)
		} else {
			fmt.Printf("%d: %s\n", i, entry.name)
		}
	}
	return nil
}
