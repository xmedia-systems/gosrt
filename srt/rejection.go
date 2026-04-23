// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srt

import (
	"fmt"

	"github.com/xmedia-systems/gosrt/srtapi"
)

// RejectionError is returned when an SRT connection is rejected during the
// handshake. Callers can type-assert to this type and inspect Reason against
// the srtapi.Rej* constants to determine the specific cause.
type RejectionError struct {
	Reason int    // one of srtapi.Rej* constants
	Msg    string // human-readable reason string from libsrt
}

func (e *RejectionError) Error() string {
	return fmt.Sprintf("connection rejected: %s (reason %d)", e.Msg, e.Reason)
}

func rejectionError(fd int) error {
	reason := srtapi.GetRejectReason(fd)
	return &RejectionError{
		Reason: reason,
		Msg:    srtapi.RejectReasonStr(reason),
	}
}
