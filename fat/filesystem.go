package fat

import (
	"github.com/mitchellh/go-fs"
)

// FileSystem is the implementation of fs.FileSystem that can read a
// FAT filesystem.
type FileSystem struct {
	bs *BootSectorCommon
	fat *FAT
}

// New returns a new FileSystem for accessing a previously created
// FAT filesystem.
func New(device fs.BlockDevice) (*FileSystem, error) {
	bs, err := DecodeBootSector(device)
	if err != nil {
		return nil, err
	}

	fat, err := DecodeFAT(device, bs, 0)
	if err != nil {
		return nil, err
	}

	result := &FileSystem{
		bs: bs,
		fat: fat,
	}

	return result, nil
}

func (f *FileSystem) RootDir() (fs.Directory, error) {
	return nil, nil
}
