// Copyright 2013 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// A simple utility for detecting duplicate files.
// Pass directories as arguments. Dupes will walk those directories scanning for duplicates.
// Returns comma-separated lists of duplicate files.
//
// Example usage:
//   ./dupes ~/Dropbox
package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var lens = make(map[int64][]string)
var dupeLens []int64
var dedupeSz int
var totalSz int

func walk(root string) error {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		return sameLength(path)
	}
	return filepath.Walk(root, walkFn)
}

func sameLength(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	sz := info.Size()
	lens[sz] = append(lens[sz], path)
	if len(lens[sz]) == 2 {
		dupeLens = append(dupeLens, sz)
	}
	return nil
}

func sameHash(paths []string) ([]string, error) {
	var hashes = make(map[uint32][]string)
	var dupes []uint32
	var dupesSz []int
	var dupePaths []string
	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return dupePaths, err
		}
		check := crc32.ChecksumIEEE(data)
		hashes[check] = append(hashes[check], path)
		if len(hashes[check]) == 2 {
			dupes = append(dupes, check)
			dedupeSz += len(data)
			dupesSz = append(dupesSz, len(data))
		}
	}
	for i, v := range dupes {
		dupePaths = append(dupePaths, strings.Join(hashes[v], ", "))
		totalSz += dupesSz[i] * len(hashes[v])
	}
	return dupePaths, nil
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Missing directory argument.")
		os.Exit(1)
	}
	for _, v := range args {
		err := walk(v)
		if err != nil {
			fmt.Println("Error walking from ", v, ": ", err)
		}
	}
	if len(dupeLens) < 1 {
		fmt.Println("No duplicates detected")
	} else {
		var dupes []string
		for _, v := range dupeLens {
			d, err := sameHash(lens[v])
			if err != nil {
				fmt.Println("Error reading files of len: ", v)
			} else {
				dupes = append(dupes, d...)
			}
		}
		if len(dupes) < 1 {
			fmt.Println("No duplicates detected")
		} else {
			fmt.Println("Wasted space: ", (totalSz-dedupeSz)/1024, "kb")
			for _, v := range dupes {
				fmt.Println(v)
			}
		}
	}

}
