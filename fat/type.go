package fat

// FATType is a simple enum of the available FAT filesystem types.
type FATType uint8

const (
	FAT12 FATType = iota
	FAT16
	FAT32
)
