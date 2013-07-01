package fat

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mitchellh/go-fs"
	"time"
	"unicode/utf16"
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

func (d *Directory) Entries() []fs.DirectoryEntry {
	for i, entry := range d.dirCluster.entries {
		if entry.longName != "" {
			fmt.Printf("%d: %s (LONG)\n", i, entry.longName)
		} else {
			fmt.Printf("%d: %s\n", i, entry.name)
		}
	}
	return nil
}

// DirectoryCluster represents a cluster on the disk that contains
// entries/contents.
type DirectoryCluster struct {
	entries []*DirectoryEntry
}

// DirectoryEntry is a single 32-byte entry that is part of the
// chain of entries in a directory cluster.
type DirectoryEntry struct {
	name       string
	ext        string
	attr       DirectoryAttr
	createTime time.Time
	accessTime time.Time
	writeTime  time.Time
	cluster    uint32
	fileSize   uint32

	longOrd      uint8
	longName     string
	longChecksum uint8
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
		entry, err := DecodeDirectoryEntry(entryData)
		if err != nil {
			return nil, err
		}

		if entry == nil {
			// End of the chain of entries
			break
		}

		entries = append(entries, entry)
	}

	result := &DirectoryCluster{
		entries: entries,
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
		entries: make([]*DirectoryEntry, 0, bs.RootEntryCount),
	}

	return result, nil
}

// Bytes returns the on-disk byte data for this directory structure.
func (d *DirectoryCluster) Bytes() []byte {
	result := make([]byte, cap(d.entries)*DirectoryEntrySize)

	for i, entry := range d.entries {
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
	if data[0] == 0 {
		return nil, nil
	}

	var result DirectoryEntry

	// Do the attributes so we can determine if we're dealing with long names
	result.attr = DirectoryAttr(data[11])
	if (result.attr & AttrLongName) == AttrLongName {
		result.longOrd = data[0]

		chars := make([]uint16, 13)
		for i := 0; i < 5; i++ {
			offset := 1 + (i * 2)
			chars[i] = binary.LittleEndian.Uint16(data[offset : offset+2])
		}

		for i := 0; i < 6; i++ {
			offset := 14 + (i * 2)
			chars[i+5] = binary.LittleEndian.Uint16(data[offset : offset+2])
		}

		for i := 0; i < 2; i++ {
			offset := 28 + (i * 2)
			chars[i+11] = binary.LittleEndian.Uint16(data[offset : offset+2])
		}

		result.longName = string(utf16.Decode(chars))
		result.longChecksum = data[13]
	} else {
		if data[0] == 0xE5 {
			return nil, nil
		}

		// Basic attributes
		if data[0] == 0x05 {
			data[0] = 0xE5
		}

		result.name = string(data[0:8])
		result.ext = string(data[8:11])

		// Creation time
		createTimeTenths := data[13]
		createTimeWord := binary.LittleEndian.Uint16(data[14:16])
		createDateWord := binary.LittleEndian.Uint16(data[16:18])
		result.createTime = decodeDOSTime(createDateWord, createTimeWord, createTimeTenths)

		// Access time
		accessDateWord := binary.LittleEndian.Uint16(data[18:20])
		result.accessTime = decodeDOSTime(accessDateWord, 0, 0)

		// Write time
		writeTimeWord := binary.LittleEndian.Uint16(data[22:24])
		writeDateWord := binary.LittleEndian.Uint16(data[24:26])
		result.writeTime = decodeDOSTime(writeDateWord, writeTimeWord, 0)

		// Cluster
		result.cluster = uint32(binary.LittleEndian.Uint16(data[20:22]))
		result.cluster <<= 4
		result.cluster |= uint32(binary.LittleEndian.Uint16(data[26:28]))

		// File size
		result.fileSize = binary.LittleEndian.Uint32(data[28:32])
	}

	return &result, nil
}

func decodeDOSTime(date, dosTime uint16, tenths uint8) time.Time {
	return time.Date(
		1980+int(date>>9),
		time.Month((date>>5)&0x0F),
		int(date&0x1F),
		int(dosTime>>11),
		int((dosTime>>5)&0x3F),
		int((dosTime&0x1F)*2),
		int(tenths)*10*int(time.Millisecond),
		time.Local)
}
