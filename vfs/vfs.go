package vfs

import (
	"bytes"
	"fmt"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)


// byName implements sort.Interface.
type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func ReadDir(fs Vfs, dirname string) ([]os.FileInfo, error) {
	f, err := fs.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Sort(byName(list))
	return list, nil
}


func ReadFile(fs Vfs, filename string) ([]byte, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// It's a good but not certain bet that FileInfo will tell us exactly how much to
	// read, so let's try it but be prepared for the answer to be wrong.
	var n int64

	if fi, err := f.Stat(); err == nil {
		// Don't preallocate a huge buffer, just in case.
		if size := fi.Size(); size < 1e9 {
			n = size
		}
	}
	// As initial capacity for readAll, use n + a little extra in case Size is zero,
	// and to avoid another allocation after Read has filled the buffer.  The readAll
	// call will read into its allocated internal buffer cheaply.  If the size was
	// wrong, we'll either waste some space off the end or reallocate as needed, but
	// in the overwhelmingly common case we'll get it just right.
	return readAll(f, n+bytes.MinRead)
}


// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}


// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
func ReadAll(r io.Reader) ([]byte, error) {
	return readAll(r, bytes.MinRead)
}


func WriteFile(fs Vfs, filename string, data []byte, perm os.FileMode) error {
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}


// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}


func TempFile(fs Vfs, dir, prefix string) (f File, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, prefix+nextSuffix())
		f, err = fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				randmu.Lock()
				rand = reseed()
				randmu.Unlock()
			}
			continue
		}
		break
	}
	return
}


func TempDir(fs Vfs, dir, prefix string) (name string, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		try := filepath.Join(dir, prefix+nextSuffix())
		err = fs.Mkdir(try, 0700)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				randmu.Lock()
				rand = reseed()
				randmu.Unlock()
			}
			continue
		}
		if err == nil {
			name = try
		}
		break
	}
	return
}


// Filepath separator defined by os.Separator.
const FilePathSeparator = string(filepath.Separator)

// Takes a reader and a path and writes the content


func WriteReader(fs Vfs, path string, r io.Reader) (err error) {
	dir, _ := filepath.Split(path)
	ospath := filepath.FromSlash(dir)

	if ospath != "" {
		err = fs.MkdirAll(ospath, 0777) // rwx, rw, r
		if err != nil {
			if err != os.ErrExist {
				return err
			}
		}
	}

	file, err := fs.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return
}

// Same as WriteReader but checks to see if file/directory already exists.

func SafeWriteReader(fs Vfs, path string, r io.Reader) (err error) {
	dir, _ := filepath.Split(path)
	ospath := filepath.FromSlash(dir)

	if ospath != "" {
		err = fs.MkdirAll(ospath, 0777) // rwx, rw, r
		if err != nil {
			return
		}
	}

	exists, err := Exists(fs, path)
	if err != nil {
		return
	}
	if exists {
		return fmt.Errorf("%v already exists", path)
	}

	file, err := fs.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return
}


// GetTempDir returns the default temp directory with trailing slash
// if subPath is not empty then it will be created recursively with mode 777 rwx rwx rwx
func GetTempDir(fs Vfs, subPath string) string {
	addSlash := func(p string) string {
		if FilePathSeparator != p[len(p)-1:] {
			p = p + FilePathSeparator
		}
		return p
	}
	dir := addSlash(os.TempDir())

	if subPath != "" {
		// preserve windows backslash :-(
		if FilePathSeparator == "\\" {
			subPath = strings.Replace(subPath, "\\", "____", -1)
		}
		dir = dir + UnicodeSanitize((subPath))
		if FilePathSeparator == "\\" {
			dir = strings.Replace(dir, "____", "\\", -1)
		}

		if exists, _ := Exists(fs, dir); exists {
			return addSlash(dir)
		}

		err := fs.MkdirAll(dir, 0777)
		if err != nil {
			panic(err)
		}
		dir = addSlash(dir)
	}
	return dir
}

// Rewrite string to remove non-standard path characters
func UnicodeSanitize(s string) string {
	source := []rune(s)
	target := make([]rune, 0, len(source))

	for _, r := range source {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			unicode.IsMark(r) ||
			r == '.' ||
			r == '/' ||
			r == '\\' ||
			r == '_' ||
			r == '-' ||
			r == '%' ||
			r == ' ' ||
			r == '#' {
			target = append(target, r)
		}
	}

	return string(target)
}

// Transform characters with accents into plain forms.
func NeuterAccents(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, string(s))

	return result
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}


// Check if a file contains a specified byte slice.
func FileContainsBytes(fs Vfs, filename string, subslice []byte) (bool, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	return readerContainsAny(f, subslice), nil
}


// Check if a file contains any of the specified byte slices.
func FileContainsAnyBytes(fs Vfs, filename string, subslices [][]byte) (bool, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	return readerContainsAny(f, subslices...), nil
}

// readerContains reports whether any of the subslices is within r.
func readerContainsAny(r io.Reader, subslices ...[]byte) bool {

	if r == nil || len(subslices) == 0 {
		return false
	}

	largestSlice := 0

	for _, sl := range subslices {
		if len(sl) > largestSlice {
			largestSlice = len(sl)
		}
	}

	if largestSlice == 0 {
		return false
	}

	bufflen := largestSlice * 4
	halflen := bufflen / 2
	buff := make([]byte, bufflen)
	var err error
	var n, i int

	for {
		i++
		if i == 1 {
			n, err = io.ReadAtLeast(r, buff[:halflen], halflen)
		} else {
			if i != 2 {
				// shift left to catch overlapping matches
				copy(buff[:], buff[halflen:])
			}
			n, err = io.ReadAtLeast(r, buff[halflen:], halflen)
		}

		if n > 0 {
			for _, sl := range subslices {
				if bytes.Contains(buff, sl) {
					return true
				}
			}
		}

		if err != nil {
			break
		}
	}
	return false
}

func DirExists(fs Vfs, path string) (bool, error) {
	fi, err := fs.Stat(path)
	if err == nil && fi.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}



// IsDir checks if a given path is a directory.
func IsDir(fs Vfs, path string) (bool, error) {
	fi, err := fs.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}



// IsEmpty checks if a given file or directory is empty.
func IsEmpty(fs Vfs, path string) (bool, error) {
	if b, _ := Exists(fs, path); !b {
		return false, fmt.Errorf("%q path does not exist", path)
	}
	fi, err := fs.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		f, err := fs.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()
		list, err := f.Readdir(-1)
		return len(list) == 0, nil
	}
	return fi.Size() == 0, nil
}


// Check if a file or directory exists.
func Exists(fs Vfs, path string) (bool, error) {
	_, err := fs.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func FullBaseFsPath(basePathFs *BasePathFs, relativePath string) string {
	combinedPath := filepath.Join(basePathFs.path, relativePath)
	if parent, ok := basePathFs.source.(*BasePathFs); ok {
		return FullBaseFsPath(parent, combinedPath)
	}

	return combinedPath
}







