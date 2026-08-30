package profiling

import (
	"fmt"
	"os"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

// To use in cluster:
// brew install graphviz
// kubectl port-forward -n runai deploy/pod-group-assigner 8182
// go tool pprof -http localhost:8888 http://localhost:8182/debug/pprof/profile
// -> and then wait a little bit and it will open a window in the browser

func RegisterProfiler(port string) {
	router := gin.Default()

	err := router.SetTrustedProxies(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	pprof.Register(router)

	err = router.Run(fmt.Sprintf(":%v", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
