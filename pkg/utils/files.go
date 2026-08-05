package utils

import (
	"os"
)

func FileExists(path string) bool {

	_, err := os.Stat(path)

	return err == nil

}

// WriteFileAtomic writes data to a temp file in the same directory as path
// and renames it into place, so readers never observe a partially written
// file (used for certificates and private keys on both the manager and the
// agent).
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}
