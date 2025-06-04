package lanHttp

import (
	"github.com/gin-gonic/gin"
	"github.com/shockerli/cvt"
	"github.com/uller_share/common"
	"time"
)

type Router struct {
	*gin.Engine
}

var RouterPath = make(map[string]string)

func (r *Router) SetUp(runPath string) {
	RouterPath["sync"] = "/share/sync"
	RouterPath["requestFile"] = "/share/requestFile"
	RouterPath["host"] = "/host"
	RouterPath["upload"] = "/upload"
	handle := Handle{true, runPath}
	r.Use(headAuth())
	r.Engine.POST(RouterPath["host"], handle.GetHostShareList)
	r.Engine.POST(RouterPath["upload"], handle.UploadFile)
	r.Engine.POST(RouterPath["requestFile"], handle.RequestFile)
	r.Engine.HEAD(RouterPath["requestFile"], handle.RequestFileHead)
	r.Engine.POST(RouterPath["sync"], handle.Sync)
}

func headAuth() gin.HandlerFunc {
	return func(g *gin.Context) {
		sign := g.GetHeader("sign")
		clientType := g.GetHeader("uller-client")
		t, err := cvt.Int64E(g.GetHeader("uller-client-time"))
		if err != nil {
			g.Status(401)
			g.Abort()
			return
		}
		if clientType == "h5" {
			if t < time.Now().Unix()-1800 {
				g.Status(401)
				g.Abort()
				return
			}
		} else {
			if t < time.Now().Unix()-300 {
				g.Status(401)
				g.Abort()
				return
			}
		}
		if sign != common.GetMD5Encode(cvt.String(t)+HttpSecret) {
			g.Status(401)
			g.Abort()
			return
		}
		g.Next()
	}
}
