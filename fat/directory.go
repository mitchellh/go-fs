package fat

import (
	"errors"
	"time"
)

type DirectoryAttr uint8

const (
	AttrReadOnly  DirectoryAttr = 0x01
	AttrHidden                  = 0x02
	AttrSystem                  = 0x04
	AttrVolumeId                = 0x08
	AttrDirectory               = 0x10
	AttrArchive                 = 0x20
	AttrLongName                = AttrReadOnly | AttrHidden | AttrSystem | AttrVolumeId
)

// The size in bytes of a single directory entry.
const DirectoryEntrySize = 32

// Directory represents a single directory that can contain many
// entries.
type Directory struct {
	Entries []DirectoryEntry
}

// DirectoryEntry is a single 32-byte entry that is part of the
// chain of entries in a directory cluster.
type DirectoryEntry struct {
	Name       string
	Attr       DirectoryAttr
	CreateTime time.Time
	AccessTime time.Time
}

// NewFat16RootDirectory creates a new Directory that is meant only
// to be the root directory of a FAT12/FAT16 filesystem.
func NewFat16RootDirectory(bs *BootSectorCommon) (*Directory, error) {
	if bs.RootEntryCount == 0 {
		return nil, errors.New("root entry count is 0 in boot sector")
	}

	result := &Directory{
		Entries: make([]DirectoryEntry, 0, bs.RootEntryCount),
	}

	return result, nil
}

// Bytes returns the on-disk byte data for this directory structure.
func (d *Directory) Bytes() []byte {
	result := make([]byte, cap(d.Entries)*DirectoryEntrySize)

	for i, entry := range d.Entries {
		offset := i * DirectoryEntrySize
		entryBytes := entry.Bytes()
		copy(result[offset:offset+DirectoryEntrySize], entryBytes)
	}

	return result
}

// Bytes returns the on-disk byte data for this directory entry.
func (d *DirectoryEntry) Bytes() []byte {
	var result [DirectoryEntrySize]byte
	return result[:]
}
