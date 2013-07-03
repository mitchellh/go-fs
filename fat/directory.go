package fat

import (
	"fmt"
	"github.com/mitchellh/go-fs"
	"strings"
)

// Directory implements fs.Directory and is used to interface with
// a directory on a FAT filesystem.
type Directory struct {
	device     fs.BlockDevice
	dirCluster *DirectoryCluster
	fat        *FAT
}

// DirectoryEntry implements fs.DirectoryEntry and represents a single
// file/folder within a directory in a FAT filesystem. Note that the
// underlying directory entry data structures on the disk may be more
// than one to accomodate for long filenames.
type DirectoryEntry struct {
	dir        *Directory
	lfnEntries []*DirectoryClusterEntry
	entry      *DirectoryClusterEntry

	name string
}

// DecodeDirectoryEntry takes a list of entries, decodes the next full
// DirectoryEntry, and returns the newly created entry, the remaining
// entries, and an error, if there was one.
func DecodeDirectoryEntry(d *Directory, entries []*DirectoryClusterEntry) (*DirectoryEntry, []*DirectoryClusterEntry, error) {
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
		nameBytes = make([]rune, 13*len(lfnEntries))
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
		dir:        d,
		lfnEntries: lfnEntries,
		entry:      entry,
		name:       name,
	}

	return result, entries, nil
}

func (d *DirectoryEntry) Dir() (fs.Directory, error) {
	if !d.IsDir() {
		panic("not a directory")
	}

	dirCluster, err := DecodeDirectoryCluster(
		d.entry.cluster, d.dir.device, d.dir.fat)
	if err != nil {
		return nil, err
	}

	result := &Directory{
		device:     d.dir.device,
		dirCluster: dirCluster,
		fat:        d.dir.fat,
	}

	return result, nil
}

func (d *DirectoryEntry) IsDir() bool {
	return (d.entry.attr & AttrDirectory) == AttrDirectory
}

func (d *DirectoryEntry) Name() string {
	return d.name
}

func (d *Directory) AddDirectory(name string) (fs.DirectoryEntry, error) {
	name = strings.TrimSpace(name)

	for _, entry := range d.Entries() {
		// TODO(mitchellh): case sensitivity? I think fat ISNT sensitive.
		if entry.Name() == name {
			return nil, fmt.Errorf("name already exists: %s", name)
		}
	}

	// TODO(mitchellh:
	// * make the short name
	// * allocate cluster space
	// * create the entry
	// * create the ., .. entries in the new directory
	return nil, nil
}

func (d *Directory) Entries() []fs.DirectoryEntry {
	entries := d.dirCluster.entries
	result := make([]fs.DirectoryEntry, 0, len(entries)/2)
	for len(entries) > 0 {
		var entry *DirectoryEntry
		entry, entries, _ = DecodeDirectoryEntry(d, entries)
		if entry != nil {
			result = append(result, entry)
		}
	}

	return result
}
