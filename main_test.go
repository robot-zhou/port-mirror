package main

import (
	"net"
	"testing"
)

func TestRandomString(t *testing.T) {
	t.Logf("random string: %s", RandomString())
}

func TestDailTarget(t *testing.T) {
	_, err := net.Dial("tcp", "baidu.com:80")
	t.Logf("err: %v", err)
}
