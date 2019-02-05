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
	"rasm/file"
	"regexp"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------

// Parser token definitions.
const (
	constStartToken string = "["
	incToken        string = "<"
	dataLineToken   string = "$"
)

// Namespace delimiter definition.
const namespaceDlm string = "."

// Built-in, immutable preprocessor constant definitions.
var defaultConsts = map[string]string{
	// Special Addresses
	"[SP]":   "FFB0", // Stack Pointer
	"[IO]":   "FFB2", // Subroutine I/O
	"[PC]":   "FFB4", // Program Counter
	"[ST]":   "FFB6", // CPU Status
	"[UN0]":  "FFB8", // UNDEFINED0
	"[UN1]":  "FFBA", // UNDEFINED1
	"[UN2]":  "FFBC", // UNDEFINED2
	"[UN3]":  "FFBE", // UNDEFINED3
	"[IRQ0]": "FFC0", // INTERRUPT0
	"[IRQ1]": "FFC2", // INTERRUPT1
	"[IRQ2]": "FFC4", // INTERRUPT2
	"[IRQ3]": "FFC6", // INTERRUPT3
	"[IRQ4]": "FFC8", // INTERRUPT4
	"[IRQ5]": "FFCA", // INTERRUPT5
	"[IRQ6]": "FFCC", // INTERRUPT6
	"[IRQ7]": "FFCE", // INTERRUPT7
	"[IN0]":  "FFD0", // INPUT0
	"[IN1]":  "FFD2", // INPUT1
	"[IN2]":  "FFD4", // INPUT2
	"[IN3]":  "FFD6", // INPUT3
	"[IN4]":  "FFD8", // INPUT4
	"[IN5]":  "FFDA", // INPUT5
	"[IN6]":  "FFDC", // INPUT6
	"[IN7]":  "FFDE", // INPUT7
	"[OUT0]": "FFE0", // OUTPUT0
	"[OUT1]": "FFE2", // OUTPUT1
	"[OUT2]": "FFE4", // OUTPUT2
	"[OUT3]": "FFE6", // OUTPUT3
	"[OUT4]": "FFE8", // OUTPUT4
	"[OUT5]": "FFEA", // OUTPUT5
	"[OUT6]": "FFEC", // OUTPUT6
	"[OUT7]": "FFEE", // OUTPUT7
	"[GP0]":  "FFF0", // GENERAL0
	"[GP0L]": "FFF1", // GENERAL0 LOW BYTE
	"[GP1]":  "FFF2", // GENERAL1
	"[GP1L]": "FFF3", // GENERAL1 LOW BYTE
	"[GP2]":  "FFF4", // GENERAL2
	"[GP2L]": "FFF5", // GENERAL2 LOW BYTE
	"[GP3]":  "FFF6", // GENERAL3
	"[GP3L]": "FFF7", // GENERAL3 LOW BYTE
	"[GP4]":  "FFF8", // GENERAL4
	"[GP4L]": "FFF9", // GENERAL4 LOW BYTE
	"[GP5]":  "FFFA", // GENERAL5
	"[GP5L]": "FFFB", // GENERAL5 LOW BYTE
	"[GP6]":  "FFFC", // GENERAL6
	"[GP6L]": "FFFD", // GENERAL6 LOW BYTE
	"[GP7]":  "FFFE", // GENERAL7
	"[GP7L]": "FFFF", // GENERAL7 LOW BYTE

	// Magic Values
	"[TRUE]":  "0001",
	"[FALSE]": "FFFF",
	"[NULL]":  "0000",
}

// -----------------------------------------------------------------------------

// cleanSrc removes comments and extraneous whitespace from the source code.
func cleanSrc(srcLines []string) []string {
	var cleanSrcLines []string

	reComments := regexp.MustCompile("#.*")
	reDoubleSpace := regexp.MustCompile(`[\s\p{Zs}]{2,}`)

	for _, srcLine := range srcLines {
		cleanLine := reComments.ReplaceAllLiteralString(srcLine, "")
		cleanLine = reDoubleSpace.ReplaceAllLiteralString(cleanLine, " ")
		cleanLine = strings.TrimSpace(cleanLine)

		cleanSrcLines = append(cleanSrcLines, cleanLine)
	}

	return cleanSrcLines
}

// -----------------------------------------------------------------------------

// expandConsts translates preprocessor constants to their values.
func expandConsts(srcLines []string) ([]string, error) {
	var expandedSrcLines []string
	var expandedLine string

	expandedConsts, err := getConsts(srcLines)
	if err != nil {
		return nil, err
	}

	reConstName := regexp.MustCompile(`\[.+\]`)

	for lineNum, srcLine := range srcLines {
		expandedLine = srcLine

		if srcLine != "" {
			if srcLine[:1] != constStartToken {
				for constName, constValue := range expandedConsts {
					expandedLine = strings.Replace(expandedLine, constName, constValue, -1)
				}

				foundUnmatched := reConstName.FindString(expandedLine)
				if foundUnmatched != "" {
					return nil, errors.New(strconv.Itoa(lineNum+1) + ": Preprocessor constant " + foundUnmatched + " not defined")
				}
			} else {
				expandedLine = ""
			}
		}

		expandedSrcLines = append(expandedSrcLines, expandedLine)
	}

	return expandedSrcLines, nil
}

// -----------------------------------------------------------------------------

// getConsts finds non-default preprocessor constants in the source code.
func getConsts(srcLines []string) (map[string]string, error) {
	consts := defaultConsts

	reConstName := regexp.MustCompile(`\[.+\]`)
	reConstValue := regexp.MustCompile(`\].+`)

	for lineNum, srcLine := range srcLines {
		if srcLine != "" && srcLine[:1] == "[" {
			constName := reConstName.FindString(srcLine)

			if _, exists := consts[constName]; exists {
				return nil, errors.New(strconv.Itoa(lineNum+1) + ": Cannot redefine preprocessor constant " + constName)
			}

			constValue := reConstValue.FindString(srcLine)
			constValue = constValue[1:]
			constValue = strings.TrimSpace(constValue)

			consts[constName] = constValue
		}
	}

	if DEBUG {
		fmt.Print("Preprocessor constants: ")
		fmt.Println(consts)
	}

	return consts, nil
}

// -----------------------------------------------------------------------------

// addSrcLabelNamespaces prefixes source code labels with namespaces based on
// the source/include file they occur in.
func addSrcLabelNamespaces(srcLines []string, srcName string) []string {
	namespace := strings.SplitN(srcName, namespaceDlm, 2)[0]

	var namespacedSrcLines []string

	reSrcLabel := regexp.MustCompile(`([\w.]{5,})`)

	for _, srcLine := range srcLines {
		namespacedLine := srcLine

		if srcLine != "" && srcLine[:1] != incToken && srcLine[:1] != dataLineToken {
			namespacedLine = reSrcLabel.ReplaceAllStringFunc(srcLine, func(s string) string {
				if strings.Contains(s, namespaceDlm) {
					return s
				}

				return namespace + namespaceDlm + s
			})
		}

		namespacedSrcLines = append(namespacedSrcLines, namespacedLine)
	}

	return namespacedSrcLines
}

// -----------------------------------------------------------------------------

// addIncludes reads rasm include files referenced in the main source file,
// processes them and returns the final, complete source code.
func addIncludes(srcLines []string) ([]string, error) {
	var allSrcLines []string

	for _, srcLine := range srcLines {
		if srcLine != "" && srcLine[:1] == incToken {
			incName := srcLine[1:]
			incName = strings.TrimSpace(incName)
			incName += file.IncExt

			rawIncLines, err := file.ReadSrc(incName)
			if err != nil {
				return nil, err
			}
			printSrc("", rawIncLines)

			rawIncLines = cleanSrc(rawIncLines)
			printSrc("Removed comments and extraneous whitespace", rawIncLines)

			rawIncLines, err = expandConsts(rawIncLines)
			if err != nil {
				return nil, err
			}
			printSrc("Expanded preprocessor constants", rawIncLines)

			if hasInclude(rawIncLines) {
				return nil, errors.New("Inc file " + incName + " cannot contain includes of its own")
			}

			rawIncLines = addSrcLabelNamespaces(rawIncLines, incName)
			printSrc("Added label namespaces", rawIncLines)

			allSrcLines = append(allSrcLines, rawIncLines...)
		} else {
			allSrcLines = append(allSrcLines, srcLine)
		}
	}

	return allSrcLines, nil
}

// -----------------------------------------------------------------------------

// hasInclude checks whether an include file contains include files of its own.
func hasInclude(srcLines []string) bool {
	for _, srcLine := range srcLines {
		if srcLine != "" && srcLine[:1] == incToken {
			return true
		}
	}

	return false
}

// -----------------------------------------------------------------------------

// hasDupeSrcLabels checks whether the source code contains duplicate labels.
func hasDupeSrcLabels(srcLines []string) (bool, string, int) {
	srcLabels := make(map[string]bool)

	for lineNum, srcLine := range srcLines {
		if srcLine != "" && isSrcLabel(srcLine) {
			if _, exists := srcLabels[srcLine]; exists {
				return true, srcLine, lineNum
			}

			srcLabels[srcLine] = true
		}
	}

	return false, "", 0
}

// -----------------------------------------------------------------------------

// printSrc prints unstructured source code for debugging purposes.
func printSrc(message string, srcLines []string) {
	if DEBUG {
		fmt.Println(message)

		for lineNum, srcLine := range srcLines {
			fmt.Print(lineNum + 1)
			fmt.Println("\t" + srcLine)
		}
	}
}
