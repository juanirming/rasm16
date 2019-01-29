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
	"fmt"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------

// RELIC-16 binary executable file magic header.
var binMagicHeader []byte = []byte{0x12, 0x31, 0x1C, 0x16} // 0x12311C16 == RELIC16

// -----------------------------------------------------------------------------

// buildBinSrcLines constructs the binary instructions from a slice of
// processed source code.
func buildBinSrcLines(srcLines []srcLine) []srcLine {
	var binSrcLines []srcLine

	for _, srcLine := range srcLines {
		binSrcLine := srcLine

		if isValidDataDirective(srcLine.mnemonic) {
			binSrcLine = buildData(binSrcLine)
		} else {
			binSrcLine = buildInstr(binSrcLine)
		}

		binSrcLines = append(binSrcLines, binSrcLine)
	}

	return binSrcLines
}

// -----------------------------------------------------------------------------

// buildData builds out the binary values from a data directive.
func buildData(srcLine srcLine) srcLine {
	binSrcLine := srcLine

	var data64 uint64

	splitData := strings.Split(srcLine.data, dataDlm)

	if srcLine.mnemonic == directiveTokens[data8BitDirective] {
		for _, data := range splitData {
			data64, _ = strconv.ParseUint(data, 16, 8)
			binSrcLine.bin = append(binSrcLine.bin, byte(data64))
		}
	} else if srcLine.mnemonic == directiveTokens[data16BitDirective] {
		for _, data := range splitData {
			data64, _ = strconv.ParseUint(data, 16, 16)
			binSrcLine.bin = appendUint16(binSrcLine.bin, uint16(data64))
		}
	}

	return binSrcLine
}

// -----------------------------------------------------------------------------

// buildInstr constructs the opcode and operands for a line of source code.
func buildInstr(srcLine srcLine) srcLine {
	binSrcLine := buildOpcode(srcLine)
	binSrcLine = buildOps(binSrcLine)

	return binSrcLine
}

// -----------------------------------------------------------------------------

// buildOpcode creates binary opcodes from a line of source code.
func buildOpcode(srcLine srcLine) srcLine {
	opcodeSrcLine := srcLine

	opcode := mnemonics[srcLine.mnemonic].opcode
	opcode <<= 3

	if srcLine.op1Type == literalOp && srcLine.op2Type == addressOp {
		opcode |= 0x00
	} else if srcLine.op1Type == literalOp && srcLine.op2Type == pointerOp {
		opcode |= 0x01
	} else if srcLine.op1Type == addressOp && srcLine.op2Type == addressOp {
		opcode |= 0x02
	} else if srcLine.op1Type == addressOp && srcLine.op2Type == pointerOp {
		opcode |= 0x03
	} else if srcLine.op1Type == pointerOp && srcLine.op2Type == addressOp {
		opcode |= 0x04
	} else if srcLine.op1Type == pointerOp && srcLine.op2Type == pointerOp {
		opcode |= 0x05
	}

	opcodeSrcLine.bin = append(opcodeSrcLine.bin, opcode)

	return opcodeSrcLine
}

// -----------------------------------------------------------------------------

// buildOps creates binary operands from a line of source code.
func buildOps(srcLine srcLine) srcLine {
	opsSrcLine := srcLine

	var data64 uint64

	switch mnemonics[srcLine.mnemonic].instrLength {
	case 3:
		data64, _ = strconv.ParseUint(srcLine.op1, 16, 16)
		opsSrcLine.bin = appendUint16(opsSrcLine.bin, uint16(data64))
	case 5:
		data64, _ = strconv.ParseUint(srcLine.op1, 16, 16)
		opsSrcLine.bin = appendUint16(opsSrcLine.bin, uint16(data64))

		data64, _ = strconv.ParseUint(srcLine.op2, 16, 16)
		opsSrcLine.bin = appendUint16(opsSrcLine.bin, uint16(data64))
	}

	return opsSrcLine
}

// -----------------------------------------------------------------------------

// appendUint16 appends a 16-bit value to a byte slice.
func appendUint16(bin []byte, srcInt uint16) []byte {
	high8, low8 := splitUint16(srcInt)

	bin = append(bin, high8)
	bin = append(bin, low8)

	return bin
}

// -----------------------------------------------------------------------------

// splitUint16 breaks down a 16-bit value to two 8-bit values.
func splitUint16(srcInt uint16) (byte, byte) {
	low8 := byte(srcInt)
	srcInt >>= 8
	high8 := byte(srcInt)

	return high8, low8
}

// -----------------------------------------------------------------------------

// buildBin constructs the final binary executable from the binary data in each
// structured and processed binary line of source code.
func buildBin(srcLines []srcLine, programOffset uint16) []byte {
	bin := binMagicHeader
	bin = appendUint16(bin, programOffset)

	for _, srcLine := range srcLines {
		bin = append(bin, srcLine.bin...)
	}

	return bin
}

// -----------------------------------------------------------------------------

// printBin outputs the final binary for debugging purposes.
func printBin(message string, bin []byte) {
	if DEBUG {
		fmt.Println(message)

		x := 0

		for _, currentByte := range bin {
			fmt.Print(strings.ToUpper(fmt.Sprintf("%02x", currentByte)) + " ")

			x++
			if x%8 == 0 {
				fmt.Println()
			}
		}

		fmt.Println()
	}
}
