package main

import (
	"fmt"
	"net/url"
	"testing"
)

func TestUrl(t *testing.T) {
	u, e := url.Parse("https://abc.test.com/sdfsdfd")
	fmt.Println(u, e)
	fmt.Println(u.Host)
	fmt.Println(u.String())
}

func TestHttproxy(t *testing.T) {

}
