package fat

import "github.com/mitchellh/go-fs"

// SuperFloppyConfig is the configuration for various properties of
// a new super floppy formatted block device. Once this configuration is used
// to format a device, it must not be modified.
type SuperFloppyConfig struct {
	// The type of FAT filesystem to use.
	FATType FATType

	// The label of the drive. Defaults to "NONAME"
	Label string

	// The OEM name for the FAT filesystem. Defaults to "gofs" if not set.
	OEMName string
}

// Formats an fs.BlockDevice with the "super floppy" format according
// to the given configuration. The "super floppy" standard means that the
// device will be formatted so that it does not contain a partition table.
// Instead, the entire device holds a single FAT file system.
func FormatSuperFloppy(device fs.BlockDevice, config *SuperFloppyConfig) error {
	formatter := &superFloppyFormatter{
		config: config,
		device: device,
	}

	return formatter.format()
}

// An internal struct that helps maintain state and perform calculations
// during a single formatting pass.
type superFloppyFormatter struct {
	config *SuperFloppyConfig
	device fs.BlockDevice
}

func (f *superFloppyFormatter) format() error {
	return nil
}
