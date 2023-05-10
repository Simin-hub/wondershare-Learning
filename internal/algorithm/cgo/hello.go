package cgo

//#cgo LDFLAGS: -L. -lhello
//#include "hello.h"
import "C"
import "fmt"

func StartProcess(s string) {
	fmt.Printf("starting %s\n", s)
	C.Hello()
}
