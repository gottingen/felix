package filesystem

import "testing"

func checkSizePath(t *testing.T, path string, size int64) {
	dir, err := testFS.Stat(path)
	if err != nil {
		t.Fatalf("Stat %q (looking for size %d): %s", path, size, err)
	}
	if dir.Size() != size {
		t.Errorf("Stat %q: size %d want %d", path, dir.Size(), size)
	}
}

func TestReadDir(t *testing.T) {
	testFS = &MemMapFs{}
	testFS.Mkdir("/i-am-a-dir", 0777)
	testFS.Create("/this_exists.go")
	dirname := "rumpelstilzchen"
	_, err := ReadDir(testFS, dirname)
	if err == nil {
		t.Fatalf("ReadDir %s: error expected, none found", dirname)
	}

	dirname = ".."
	list, err := ReadDir(testFS, dirname)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dirname, err)
	}

	foundFile := false
	foundSubDir := false
	for _, dir := range list {
		switch {
		case !dir.IsDir() && dir.Name() == "this_exists.go":
			foundFile = true
		case dir.IsDir() && dir.Name() == "i-am-a-dir":
			foundSubDir = true
		}
	}
	if !foundFile {
		t.Fatalf("ReadDir %s: this_exists.go file not found", dirname)
	}
	if !foundSubDir {
		t.Fatalf("ReadDir %s: i-am-a-dir directory not found", dirname)
	}
}


