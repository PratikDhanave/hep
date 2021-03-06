// Copyright 2018 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riofs

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"go-hep.org/x/hep/xrootd/xrdio"
)

func openFile(path string) (Reader, error) {
	switch {
	case strings.HasPrefix(path, "http://"), strings.HasPrefix(path, "https://"):
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		f, err := ioutil.TempFile("", "riofs-remote-")
		if err != nil {
			return nil, err
		}
		_, err = io.CopyBuffer(f, resp.Body, make([]byte, 16*1024*1024))
		if err != nil {
			f.Close()
			return nil, err
		}
		_, err = f.Seek(0, 0)
		if err != nil {
			f.Close()
			return nil, err
		}
		return &tmpFile{f}, nil

	case strings.HasPrefix(path, "file://"):
		return os.Open(path[len("file://"):])

	case strings.HasPrefix(path, "xroot://"), strings.HasPrefix(path, "root://"):
		return xrdio.Open(path)

	default:
		return os.Open(path)
	}
}

// tmpFile wraps a regular os.File to automatically remove it when closed.
type tmpFile struct {
	*os.File
}

func (f *tmpFile) Close() error {
	err1 := f.File.Close()
	err2 := os.Remove(f.File.Name())
	if err1 != nil {
		return err1
	}
	return err2
}

var (
	_ Reader = (*tmpFile)(nil)
	_ Writer = (*tmpFile)(nil)

	_ Reader = (*xrdio.File)(nil)
	_ Writer = (*xrdio.File)(nil)
)
