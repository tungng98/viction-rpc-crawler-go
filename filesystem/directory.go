package filesystem

import "os"

func CreateDirectoryRecursive(dPath string) error {
	return os.MkdirAll(dPath, 0755)
}
