package main

import (
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"strconv"
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

type UrlAddr struct {
	Url         *url.URL
	DefaultPort int
}

func (adr *UrlAddr) Network() string {
	return adr.Url.Scheme
}

func (adr *UrlAddr) String() string {
	if adr.Url.Port() == "" && adr.DefaultPort > 0 {
		return adr.Url.Hostname() + ":" + strconv.Itoa(adr.DefaultPort)
	}
	return adr.Url.Host
}

func (adr *UrlAddr) PortString() string {
	if adr.Url.Port() == "" && adr.DefaultPort > 0 {
		return ":" + strconv.Itoa(adr.DefaultPort)
	}
	return ":" + adr.Url.Port()
}

func ParseUrlAddr(a string) (*UrlAddr, error) {

	re := regexp.MustCompile(`^[a-zA-z][a-zA-Z0-9]*://`)
	if re.FindString(a) == "" {
		a = "//" + a
	}

	urlobj, err := url.Parse(a)
	if err != nil {
		return nil, &net.AddrError{Addr: a, Err: "invald url:" + err.Error()}
	}

	host, port, err := net.SplitHostPort(urlobj.Host)
	if err != nil {
		return nil, err
	}

	if host == "" && port == "" {
		return nil, &net.AddrError{Addr: a, Err: "invalid host or port"}
	}

	return &UrlAddr{Url: urlobj}, nil
}
