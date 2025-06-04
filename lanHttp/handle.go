package lanHttp

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shockerli/cvt"
	"github.com/sirupsen/logrus"
	"github.com/uller_share/common"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"uller_share/sqlite"
)

type Handle struct {
	Encrypt bool
	RunPath string
}

// 下载文件
func (h *Handle) RequestFileHead(g *gin.Context) {
	var param = g.Query("shareId")
	shareId, err := cvt.Int64E(param)
	if err != nil || shareId == 0 {
		h.Status(g, 400)
		return
	}
	share, err := common.NetShare.Get(shareId, common.LocalIp.String())
	if err != nil {
		h.Status(g, 404)
		return
	}
	if share.ShareId == 0 {
		h.Status(g, 404)
		return
	}
	if share.ExpireTime > 0 && time.Now().Unix() > share.ExpireTime {
		h.Status(g, 410)
		return
	}
	fileInfo, err := os.Stat(share.LocalPath)
	if err != nil {
		_ = common.NetShare.Del(share.ShareId, common.LocalIp.String())
		_ = sqlite.Exec("delete from share where share_id=" + cvt.String(share.ShareId) + ";")
		h.Status(g, 404)
		return
	}
	if fileInfo.IsDir() {
		_ = common.NetShare.Del(share.ShareId, common.LocalIp.String())
		_ = sqlite.Exec("delete from share where share_id=" + cvt.String(share.ShareId) + ";")
		h.Status(g, 404)
		return
	}
	//g.Header("ETag", common.GetMD5Encode(cvt.String(fileInfo.ModTime().Unix())))
	g.Header("Last-Modified", cvt.String(fileInfo.ModTime()))
	g.Header("Accept-Ranges", "bytes")
	g.Header("Connection", "keep-alive")
	g.Header("Content-Type", mime.TypeByExtension(".exe"))
	g.Header("Content-Length", cvt.String(fileInfo.Size()))
}

// 下载文件
func (h *Handle) RequestFile(g *gin.Context) {
	var requestFile RequestFile
	err := g.ShouldBindJSON(&requestFile)
	if err != nil {
		h.Status(g, 400)
		return
	}
	share := common.Share{}
	share, err = common.NetShare.Get(requestFile.ShareId, common.LocalIp.String())
	if err != nil {
		h.Status(g, 404)
		return
	}
	if share.ShareId == 0 {
		h.Status(g, 404)
		return
	}
	if share.ExpireTime > 0 && time.Now().Unix() > share.ExpireTime {
		h.Status(g, 410)
		return
	}
	fileInfo, err := os.Stat(share.LocalPath)
	if err != nil {
		_ = common.NetShare.Del(share.ShareId, common.LocalIp.String())
		_ = sqlite.Exec("delete from share where share_id=" + cvt.String(share.ShareId) + ";")
		h.Status(g, 404)
		return
	}
	if fileInfo.IsDir() {
		_ = common.NetShare.Del(share.ShareId, common.LocalIp.String())
		_ = sqlite.Exec("delete from share where share_id=" + cvt.String(share.ShareId) + ";")
		h.Status(g, 404)
		return
	}
	// 打开文件
	file, err := os.Open(share.LocalPath)
	if err != nil {
		h.Status(g, 404)
		return
	}
	defer file.Close()
	SetHeaderCors(g)
	g.Header("Accept-Ranges", "bytes")
	g.Header("Last-Modified", cvt.String(fileInfo.ModTime()))
	g.Header("Connection", "keep-alive")
	g.Header("Content-Type", "application/octet-stream")
	// 获取请求头中的 Range 字段
	rangeHeader := g.Request.Header.Get("Range")
	var start, end int64
	start = 0
	end = fileInfo.Size()
	if rangeHeader != "" {
		// 解析 Range 字段
		parts := strings.Split(rangeHeader, "=")
		if len(parts) == 2 && parts[0] == "bytes" {
			ranges := strings.Split(parts[1], "-")
			if len(ranges) == 2 {
				start, err = strconv.ParseInt(ranges[0], 10, 64)
				if err != nil || start < 0 || start > fileInfo.Size() {
					h.AbortWithError(g, errors.New("invalid range"))
					return
				}
				if ranges[1] != "" {
					end, err = strconv.ParseInt(ranges[1], 10, 64)
					if err != nil || end < start || end > fileInfo.Size() {
						h.AbortWithError(g, errors.New("invalid range"))
						return
					}
				}

			}
		} else {
			g.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", share.Title))
		}
		var n int
		var buffSize int64 = 1024 * 1024
		var m = end - start
		_, _ = file.Seek(start, 0)
		g.Header("Content-Length", cvt.String(end-start))
		if m <= buffSize {
			buffer := make([]byte, end-start)
			n, err = file.Read(buffer)
			g.Writer.Write(buffer[:n])
			g.Writer.Flush()
		} else {
			buffer := make([]byte, buffSize)
			for i := int64(0); i < m; i++ {
				n, err = file.Read(buffer)
				if err != nil && err != io.EOF {
					h.AbortWithError(g, err)
					return
				}
				if n == 0 {
					break
				}
				_, _ = g.Writer.Write(buffer[:n])
				g.Writer.Flush()
			}
			if (end-start)-(buffSize*m) > 0 {
				bufferLast := make([]byte, (end-start)-(buffSize*m))
				_, _ = file.Seek(start+buffSize*m, 0)
				n, err = file.Read(bufferLast)
				_, _ = g.Writer.Write(bufferLast[:n])
				g.Writer.Flush()
			}
		}
		return
	}
}

func (h *Handle) UploadFile(g *gin.Context) {
	file, err := g.FormFile("file")
	lastModified, err := cvt.Int64E(g.PostForm("lastModified"))
	if err != nil {
		h.JsonResult(g, 1, "未获取到上传文件。", nil)
		return
	}
	disk, err := common.GetDiskInfo(h.RunPath)
	if err != nil {
		h.JsonResult(g, 1, "获取硬盘信息错误。", nil)
		return
	}
	if file.Size > cvt.Int64(disk.Free) {
		h.JsonResult(g, 1, "电脑硬盘空间不足。", nil)
		return
	}
	localPath := h.RunPath + "/share/"
	ext := filepath.Ext(file.Filename)
	fileName := filepath.Base(file.Filename)
	saveFile := localPath + fileName
	var fileInfo os.FileInfo
	fileInfo, err = os.Stat(saveFile)
	if err == nil {
		if fileInfo.ModTime().UnixMilli() == lastModified && fileInfo.Size() == file.Size {
			h.JsonResult(g, 0, "", nil)
			return
		} else {
			for i := 1; i < 100; i++ {
				fileName = fileName[:len(fileName)-len(ext)] + "(" + cvt.String(i) + ")" + ext
				saveFile = localPath + fileName
				fileInfo, err = os.Stat(saveFile)
				if err != nil {
					break
				} else if err == nil && fileInfo.Size() == file.Size {
					h.JsonResult(g, 0, "", nil)
					return
				}
			}
			err = saveUpFile(g, file, saveFile, lastModified, ext, fileName)
			if err != nil {
				h.JsonResult(g, 1, "保存文件错误。", nil)
				return
			}
		}
	} else {
		err = saveUpFile(g, file, saveFile, lastModified, ext, fileName)
		if err != nil {
			h.JsonResult(g, 1, "保存文件错误。", nil)
			return
		}
	}
	h.JsonResult(g, 0, "", nil)
}

func saveUpFile(g *gin.Context, file *multipart.FileHeader, localPath string, lastModified int64, ext string, fileName string) (err error) {
	err = g.SaveUploadedFile(file, localPath)
	if err != nil {
		logrus.Error("保存其他主机上传文件失败:", err)
		return
	}
	err = saveUploadHistory(g, file, localPath, ext, fileName)
	if err != nil {
		logrus.Error("保存其他主机上传数据失败:", err)
	}
	err = os.Chtimes(localPath, time.Now(), time.UnixMilli(lastModified))
	if err != nil {
		logrus.Error("文件上传修改文件属性失败:", err)
	}
	return
}

func saveUploadHistory(g *gin.Context, file *multipart.FileHeader, localPath string, ext string, fileName string) (err error) {
	timeNow := cvt.String(time.Now().Format("2006-01-02 15:04:05"))
	sqlStr := "insert into download_history values (" + cvt.String(common.GetSnowFlakeId()) + "," +
		"0," +
		"'" + fileName + "'," +
		"'" + filepath.FromSlash(localPath) + "'," +
		"'" + g.ClientIP() + "'," +
		"'" + ext + "'," +
		"" + cvt.String(file.Size) + "," +
		"" + cvt.String(file.Size) + "," +
		"2," +
		"'" + timeNow + "'," +
		"'" + timeNow + "');"
	err = sqlite.Exec(sqlStr)
	return
}

func (h *Handle) GetHostShareList(g *gin.Context) {
	h.JsonResult(g, 0, "", common.NetShare)
}

func (h *Handle) Sync(g *gin.Context) {
	hostData := common.HostData{}
	err := g.ShouldBindJSON(&hostData)
	if err != nil {
		h.JsonResult(g, 1, "参数错误。", nil)
	}
	if hostData.Ip == common.LocalIp.String() {
		h.JsonResult(g, 0, "", common.NetShare.Host[common.LocalIp.String()])
		return
	}
	common.NetShare.Lock()
	common.NetShare.Host[hostData.Ip] = &hostData
	common.NetShare.Unlock()
	eventData := make(map[string]string)
	eventData["ip"] = hostData.Ip
	eventData["nick"] = hostData.Nick
	eventData["shareCount"] = cvt.String(len(hostData.Share))
	wailsruntime.EventsEmit(common.AppCtx, "FindHost", eventData)
	h.JsonResult(g, 0, "", common.NetShare.Host[common.LocalIp.String()])
}

func (h *Handle) JsonResult(g *gin.Context, code int, msg string, data interface{}, total ...int64) {
	SetHeaderCors(g)
	var hasTotal *int64 = nil
	if len(total) > 0 {
		hasTotal = &total[0]
	}
	g.AbortWithStatusJSON(200, common.JsonResultData{code, msg, data, hasTotal})
	g.Abort()
}

func (h *Handle) Status(g *gin.Context, code int) {
	SetHeaderCors(g)
	g.Status(code)
	g.Abort()
}

func (h *Handle) AbortWithError(g *gin.Context, err error) {
	SetHeaderCors(g)
	_ = g.AbortWithError(http.StatusInternalServerError, err)
	g.Abort()
}
