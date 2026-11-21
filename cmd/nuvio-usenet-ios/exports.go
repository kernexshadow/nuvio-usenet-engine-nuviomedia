package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"unsafe"
)

// The C API declared in nuvio_usenet.h.

//export nuvio_usenet_start
func nuvio_usenet_start(record *C.char) *C.char {
	out, err := start([]byte(C.GoString(record)))
	if err != nil {
		out, _ = json.Marshal(map[string]string{"error": err.Error()})
	}
	return C.CString(string(out))
}

//export nuvio_usenet_stop
func nuvio_usenet_stop() {
	stop()
}

//export nuvio_usenet_free
func nuvio_usenet_free(record *C.char) {
	C.free(unsafe.Pointer(record))
}

// main is required by -buildmode=c-archive and never runs.
func main() {}
