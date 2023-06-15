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

type Addr struct {
	Name string
	Host string
	Port int
}

func (adr *Addr) Network() string {
	if adr == nil {
		return "<nil>"
	}
	return adr.Name
}

func (adr *Addr) String() string {
	if adr == nil {
		return "<nil>"
	}

	return net.JoinHostPort(adr.Host, strconv.Itoa(adr.Port))
}

func (adr *Addr) PortString() string {
	return ":" + strconv.Itoa(adr.Port)
}

func Url2Addr(a string, defNetwork ...string) (*Addr, error) {
	u, err := url.Parse(a)
	if err != nil {
		return nil, err
	}

	host, portstr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portstr)
	if err != nil {
		return nil, &net.AddrError{Addr: u.Host, Err: "invalid port number"}
	}

	adr := &Addr{Name: u.Scheme, Host: host, Port: port}
	if adr.Name == "" && len(defNetwork) > 0 {
		adr.Name = defNetwork[0]
	}

	return adr, nil
}
