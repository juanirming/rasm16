/*
Package assemble provides functions for processing raw rasm source code,
converting it to structured form, processing the structured source and finally
converting it to a binary executable.

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
package assemble

import (
	"errors"
	"fmt"
	"strconv"
)

// -----------------------------------------------------------------------------

const DEBUG bool = true

// -----------------------------------------------------------------------------

// Raw orchestrates the complete assembly process, turning a string slice into
// a byte slice via the following steps, in order:
//
// Process source
//      Clean-up
//      Expand constants
//      Namespacing
//      Includes
//      Validate labels
// Convert to struct
// Process struct
//      Unalias mnemonics
//      Translate data strings to hex
//      Expand data null repeats
//      Validate mnemonics
//      Validate data directives
//      Calculate addresses
//      Expand labels
//      Validate operands
// Convert to binary
func Raw(rawSrcLines []string, srcName string, programOffset uint16) ([]byte, error) {
	var err error

	printSrc("", rawSrcLines)

	rawSrcLines = cleanSrc(rawSrcLines)
	printSrc("Removed comments and extraneous whitespace", rawSrcLines)

	rawSrcLines, err = expandConsts(rawSrcLines)
	if err != nil {
		return nil, err
	}
	printSrc("Expanded constants", rawSrcLines)

	rawSrcLines = addSrcLabelNamespaces(rawSrcLines, srcName)
	printSrc("Added label namespaces", rawSrcLines)

	rawSrcLines, err = addIncludes(rawSrcLines)
	if err != nil {
		return nil, err
	}
	printSrc("Added include files", rawSrcLines)

	hasDupeSrcLabels, srcLabel, lineNum := hasDupeSrcLabels(rawSrcLines)
	if hasDupeSrcLabels {
		return nil, errors.New("Duplicate label found on line " + strconv.Itoa(lineNum+1) + ": " + srcLabel)
	}

	srcLines := buildStructSrc(rawSrcLines)
	printStructSrc("Built structured source", srcLines)

	srcLines = unaliasMnemonics(srcLines)
	printStructSrc("Unaliased mnemonics", srcLines)

	srcLines = convDataStringsToHex(srcLines)
	printStructSrc("Converted data strings to hex", srcLines)

	srcLines, err = expandDataNullRepeats(srcLines)
	if err != nil {
		return nil, err
	}
	printStructSrc("Expanded data null repeats", srcLines)

	_, err = validateMnemonics(srcLines)
	if err != nil {
		return nil, err
	}

	_, err = validateDataDirectives(srcLines)
	if err != nil {
		return nil, err
	}

	srcLines, err = calcAddresses(srcLines, programOffset)
	if err != nil {
		return nil, err
	}
	printStructSrc("Calculated addresses", srcLines)

	labelAddresses := getLabelAddresses(srcLines)

	if DEBUG {
		fmt.Println("Found label addresses", labelAddresses)
	}

	srcLines, err = expandLabels(srcLines, labelAddresses)
	if err != nil {
		return nil, err
	}
	printStructSrc("Expanded labels", srcLines)

	_, err = validateOps(srcLines)
	if err != nil {
		return nil, err
	}

	srcLines = buildBinSrcLines(srcLines)
	printStructSrc("Built structured binary", srcLines)

	bin := buildBin(srcLines, programOffset)
	printBin("Built final binary", bin)

	return bin, nil
}
