package vfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func tmpFile(fs Vfs) File {
	x, err := TempFile(fs, "", "felix")

	if err != nil {
		panic(fmt.Sprint("unable to work with temp file", err))
	}

	testRegistry[fs] = append(testRegistry[fs], x.Name())

	return x
}

//Read with length 0 should not return EOF.
func TestRead0(t *testing.T) {
	for _, fs := range Fss {
		f := tmpFile(fs)
		defer f.Close()
		f.WriteString("Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")

		var b []byte
		// b := make([]byte, 0)
		n, err := f.Read(b)
		if n != 0 || err != nil {
			t.Errorf("%v: Read(0) = %d, %v, want 0, nil", fs.Name(), n, err)
		}
		f.Seek(0, 0)
		b = make([]byte, 100)
		n, err = f.Read(b)
		if n <= 0 || err != nil {
			t.Errorf("%v: Read(100) = %d, %v, want >0, nil", fs.Name(), n, err)
		}
	}
}

func TestOpenFile(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		tmp := TestDir(fs)
		path := filepath.Join(tmp, testName)

		f, err := fs.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			t.Error(fs.Name(), "OpenFile (O_CREATE) failed:", err)
			continue
		}
		io.WriteString(f, "initial")
		f.Close()

		f, err = fs.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			t.Error(fs.Name(), "OpenFile (O_APPEND) failed:", err)
			continue
		}
		io.WriteString(f, "|append")
		f.Close()

		f, err = fs.OpenFile(path, os.O_RDONLY, 0600)
		contents, _ := ioutil.ReadAll(f)
		expectedContents := "initial|append"
		if string(contents) != expectedContents {
			t.Errorf("%v: appending, expected '%v', got: '%v'", fs.Name(), expectedContents, string(contents))
		}
		f.Close()

		f, err = fs.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			t.Error(fs.Name(), "OpenFile (O_TRUNC) failed:", err)
			continue
		}
		contents, _ = ioutil.ReadAll(f)
		if string(contents) != "" {
			t.Errorf("%v: expected truncated file, got: '%v'", fs.Name(), string(contents))
		}
		f.Close()
	}
}

func TestCreate(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		tmp := TestDir(fs)
		path := filepath.Join(tmp, testName)

		f, err := fs.Create(path)
		if err != nil {
			t.Error(fs.Name(), "Create failed:", err)
			f.Close()
			continue
		}
		io.WriteString(f, "initial")
		f.Close()

		f, err = fs.Create(path)
		if err != nil {
			t.Error(fs.Name(), "Create failed:", err)
			f.Close()
			continue
		}
		secondContent := "second create"
		io.WriteString(f, secondContent)
		f.Close()

		f, err = fs.Open(path)
		if err != nil {
			t.Error(fs.Name(), "Open failed:", err)
			f.Close()
			continue
		}
		buf, err := ReadAll(f)
		if err != nil {
			t.Error(fs.Name(), "ReadAll failed:", err)
			f.Close()
			continue
		}
		if string(buf) != secondContent {
			t.Error(fs.Name(), "Content should be", "\""+secondContent+"\" but is \""+string(buf)+"\"")
			f.Close()
			continue
		}
		f.Close()
	}
}

func TestMemFileRead(t *testing.T) {
	f := tmpFile(new(MemMapFs))
	// f := MemFileCreate("testfile")
	f.WriteString("abcd")
	f.Seek(0, 0)
	b := make([]byte, 8)
	n, err := f.Read(b)
	if n != 4 {
		t.Errorf("didn't read all bytes: %v %v %v", n, err, b)
	}
	if err != nil {
		t.Errorf("err is not nil: %v %v %v", n, err, b)
	}
	n, err = f.Read(b)
	if n != 0 {
		t.Errorf("read more bytes: %v %v %v", n, err, b)
	}
	if err != io.EOF {
		t.Errorf("error is not EOF: %v %v %v", n, err, b)
	}
}

func TestRename(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		tDir := TestDir(fs)
		from := filepath.Join(tDir, "/renamefrom")
		to := filepath.Join(tDir, "/renameto")
		exists := filepath.Join(tDir, "/renameexists")
		file, err := fs.Create(from)
		if err != nil {
			t.Fatalf("%s: open %q failed: %v", fs.Name(), to, err)
		}
		if err = file.Close(); err != nil {
			t.Errorf("%s: close %q failed: %v", fs.Name(), to, err)
		}
		file, err = fs.Create(exists)
		if err != nil {
			t.Fatalf("%s: open %q failed: %v", fs.Name(), to, err)
		}
		if err = file.Close(); err != nil {
			t.Errorf("%s: close %q failed: %v", fs.Name(), to, err)
		}
		err = fs.Rename(from, to)
		if err != nil {
			t.Fatalf("%s: rename %q, %q failed: %v", fs.Name(), to, from, err)
		}
		file, err = fs.Create(from)
		if err != nil {
			t.Fatalf("%s: open %q failed: %v", fs.Name(), to, err)
		}
		if err = file.Close(); err != nil {
			t.Errorf("%s: close %q failed: %v", fs.Name(), to, err)
		}
		err = fs.Rename(from, exists)
		if err != nil {
			t.Errorf("%s: rename %q, %q failed: %v", fs.Name(), exists, from, err)
		}
		names, err := readDirNames(fs, tDir)
		if err != nil {
			t.Errorf("%s: readDirNames error: %v", fs.Name(), err)
		}
		found := false
		for _, e := range names {
			if e == "renamefrom" {
				t.Error("File is still called renamefrom")
			}
			if e == "renameto" {
				found = true
			}
		}
		if !found {
			t.Error("File was not renamed to renameto")
		}

		_, err = fs.Stat(to)
		if err != nil {
			t.Errorf("%s: stat %q failed: %v", fs.Name(), to, err)
		}
	}
}

func TestRemove(t *testing.T) {
	for _, fs := range Fss {

		x, err := TempFile(fs, "", "felix")
		if err != nil {
			t.Error(fmt.Sprint("unable to work with temp file", err))
		}

		path := x.Name()
		x.Close()

		tDir := filepath.Dir(path)

		err = fs.Remove(path)
		if err != nil {
			t.Errorf("%v: Remove() failed: %v", fs.Name(), err)
			continue
		}

		_, err = fs.Stat(path)
		if !os.IsNotExist(err) {
			t.Errorf("%v: Remove() didn't remove file", fs.Name())
			continue
		}

		// Deleting non-existent file should raise error
		err = fs.Remove(path)
		if !os.IsNotExist(err) {
			t.Errorf("%v: Remove() didn't raise error for non-existent file", fs.Name())
		}

		f, err := fs.Open(tDir)
		if err != nil {
			t.Error("TestDir should still exist:", err)
		}

		names, err := f.Readdirnames(-1)
		if err != nil {
			t.Error("Readdirnames failed:", err)
		}

		for _, e := range names {
			if e == testName {
				t.Error("File was not removed from parent directory")
			}
		}
	}
}

func TestTruncate(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		f := tmpFile(fs)
		defer f.Close()

		checkSize(t, f, 0)
		f.Write([]byte("hello, world\n"))
		checkSize(t, f, 13)
		f.Truncate(10)
		checkSize(t, f, 10)
		f.Truncate(1024)
		checkSize(t, f, 1024)
		f.Truncate(0)
		checkSize(t, f, 0)
		_, err := f.Write([]byte("surprise!"))
		if err == nil {
			checkSize(t, f, 13+9) // wrote at offset past where hello, world was.
		}
	}
}

func TestSeek(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		f := tmpFile(fs)
		defer f.Close()

		const data = "hello, world\n"
		io.WriteString(f, data)

		type test struct {
			in     int64
			whence int
			out    int64
		}
		var tests = []test{
			{0, 1, int64(len(data))},
			{0, 0, 0},
			{5, 0, 5},
			{0, 2, int64(len(data))},
			{0, 0, 0},
			{-1, 2, int64(len(data)) - 1},
			{1 << 33, 0, 1 << 33},
			{1 << 33, 2, 1<<33 + int64(len(data))},
		}
		for i, tt := range tests {
			off, err := f.Seek(tt.in, tt.whence)
			if off != tt.out || err != nil {
				if e, ok := err.(*os.PathError); ok && e.Err == syscall.EINVAL && tt.out > 1<<32 {
					// Reiserfs rejects the big seeks.
					// http://code.google.com/p/go/issues/detail?id=91
					break
				}
				t.Errorf("#%d: Seek(%v, %v) = %v, %v want %v, nil", i, tt.in, tt.whence, off, err, tt.out)
			}
		}
	}
}

func TestReadAt(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		f := tmpFile(fs)
		defer f.Close()

		const data = "hello, world\n"
		io.WriteString(f, data)

		b := make([]byte, 5)
		n, err := f.ReadAt(b, 7)
		if err != nil || n != len(b) {
			t.Fatalf("ReadAt 7: %d, %v", n, err)
		}
		if string(b) != "world" {
			t.Fatalf("ReadAt 7: have %q want %q", string(b), "world")
		}
	}
}

func TestWriteAt(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		f := tmpFile(fs)
		defer f.Close()

		const data = "hello, world\n"
		io.WriteString(f, data)

		n, err := f.WriteAt([]byte("WORLD"), 7)
		if err != nil || n != 5 {
			t.Fatalf("WriteAt 7: %d, %v", n, err)
		}

		f2, err := fs.Open(f.Name())
		if err != nil {
			t.Fatalf("%v: ReadFile %s: %v", fs.Name(), f.Name(), err)
		}
		defer f2.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(f2)
		b := buf.Bytes()
		if string(b) != "hello, WORLD\n" {
			t.Fatalf("after write: have %q want %q", string(b), "hello, WORLD\n")
		}

	}
}

func TestReaddirnames(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		testSubDir := SetupTestDir(t, fs)
		tDir := filepath.Dir(testSubDir)

		root, err := fs.Open(tDir)
		if err != nil {
			t.Fatal(fs.Name(), tDir, err)
		}
		defer root.Close()

		namesRoot, err := root.Readdirnames(-1)
		if err != nil {
			t.Fatal(fs.Name(), namesRoot, err)
		}

		sub, err := fs.Open(testSubDir)
		if err != nil {
			t.Fatal(err)
		}
		defer sub.Close()

		namesSub, err := sub.Readdirnames(-1)
		if err != nil {
			t.Fatal(fs.Name(), namesSub, err)
		}

		findNames(fs, t, tDir, testSubDir, namesRoot, namesSub)
	}
}

func TestReaddirSimple(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		testSubDir := SetupTestDir(t, fs)
		tDir := filepath.Dir(testSubDir)

		root, err := fs.Open(tDir)
		if err != nil {
			t.Fatal(err)
		}
		defer root.Close()

		rootInfo, err := root.Readdir(1)
		if err != nil {
			t.Log(myFileInfo(rootInfo))
			t.Error(err)
		}

		rootInfo, err = root.Readdir(5)
		if err != io.EOF {
			t.Log(myFileInfo(rootInfo))
			t.Error(err)
		}

		sub, err := fs.Open(testSubDir)
		if err != nil {
			t.Fatal(err)
		}
		defer sub.Close()

		subInfo, err := sub.Readdir(5)
		if err != nil {
			t.Log(myFileInfo(subInfo))
			t.Error(err)
		}
	}
}

func TestReaddir(t *testing.T) {
	defer removeAllTestFiles(t)
	for num := 0; num < 6; num++ {
		outputs := make([]string, len(Fss))
		infos := make([]string, len(Fss))
		for i, fs := range Fss {
			testSubDir := SetupTestDir(t, fs)
			//tDir := filepath.Dir(testSubDir)
			root, err := fs.Open(testSubDir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()

			for j := 0; j < 6; j++ {
				info, err := root.Readdir(num)
				outputs[i] += fmt.Sprintf("%v  Error: %v\n", myFileInfo(info), err)
				infos[i] += fmt.Sprintln(len(info), err)
			}
		}

		fail := false
		for i, o := range infos {
			if i == 0 {
				continue
			}
			if o != infos[i-1] {
				fail = true
				break
			}
		}
		if fail {
			t.Log("Readdir outputs not equal for Readdir(", num, ")")
			for i, o := range outputs {
				t.Log(Fss[i].Name())
				t.Log(o)
			}
			t.Fail()
		}
	}
}

func TestReaddirRegularFile(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		f := tmpFile(fs)
		defer f.Close()

		_, err := f.Readdirnames(-1)
		if err == nil {
			t.Fatal("Expected error")
		}

		_, err = f.Readdir(-1)
		if err == nil {
			t.Fatal("Expected error")
		}
	}
}

type myFileInfo []os.FileInfo

func (m myFileInfo) String() string {
	out := "Fileinfos:\n"
	for _, e := range m {
		out += "  " + e.Name() + "\n"
	}
	return out
}

func TestReaddirAll(t *testing.T) {
	defer removeAllTestFiles(t)
	for _, fs := range Fss {
		testSubDir := SetupTestDir(t, fs)
		tDir := filepath.Dir(testSubDir)

		root, err := fs.Open(tDir)
		if err != nil {
			t.Fatal(err)
		}
		defer root.Close()

		rootInfo, err := root.Readdir(-1)
		if err != nil {
			t.Fatal(err)
		}
		var namesRoot = []string{}
		for _, e := range rootInfo {
			namesRoot = append(namesRoot, e.Name())
		}

		sub, err := fs.Open(testSubDir)
		if err != nil {
			t.Fatal(err)
		}
		defer sub.Close()

		subInfo, err := sub.Readdir(-1)
		if err != nil {
			t.Fatal(err)
		}
		var namesSub = []string{}
		for _, e := range subInfo {
			namesSub = append(namesSub, e.Name())
		}

		findNames(fs, t, tDir, testSubDir, namesRoot, namesSub)
	}
}

func findNames(fs Vfs, t *testing.T, tDir, testSubDir string, root, sub []string) {
	var foundRoot bool
	for _, e := range root {
		f, err := fs.Open(filepath.Join(tDir, e))
		if err != nil {
			t.Error("Open", filepath.Join(tDir, e), ":", err)
		}
		defer f.Close()

		if equal(e, "we") {
			foundRoot = true
		}
	}
	if !foundRoot {
		t.Logf("Names root: %v", root)
		t.Logf("Names sub: %v", sub)
		t.Error("Didn't find subdirectory we")
	}

	var found1, found2 bool
	for _, e := range sub {
		f, err := fs.Open(filepath.Join(testSubDir, e))
		if err != nil {
			t.Error("Open", filepath.Join(testSubDir, e), ":", err)
		}
		defer f.Close()

		if equal(e, "testfile1") {
			found1 = true
		}
		if equal(e, "testfile2") {
			found2 = true
		}
	}

	if !found1 {
		t.Logf("Names root: %v", root)
		t.Logf("Names sub: %v", sub)
		t.Error("Didn't find testfile1")
	}
	if !found2 {
		t.Logf("Names root: %v", root)
		t.Logf("Names sub: %v", sub)
		t.Error("Didn't find testfile2")
	}
}

func removeAllTestFiles(t *testing.T) {
	for fs, list := range testRegistry {
		for _, path := range list {
			if err := fs.RemoveAll(path); err != nil {
				t.Error(fs.Name(), err)
			}
		}
	}
	testRegistry = make(map[Vfs][]string)
}

func equal(name1, name2 string) (r bool) {
	switch runtime.GOOS {
	case "windows":
		r = strings.ToLower(name1) == strings.ToLower(name2)
	default:
		r = name1 == name2
	}
	return
}

func checkSize(t *testing.T, f File, size int64) {
	dir, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat %q (looking for size %d): %s", f.Name(), size, err)
	}
	if dir.Size() != size {
		t.Errorf("Stat %q: size %d want %d", f.Name(), dir.Size(), size)
	}
}



func TestDirExists(t *testing.T) {
	type test struct {
		input    string
		expected bool
	}

	// First create a couple directories so there is something in the filesystem
	//testFS := new(MemMapFs)
	testFS.MkdirAll("/foo/bar", 0777)

	data := []test{
		{".", true},
		{"./", true},
		{"..", true},
		{"../", true},
		{"./..", true},
		{"./../", true},
		{"/foo/", true},
		{"/foo", true},
		{"/foo/bar", true},
		{"/foo/bar/", true},
		{"/", true},
		{"/some-really-random-directory-name", false},
		{"/some/really/random/directory/name", false},
		{"./some-really-random-local-directory-name", false},
		{"./some/really/random/local/directory/name", false},
	}

	for i, d := range data {
		exists, _ := DirExists(testFS, filepath.FromSlash(d.input))
		if d.expected != exists {
			t.Errorf("Test %d %q failed. Expected %t got %t", i, d.input, d.expected, exists)
		}
	}
}

func TestIsDir(t *testing.T) {
	testFS = new(MemMapFs)

	type test struct {
		input    string
		expected bool
	}
	data := []test{
		{"./", true},
		{"/", true},
		{"./this-directory-does-not-existi", false},
		{"/this-absolute-directory/does-not-exist", false},
	}

	for i, d := range data {

		exists, _ := IsDir(testFS, d.input)
		if d.expected != exists {
			t.Errorf("Test %d failed. Expected %t got %t", i, d.expected, exists)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	testFS = new(MemMapFs)

	zeroSizedFile, _ := createZeroSizedFileInTempDir()
	defer deleteFileInTempDir(zeroSizedFile)
	nonZeroSizedFile, _ := createNonZeroSizedFileInTempDir()
	defer deleteFileInTempDir(nonZeroSizedFile)
	emptyDirectory, _ := createEmptyTempDir()
	defer deleteTempDir(emptyDirectory)
	nonEmptyZeroLengthFilesDirectory, _ := createTempDirWithZeroLengthFiles()
	defer deleteTempDir(nonEmptyZeroLengthFilesDirectory)
	nonEmptyNonZeroLengthFilesDirectory, _ := createTempDirWithNonZeroLengthFiles()
	defer deleteTempDir(nonEmptyNonZeroLengthFilesDirectory)
	nonExistentFile := os.TempDir() + "/this-file-does-not-exist.txt"
	nonExistentDir := os.TempDir() + "/this/direcotry/does/not/exist/"

	fileDoesNotExist := fmt.Errorf("%q path does not exist", nonExistentFile)
	dirDoesNotExist := fmt.Errorf("%q path does not exist", nonExistentDir)

	type test struct {
		input          string
		expectedResult bool
		expectedErr    error
	}

	data := []test{
		{zeroSizedFile.Name(), true, nil},
		{nonZeroSizedFile.Name(), false, nil},
		{emptyDirectory, true, nil},
		{nonEmptyZeroLengthFilesDirectory, false, nil},
		{nonEmptyNonZeroLengthFilesDirectory, false, nil},
		{nonExistentFile, false, fileDoesNotExist},
		{nonExistentDir, false, dirDoesNotExist},
	}
	for i, d := range data {
		exists, err := IsEmpty(testFS, d.input)
		if d.expectedResult != exists {
			t.Errorf("Test %d %q failed exists. Expected result %t got %t", i, d.input, d.expectedResult, exists)
		}
		if d.expectedErr != nil {
			if d.expectedErr.Error() != err.Error() {
				t.Errorf("Test %d failed with err. Expected %q(%#v) got %q(%#v)", i, d.expectedErr, d.expectedErr, err, err)
			}
		} else {
			if d.expectedErr != err {
				t.Errorf("Test %d failed. Expected error %q(%#v) got %q(%#v)", i, d.expectedErr, d.expectedErr, err, err)
			}
		}
	}
}

func TestReaderContains(t *testing.T) {
	for i, this := range []struct {
		v1     string
		v2     [][]byte
		expect bool
	}{
		{"abc", [][]byte{[]byte("a")}, true},
		{"abc", [][]byte{[]byte("b")}, true},
		{"abcdefg", [][]byte{[]byte("efg")}, true},
		{"abc", [][]byte{[]byte("d")}, false},
		{"abc", [][]byte{[]byte("d"), []byte("e")}, false},
		{"abc", [][]byte{[]byte("d"), []byte("a")}, true},
		{"abc", [][]byte{[]byte("b"), []byte("e")}, true},
		{"", nil, false},
		{"", [][]byte{[]byte("a")}, false},
		{"a", [][]byte{[]byte("")}, false},
		{"", [][]byte{[]byte("")}, false}} {
		result := readerContainsAny(strings.NewReader(this.v1), this.v2...)
		if result != this.expect {
			t.Errorf("[%d] readerContains: got %t but expected %t", i, result, this.expect)
		}
	}

	if readerContainsAny(nil, []byte("a")) {
		t.Error("readerContains with nil reader")
	}

	if readerContainsAny(nil, nil) {
		t.Error("readerContains with nil arguments")
	}
}

func createZeroSizedFileInTempDir() (File, error) {
	filePrefix := "_path_test_"
	f, e := TempFile(testFS, "", filePrefix) // dir is os.TempDir()
	if e != nil {
		// if there was an error no file was created.
		// => no requirement to delete the file
		return nil, e
	}
	return f, nil
}

func createNonZeroSizedFileInTempDir() (File, error) {
	f, err := createZeroSizedFileInTempDir()
	if err != nil {
		// no file ??
	}
	byteString := []byte("byteString")
	err = WriteFile(testFS, f.Name(), byteString, 0644)
	if err != nil {
		// delete the file
		deleteFileInTempDir(f)
		return nil, err
	}
	return f, nil
}

func deleteFileInTempDir(f File) {
	err := testFS.Remove(f.Name())
	if err != nil {
		// now what?
	}
}

func createEmptyTempDir() (string, error) {
	dirPrefix := "_dir_prefix_"
	d, e := TempDir(testFS, "", dirPrefix) // will be in os.TempDir()
	if e != nil {
		// no directory to delete - it was never created
		return "", e
	}
	return d, nil
}

func createTempDirWithZeroLengthFiles() (string, error) {
	d, dirErr := createEmptyTempDir()
	if dirErr != nil {
		//now what?
	}
	filePrefix := "_path_test_"
	_, fileErr := TempFile(testFS, d, filePrefix) // dir is os.TempDir()
	if fileErr != nil {
		// if there was an error no file was created.
		// but we need to remove the directory to clean-up
		deleteTempDir(d)
		return "", fileErr
	}
	// the dir now has one, zero length file in it
	return d, nil

}

func createTempDirWithNonZeroLengthFiles() (string, error) {
	d, dirErr := createEmptyTempDir()
	if dirErr != nil {
		//now what?
	}
	filePrefix := "_path_test_"
	f, fileErr := TempFile(testFS, d, filePrefix) // dir is os.TempDir()
	if fileErr != nil {
		// if there was an error no file was created.
		// but we need to remove the directory to clean-up
		deleteTempDir(d)
		return "", fileErr
	}
	byteString := []byte("byteString")
	fileErr = WriteFile(testFS, f.Name(), byteString, 0644)
	if fileErr != nil {
		// delete the file
		deleteFileInTempDir(f)
		// also delete the directory
		deleteTempDir(d)
		return "", fileErr
	}

	// the dir now has one, zero length file in it
	return d, nil

}

func TestExists(t *testing.T) {
	zeroSizedFile, _ := createZeroSizedFileInTempDir()
	defer deleteFileInTempDir(zeroSizedFile)
	nonZeroSizedFile, _ := createNonZeroSizedFileInTempDir()
	defer deleteFileInTempDir(nonZeroSizedFile)
	emptyDirectory, _ := createEmptyTempDir()
	defer deleteTempDir(emptyDirectory)
	nonExistentFile := os.TempDir() + "/this-file-does-not-exist.txt"
	nonExistentDir := os.TempDir() + "/this/direcotry/does/not/exist/"

	type test struct {
		input          string
		expectedResult bool
		expectedErr    error
	}

	data := []test{
		{zeroSizedFile.Name(), true, nil},
		{nonZeroSizedFile.Name(), true, nil},
		{emptyDirectory, true, nil},
		{nonExistentFile, false, nil},
		{nonExistentDir, false, nil},
	}
	for i, d := range data {
		exists, err := Exists(testFS, d.input)
		if d.expectedResult != exists {
			t.Errorf("Test %d failed. Expected result %t got %t", i, d.expectedResult, exists)
		}
		if d.expectedErr != err {
			t.Errorf("Test %d failed. Expected %q got %q", i, d.expectedErr, err)
		}
	}

}

func TestSafeWriteToDisk(t *testing.T) {
	emptyFile, _ := createZeroSizedFileInTempDir()
	defer deleteFileInTempDir(emptyFile)
	tmpDir, _ := createEmptyTempDir()
	defer deleteTempDir(tmpDir)

	randomString := "This is a random string!"
	reader := strings.NewReader(randomString)

	fileExists := fmt.Errorf("%v already exists", emptyFile.Name())

	type test struct {
		filename    string
		expectedErr error
	}

	now := time.Now().Unix()
	nowStr := strconv.FormatInt(now, 10)
	data := []test{
		{emptyFile.Name(), fileExists},
		{tmpDir + "/" + nowStr, nil},
	}

	for i, d := range data {
		e := SafeWriteReader(testFS, d.filename, reader)
		if d.expectedErr != nil {
			if d.expectedErr.Error() != e.Error() {
				t.Errorf("Test %d failed. Expected error %q but got %q", i, d.expectedErr.Error(), e.Error())
			}
		} else {
			if d.expectedErr != e {
				t.Errorf("Test %d failed. Expected %q but got %q", i, d.expectedErr, e)
			}
			contents, _ := ReadFile(testFS, d.filename)
			if randomString != string(contents) {
				t.Errorf("Test %d failed. Expected contents %q but got %q", i, randomString, string(contents))
			}
		}
		reader.Seek(0, 0)
	}
}

func TestWriteToDisk(t *testing.T) {
	emptyFile, _ := createZeroSizedFileInTempDir()
	defer deleteFileInTempDir(emptyFile)
	tmpDir, _ := createEmptyTempDir()
	defer deleteTempDir(tmpDir)

	randomString := "This is a random string!"
	reader := strings.NewReader(randomString)

	type test struct {
		filename    string
		expectedErr error
	}

	now := time.Now().Unix()
	nowStr := strconv.FormatInt(now, 10)
	data := []test{
		{emptyFile.Name(), nil},
		{tmpDir + "/" + nowStr, nil},
	}

	for i, d := range data {
		e := WriteReader(testFS, d.filename, reader)
		if d.expectedErr != e {
			t.Errorf("Test %d failed. WriteToDisk Error Expected %q but got %q", i, d.expectedErr, e)
		}
		contents, e := ReadFile(testFS, d.filename)
		if e != nil {
			t.Errorf("Test %d failed. Could not read file %s. Reason: %s\n", i, d.filename, e)
		}
		if randomString != string(contents) {
			t.Errorf("Test %d failed. Expected contents %q but got %q", i, randomString, string(contents))
		}
		reader.Seek(0, 0)
	}
}

func TestGetTempDir(t *testing.T) {
	dir := os.TempDir()
	if FilePathSeparator != dir[len(dir)-1:] {
		dir = dir + FilePathSeparator
	}
	testDir := "hugoTestFolder" + FilePathSeparator
	tests := []struct {
		input    string
		expected string
	}{
		{"", dir},
		{testDir + "  Foo bar  ", dir + testDir + "  Foo bar  " + FilePathSeparator},
		{testDir + "Foo.Bar/foo_Bar-Foo", dir + testDir + "Foo.Bar/foo_Bar-Foo" + FilePathSeparator},
		{testDir + "fOO,bar:foo%bAR", dir + testDir + "fOObarfoo%bAR" + FilePathSeparator},
		{testDir + "FOo/BaR.html", dir + testDir + "FOo/BaR.html" + FilePathSeparator},
		{testDir + "трям/трям", dir + testDir + "трям/трям" + FilePathSeparator},
		{testDir + "은행", dir + testDir + "은행" + FilePathSeparator},
		{testDir + "Банковский кассир", dir + testDir + "Банковский кассир" + FilePathSeparator},
	}

	for _, test := range tests {
		output := GetTempDir(new(MemMapFs), test.input)
		if output != test.expected {
			t.Errorf("Expected %#v, got %#v\n", test.expected, output)
		}
	}
}

// This function is very dangerous. Don't use it.
func deleteTempDir(d string) {
	err := os.RemoveAll(d)
	if err != nil {
		// now what?
	}
}

func TestFullBaseFsPath(t *testing.T) {
	type dirSpec struct {
		Dir1, Dir2, Dir3 string
	}
	dirSpecs := []dirSpec{
		dirSpec{Dir1: "/", Dir2: "/", Dir3: "/"},
		dirSpec{Dir1: "/", Dir2: "/path2", Dir3: "/"},
		dirSpec{Dir1: "/path1/dir", Dir2: "/path2/dir/", Dir3: "/path3/dir"},
		dirSpec{Dir1: "C:/path1", Dir2: "path2/dir", Dir3: "/path3/dir/"},
	}

	for _, ds := range dirSpecs {
		memFs := NewMemMapFs()
		level1Fs := NewBasePathFs(memFs, ds.Dir1)
		level2Fs := NewBasePathFs(level1Fs, ds.Dir2)
		level3Fs := NewBasePathFs(level2Fs, ds.Dir3)

		type spec struct {
			BaseFs       Vfs
			FileName     string
			ExpectedPath string
		}
		specs := []spec{
			spec{BaseFs: level3Fs, FileName: "f.txt", ExpectedPath: filepath.Join(ds.Dir1, ds.Dir2, ds.Dir3, "f.txt")},
			spec{BaseFs: level3Fs, FileName: "", ExpectedPath: filepath.Join(ds.Dir1, ds.Dir2, ds.Dir3, "")},
			spec{BaseFs: level2Fs, FileName: "f.txt", ExpectedPath: filepath.Join(ds.Dir1, ds.Dir2, "f.txt")},
			spec{BaseFs: level2Fs, FileName: "", ExpectedPath: filepath.Join(ds.Dir1, ds.Dir2, "")},
			spec{BaseFs: level1Fs, FileName: "f.txt", ExpectedPath: filepath.Join(ds.Dir1, "f.txt")},
			spec{BaseFs: level1Fs, FileName: "", ExpectedPath: filepath.Join(ds.Dir1, "")},
		}

		for _, s := range specs {
			if actualPath := FullBaseFsPath(s.BaseFs.(*BasePathFs), s.FileName); actualPath != s.ExpectedPath {
				t.Errorf("Expected \n%s got \n%s", s.ExpectedPath, actualPath)
			}
		}
	}
}



