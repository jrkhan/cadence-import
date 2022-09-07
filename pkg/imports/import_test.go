package imports

import (
	_ "embed"
	"errors"
	"io/fs"
	"testing"

	"github.com/onflow/cadence/runtime/parser"
	"github.com/stretchr/testify/assert"
)

//go:embed compiled_registry.json
var testRegistry []byte

//go:embed testdata/example_flow.json
var testFlow []byte

type mockReaderWriter struct {
	files map[string][]byte
}

func (r mockReaderWriter) ReadFile(name string) ([]byte, error) {
	val, has := r.files[name]
	if !has {
		return nil, errors.New("file not found in mock file system")
	}
	return val, nil
}

func (r mockReaderWriter) WriteFile(name string, data []byte, perm fs.FileMode) error {
	r.files[name] = data
	return nil
}

func TestGetImport(t *testing.T) {
	files := map[string][]byte{
		"./registry.json": testRegistry,
		"./flow.json":     testFlow,
	}
	err := GetImport(mockReaderWriter{
		files: files,
	}, "testnet", "TopShot")
	assert.NoError(t, err)

	t.Run("should have expected dependencies added", func(t *testing.T) {
		// should be safe to assume TopShot will always require several standard contracts
		// if these dependencies change, consider refactor this to use a mock network
		var expected = []string{
			"./imports/FungibleToken.cdc",
			"./imports/NonFungibleToken.cdc",
			"./imports/TopShot.cdc",
		}
		for _, expectedFile := range expected {
			assert.Contains(t, files, expectedFile)
		}
	})

	t.Run("should have mapped imports to point to local files", func(t *testing.T) {
		var src = files["./imports/TopShot.cdc"]
		p, err := parser.ParseProgram(string(src), nil)
		assert.NoError(t, err)
		for _, dec := range p.ImportDeclarations() {
			for _, id := range dec.Identifiers {
				importName := id.String()
				if importName == "NonFungibleToken" {
					assert.Equal(t, "NonFungibleToken.cdc", dec.Location.String())
				}
			}
		}
	})
}
