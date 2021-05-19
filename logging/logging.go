// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package logging

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
// #include "udt_wrapper.h"
/*
   void logHandler_cgo(void* opaque, int level, const char* file, int line, const char* area, const char* message);
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/xmedia-systems/gosrt/conf"
	"github.com/xmedia-systems/gosrt/srtapi"
)

// HandlerFunc logging handler function type
type HandlerFunc func(level int, file string, line int, area string, message string)

var handler HandlerFunc

//export logHandler
func logHandler(opaque unsafe.Pointer, level C.int, file *C.char, line C.int, area *C.char, message *C.char) {
	if handler != nil {
		handler(int(level), C.GoString(file), int(line), C.GoString(area), C.GoString(message))
	} else {
		now := time.Now()
		buf := fmt.Sprintf("[%v, %s:%d(%s)]{%d} %s", now, C.GoString(file), line, C.GoString(area), level, C.GoString(message))
		println(buf)
	}
}

// Init initialize logging function
func Init() {
	srtapi.SetLogLevel(conf.SystemConf().LogLevel())
	for fa := range conf.SystemConf().LogFAs() {
		srtapi.AddLogFA(fa)
	}
	NAME := C.CString("SRTLIB")
	defer C.free(unsafe.Pointer(NAME))
	if conf.SystemConf().LogInternal() {
		srtapi.SetLogFlags(0 | srtapi.LogFlagDisableTime | srtapi.LogFlagDisableSeverity | srtapi.LogFlagDisableThreadname | srtapi.LogFlagDisableEOF)
		C.srt_setloghandler(unsafe.Pointer(NAME), (*C.SRT_LOG_HANDLER_FN)(C.logHandler_cgo))
	} else if logFile := conf.SystemConf().LogFile(); logFile != "" {
		p := C.CString(logFile)
		defer C.free(unsafe.Pointer(p))
		C.udtSetLogStream(p)
	}
}

// SetHandler set handler
func SetHandler(h HandlerFunc) {
	handler = h
}
