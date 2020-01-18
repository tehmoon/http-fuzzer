package main

import (
	"os"
	"regexp"
	"github.com/tehmoon/errors"
	"strconv"
	"io"
	"bufio"
	"fmt"
	"strings"
)

func StringArrayToRegex(arrStr []string) (arrRe []*regexp.Regexp, err error) {
	arrRe = make([]*regexp.Regexp, len(arrStr))

	for i, str := range arrStr {
		arrRe[i], err = regexp.Compile(str)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing %q to regex", str)
		}
	}

	return arrRe, nil
}

func StringArrayToInt(arrStr []string, base, bitSize int) (arrInt []int, err error) {
	arrInt = make([]int, len(arrStr))

	for i, str := range arrStr {
		i64, err := strconv.ParseInt(str, base, bitSize)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing %q to integer", str)
		}

		arrInt[i] = int(i64)
	}

	return arrInt, nil
}

func findIntArray(a []int, i int) bool {
	for _, ii := range a {
		if i == ii {
			return true
		}
	}

	return false
}

func formatIntArray(a []int, base int) (s string) {
	var b strings.Builder

	for _, ii := range a {
		fmt.Fprintf(&b, "%s,", strconv.FormatInt(int64(ii), base))
	}

	// remove last `,`
	s = b.String()
	if len(a) > 0 {
		s = s[:len(s) - 1]
	}

	return s
}

func matchReInBytes(regexes []*regexp.Regexp, data []byte) bool {
	for _, re := range regexes {
		if re.Match(data) {
			return true
		}
	}

	return false
}

func countLines(reader io.Reader) (count int, err error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		count++
	}
	err = scanner.Err()
	if err != nil {
		return 0, errors.Wrap(err, "Error scanning reader for lines")
	}

	return count, nil
}

func countWords(reader io.Reader) (count int, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
	    count++
	}
	err = scanner.Err()
	if err != nil {
		return 0, errors.Wrap(err, "Error scanning reader for words")
	}

	return count, nil
}

func scanLinesFile(p string) (lines []string, err error) {
	file, err := os.Open(p)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening file")
	}

	lines = make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		lines = append(lines, line)
	}
	err = scanner.Err()
	if err != nil {
		return nil, errors.Wrap(err, "Error scanning file for lines")
	}

	return lines, nil
}
