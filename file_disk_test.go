package fs

import (
	"os"
	"testing"
)

func TestFileDisk_NewDiskFile_Dir(t *testing.T) {
	f, err := os.Open(os.TempDir())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	_, err = NewFileDisk(f)
	if err == nil {
		t.Fatal("should error if directory")
	}
}
