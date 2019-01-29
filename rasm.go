/*
Package main provides the executable kick-off point for the rasm assembler.

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
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"github.com/juanirming/rasm16/assemble"
	"github.com/juanirming/rasm16/file"
	"strconv"
)

// -----------------------------------------------------------------------------

// Basic application information.
const (
	appName    string = "rasm16"
	appVersion string = "1.0.0 alpha"
	appAuthor  string = "Juan Irming"
)

// -----------------------------------------------------------------------------

// Main reads a source file, kicks off the assembly process and writes the final
// binary to disk.
func main() {
	printAppInfo()

	programOffsetPtr := flag.String("o", "0000", "16-bit hexadecimal program offset")

	flag.Parse()

	programOffset, err := strconv.ParseUint(*programOffsetPtr, 16, 16)
	if err != nil {
		fmt.Println(err)

		return
	}

	srcName, binName, err := getFilenames()
	if err != nil {
		fmt.Println(err)
	} else {
		rawSrcLines, err := file.ReadSrc(srcName)
		if err != nil {
			fmt.Println(err)

			return
		}

		var bin []byte

		bin, err = assemble.Raw(rawSrcLines, srcName, uint16(programOffset))
		if err != nil {
			fmt.Println(err)

			return
		}

		err = file.WriteBin(bin, binName)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// -----------------------------------------------------------------------------

// printAppInfo outputs basic application information.
func printAppInfo() {
	fmt.Println(appName + " v" + appVersion + " by " + appAuthor)
}

// -----------------------------------------------------------------------------

// getFilenames returns the input- and output filenames based on the first
// command line argument passed into rasm.
func getFilenames() (string, string, error) {
	var srcName string
	var binName string

	if len(os.Args) > 1 {
		srcName = flag.Args()[0] + file.SrcExt
		binName = flag.Args()[0] + file.BinExt
	} else {
		return "", "", errors.New("Need source filename (without " + file.SrcExt + " extension) as first argument")
	}

	return srcName, binName, nil
}
