package l2

import (
	"fmt"
	"net/http"

	//pprof http handler
	_ "net/http/pprof"
)

//StartPProf opens the http pprof handler to desired port
func StartPProf(port int) {
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
}
