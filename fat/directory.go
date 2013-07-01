package fat

import (
	"errors"
	"github.com/mitchellh/go-fs"
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

// Directory implements fs.Directory and is used to interface with
// a directory on a FAT filesystem.
type Directory struct {
	dirCluster *DirectoryCluster
	fat        *FAT
}

// DirectoryCluster represents a cluster on the disk that contains
// entries/contents.
type DirectoryCluster struct {
	Entries []*DirectoryEntry
}

// DirectoryEntry is a single 32-byte entry that is part of the
// chain of entries in a directory cluster.
type DirectoryEntry struct {
	Name       string
	Attr       DirectoryAttr
	CreateTime time.Time
	AccessTime time.Time
}

// DecodeFAT16RootDirectory decodes the FAT16 root directory structure
// from the device.
func DecodeFAT16RootDirectoryCluster(device fs.BlockDevice, bs *BootSectorCommon) (*DirectoryCluster, error) {
	data := make([]byte, DirectoryEntrySize*bs.RootEntryCount)
	if _, err := device.ReadAt(data, int64(bs.RootDirOffset())); err != nil {
		return nil, err
	}

	entries := make([]*DirectoryEntry, 0, bs.RootEntryCount)
	for i := uint16(0); i < bs.RootEntryCount; i++ {
		offset := i * DirectoryEntrySize
		entryData := data[offset : offset+DirectoryEntrySize]
		if entryData[0] == 0 {
			// We're done if the first byte is nul
			break
		}

		entry, err := DecodeDirectoryEntry(entryData)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	result := &DirectoryCluster{
		Entries: entries,
	}

	return result, nil
}

// NewFat16RootDirectory creates a new DirectoryCluster that is meant only
// to be the root directory of a FAT12/FAT16 filesystem.
func NewFat16RootDirectoryCluster(bs *BootSectorCommon) (*DirectoryCluster, error) {
	if bs.RootEntryCount == 0 {
		return nil, errors.New("root entry count is 0 in boot sector")
	}

	result := &DirectoryCluster{
		Entries: make([]*DirectoryEntry, 0, bs.RootEntryCount),
	}

	return result, nil
}

// Bytes returns the on-disk byte data for this directory structure.
func (d *DirectoryCluster) Bytes() []byte {
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

// DecodeDirectoryEntry decodes a single directory entry in the
// Directory structure.
func DecodeDirectoryEntry(data []byte) (*DirectoryEntry, error) {
	return &DirectoryEntry{}, nil
}
