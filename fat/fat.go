package fat

// FAT is the actual file allocation table data structure that is
// stored on disk to describe the various clusters on the disk.
type FAT struct {
	bs      *BootSectorCommon
	fatType FATType
	entries []uint32
}

// NewFAT creates a new FAT data structure, properly initialized.
func NewFAT(bs *BootSectorCommon) (*FAT, error) {
	// Determine what FAT-type we're using. This is done using a calculation
	// pulled directly from the specification. This is the ONLY way to PROPERLY
	// calculate this.
	var rootDirSectors uint32
	rootDirSectors = (uint32(bs.RootEntryCount) * 32) + (uint32(bs.BytesPerSector) - 1)
	rootDirSectors /= uint32(bs.BytesPerSector)
	dataSectors := bs.SectorsPerFat * uint32(bs.NumFATs)
	dataSectors += uint32(bs.ReservedSectorCount)
	dataSectors += rootDirSectors
	dataSectors = bs.TotalSectors - dataSectors
	countClusters := dataSectors / uint32(bs.SectorsPerCluster)

	var fatType FATType
	switch {
	case countClusters < 4085:
		fatType = FAT12
	case countClusters < 65525:
		fatType = FAT16
	default:
		fatType = FAT32
	}

	// Determine the number of entries that'll go in the FAT.
	var entryCount uint32 = bs.SectorsPerFat * uint32(bs.BytesPerSector)
	switch fatType {
	case FAT12:
		entryCount = uint32((uint64(entryCount) * 8) / 12)
	case FAT16:
		entryCount /= 2
	case FAT32:
		entryCount /= 4
	default:
		panic("impossible fat type")
	}

	result := &FAT{
		bs:      bs,
		fatType: fatType,
		entries: make([]uint32, entryCount),
	}

	// Set the initial two entries according to spec
	result.entries[0] = (uint32(bs.Media) & 0xFF) |
		(0xFFFFFF00 & result.entryMask())
	result.entries[1] = 0xFFFFFFFF & result.entryMask()

	return result, nil
}

// Bytes returns the raw bytes for the FAT that should be written to
// the block device.
func (f *FAT) Bytes() []byte {
	result := make([]byte, f.bs.SectorsPerFat*uint32(f.bs.BytesPerSector))

	for i, entry := range f.entries {
		switch f.fatType {
		case FAT12:
			f.writeEntry12(result, i, entry)
		case FAT16:
			f.writeEntry16(result, i, entry)
		default:
			f.writeEntry32(result, i, entry)
		}
	}

	return result
}

func (f *FAT) entryMask() uint32 {
	switch f.fatType {
	case FAT12:
		return 0x0FFF
	case FAT16:
		return 0xFFFF
	default:
		return 0x0FFFFFFF
	}
}

func (f *FAT) writeEntry12(data []byte, idx int, entry uint32) {
	idx += idx / 2

	if idx%2 == 0 {
		// Cluster number is EVEN
		data[idx] = byte(entry & 0xFF)
		data[idx+1] = byte((entry >> 8) & 0x0F)
	} else {
		// Cluster number is ODD
		data[idx] |= byte((entry & 0x0F) << 4)
		data[idx+1] = byte((entry >> 4) & 0xFF)
	}
}

func (f *FAT) writeEntry16(data []byte, idx int, entry uint32) {
	idx <<= 1
	data[idx] = byte(entry & 0xFF)
	data[idx+1] = byte((entry >> 8) & 0xFF)
}

func (f *FAT) writeEntry32(data []byte, idx int, entry uint32) {
	idx <<= 2
	data[idx] = byte(entry & 0xFF)
	data[idx+1] = byte((entry >> 8) & 0xFF)
	data[idx+2] = byte((entry >> 16) & 0xFF)
	data[idx+3] = byte((entry >> 24) & 0xFF)
}
