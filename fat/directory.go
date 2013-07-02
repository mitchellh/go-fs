package fat

import (
	"fmt"
	"github.com/mitchellh/go-fs"
	"strings"
)

// Directory implements fs.Directory and is used to interface with
// a directory on a FAT filesystem.
type Directory struct {
	dirCluster *DirectoryCluster
	fat        *FAT
}

// DirectoryEntry implements fs.DirectoryEntry and represents a single
// file/folder within a directory in a FAT filesystem. Note that the
// underlying directory entry data structures on the disk may be more
// than one to accomodate for long filenames.
type DirectoryEntry struct {
	lfnEntries []*DirectoryClusterEntry
	entry *DirectoryClusterEntry

	name string
}

// DecodeDirectoryEntry takes a list of entries, decodes the next full
// DirectoryEntry, and returns the newly created entry, the remaining
// entries, and an error, if there was one.
func DecodeDirectoryEntry(entries []*DirectoryClusterEntry) (*DirectoryEntry, []*DirectoryClusterEntry, error) {
	var lfnEntries []*DirectoryClusterEntry
	var entry *DirectoryClusterEntry
	var name string

	// Skip all the deleted entries
	for len(entries) > 0 && entries[0].deleted {
		entries = entries[1:]
	}

	if len(entries) == 0 {
		return nil, entries, nil
	}

	// We have a long entry, so we have to traverse to the point where
	// we're done. Also, calculate out the name and such.
	if entries[0].IsLong() {
		lfnEntries := make([]*DirectoryClusterEntry, 0, 3)
		for entries[0].IsLong() {
			lfnEntries = append(lfnEntries, entries[0])
			entries = entries[1:]
		}

		var nameBytes []rune
		nameBytes = make([]rune, 13 * len(lfnEntries))
		for i := len(lfnEntries) - 1; i >= 0; i-- {
			for _, char := range lfnEntries[i].longName {
				nameBytes = append(nameBytes, char)
			}
		}

		name = string(nameBytes)
	}

	// Get the short entry
	entry = entries[0]
	entries = entries[1:]

	// If the short entry is deleted, ignore everything
	if entry.deleted {
		return nil, entries, nil
	}

	if name == "" {
		name = strings.TrimSpace(entry.name)
		ext := strings.TrimSpace(entry.ext)
		if ext != "" {
			name = fmt.Sprintf("%s.%s", name, ext)
		}
	}

	result := &DirectoryEntry{
		lfnEntries: lfnEntries,
		entry: entry,
		name: name,
	}

	return result, entries, nil
}

func (d *DirectoryEntry) Name() string {
	return d.name
}

func (d *Directory) Entries() []fs.DirectoryEntry {
	entries := d.dirCluster.entries
	result := make([]fs.DirectoryEntry, 0, len(entries) / 2)
	for len(entries) > 0 {
		var entry *DirectoryEntry
		entry, entries, _ = DecodeDirectoryEntry(entries)
		if entry != nil {
			result = append(result, entry)
		}
	}

	return result
}
