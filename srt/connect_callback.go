// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srt

import (
	"context"
	"net"
	"syscall"

	"github.com/xmedia-systems/gosrt/srtapi"
)

// ConnectCallbackFunc is called on a listener for every connection attempt that
// completes — whether accepted or rejected during the SRT handshake. peerAddr is
// the remote endpoint; rejection is non-nil when the handshake was rejected (e.g.
// encryption mismatch). The rejection value carries the structured reason code and
// a human-readable message from libsrt.
type ConnectCallbackFunc func(peerAddr net.Addr, rejection *RejectionError)

// connectCallbackContextKey is the type of contextKeys used for connectCallback.
type connectCallbackContextKey struct{}

// WithConnectCallback returns a new context.Context carrying the given callback.
// Pass this context to Listen so that every incoming connection attempt (including
// rejections that never surface through Accept) triggers the callback.
func WithConnectCallback(ctx context.Context, callback ConnectCallbackFunc) context.Context {
	return context.WithValue(ctx, connectCallbackContextKey{}, callback)
}

func connectCallbackValue(ctx context.Context) ConnectCallbackFunc {
	callback, _ := ctx.Value(connectCallbackContextKey{}).(ConnectCallbackFunc)
	return callback
}

// toSrtapiConnectCallback wraps the high-level ConnectCallbackFunc into the
// srtapi-level type, constructing a RejectionError while ns is still live.
func toSrtapiConnectCallback(cbk ConnectCallbackFunc) srtapi.SrtConnectCallbackFunc {
	return func(ns int, errorcode int, peeraddr syscall.Sockaddr, token int) {
		addr := sockaddrToSRT(peeraddr)
		var rejection *RejectionError
		if errorcode != 0 {
			reason := srtapi.GetRejectReason(ns)
			rejection = &RejectionError{
				Reason: reason,
				Msg:    srtapi.RejectReasonStr(reason),
			}
		}
		cbk(addr, rejection)
	}
}
