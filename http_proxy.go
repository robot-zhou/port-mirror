package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

func dialContext(ctx context.Context, d proxy.Dialer, network, address string) (net.Conn, error) {
	var (
		conn net.Conn
		done = make(chan struct{}, 1)
		err  error
	)
	go func() {
		conn, err = d.Dial(network, address)
		close(done)
		if conn != nil && ctx.Err() != nil {
			conn.Close()
		}
	}()
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-done:
	}
	return conn, err
}

type HttpProxyDailer struct {
	proxyAddr *UrlAddr
	haveSSL   bool
	haveAuth  bool
	username  string
	password  string

	forwardDail func(context.Context, string, string) (net.Conn, error)
}

func newHttpProxyDailer(uri *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	d := new(HttpProxyDailer)

	host := uri.Hostname()
	port := uri.Port()

	if uri.Scheme == "https" {
		d.haveSSL = true
		if port == "" {
			port = "443"
		}
	} else {
		if port == "" {
			port = "80"
		}
	}

	d.proxyAddr, _ = ParseUrlAddr(uri.Scheme + "://" + host + ":" + port)

	if forward != nil {
		if f, ok := forward.(proxy.ContextDialer); ok {
			d.forwardDail = func(ctx context.Context, network string, address string) (net.Conn, error) {
				return f.DialContext(ctx, network, address)
			}
		} else {
			d.forwardDail = func(ctx context.Context, network string, address string) (net.Conn, error) {
				return dialContext(ctx, forward, network, address)
			}
		}
	}

	if uri.User != nil {
		d.haveAuth = true
		d.username = uri.User.Username()
		d.password, _ = uri.User.Password()
	}

	return d, nil
}

func (d *HttpProxyDailer) connect(c net.Conn, address string) (net.Conn, error) {
	req, err := http.NewRequest(http.MethodConnect, "//"+address, nil)
	if err != nil {
		c.Close()
		return nil, err
	}

	req.Close = false
	req.Header.Set("User-Agent", "port-mirror")
	if d.haveAuth {
		req.SetBasicAuth(d.username, d.password)
		req.Header.Set("Proxy-Authorization", req.Header.Get("Authorization"))
	}

	if err = req.Write(c); err != nil {
		c.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(c), req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		c.Close()
		return nil, err
	}

	if resp == nil {
		c.Close()
		return nil, errors.New("connect with proxy error: invalid response")
	}

	if resp.StatusCode != 200 {
		c.Close()
		return nil, fmt.Errorf("connect with proxy error, StatusCode(%d)", resp.StatusCode)
	}

	return c, nil
}

func (d *HttpProxyDailer) connectWithContext(ctx context.Context, c net.Conn, address string) (net.Conn, error) {
	var (
		conn net.Conn
		done = make(chan struct{}, 1)
		err  error
	)

	if deadline, ok := ctx.Deadline(); ok && !deadline.IsZero() {
		c.SetDeadline(deadline)
		defer c.SetDeadline(time.Time{})
	}

	go func() {
		conn, err = d.connect(c, address)
		close(done)
		if conn != nil && ctx.Err() != nil {
			conn.Close()
		}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-done:
	}

	return conn, err
}

func (d *HttpProxyDailer) validateTarget(network, address string) (net.Addr, error) {
	switch network {
	case "tcp", "tcp6", "tcp4":
	default:
		return nil, errors.New("network not implemented")
	}

	return ParseUrlAddr(address)
}

func (d *HttpProxyDailer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var (
		err       error
		targetAdr net.Addr
		con       net.Conn
	)

	if targetAdr, err = d.validateTarget(network, address); err != nil {
		return nil, &net.OpError{Op: "http proxy", Net: network, Source: d.proxyAddr, Addr: targetAdr, Err: err}
	}

	if d.forwardDail != nil {
		con, err = d.forwardDail(ctx, "tcp", d.proxyAddr.String())
	} else {
		dd := proxy.Direct
		con, err = dd.DialContext(ctx, "tcp", d.proxyAddr.String())
	}

	if err != nil {
		return nil, &net.OpError{Op: "http proxy", Net: network, Source: d.proxyAddr, Addr: targetAdr, Err: err}
	}

	if d.haveSSL {
		sslconn := tls.Client(con, &tls.Config{InsecureSkipVerify: true})
		if err := sslconn.HandshakeContext(ctx); err != nil {
			con.Close()
			return nil, &net.OpError{Op: "https proxy", Net: network, Source: d.proxyAddr, Addr: targetAdr, Err: err}
		}
		con = sslconn
	}

	return d.connectWithContext(ctx, con, targetAdr.String())
}

func (d *HttpProxyDailer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.TODO(), network, address)
}

func init() {
	proxy.RegisterDialerType("http", newHttpProxyDailer)
	proxy.RegisterDialerType("https", newHttpProxyDailer)
}
