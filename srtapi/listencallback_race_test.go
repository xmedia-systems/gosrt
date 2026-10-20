package srtapi

import (
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
)

// TestListenCallbackConcurrentClose reproduces
// "fatal error: concurrent map read and map write":
// srtListenCallback reads listenCallbackMap from a native libsrt thread for
// every incoming connection, while Close deletes from it on every socket close.
func TestListenCallbackConcurrentClose(t *testing.T) {
	const (
		dialers         = 4
		dialsPerDialer  = 100
		closers         = 4
		closesPerCloser = 2000
	)

	if err := Startup(); err != nil {
		t.Fatalf("Startup: %v", err)
	}
	t.Cleanup(func() { _ = Cleanup() })

	listener, err := Socket()
	if err != nil {
		t.Fatalf("Socket: %v", err)
	}
	t.Cleanup(func() { _ = Close(listener) })

	if err := Bind(listener, &syscall.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	sa, err := Getsockname(listener)
	if err != nil {
		t.Fatalf("Getsockname: %v", err)
	}
	target := &syscall.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}, Port: sa.(*syscall.SockaddrInet4).Port}

	var callbacks int64
	if err := ListenCallback(listener, func(int, int, syscall.Sockaddr, string) int {
		atomic.AddInt64(&callbacks, 1)
		return 0
	}); err != nil {
		t.Fatalf("ListenCallback: %v", err)
	}
	if err := Listen(listener, 128); err != nil {
		t.Fatalf("Listen: %v", err)
	}

	var accepted sync.WaitGroup
	accepted.Add(1)
	go func() {
		defer accepted.Done()
		for {
			fd, _, err := Accept(listener)
			if err != nil {
				return
			}
			_ = Close(fd)
		}
	}()

	var connected int64
	var wg sync.WaitGroup
	for i := 0; i < dialers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < dialsPerDialer; j++ {
				fd, err := Socket()
				if err != nil {
					t.Errorf("Socket: %v", err)
					return
				}
				if err := Connect(fd, target); err == nil {
					atomic.AddInt64(&connected, 1)
				}
				_ = Close(fd)
			}
		}()
	}
	for i := 0; i < closers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < closesPerCloser; j++ {
				fd, err := Socket()
				if err != nil {
					t.Errorf("Socket: %v", err)
					return
				}
				_ = Close(fd)
			}
		}()
	}
	wg.Wait()

	// Closing the listener unblocks Accept.
	_ = Close(listener)
	accepted.Wait()

	if atomic.LoadInt64(&connected) == 0 {
		t.Fatal("no client managed to connect")
	}
	if atomic.LoadInt64(&callbacks) < atomic.LoadInt64(&connected) {
		t.Errorf("listen callback invoked %d times for %d connections", atomic.LoadInt64(&callbacks), atomic.LoadInt64(&connected))
	}
}
