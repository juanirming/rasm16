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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------

// Maximum address space limit.
const maxAddressSpace int = 0xFFB0 - 2*128 // Leaving space for call stack below Stack Pointer.

// Parser token definitions.
const (
	srcStringToken       string = `"`
	nullRepeatStartToken string = "("
	nullRepeatEndToken   string = ")"
)

// Data directive value delimiter definition.
const dataDlm string = ","

type directiveType int

// Data directive type definitions.
const (
	invalidDirective   directiveType = 0
	data8BitDirective  directiveType = 1
	data16BitDirective directiveType = 2
)

// Data directive token definitions.
var directiveTokens = map[directiveType]string{
	data8BitDirective:  "$8",
	data16BitDirective: "$16",
}

type opType int

// Operand type definitions.
const (
	invalidOp opType = 0
	addressOp opType = 1
	literalOp opType = 2
	pointerOp opType = 3
)

// Parser operand token definitions.
var opTokens = map[opType]string{
	literalOp: "$",
	pointerOp: "*",
}

// Human-readable operand descriptions.
var opDescr = map[opType]string{
	literalOp: "LITERAL",
	pointerOp: "POINTER",
}

// Human-readable operand description.
var defaultOpDescr string = "ADDRESS"

// Mnemonic alias definitions.
var mnemonicAliases = map[string]string{
	// Instructions
	"CO": "CO16",
	"AD": "AD16",
	"SU": "SU16",
	"MU": "MU16",
	"DV": "DV16",
	"ND": "ND16",
	"OR": "OR16",
	"XR": "XR16",
	"SL": "SL16",
	"SR": "SR16",
	"CM": "CM16",

	// Directives
	"$": "$16",
}

// Mnemonic type definition.
type mnemonic struct {
	descr       string
	opcode      byte
	numOps      int
	instrLength int
}

// Mnemonic definitions (instructions and directives).
var mnemonics = map[string]mnemonic{
	// Instructions
	"NO":   {descr: "NO OPERATION", opcode: 0x00, numOps: 0, instrLength: 1},
	"CO8":  {descr: "COPY", opcode: 0x01, numOps: 2, instrLength: 5},
	"CO16": {descr: "COPY", opcode: 0x02, numOps: 2, instrLength: 5},
	"AD8":  {descr: "ADD", opcode: 0x03, numOps: 2, instrLength: 5},
	"AD16": {descr: "ADD", opcode: 0x04, numOps: 2, instrLength: 5},
	"SU8":  {descr: "SUBTRACT", opcode: 0x05, numOps: 2, instrLength: 5},
	"SU16": {descr: "SUBTRACT", opcode: 0x06, numOps: 2, instrLength: 5},
	"MU8":  {descr: "MULTIPLY", opcode: 0x07, numOps: 2, instrLength: 5},
	"MU16": {descr: "MULTIPLY", opcode: 0x08, numOps: 2, instrLength: 5},
	"DV8":  {descr: "DIVIDE", opcode: 0x09, numOps: 2, instrLength: 5},
	"DV16": {descr: "DIVIDE", opcode: 0x0A, numOps: 2, instrLength: 5},
	"ND8":  {descr: "BITWISE AND", opcode: 0x0B, numOps: 2, instrLength: 5},
	"ND16": {descr: "BITWISE AND", opcode: 0x0C, numOps: 2, instrLength: 5},
	"OR8":  {descr: "BITWISE OR", opcode: 0x0D, numOps: 2, instrLength: 5},
	"OR16": {descr: "BITWISE OR", opcode: 0x0E, numOps: 2, instrLength: 5},
	"XR8":  {descr: "BITWISE XOR", opcode: 0x0F, numOps: 2, instrLength: 5},
	"XR16": {descr: "BITWISE XOR", opcode: 0x10, numOps: 2, instrLength: 5},
	"SL8":  {descr: "BITWISE SHIFT LEFT", opcode: 0x11, numOps: 2, instrLength: 5},
	"SL16": {descr: "BITWISE SHIFT LEFT", opcode: 0x12, numOps: 2, instrLength: 5},
	"SR8":  {descr: "BITWISE SHIFT RIGHT", opcode: 0x13, numOps: 2, instrLength: 5},
	"SR16": {descr: "BITWISE SHIFT RIGHT", opcode: 0x15, numOps: 2, instrLength: 5},
	"CM8":  {descr: "COMPARE", opcode: 0x15, numOps: 2, instrLength: 5},
	"CM16": {descr: "COMPARE", opcode: 0x16, numOps: 2, instrLength: 5},
	"EQ":   {descr: "JUMP IF EQUAL", opcode: 0x17, numOps: 1, instrLength: 3},
	"NE":   {descr: "JUMP IF NOT EQUAL", opcode: 0x18, numOps: 1, instrLength: 3},
	"LT":   {descr: "JUMP IF LESS THAN", opcode: 0x19, numOps: 1, instrLength: 3},
	"GT":   {descr: "JUMP IF GREATER THAN", opcode: 0x1A, numOps: 1, instrLength: 3},
	"EL":   {descr: "JUMP IF EQUAL OR LESS THAN", opcode: 0x1B, numOps: 1, instrLength: 3},
	"EG":   {descr: "JUMP IF EQUAL OR GREATER THAN", opcode: 0x1C, numOps: 1, instrLength: 3},
	"JM":   {descr: "JUMP", opcode: 0x1D, numOps: 0, instrLength: 1},
	"JS":   {descr: "JUMP TO SUBROUTINE", opcode: 0x1E, numOps: 2, instrLength: 5},
	"RT":   {descr: "RETURN", opcode: 0x1F, numOps: 1, instrLength: 3},

	// Directives
	"$8":  {descr: "DATA DIRECTIVE", opcode: 0x00, numOps: 0, instrLength: 0},
	"$16": {descr: "DATA DIRECTIVE", opcode: 0x00, numOps: 0, instrLength: 0},
}

// -----------------------------------------------------------------------------

// unaliasMnemonics replaces mnemonic aliases with their corresponding base
// mnemonics.
func unaliasMnemonics(srcLines []srcLine) []srcLine {
	var unaliasedSrcLines []srcLine

	for _, srcLine := range srcLines {
		currentSrcLine := srcLine

		if _, exists := mnemonicAliases[currentSrcLine.mnemonic]; exists {
			currentSrcLine.mnemonic = mnemonicAliases[srcLine.mnemonic]
		}

		unaliasedSrcLines = append(unaliasedSrcLines, currentSrcLine)
	}

	return unaliasedSrcLines
}

// -----------------------------------------------------------------------------

// expandDataNullRepeats translates data directive null repeat syntax to full
// data directive value lists.
func expandDataNullRepeats(srcLines []srcLine) ([]srcLine, error) {
	var expandedSrcLines []srcLine

	for _, srcLine := range srcLines {
		currentSrcLine := srcLine

		if isValidDataDirective(srcLine.mnemonic) && isDataNullRepeat(srcLine.data) {
			num_repeats, err := strconv.Atoi(srcLine.data[1 : len(srcLine.data)-1])
			if err != nil {
				return nil, errors.New(strconv.Itoa(srcLine.lineNum+1) + ":\tInvalid data null repeat " + srcLine.data)
			}

			currentSrcLine.data = ""
			for i := 1; i < num_repeats; i++ {
				currentSrcLine.data += defaultConsts["[NULL]"] + dataDlm
			}
			currentSrcLine.data += defaultConsts["[NULL]"]
		}

		expandedSrcLines = append(expandedSrcLines, currentSrcLine)
	}

	return expandedSrcLines, nil
}

// -----------------------------------------------------------------------------

// isDataNullRepeat checks whether a data directive contains null repeat syntax.
func isDataNullRepeat(data string) bool {
	return len(data) > 0 &&
		data[:1] == nullRepeatStartToken &&
		data[len(data)-1:] == nullRepeatEndToken
}

// -----------------------------------------------------------------------------

// convDataStringsToHex converts data directive strings to value lists.
func convDataStringsToHex(srcLines []srcLine) []srcLine {
	var convSrcLines []srcLine

	for _, srcLine := range srcLines {
		currentSrcLine := srcLine

		if srcLine.mnemonic == directiveTokens[data8BitDirective] && isDataString(srcLine.data) {
			currentSrcLine.data = dataStringToHex(srcLine.data)
		}

		convSrcLines = append(convSrcLines, currentSrcLine)
	}

	return convSrcLines
}

// -----------------------------------------------------------------------------

// isDataString checks whether a data directive contains a string.
func isDataString(data string) bool {
	return len(data) > 0 &&
		data[:1] == srcStringToken &&
		data[len(data)-1:] == srcStringToken
}

// -----------------------------------------------------------------------------

// dataStringToHex converts a single data directive string to a value list.
func dataStringToHex(dataString string) string {
	bytes := []byte(dataString)

	var hex []string
	for _, byte := range bytes {
		hex = append(hex, strings.ToUpper(fmt.Sprintf("%x", byte)))
	}

	hexData := strings.Join(hex, dataDlm)

	return hexData
}

// -----------------------------------------------------------------------------

// validateMnemonics checks whether any invalid mnemonics exist.
func validateMnemonics(srcLines []srcLine) (bool, error) {
	for _, srcLine := range srcLines {
		if _, exists := mnemonics[srcLine.mnemonic]; !exists {
			return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + ":\tInvalid mnemonic " + srcLine.mnemonic)
		}
	}

	return true, nil
}

// -----------------------------------------------------------------------------

// validateDataDirectives checks whether any invalid data directives exist.
func validateDataDirectives(srcLines []srcLine) (bool, error) {
	errMessageStart := ":\tInvalid "
	errMessageEnd := "-bit data in directive"

	for _, srcLine := range srcLines {
		if srcLine.mnemonic == directiveTokens[data8BitDirective] {
			splitData := strings.Split(srcLine.data, dataDlm)

			if !is8BitHexStrings(splitData) {
				return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + errMessageStart + "8" + errMessageEnd)
			}
		} else if srcLine.mnemonic == directiveTokens[data16BitDirective] {
			splitData := strings.Split(srcLine.data, dataDlm)

			if !is16BitHexStrings(splitData) {
				return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + errMessageStart + "16" + errMessageEnd)
			}
		}
	}

	return true, nil
}

// -----------------------------------------------------------------------------

// is16BitHexStrings checks whether a string slice contains valid 16-bit
// hexadecimal values.
func is16BitHexStrings(hex []string) bool {
	for _, data := range hex {
		if !is16BitHexString(data) {
			return false
		}
	}

	return true
}

// -----------------------------------------------------------------------------

// is16BitHexString checks whether a string is a valid 16-bit hexadecimal value.
func is16BitHexString(hex string) bool {
	re16BitHex := regexp.MustCompile(`^[0-9A-Fa-f]{1,4}$`)

	return re16BitHex.MatchString(hex)
}

// -----------------------------------------------------------------------------

// is8BitHexStrings checks whether a string slice contains valid 8-bit
// hexadecimal values.
func is8BitHexStrings(hex []string) bool {
	for _, data := range hex {
		if !is8BitHexString(data) {
			return false
		}
	}

	return true
}

// -----------------------------------------------------------------------------

// is8BitHexString checks whether a string is a valid 8-bit hexadecimal value.
func is8BitHexString(hex string) bool {
	re8BitHex := regexp.MustCompile(`^[0-9A-Fa-f]{1,2}$`)

	return re8BitHex.MatchString(hex)
}

// -----------------------------------------------------------------------------

// calcAddresses calculates the address for each instruction/directive based
// on the program offset and instruction/data lengths.
func calcAddresses(srcLines []srcLine, programOffset uint16) ([]srcLine, error) {
	var addressSrcLines []srcLine

	programCounter := int(programOffset)

	for _, srcLine := range srcLines {
		currentSrcLine := srcLine

		currentSrcLine.address = programCounter

		addressSrcLines = append(addressSrcLines, currentSrcLine)

		if isValidDataDirective(srcLine.mnemonic) {
			splitData := strings.Split(srcLine.data, dataDlm)

			if srcLine.mnemonic == directiveTokens[data8BitDirective] {
				programCounter += len(splitData)
			} else if srcLine.mnemonic == directiveTokens[data16BitDirective] {
				programCounter += 2 * len(splitData)
			}
		} else {
			programCounter += mnemonics[srcLine.mnemonic].instrLength
		}

		if programCounter >= maxAddressSpace {
			return nil, errors.New(strconv.Itoa(srcLine.lineNum+1) + ":\tAddress out of range")
		}
	}

	return addressSrcLines, nil
}

// -----------------------------------------------------------------------------

// isValidDataDirective checks whether a mnemonic is a data directive.
func isValidDataDirective(mnemonic string) bool {
	return mnemonic == directiveTokens[data8BitDirective] || mnemonic == directiveTokens[data16BitDirective]
}

// -----------------------------------------------------------------------------

// getLabelAddresses finds all source labels and returns them along with their
// addresses.
func getLabelAddresses(srcLines []srcLine) map[string]int {
	labelAddresses := make(map[string]int)

	for _, srcLine := range srcLines {
		if srcLine.label != "" {
			labelAddresses[srcLine.label] = srcLine.address
		}
	}

	return labelAddresses
}

// -----------------------------------------------------------------------------

// expandLabels translates source labels into final addresses.
func expandLabels(srcLines []srcLine, labelAddresses map[string]int) ([]srcLine, error) {
	var expandedSrcLines []srcLine

	errMessageStart := ":\tLabel "
	errMessageEnd := " not defined"

	for _, srcLine := range srcLines {
		currentSrcLine := srcLine

		if srcLine.op1 != "" {
			op1Label := getOpLabel(srcLine.op1)

			if op1Label != "" {
				if _, exists := labelAddresses[op1Label]; exists {
					currentSrcLine.op1 = strings.Replace(currentSrcLine.op1, op1Label, strings.ToUpper(fmt.Sprintf("%04x", labelAddresses[op1Label])), 1)
				} else {
					return nil, errors.New(strconv.Itoa(srcLine.lineNum+1) + errMessageStart + op1Label + errMessageEnd)
				}
			}
		}

		if srcLine.op2 != "" {
			op2Label := getOpLabel(srcLine.op2)

			if op2Label != "" {
				if _, exists := labelAddresses[op2Label]; exists {
					currentSrcLine.op2 = strings.Replace(currentSrcLine.op2, op2Label, strings.ToUpper(fmt.Sprintf("%04x", labelAddresses[op2Label])), 1)
				} else {
					return nil, errors.New(strconv.Itoa(srcLine.lineNum+1) + errMessageStart + op2Label + errMessageEnd)
				}
			}
		}

		expandedSrcLines = append(expandedSrcLines, currentSrcLine)
	}

	return expandedSrcLines, nil
}

// -----------------------------------------------------------------------------

// getOpLabel finds a source label in an operand.
func getOpLabel(op string) string {
	reSrcLabel := regexp.MustCompile(`([\w.]{5,})`)
	label := reSrcLabel.FindString(op)

	return label
}

// -----------------------------------------------------------------------------

// validateOps checks whether any erroneous operands exist.
func validateOps(srcLines []srcLine) (bool, error) {
	errMessage := ":\tInvalid operand "

	for _, srcLine := range srcLines {
		if !isValidDataDirective(srcLine.mnemonic) {
			switch mnemonics[srcLine.mnemonic].numOps {
			case 0:
				if srcLine.op1 != "" || srcLine.op2 != "" {
					return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + ":\t" + srcLine.mnemonic + " needs no operands")
				}
			case 1:
				if srcLine.op1 == "" || srcLine.op2 != "" {
					return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + ":\t" + srcLine.mnemonic + " needs one operand")
				}
			case 2:
				if srcLine.op1 == "" || srcLine.op2 == "" {
					return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + ":\t" + srcLine.mnemonic + " needs two operands")
				}
			}

			if srcLine.op1 != "" && !isValidHexString(srcLine.op1) {
				return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + errMessage + srcLine.op1)
			}

			if srcLine.op2 != "" && !isValidHexString(srcLine.op2) {
				return false, errors.New(strconv.Itoa(srcLine.lineNum+1) + errMessage + srcLine.op2)
			}
		}
	}

	return true, nil
}

// -----------------------------------------------------------------------------

// isValidHexString checks whether a string is a valid hexadecimal number.
func isValidHexString(hex string) bool {
	if is16BitHexString(hex) || is8BitHexString(hex) {
		return true
	}

	return false
}

// -----------------------------------------------------------------------------

// printStructSrc prints out structured source code for debugging purposes.
func printStructSrc(message string, srcLines []srcLine) {
	if DEBUG {
		fmt.Println(message)

		for _, srcLine := range srcLines {
			fmt.Print(srcLine.lineNum)
			fmt.Print("\t")
			fmt.Print(strings.ToUpper(fmt.Sprintf("%04x", srcLine.address)))
			fmt.Print("\t")
			if srcLine.label != "" {
				fmt.Println(srcLine.label)
				fmt.Print("\t\t")
			}
			fmt.Print(srcLine.mnemonic + "\t")
			if srcLine.op1 != "" {
				if srcLine.op1Type != invalidOp {
					fmt.Print("(" + getOpDescr(srcLine.op1Type) + ")")
				}
				fmt.Print(srcLine.op1)

				if srcLine.op2 != "" {
					fmt.Print(opDlm)
					if srcLine.op2Type != invalidOp {
						fmt.Print("(" + getOpDescr(srcLine.op2Type) + ")")
					}
					fmt.Print(srcLine.op2)
				}
			} else {
				fmt.Print(srcLine.data)
			}
			if len(srcLine.bin) > 0 {
				fmt.Print(" -> ")

				for _, currentByte := range srcLine.bin {
					fmt.Print(strings.ToUpper(fmt.Sprintf("%02x", currentByte)) + " ")
				}
			}
			fmt.Println()
		}
	}
}

// -----------------------------------------------------------------------------

// getOpDescr returns a human-readable operand type description.
func getOpDescr(op opType) string {
	if _, exists := opDescr[op]; exists {
		return opDescr[op]
	}

	return defaultOpDescr
}
