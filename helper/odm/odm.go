package odm

/*
#cgo LDFLAGS: -lodm -lcfg
#define _THREAD_SAFE_ERRNO
#include <odmi.h>
#include <sys/cfgodm.h>
*/
import "C"

import (
	"regexp"
	"runtime"
	"strings"
	"unsafe"
)

type OdmError struct {
	odmerrno C.int
}

var whitespace = regexp.MustCompile("\\s+")

func (e OdmError) Error() string {
	var message *C.char
	C.odm_err_msg(e.odmerrno, &message)
	return "ODM subroutine failed: " + strings.TrimSpace(whitespace.ReplaceAllLiteralString(C.GoString(message), " "))
}

func GetAttributeMap(attribute string) (map[string]string, error) {
	m := make(map[string]string)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if C.odm_initialize() == 0 {
		defer C.odm_terminate()
		class := C.odm_open_class_rdonly((C.CLASS_SYMBOL)(unsafe.Pointer(&C.CuAt_CLASS)))
		if ^uintptr(unsafe.Pointer(class)) != 0 {
			defer C.odm_close_class(class)
			var obj C.struct_CuAt
			obj_ptr := unsafe.Pointer(&obj)
			query := C.CString("attribute = " + attribute)
			for res := C.odm_get_first(class, query, obj_ptr); ^uintptr(res) != 0; res = C.odm_get_next(class, obj_ptr) {
				if res != nil {
					m[C.GoString(&obj.name[0])] = C.GoString(&obj.value[0])
				} else {
					return m, nil
				}
			}
		}
	}
	return m, OdmError{C.odmerrno}
}
