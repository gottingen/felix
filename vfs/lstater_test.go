package vfs


import (
	"os"
	"path/filepath"
	"testing"
)

func TestLstatIfPossible(t *testing.T) {
	wd, _ := os.Getwd()
	defer func() {
		os.Chdir(wd)
	}()

	osFs := &OsFs{}

	workDir, err := TempDir(osFs, "", "felix-lstate")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		osFs.RemoveAll(workDir)
	}()

	memWorkDir := "/lstate"

	memFs := NewMemMapFs()
	overlayFs1 := &CopyOnWriteFs{base: osFs, layer: memFs}
	overlayFs2 := &CopyOnWriteFs{base: memFs, layer: osFs}
	overlayFsMemOnly := &CopyOnWriteFs{base: memFs, layer: NewMemMapFs()}
	basePathFs := &BasePathFs{source: osFs, path: workDir}
	basePathFsMem := &BasePathFs{source: memFs, path: memWorkDir}
	roFs := &ReadOnlyFs{source: osFs}
	roFsMem := &ReadOnlyFs{source: memFs}

	pathFileMem := filepath.Join(memWorkDir, "felixm.txt")

	WriteFile(osFs, filepath.Join(workDir, "felix.txt"), []byte("Hi, Felix!"), 0777)
	WriteFile(memFs, filepath.Join(pathFileMem), []byte("Hi, Felix!"), 0777)

	os.Chdir(workDir)
	if err := os.Symlink("felix.txt", "symfelix.txt"); err != nil {
		t.Fatal(err)
	}

	pathFile := filepath.Join(workDir, "felix.txt")
	pathSymlink := filepath.Join(workDir, "symfelix.txt")

	checkLstat := func(l Lstater, name string, shouldLstat bool) os.FileInfo {
		statFile, isLstat, err := l.LstatIfPossible(name)
		if err != nil {
			t.Fatalf("Lstat check failed: %s", err)
		}
		if isLstat != shouldLstat {
			t.Fatalf("Lstat status was %t for %s", isLstat, name)
		}
		return statFile
	}

	testLstat := func(l Lstater, pathFile, pathSymlink string) {
		shouldLstat := pathSymlink != ""
		statRegular := checkLstat(l, pathFile, shouldLstat)
		statSymlink := checkLstat(l, pathSymlink, shouldLstat)
		if statRegular == nil || statSymlink == nil {
			t.Fatal("got nil FileInfo")
		}

		symSym := statSymlink.Mode()&os.ModeSymlink == os.ModeSymlink
		if symSym == (pathSymlink == "") {
			t.Fatal("expected the FileInfo to describe the symlink")
		}

		_, _, err := l.LstatIfPossible("this-should-not-exist.txt")
		if err == nil || !os.IsNotExist(err) {
			t.Fatalf("expected file to not exist, got %s", err)
		}
	}

	testLstat(osFs, pathFile, pathSymlink)
	testLstat(overlayFs1, pathFile, pathSymlink)
	testLstat(overlayFs2, pathFile, pathSymlink)
	testLstat(basePathFs, "felix.txt", "symfelix.txt")
	testLstat(overlayFsMemOnly, pathFileMem, "")
	testLstat(basePathFsMem, "felixm.txt", "")
	testLstat(roFs, pathFile, pathSymlink)
	testLstat(roFsMem, pathFileMem, "")
}

