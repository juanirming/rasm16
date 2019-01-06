/*
Package file provides functions for reading raw rasm source- and include
files from disk plus writing assembled binary files to disk.

It also defines filename extensions for the above.

Copyright 2018-2019 Juan Irming

This file is part of rasm16.

rasm16 is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

rasm16 is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with rasm16.  If not, see <http://www.gnu.org/licenses/>.
*/
package file

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

// -----------------------------------------------------------------------------

const DEBUG bool = true

// -----------------------------------------------------------------------------

// Filename extensions for input- and output files.
const (
	SrcExt string = ".rasm"
	IncExt string = "._rasm"
	BinExt string = ".r16"
)

// -----------------------------------------------------------------------------

// ReadSrc reads a source file from disk into a slice, one line per element.
func ReadSrc(srcName string) ([]string, error) {
	if DEBUG {
		fmt.Println("Reading " + srcName)
	}

	f, err := os.Open(srcName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	fmt.Println("Read " + srcName)

	return lines, scanner.Err()
}

// -----------------------------------------------------------------------------

// WriteBin writes a byte slice to disk as a binary file.
func WriteBin(bin []byte, binName string) error {
	if DEBUG {
		fmt.Println("Writing " + binName)
	}

	err := ioutil.WriteFile(binName, bin, 0666)
	if err != nil {
		return err
	}

	fmt.Println("Wrote", len(bin), "bytes to " + binName)

	return nil
}
