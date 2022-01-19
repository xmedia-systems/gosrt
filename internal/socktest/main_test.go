// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package socktest_test

import (
	"os"
	"sync"
	"testing"

	socktest "github.com/xmedia-systems/gosrt/internal/socktest"
)

var sw socktest.Switch

func TestMain(m *testing.M) {
	installTestHooks()

	st := m.Run()

	for s := range sw.Sockets() {
		closeFunc(s)
	}
	uninstallTestHooks()
	os.Exit(st)
}

func TestSwitch(t *testing.T) {
	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			socketFunc()
		}()
	}
	wg.Wait()
}

func TestSocket(t *testing.T) {
	for _, f := range []socktest.Filter{
		func(st *socktest.Status) (socktest.AfterFilter, error) { return nil, nil },
		nil,
	} {
		sw.Set(socktest.FilterSocket, f)
		socketFunc()

	}
}
