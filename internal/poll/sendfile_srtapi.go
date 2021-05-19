// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package poll

import (
	"io"

	"github.com/xmedia-systems/gosrt/srtapi"
)

// maxSendfileSize is the largest chunk size we ask the kernel to copy
// at a time.
const maxSendfileSize int = 4 << 20

// SendFile wraps the sendfile system call.
func SendFile(dstFD *FD, r io.Reader, remain int64) (int64, error) {
	if err := dstFD.writeLock(); err != nil {
		return 0, err
	}
	defer dstFD.writeUnlock()

	dst := int(dstFD.Sysfd)
	var written int64
	var err error
	for remain > 0 {
		n := maxSendfileSize
		if int64(n) > remain {
			n = int(remain)
		}
		n, err1 := srtapi.Sendfile(dst, r, nil, n)
		if n > 0 {
			written += int64(n)
			remain -= int64(n)
		}
		if n == 0 && err1 == nil {
			break
		}
		if err1 == srtapi.EASYNCSND {
			if err1 = dstFD.pd.waitWrite(); err1 == nil {
				continue
			}
		}
		if err1 != nil {
			err = err1
			break
		}
	}
	return written, err
}
