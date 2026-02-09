package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlameAutoStdin(t *testing.T) {
	r := require.New(t)

	// Read test input and expected output
	inputData, err := os.ReadFile(filepath.Join("testdata", "deploy.yaml"))
	r.NoError(err)
	expected, err := os.ReadFile(filepath.Join("testdata", "deploy.yaml.txt"))
	r.NoError(err)

	// Create a pipe to simulate piped stdin
	pr, pw, err := os.Pipe()
	r.NoError(err)

	_, err = pw.Write(inputData)
	r.NoError(err)
	pw.Close()

	// Swap os.Stdin with the pipe
	oldStdin := os.Stdin
	os.Stdin = pr
	t.Cleanup(func() { os.Stdin = oldStdin })

	var buf bytes.Buffer
	opts := &Options{
		inputFile:  "auto",
		timeFormat: TimeFormatNone,
		out:        &buf,
	}
	err = opts.Run()
	r.NoError(err)
	r.Equal(string(expected), buf.String())
}

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
		"multi-manager-list-item": {
			inputFile: "multi-manager-list-item.yaml",
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
