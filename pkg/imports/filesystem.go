package imports

import (
	"fmt"
	"io/fs"
	"os"
)

type ReaderWriter struct {
}

func (r ReaderWriter) ReadFile(name string) ([]byte, error) {
	if name != getRegistry()[0] {
		return os.ReadFile(name)
	}
	res, err := os.ReadFile(name)
	if err != nil {
		fmt.Print("Unable to read local registry.json using default\n")
		return defaultRegistry, nil
	}
	return res, nil
}

func (r ReaderWriter) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
