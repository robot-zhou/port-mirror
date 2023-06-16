package main

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/proxy"
)

func TestHttproxy(t *testing.T) {
	u, _ := url.Parse("http://http:http123@localhost:6080")
	dailer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		t.Errorf("create dailer fail: %s", err)
		return
	}

	conn, err := dailer.Dial("tcp", "www.google.com:80")
	if err != nil {
		t.Errorf("connect fail: %s", err)
		return
	}

	defer conn.Close()
	n, err := conn.Write([]byte("GET / HTTP/1.1\r\n\r\n"))
	if err != nil {
		t.Errorf("write fail: %s", err.Error())
		return
	}

	t.Logf("write count: %d", n)
	var buf [1024]byte

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		n, err := conn.Read(buf[:])
		if n > 0 {
			fmt.Print(string(buf[0:n]))
		}
		if err != nil {
			t.Logf("read err: %v", err)
			break
		}
	}
}
