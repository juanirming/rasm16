/*
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
	"regexp"
	"strings"
)

// -----------------------------------------------------------------------------

// Parser delimiter definitions.
const (
	mnemonicOpDlm string = " "
	opDlm         string = ","
)

// Minimum number of characters allowed in a source label.
const srcLabelMinLen = 5

// Structured source line definition.
type srcLine struct {
	lineNum  int
	label    string
	address  int
	mnemonic string
	op1Type  opType
	op1      string
	op2Type  opType
	op2      string
	data     string
	bin      []byte
}

// -----------------------------------------------------------------------------

// buildStructSrc converts processed source lines to structured source code.
func buildStructSrc(srcLines []string) []srcLine {
	var structSrcLines []srcLine

	for lineNum, srcLineString := range srcLines {
		if srcLineString != "" && !isSrcLabel(srcLineString) {
			srcLabel := getSrcLabel(srcLines, lineNum)
			currentSrcLine := srcLine{}

			mnemonic, op1, op2, data := "", "", "", ""
			var op1Type, op2Type opType

			if isSrcDataLine(srcLineString) {
				mnemonic, data = splitSrcDataLine(srcLineString)
			} else {
				mnemonic, op1, op2 = splitSrcCodeLine(srcLineString)
				mnemonic = strings.ToUpper(mnemonic)

				op1Type, op1 = splitOp(op1)
				op2Type, op2 = splitOp(op2)
			}

			currentSrcLine = srcLine{
				lineNum:  lineNum,
				label:    srcLabel,
				mnemonic: mnemonic,
				op1Type:  op1Type,
				op1:      op1,
				op2Type:  op2Type,
				op2:      op2,
				data:     data,
			}

			structSrcLines = append(structSrcLines, currentSrcLine)
		}
	}

	return structSrcLines
}

// -----------------------------------------------------------------------------

// splitSrcCodeLine breaks down a line of non-data directive source code.
func splitSrcCodeLine(srcLine string) (string, string, string) {
	splitLine := strings.SplitN(srcLine, mnemonicOpDlm, 2)

	if len(splitLine) > 1 {
		splitOps := strings.SplitN(splitLine[1], opDlm, 2)

		if len(splitOps) > 1 {
			return splitLine[0], splitOps[0], splitOps[1]
		}

		return splitLine[0], splitOps[0], ""
	}

	return splitLine[0], "", ""
}

// -----------------------------------------------------------------------------

// isSrcDataLine checks whether a line of source code is a data directive.
func isSrcDataLine(srcLine string) bool {
	return srcLine != "" && srcLine[:1] == dataLineToken
}

// -----------------------------------------------------------------------------

// splitSrcCodeLine breaks down a line of source code containing a data
// directive.
func splitSrcDataLine(srcLine string) (string, string) {
	splitLine := strings.SplitN(srcLine, mnemonicOpDlm, 2)

	if len(splitLine) > 1 {
		return splitLine[0], splitLine[1]
	}

	return splitLine[0], ""
}

// -----------------------------------------------------------------------------

// splitOp breaks down an operand into type and value.
func splitOp(op string) (opType, string) {
	cleanOp := strings.TrimSpace(op)

	opType := getOpType(cleanOp)

	if opType == literalOp || opType == pointerOp {
		return opType, cleanOp[1:len(cleanOp)]
	}

	return opType, cleanOp
}

// -----------------------------------------------------------------------------

// getOpType determines the type of an operand.
func getOpType(op string) opType {
	if op != "" {
		firstChar := op[:1]

		if firstChar == opTokens[literalOp] {
			return literalOp
		} else if firstChar == opTokens[pointerOp] {
			return pointerOp
		} else {
			return addressOp
		}
	}

	return invalidOp
}

// -----------------------------------------------------------------------------

// getSrcLabel determines the label, if any, of a line of source code.
func getSrcLabel(srcLines []string, lineNum int) string {
	currentLineNum := lineNum - 1

	for currentLineNum >= 0 {
		if srcLines[currentLineNum] != "" {
			if isSrcLabel(srcLines[currentLineNum]) {
				return srcLines[currentLineNum]
			} else {
				return ""
			}
		}

		currentLineNum--
	}

	return ""
}

// -----------------------------------------------------------------------------

// isSrcLabel checks whether a string is a source label.
func isSrcLabel(srcLine string) bool {
	reSrcLabel := regexp.MustCompile(`^[\w.]+$`)

	return len(srcLine) >= srcLabelMinLen && reSrcLabel.MatchString(srcLine)
}
