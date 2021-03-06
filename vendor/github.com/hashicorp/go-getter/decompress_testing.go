package getter

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// TestDecompressCase is a single test case for testing decompressors
type TestDecompressCase struct {
	Input   string   // Input is the complete path to the input file
	Dir     bool     // Dir is whether or not we're testing directory mode
	Err     bool     // Err is whether we expect an error or not
	DirList []string // DirList is the list of files for Dir mode
	FileMD5 string   // FileMD5 is the expected MD5 for a single file
}

// TestDecompressor is a helper function for testing generic decompressors.
func TestDecompressor(t *testing.T, d Decompressor, cases []TestDecompressCase) {
	for _, tc := range cases {
		t.Logf("Testing: %s", tc.Input)

		// Temporary dir to store stuff
		td, err := ioutil.TempDir("", "getter")
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		// Destination is always joining result so that we have a new path
		dst := filepath.Join(td, "subdir", "result")

		// We use a function so defers work
		func() {
			defer os.RemoveAll(td)

			// Decompress
			err := d.Decompress(dst, tc.Input, tc.Dir)
			if (err != nil) != tc.Err {
				t.Fatalf("err %s: %s", tc.Input, err)
			}
			if tc.Err {
				return
			}

			// If it isn't a directory, then check for a single file
			if !tc.Dir {
				fi, err := os.Stat(dst)
				if err != nil {
					t.Fatalf("err %s: %s", tc.Input, err)
				}
				if fi.IsDir() {
					t.Fatalf("err %s: expected file, got directory", tc.Input)
				}
				if tc.FileMD5 != "" {
					actual := testMD5(t, dst)
					expected := tc.FileMD5
					if actual != expected {
						t.Fatalf("err %s: expected MD5 %s, got %s", tc.Input, expected, actual)
					}
				}

				return
			}

			// Directory, check for the correct contents
			actual := testListDir(t, dst)
			expected := tc.DirList
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("bad %s\n\n%#v\n\n%#v", tc.Input, actual, expected)
			}
		}()
	}
}

func testListDir(t *testing.T, path string) []string {
	var result []string
	err := filepath.Walk(path, func(sub string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		sub = strings.TrimPrefix(sub, path)
		if sub == "" {
			return nil
		}
		sub = sub[1:] // Trim the leading path sep.

		// If it is a dir, add trailing sep
		if info.IsDir() {
			sub += "/"
		}

		result = append(result, sub)
		return nil
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	sort.Strings(result)
	return result
}

func testMD5(t *testing.T, path string) string {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result := h.Sum(nil)
	return hex.EncodeToString(result)
}
