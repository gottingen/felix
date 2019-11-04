package felix

import (
	"github.com/gottingen/felix/filesystem"
	"testing"
)

var testFS = new(filesystem.MemMapFs)

func checkSizePath(t *testing.T, path string, size int64) {
	dir, err := testFS.Stat(path)
	if err != nil {
		t.Fatalf("Stat %q (looking for size %d): %s", path, size, err)
	}
	if dir.Size() != size {
		t.Errorf("Stat %q: size %d want %d", path, dir.Size(), size)
	}
}

func TestReadFile(t *testing.T) {
	testFS = &filesystem.MemMapFs{}
	fsutil := &Felix{Fs: testFS}

	testFS.Create("this_exists.go")
	filename := "rumpelstilzchen"
	contents, err := fsutil.ReadFile(filename)
	if err == nil {
		t.Fatalf("ReadFile %s: error expected, none found", filename)
	}

	filename = "this_exists.go"
	contents, err = fsutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", filename, err)
	}

	checkSizePath(t, filename, int64(len(contents)))
}

func TestWriteFile(t *testing.T) {
	testFS = &filesystem.MemMapFs{}
	fsutil := &Felix{Fs: testFS}
	f, err := fsutil.TempFile("", "ioutil-test")
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	data := "Programming today is a race between software engineers striving to " +
		"build bigger and better idiot-proof programs, and the Universe trying " +
		"to produce bigger and better idiots. So far, the Universe is winning."

	if err := fsutil.WriteFile(filename, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", filename, err)
	}

	contents, err := fsutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", filename, err)
	}

	if string(contents) != data {
		t.Fatalf("contents = %q\nexpected = %q", string(contents), data)
	}

	// cleanup
	f.Close()
	testFS.Remove(filename) // ignore error
}



