package vfs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var testRegistry map[Vfs][]string = make(map[Vfs][]string)

var testName = "test.txt"
var Fss = []Vfs{&MemMapFs{}, &OsFs{}}
var testFS = new(MemMapFs)

func RemoveAllTestFiles(t *testing.T) {
	for fs, list := range testRegistry {
		for _, path := range list {
			if err := fs.RemoveAll(path); err != nil {
				t.Error(fs.Name(), err)
			}
		}
	}
	testRegistry = make(map[Vfs][]string)
}

func TestDir(fs Vfs) string {
	name, err := TempDir(fs, "", "felix")
	if err != nil {
		panic(fmt.Sprint("unable to work with test dir", err))
	}
	testRegistry[fs] = append(testRegistry[fs], name)

	return name
}


func SetupTestDir(t *testing.T, fs Vfs) string {
	path := TestDir(fs)
	return SetupTestFiles(t, fs, path)
}

func SetupTestDirRoot(t *testing.T, fs Vfs) string {
	path := TestDir(fs)
	SetupTestFiles(t, fs, path)
	return path
}

func SetupTestDirReusePath(t *testing.T, fs Vfs, path string) string {
	testRegistry[fs] = append(testRegistry[fs], path)
	return SetupTestFiles(t, fs, path)
}

func SetupTestFiles(t *testing.T, fs Vfs, path string) string {
	testSubDir := filepath.Join(path, "more", "subdirectories", "for", "testing", "we")
	err := fs.MkdirAll(testSubDir, 0700)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}

	f, err := fs.Create(filepath.Join(testSubDir, "testfile1"))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("Testfile 1 content")
	f.Close()

	f, err = fs.Create(filepath.Join(testSubDir, "testfile2"))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("Testfile 2 content")
	f.Close()

	f, err = fs.Create(filepath.Join(testSubDir, "testfile3"))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("Testfile 3 content")
	f.Close()

	f, err = fs.Create(filepath.Join(testSubDir, "testfile4"))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("Testfile 4 content")
	f.Close()
	return testSubDir
}

