package fs

// Directory is an entry in a filesystem that stores files.
type Directory interface {
	Entries() []DirectoryEntry
	AddDirectory(name string) (DirectoryEntry, error)
}

// DirectoryEntry represents a single entry within a directory,
// which can be either another Directory or a File.
type DirectoryEntry interface {
	Name() string
	IsDir() bool
	Dir() (Directory, error)
}
