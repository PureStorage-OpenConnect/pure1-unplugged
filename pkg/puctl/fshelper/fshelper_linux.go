// Copyright 2017, Pure Storage Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fshelper

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var _ FileSystem = (*LinuxFileSystem)(nil)

// Glob globs things
func (l *LinuxFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// ReadFile reads a file into a byte slice
func (l *LinuxFileSystem) ReadFile(filepath string) ([]byte, error) {
	return ioutil.ReadFile(filepath)
}

// WriteToFile writes a bytestream to a file.
func (l *LinuxFileSystem) WriteToFile(filepath string, data []byte) error {
	log.Debugf("Writing to file: path='%s', data='%s'", filepath, data)
	f, err := os.OpenFile(filepath, os.O_WRONLY, 0755)
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

// PathExists returns a bool reflecting if
// a path exists, or not
func (l *LinuxFileSystem) PathExists(path string) (bool, error) {
	log.WithFields(log.Fields{
		"path": path,
	}).Debug("Checking if path exists")
	_, err := os.Stat(path)
	if err == nil {
		log.Debugf("Path %s found!", path)
		return true, nil
	}
	if err != nil && !os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err,
		}).Error("Error checking path")
		return false, err
	}
	log.WithFields(log.Fields{
		"path": path,
	}).Debug("Path not found!")
	return false, nil
}

// EvalSymlinks is a shim implementation on top of `filepath.EvalSymlinks`
func (l *LinuxFileSystem) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

// ReadDir is a shim for ioutil.ReadDir
func (l *LinuxFileSystem) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}
