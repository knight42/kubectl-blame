package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlameLocalFile(t *testing.T) {
	testCases := map[string]struct {
		inputFile string
	}{
		"yaml": {
			inputFile: "deploy.yaml",
		},
		"json": {
			inputFile: "deploy.json",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			inputFile := filepath.Join("testdata", tc.inputFile)
			expectedFile := inputFile + ".txt"
			expected, err := os.ReadFile(expectedFile)
			r.NoError(err)
			var buf bytes.Buffer
			opts := &Options{
				inputFile:  inputFile,
				timeFormat: TimeFormatNone,
				out:        &buf,
			}
			err = opts.Run()
			r.NoError(err)
			r.Equal(string(expected), buf.String())
		})
	}
}
