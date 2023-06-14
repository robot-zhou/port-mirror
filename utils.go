package main

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

func RandomString(a ...int) string {
	const (
		strtable = "012345678901234567890123456789abcdefghijklmnopqrstuvwxyz"
	)

	var (
		n         = append(a, 8)[0]
		table_len = len(strtable)
		s         string
	)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		m := r.Int() % table_len
		s += strtable[m : m+1]
	}

	return s
}

func SplitString(a string, f ...string) []string {
	re := regexp.MustCompile(append(f, `[,;]`)[0])
	return re.Split(a, -1)
}

func SplitStringTrim(a string, f ...string) []string {
	var res []string

	tmp := SplitString(a, f...)
	for _, v := range tmp {
		v = strings.Trim(v, " \t")
		if v != "" {
			res = append(res, v)
		}
	}

	return res
}
