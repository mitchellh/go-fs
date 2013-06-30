package fat

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode"
)

type MediaType uint8

// The standard value for "fixed", non-removable media, directly
// from the FAT specification.
const MediaFixed MediaType = 0xF8

type bootSectorCommon struct {
	OEMName string
	BytesPerSector uint16
	SectorsPerCluster uint8
	ReservedSectorCount uint16
	NumFATs uint8
	RootEntryCount uint16
	TotalSectors uint32
	Media MediaType
	SectorsPerFat uint32
	SectorsPerTrack uint16
	NumHeads uint16
}

func (b *bootSectorCommon) Bytes() ([]byte, error) {
	var sector [512]byte

	// BS_jmpBoot
	sector[0] = 0xEB
	sector[1] = 0x3C
	sector[2] = 0x90

	// BS_OEMName
	if len(b.OEMName) > 8 {
		return nil, errors.New("OEMName must be 8 bytes or less")
	}

	for i, r := range b.OEMName {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("'%s' in OEM name not a valid ASCII char. Must be ASCII.", r)
		}

		sector[0x3+i] = byte(r)
	}

	// BPB_BytsPerSec
	binary.LittleEndian.PutUint16(sector[11:13], b.BytesPerSector)

	// BPB_SecPerClus
	sector[13] = uint8(b.SectorsPerCluster)

	// BPB_RsvdSecCnt
	binary.LittleEndian.PutUint16(sector[14:16], b.ReservedSectorCount)

	// BPB_NumFATs
	sector[16] = b.NumFATs

	// BPB_RootEntCnt
	binary.LittleEndian.PutUint16(sector[17:19], b.RootEntryCount)

	// BPB_Media
	sector[21] = byte(b.Media)

	// BPB_SecPerTrk
	binary.LittleEndian.PutUint16(sector[24:26], b.SectorsPerTrack)

	// BPB_Numheads
	binary.LittleEndian.PutUint16(sector[26:28], b.NumHeads)

	// BPB_Hiddsec
	// sector[28:32] - it is always set to 0 because we don't partition drives yet.

	// Important signature of every FAT boot sector
	sector[510] = 0x55
	sector[511] = 0xAA

	return sector[:],nil
}

type BootSectorFat16 struct {
	bootSectorCommon

	DriveNumber uint8
	VolumeID uint32
	VolumeLabel string
	FileSystemTypeLabel string
}

func (b *BootSectorFat16) Bytes() ([]byte, error) {
	sector, err := b.bootSectorCommon.Bytes()
	if err != nil {
		return nil, err
	}

	// BPB_TotSec16 AND BPB_TotSec32
	if b.TotalSectors < 0x10000 {
		binary.LittleEndian.PutUint16(sector[19:21], uint16(b.TotalSectors))
	} else {
		binary.LittleEndian.PutUint32(sector[32:36], b.TotalSectors)
	}

	// BPB_FATSz16
	if b.SectorsPerFat > 0x10000 {
		return nil, fmt.Errorf("SectorsPerFat value too big for non-FAT32: %d", b.SectorsPerFat)
	}

	binary.LittleEndian.PutUint16(sector[22:24], uint16(b.SectorsPerFat))

	// BS_DrvNum
	sector[36] = b.DriveNumber

	// BS_BootSig
	sector[38] = 0x29

	// BS_VolID
	binary.LittleEndian.PutUint32(sector[39:43], b.VolumeID)

	// BS_VolLab
	if len(b.VolumeLabel) > 11 {
		return nil, errors.New("VolumeLabel must be 11 bytes or less")
	}

	for i, r := range b.VolumeLabel {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("'%s' in VolumeLabel not a valid ASCII char. Must be ASCII.", r)
		}

		sector[43+i] = byte(r)
	}

	// BS_FilSysType
	if len(b.FileSystemTypeLabel) > 8 {
		return nil, errors.New("FileSystemTypeLabel must be 8 bytes or less")
	}

	for i, r := range b.FileSystemTypeLabel {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("'%s' in FileSystemTypeLabel not a valid ASCII char. Must be ASCII.", r)
		}

		sector[54+i] = byte(r)
	}

	return sector, nil
}
