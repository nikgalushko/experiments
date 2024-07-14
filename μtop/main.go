package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unicode"

	"github.com/rodaine/table"
)

var pagesize int

func main() {
	pagesize = syscall.Getpagesize()

	dirs, err := os.ReadDir("/proc")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read /proc: %s", err)
		os.Exit(1)
	}

	total := 0
	tbl := table.New("PID", "VIRT", "RES")

	for _, d := range dirs {
		if n := d.Name(); isPID(n) {
			total++
			mem, err := statm(n)
			if err != nil {
				fmt.Fprintf(os.Stderr, "get memory stat for %s: %s", n, err)
				continue
			}
			tbl.AddRow(n, prettyBytes(mem.size*pagesize), prettyBytes(mem.resident*pagesize))
		}
	}

	fmt.Println("\nTotal:", total)
	tbl.Print()
}

func isPID(n string) bool {
	for _, r := range []rune(n) {
		if !unicode.IsNumber(r) {
			return false
		}
	}

	return true
}

type memory struct {
	size     int
	resident int
}

func statm(pid string) (memory, error) {
	ret := memory{}

	data, err := os.ReadFile(fmt.Sprintf("/proc/%s/statm", pid))
	if err != nil {
		return ret, fmt.Errorf("open statm file: %s", err)
	}

	start := -1
	part := 0
	for i := 0; i < len(data); i++ {
		if isNumber(data[i]) {
			if start == -1 {
				start = i
			}
			continue
		} else if start == -1 {
			continue
		} else {
			v, err := strconv.Atoi(string(data[start:i]))
			if err != nil {
				return ret, fmt.Errorf("parse statm file: %s", err)
			}

			if part == 0 {
				ret.size = v
			} else if part == 1 {
				ret.resident = v
			}
			part++

			if part == 2 {
				break
			}
			start = -1
		}
	}

	return ret, nil
}

func isNumber(b byte) bool {
	return b >= '0' && b <= '9'
}

func prettyBytes(n int) string {
	p := "B"
	if n >= 1024 && n < 1048576 {
		n = n / 1024
		p = "Kb"
	} else if n >= 1048576 {
		n = (n / 1024) / 1024
		p = "Mb"
	}

	return fmt.Sprintf("%d%s", n, p)
}
