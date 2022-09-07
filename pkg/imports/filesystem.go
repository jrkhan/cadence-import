package imports

import (
	"fmt"
	"io/fs"
	"os"
)

type readerWriter struct {
}

func (r readerWriter) ReadFile(name string) ([]byte, error) {
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

func (r readerWriter) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
