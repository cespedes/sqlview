package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

func writeOrgTable(w io.Writer, columns []string, data [][]string) {
	widths := make([]int, len(columns))
	for i, x := range columns {
		widths[i] = utf8.RuneCountInString(x)
	}
	for _, x := range data {
		for i, y := range x {
			if utf8.RuneCountInString(y) > widths[i] {
				widths[i] = utf8.RuneCountInString(y)
			}
		}
	}
	line := fmt.Sprint("|", strings.Repeat("-", widths[0]+2))
	for i := range columns[1:] {
		line += "+" + strings.Repeat("-", widths[i+1]+2)
	}
	line += "|"
	fmt.Fprint(w, line, "\n|")
	for i, x := range columns {
		fmt.Fprintf(w, " %-*s |", widths[i], x)
	}
	fmt.Fprint(w, "\n", line, "\n")
	for _, x := range data {
		fmt.Fprintf(w, "|")
		for i, y := range x {
			fmt.Fprintf(w, " %-*s |", widths[i], y)
		}
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintln(w, line)
}

func readOrgLine(line string) []string {
	s := strings.Split(line, "|")
	if len(s) < 3 {
		return nil
	}
	s = s[1 : len(s)-1]
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	return s
}

func readOrgTable(r io.Reader, columns []string) (data [][]string, err error) {
	lineNo := 0
	s := bufio.NewScanner(r)
	for s.Scan() {
		lineNo++
		if strings.Contains(s.Text(), `|---`) {
			break
		}
	}
	if !s.Scan() {
		return nil, fmt.Errorf("no table found after reading %d lines of text", lineNo)
	}
	lineNo++
	newCols := readOrgLine(s.Text())
	if len(newCols) != len(columns) {
		return nil, fmt.Errorf("wrong header for table in line %d", lineNo)
	}
	if !s.Scan() {
		return nil, fmt.Errorf("no table found after header in line %d", lineNo)
	}
	lineNo++
	if !strings.Contains(s.Text(), `|---`) {
		return nil, fmt.Errorf("wrong table found after header in line %d", lineNo)
	}
	for s.Scan() {
		lineNo++
		line := s.Text()
		if strings.Contains(line, `|---`) {
			break
		}
		s := readOrgLine(s.Text())
		if len(s) != len(columns) {
			return nil, fmt.Errorf("wrong number of columns in line %d", lineNo)
		}
		data = append(data, s)
	}
	return data, nil
}

func writeINI(w io.Writer, columns []string, data [][]string) {
	if len(columns) == 0 {
		panic("writeINI: columns = 0")
	}
	for i, entry := range data {
		if len(entry) != len(columns) {
			panic("writeINI: len(entry) != len(columns)")
		}
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "[%s]\n", entry[0])
		for j := range entry[1:] {
			fmt.Fprintf(w, "%s = %s\n", columns[j+1], entry[j+1])
		}
	}
}

func readINI(r io.Reader, columns []string) (data [][]string, err error) {
	if len(columns) < 1 {
		return nil, fmt.Errorf("no columns to read?")
	}
	lineNo := 0
	s := bufio.NewScanner(r)
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		record := make([]string, len(columns))
		if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
			return nil, fmt.Errorf("wrong section header in line %d", lineNo)
		}
		record[0] = line[1 : len(line)-1]
		for s.Scan() {
			lineNo++
			line = strings.TrimSpace(s.Text())
			if line == "" {
				break
			}
			var key, value string
			if i := strings.Index(line, "="); i >= 0 {
				key = strings.TrimSpace(line[:i])
				value = strings.TrimSpace(line[i+1:])
			} else {
				return nil, fmt.Errorf("syntax error in line %d", lineNo)
			}

			found := false
			for i, c := range columns[1:] {
				if key == c {
					found = true
					record[i+1] = value
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("line %d: unknown key %q", lineNo, key)
			}
		}
		data = append(data, record)
	}
	return data, nil
}

func writeYAML(w io.Writer, columns []string, data [][]string) {
	if len(columns) == 0 {
		panic("writeYAML: columns = 0")
	}
	for i, entry := range data {
		if len(entry) != len(columns) {
			panic("writeYAML: len(entry) != len(columns)")
		}
		if i > 0 {
			fmt.Fprintln(w)
		}
		for j := range entry {
			fmt.Fprintf(w, "%s: %s\n", columns[j], entry[j])
		}
	}
}

func readYAML(r io.Reader, columns []string) (data [][]string, err error) {
	if len(columns) < 1 {
		return nil, fmt.Errorf("no columns to read?")
	}
	lineNo := 0
	line := ""
	s := bufio.NewScanner(r)
	for s.Scan() {
		lineNo++
		line = strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		record := make([]string, len(columns))
		for {
			line = strings.TrimSpace(s.Text())
			if line == "" {
				break
			}
			var key, value string
			if i := strings.Index(line, ":"); i >= 0 {
				key = strings.TrimSpace(line[:i])
				value = strings.TrimSpace(line[i+1:])
			} else {
				return nil, fmt.Errorf("syntax error in line %d", lineNo)
			}

			found := false
			for i, c := range columns {
				if key == c {
					found = true
					record[i] = value
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("line %d: unknown key %q", lineNo, key)
			}

			if s.Scan() == false {
				break
			}
			lineNo++
		}
		data = append(data, record)
	}
	return data, nil
}
