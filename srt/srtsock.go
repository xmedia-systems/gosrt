// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/xmedia-systems/gosrt/srtapi"
)

// SRTAddr represents the address of a SRT end point.
type SRTAddr struct {
	IP   net.IP
	Port int
	Zone string // IPv6 scoped addressing zone
}

// Network returns the address's network name, "srt".
func (a *SRTAddr) Network() string { return "srt" }

func (a *SRTAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	ip := ipEmptyString(a.IP)
	if a.Zone != "" {
		return net.JoinHostPort(ip+"%"+a.Zone, itoa(a.Port))
	}
	return net.JoinHostPort(ip, itoa(a.Port))
}

func (a *SRTAddr) isWildcard() bool {
	if a == nil || a.IP == nil {
		return true
	}
	return a.IP.IsUnspecified()
}

func (a *SRTAddr) opAddr() net.Addr {
	if a == nil {
		return nil
	}
	return a
}

func (a *SRTAddr) matchAddrFamily(x net.IP) bool {
	return a.IP.To4() != nil && x.To4() != nil || a.IP.To16() != nil && a.IP.To4() == nil && x.To16() != nil && x.To4() == nil
}

// ResolveSRTAddr returns an address of SRT end point.
//
// The network must be a SRT network name.
//
// If the host in the address parameter is not a literal IP address or
// the port is not a literal port number, ResolveSRTAddr resolves the
// address to an address of SRT end point.
// Otherwise, it parses the address as a pair of literal IP address
// and port number.
// The address parameter can use a host name, but this is not
// recommended, because it will return at most one of the host name's
// IP addresses.
//
// See func Dial for a description of the network and address
// parameters.
func ResolveSRTAddr(network, address string) (*SRTAddr, error) {
	switch network {
	case "srt", "srt4", "srt6":
	case "": // a hint wildcard for Go 1.0 undocumented behavior
		network = "srt"
	default:
		return nil, net.UnknownNetworkError(network)
	}
	addrs, err := DefaultResolver.internetAddrList(context.Background(), network, address)
	if err != nil {
		return nil, err
	}
	return addrs.forResolve(network, address).(*SRTAddr), nil
}

// SRTConn is an implementation of the Conn interface for SRT network
// connections.
type SRTConn struct {
	conn
}

// ReadFrom implements the io.ReaderFrom ReadFrom method.
func (c *SRTConn) ReadFrom(r io.Reader) (int64, error) {
	if !c.ok() {
		return 0, srtapi.EINVPARAM
	}
	n, err := c.readFrom(r)
	if err != nil && err != io.EOF {
		err = &OpError{Op: "readfrom", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, err
}

func newSRTConn(fd *netFD) *SRTConn {
	c := &SRTConn{conn{fd}}
	return c
}

// DialSRT acts like Dial for SRT networks.
//
// The network must be a SRT network name; see func Dial for details.
//
// If laddr is nil, a local address is automatically chosen.
// If the IP field of raddr is nil or an unspecified IP address, the
// local system is assumed.
func DialSRT(network string, laddr, raddr *SRTAddr) (*SRTConn, error) {
	switch network {
	case "srt", "srt4", "srt6":
	default:
		return nil, &OpError{Op: "dial", Net: network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: net.UnknownNetworkError(network)}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: network, Source: laddr.opAddr(), Addr: nil, Err: errMissingAddress}
	}

	c, err := dialSRT(context.Background(), network, laddr, raddr)

	if err != nil {
		return nil, &OpError{Op: "dial", Net: network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: err}
	}
	return c, nil
}

// SRTListener is a SRT network listener. Clients should typically
// use variables of type Listener instead of assuming SRT.
type SRTListener struct {
	fd  *netFD
	ctx context.Context
}

// AcceptSRT accepts the next incoming call and returns the new
// connection.
func (l *SRTListener) AcceptSRT() (*SRTConn, error) {
	if !l.ok() {
		return nil, srtapi.EINVPARAM
	}
	c, err := l.accept()
	if err != nil {
		return nil, &OpError{Op: "accept", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return c, nil
}

// Accept implements the Accept method in the Listener interface; it
// waits for the next call and returns a generic Conn.
func (l *SRTListener) Accept() (net.Conn, error) {
	if !l.ok() {
		return nil, srtapi.EINVPARAM
	}
	c, err := l.accept()
	if err != nil {
		return nil, &OpError{Op: "accept", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return c, nil
}

// Close stops listening on the SRT address.
// Already Accepted connections are not closed.
func (l *SRTListener) Close() error {
	if !l.ok() {
		return srtapi.EINVPARAM
	}
	if err := l.close(); err != nil {
		return &OpError{Op: "close", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return nil
}

// Addr returns the listener's network address, a *SRTAddr.
// The Addr returned is shared by all invocations of Addr, so
// do not modify it.
func (l *SRTListener) Addr() net.Addr { return l.fd.laddr }

// SetDeadline sets the deadline associated with the listener.
// A zero time value disables the deadline.
func (l *SRTListener) SetDeadline(t time.Time) error {
	if !l.ok() {
		return srtapi.EINVPARAM
	}
	if err := l.fd.pfd.SetDeadline(t); err != nil {
		return &OpError{Op: "set", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return nil
}

// ListenSRT acts like Listen for SRT networks.
//
// The network must be a SRT network name; see func Dial for details.
//
// If the IP field of laddr is nil or an unspecified IP address,
// ListenSRT listens on all available unicast and anycast IP addresses
// of the local system.
// If the Port field of laddr is 0, a port number is automatically
// chosen.
func ListenSRT(network string, laddr *SRTAddr) (*SRTListener, error) {
	switch network {
	case "srt", "srt4", "srt6":
	default:
		return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr.opAddr(), Err: net.UnknownNetworkError(network)}
	}
	if laddr == nil {
		laddr = &SRTAddr{}
	}
	ln, err := listenSRT(context.Background(), network, laddr)
	if err != nil {
		return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr.opAddr(), Err: err}
	}
	return ln, nil
}
