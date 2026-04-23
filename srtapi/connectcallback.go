// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srtapi

/*

#include <srt/srt.h>

// Gateway: calls the Go export srtConnectCallback from C.
void SrtConnectCallback_cgo(void* opaq, SRTSOCKET ns, int errorcode,
    const struct sockaddr* peeraddr, int token)
{
	void srtConnectCallback(void*, SRTSOCKET, int, const struct sockaddr*, int);
	srtConnectCallback(opaq, ns, errorcode, peeraddr, token);
}
*/
import "C"
