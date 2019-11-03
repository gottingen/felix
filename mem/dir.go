package mem

type Dir interface {
	Len() int
	Names() []string
	Files() []*FileData
	Add(*FileData)
	Remove(*FileData)
}

func RemoveFromMemDir(dir *FileData, f *FileData) {
	dir.memDir.Remove(f)
}

func AddToMemDir(dir *FileData, f *FileData) {
	dir.memDir.Add(f)
}

func InitializeDir(d *FileData) {
	if d.memDir == nil {
		d.dir = true
		d.memDir = &DirMap{}
	}
}
