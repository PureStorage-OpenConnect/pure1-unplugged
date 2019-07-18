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

import "os"

// FileSystem is an interface for interacting with the filesystem
type FileSystem interface {
	Glob(pattern string) ([]string, error)
	PathExists(path string) (bool, error)
	ReadFile(filepath string) ([]byte, error)
	WriteToFile(filepath string, data []byte) error
	EvalSymlinks(path string) (string, error)
	ReadDir(dirname string) ([]os.FileInfo, error)
}

// LinuxFileSystem is an implementation of Filesystem
// available for use from this package
type LinuxFileSystem struct {
}
