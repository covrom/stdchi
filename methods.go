package stdchi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type methodTyp uint

const (
	mSTUB methodTyp = 1 << iota
	mCONNECT
	mDELETE
	mGET
	mHEAD
	mOPTIONS
	mPATCH
	mPOST
	mPUT
	mTRACE
)

var mALL = mCONNECT | mDELETE | mGET | mHEAD |
	mOPTIONS | mPATCH | mPOST | mPUT | mTRACE

var methodMap = map[string]methodTyp{
	http.MethodConnect: mCONNECT,
	http.MethodDelete:  mDELETE,
	http.MethodGet:     mGET,
	http.MethodHead:    mHEAD,
	http.MethodOptions: mOPTIONS,
	http.MethodPatch:   mPATCH,
	http.MethodPost:    mPOST,
	http.MethodPut:     mPUT,
	http.MethodTrace:   mTRACE,
}

// RegisterMethod adds support for custom HTTP method handlers, available
// via Router#Method and Router#MethodFunc
func RegisterMethod(method string) {
	if method == "" {
		return
	}
	method = strings.ToUpper(method)
	if _, ok := methodMap[method]; ok {
		return
	}
	n := len(methodMap)
	if n > strconv.IntSize-2 {
		panic(fmt.Sprintf("chi: max number of methods reached (%d)", strconv.IntSize))
	}
	mt := methodTyp(2 << n)
	methodMap[method] = mt
	mALL |= mt
}
