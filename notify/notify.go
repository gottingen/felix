// Copyright 2012 The Go Authors. All rights reserved.
// Copyright 2019 lijippy@163.com
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package notify

import (
	"bytes"
	"errors"
	"fmt"
)

type Event struct {
	Name string
	Op   Op
}

type Op int32

const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

func (op Op) String() string {
	// Use a buffer for efficient string concatenation
	var buffer bytes.Buffer

	if op&Create == Create {
		buffer.WriteString("|CREATE")
	}
	if op&Remove == Remove {
		buffer.WriteString("|REMOVE")
	}
	if op&Write == Write {
		buffer.WriteString("|WRITE")
	}
	if op&Rename == Rename {
		buffer.WriteString("|RENAME")
	}
	if op&Chmod == Chmod {
		buffer.WriteString("|CHMOD")
	}
	if buffer.Len() == 0 {
		return ""
	}
	return buffer.String()[1:] // Strip leading pipe
}

func (e Event) String() string {
	return fmt.Sprintf("%q: %s", e.Name, e.Op.String())
}

var (
	ErrEventOverflow = errors.New("fsnotify queue overflow")
)

type Watcher interface {
	Close() error
	Add(path string) error
	Remove(path string) error
	EventChannel() <-chan Event
	ErrorChannel() <-chan error
}
