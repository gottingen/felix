package felix

import (
	"github.com/gottingen/felix/filesystem"
)

type Felix struct {
	filesystem.Fs
}

func NewOsFs() Felix {
	f := filesystem.NewOsFs()
	return Felix{f}
}

func NewMemFs() Felix {
	f := filesystem.NewMemMapFs()
	return Felix{f}
}
