package fs

// Directory is an entry in a filesystem that stores files.
type Directory interface {
	Name() string
}
