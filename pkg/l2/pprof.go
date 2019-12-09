package l2

import (
	"fmt"
	"net/http"

	//pprof http handler
	_ "net/http/pprof"
)

func StartPProf(port int) {
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
