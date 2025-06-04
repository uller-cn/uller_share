package main

import (
	"embed"
	"github.com/sirupsen/logrus"
	"github.com/uller_share/common"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"uller_share/lanHttp"
	"uller_share/lanNet"
	"uller_share/sqlite"
	//"embed"
	//"fmt"
	//"github.com/sirupsen/logrus"
	//"github.com/uller_share/common"
	//"github.com/wailsapp/wails/v2"
	//"github.com/wailsapp/wails/v2/pkg/options"
	//"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	//"github.com/wailsapp/wails/v2/pkg/options/windows"
	//"gopkg.in/natefinch/lumberjack.v2"
	//"uller_share/lanHttp"
	//"uller_share/lanNet"
	//"uller_share/sqlite"
)

//go:embed all:gui/dist
var assets embed.FS

//go:embed all:webui/dist
var h5ui embed.FS

func main() {
	runPath := common.GetCurrentDirectory()
	common.NetShare.Host = make(map[string]*common.HostData)
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   runPath + "\\log\\log.log", // 日志文件的路径
		MaxSize:    5,                          // 每个日志文件的最大大小（以MB为单位）
		MaxBackups: 7,                          // 保留的日志文件的最大个数
		MaxAge:     7,                          // 日志文件的最大存储天数
		Compress:   true,                       // 是否压缩旧日志文件
	})
	logrus.SetFormatter(&logrus.TextFormatter{})

	common.Iprofile = common.Profile{}

	//创建本地保存上传文件目录
	go func() {
		if !common.FileExists(runPath + "\\share") {
			_ = os.Mkdir(runPath+"\\share", 0755)
		}
	}()

	sqlite.IsRun = false
	sqlite.Db.RunPath = runPath
	go sqlite.Open()

	lanHttp.IsRun = false
	lanHttp.Progress.Task = make(map[int64]*lanHttp.TaskProgress)
	lanHttp.Progress.DownLoadIsRun = false
	lanHttp.Progress.DownLoadStop = false
	lh := lanHttp.NewLanHttp(assets, runPath)
	lh.SetTemplates(h5ui)
	go lh.Server()

	lanNet.IsRun = false
	var err error
	common.LocalIp, err = lanNet.GetLocalIP()
	if err != nil {
		logrus.Error("未获取到本机IP地址:", err)
	}
	ln, err := lanNet.NewLanNet()
	if err != nil {

	}
	ln.ListenUdp4Multicast()

	app := &App{nil, runPath, "", "", common.GetMachineId(), "", &ln}
	err = wails.Run(&options.App{
		Title:  "悠乐快传",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		EnableFraudulentWebsiteDetection: false,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "6edc613054a64f93bcee2a2788d5109d",
		},
		//Windows: &windows.Options{
		//	WebviewBrowserPath: runPath + "\\package\\Microsoft.WebView2.FixedVersionRuntime.109.0.1518.78.x64",
		//}
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		logrus.Error("程序启动错误:", err)
	}
}
