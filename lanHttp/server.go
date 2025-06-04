package lanHttp

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shockerli/cvt"
	"github.com/uller_share/common"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

var Srv http.Server
var SrvH5 http.Server
var IsRun bool
var h5ui embed.FS

const (
	HttpSecret     = "xG&*DSTf12gyu294u"
	LanHttpPort    = ":35286"
	LanHttpH5Port  = ":35287"
	LanHttpTimeOut = 5 * 60 * 60
)

type LanHttp struct {
	Assets  embed.FS
	RunPath string
}

func NewLanHttp(assets embed.FS, runPath string) LanHttp {
	lanHttp := LanHttp{assets, runPath}
	return lanHttp
}

func (l *LanHttp) SetTemplates(fs embed.FS) {
	h5ui = fs
}

func (l *LanHttp) Server() {
	ginEngine := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	ginEngine.Use(l.cors())
	router := Router{ginEngine}
	router.SetUp(l.RunPath)
	Srv = http.Server{
		Addr:         LanHttpPort,
		Handler:      router,
		ReadTimeout:  LanHttpTimeOut * time.Second,
		WriteTimeout: LanHttpTimeOut * time.Second,
	}
	_ = ginEngine.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})
	go Srv.ListenAndServe()

	ginH5Engine := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	ginH5Engine.Use(l.cors())
	routerH5 := Router{ginH5Engine}
	fsSub, err := fs.Sub(h5ui, "webui/dist")
	if err != nil {
		fmt.Println(err)
	}
	routerH5.StaticFS("/", http.FS(fsSub))
	//routerH5.Static("/", "E:\\go\\uller_share\\webui\\dist")
	SrvH5 = http.Server{
		Addr:         LanHttpH5Port,
		Handler:      routerH5,
		ReadTimeout:  LanHttpTimeOut * time.Second,
		WriteTimeout: LanHttpTimeOut * time.Second,
	}
	_ = ginH5Engine.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})
	go SrvH5.ListenAndServe()
	IsRun = true
}

func StopHttp() (err error) {
	err = Srv.Close()
	err = SrvH5.Close()
	return
}

// 允许http跨域请求
func (l *LanHttp) cors() gin.HandlerFunc {
	return func(g *gin.Context) {
		if strings.ToUpper(g.Request.Method) == "OPTIONS" {
			SetHeaderCors(g)
			g.Status(200)
			g.Writer.Status()
			g.Abort()
		}
		g.Next()
	}
}

func SetHeaderCors(g *gin.Context) {
	g.Header("Access-Control-Allow-Origin", "*")
	g.Header("Access-Control-Allow-Headers", "sign,uller-client-time,uller-client,Content-Type,content-type,Content-Disposition")
	g.Header("Access-Control-Expose-Headers", ",sign,uller-client-time,uller-client,Content-Type,content-type,Content-Disposition")
	g.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,UPDATE")
}

func SignHttpHeader() (header map[string]string) {
	header = make(map[string]string)
	t := cvt.String(time.Now().Unix())
	header["sign"] = common.GetMD5Encode(t + HttpSecret)
	header["uller-client-time"] = t
	return
}
