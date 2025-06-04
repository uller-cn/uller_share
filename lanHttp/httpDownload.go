package lanHttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shockerli/cvt"
	"github.com/sirupsen/logrus"
	"github.com/uller_share/common"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	"uller_share/sqlite"
)

//多携程http下载

// 下载进度数据
var Progress progress
var SingleDownLoadQueue = NewDownLoadQueue(1000)

type progress struct {
	sync.Mutex
	DownLoadIsRun bool                    `json:"downLoadIsRun"` //下载程序是否正在运行
	DownLoadStop  bool                    `json:"downLoadStop"`  //发送停止命令
	FinishRoutine uint16                  `json:"finishRoutine"` //已完成下载的携程数
	Signal        uint8                   `json:"signal"`        //下载信号，1减少下载携程
	Routine       uint16                  `json:"routine"`       //启动的携程数
	RunRoutine    uint16                  `json:"runRoutine"`    //运行中的携程数
	Size          int64                   `json:"size"`          //总共要下载的byte
	Finish        int64                   `json:"finish"`        //下载完成byte
	Task          map[int64]*TaskProgress `json:"task"`          //各个下载任务数据
}

type TaskProgress struct {
	DownLoadHistory common.DownLoadHistory `json:"downLoadHistory"`
	Err             error                  `json:"err"`    //下载过程中发生的错误
	Signal          uint8                  `json:"signal"` //向单个下载任务发送信号，0正常下载，1停止单个任务下载
}

type Config struct {
	Retry     uint8  //下载失败重试次数
	RetryTime uint16 //下载失败间隔多少毫秒后重新下载
	TimeOut   uint32 //下载超时时间，单位秒
	CacheSize int    //缓存大小，单位bytes
}

type Task struct {
	DownLoadHistory common.DownLoadHistory
	Overwrite       uint8 //文件已存在下载过程如何处理，1删除本地文件，并重新下载，0重命名下载文件下载
}

// 获取下载文件head
func (d *Config) httpHeader(url string, header map[string]string) (size int64, lastModified string, err error) {
	var req *http.Request
	var resp *http.Response
	var i uint8
	for i = 0; i < d.Retry; i++ {
		err = nil
		req, err = http.NewRequest(http.MethodHead, url, nil)
		for key, value := range header {
			req.Header.Set(key, value)
		}
		resp, err = (&http.Client{Timeout: time.Duration(d.TimeOut) * time.Second}).Do(req)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			lastModified = resp.Header.Get("Last-Modified")
			_, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", lastModified)
			if err != nil {
				lastModified = cvt.String(time.Now().Format("2006-01-02 15:04:05"))
			}
			size, err = cvt.Int64E(resp.Header.Get("Content-Length"))
			break
		} else if resp.StatusCode == 404 || resp.StatusCode == 400 {
			err = errors.New("共享文件已删除。")
			break
		} else if resp.StatusCode == 410 {
			err = errors.New("共享文件已过期。")
			break
		} else {
			err = errors.New("获取下载文件头信息错误。")
			continue
		}
	}
	return
}

// 调整下载携程数量
func RoutineModify(routine uint16) (err error) {
	queueLen := SingleDownLoadQueue.Length()
	if queueLen == 0 {
		return nil
	} else if queueLen > 0 && queueLen < routine {
		routine = queueLen
	}
	Progress.Lock()
	Progress.Routine = routine
	Progress.Unlock()
	if Progress.RunRoutine < routine {
		config := NewHttpDownLoad()
		for i := uint16(0); i < routine-Progress.RunRoutine; i++ {
			go config.HttpDownLoad()
		}
	} else {
		Progress.Lock()
		Progress.Signal = 1
		Progress.Unlock()
		for {
			if Progress.RunRoutine == Progress.Routine {
				break
			}
		}
		Progress.Lock()
		Progress.Signal = 0
		Progress.Unlock()
	}
	return
}

func StopDownLoad() (err error) {
	Progress.Lock()
	if Progress.DownLoadIsRun {
		Progress.DownLoadStop = true
	} else {
		Progress.Unlock()
		return errors.New("下载任务未启动")
	}
	Progress.Unlock()
	for {
		if Progress.Routine == Progress.FinishRoutine {
			break
		}
	}
	for k, v := range Progress.Task {
		if v.DownLoadHistory.Status == 0 || v.DownLoadHistory.Status == 2 {
			Progress.Task[k].DownLoadHistory.Status = 4
		}
	}
	return nil
}

func StopTask(shareId int64) (err error) {
	if _, ok := Progress.Task[shareId]; !ok {
		return nil
	}
	Progress.Lock()
	Progress.Task[shareId].Signal = 1
	Progress.Unlock()
	for {
		if Progress.Task[shareId].DownLoadHistory.Status == 4 {
			break
		}
	}
	return nil
}

func NewHttpDownLoad() (config *Config) {
	config = &Config{3, 500, LanHttpTimeOut, 5 * 1024 * 1024}
	return
}

func (d *Config) HttpDownLoad() {
	var ret bool
	var task = Task{}
	var urlPath = "/share/requestFile"
	var size, start, end int64
	var n int
	var lastModified, urlStr, reTitle string
	var err error
	var fileInfo os.FileInfo
	var httpSendByte []byte
	var req *http.Request
	var resp *http.Response
	var header map[string]string
	var i, reTitleStatus uint8
	var buf = make([]byte, d.CacheSize)
	var reqBytes = &bytes.Buffer{}
	Progress.Lock()
	Progress.RunRoutine++
	Progress.Unlock()
EndFor:
	for {
		task, ret = SingleDownLoadQueue.Pop()
		if !ret {
			break EndFor
		}
		//全局停止下载
		Progress.Lock()
		if Progress.DownLoadStop {
			Progress.Unlock()
			break EndFor
		}
		//对单个任务停止下载
		if Progress.Task[task.DownLoadHistory.Share.ShareId].Signal == 1 {
			Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 4
			Progress.Unlock()
			continue
		}
		//减少下载携程
		if Progress.Signal == 1 {
			if Progress.RunRoutine > Progress.Routine {
				SingleDownLoadQueue.Push(task)
				Progress.Unlock()
				break EndFor
			}
		}
		Progress.RunRoutine += 1
		Progress.Unlock()
		reTitleStatus = 0 //下载过程如果遇到要下载的共享文件在本地已存在会对本地文件进行重命名，此值代表重命名状态，0重命名失败（会跳过下载任务），1重命名成功（以新名称保存下载文件），2重命名文件过程发现已有下载完成的文件（会跳过下载任务）
		Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 3
		err = nil
		urlStr = "http://" + task.DownLoadHistory.Share.Ip + LanHttpPort + urlPath + "?shareId=" + cvt.String(task.DownLoadHistory.Share.ShareId)
		header = SignHttpHeader()
		size, lastModified, err = d.httpHeader(urlStr, header)
		fmt.Println(size, lastModified, err)
		if err != nil {
			Progress.Lock()
			Progress.Finish += size
			Progress.Task[task.DownLoadHistory.Share.ShareId].Err = err
			if err.Error() == "共享文件已过期。" {
				Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 6
			} else if err.Error() == "共享文件已删除。" {
				Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 7
			} else {
				Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 5
			}
			Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish = 0
			Progress.Unlock()
			wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
			saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
			continue
		}
		Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Size = size
		fileInfo, err = os.Stat(task.DownLoadHistory.LocalPath)
		if err != nil { //共享文件没有下载过
			start = 0
			end = size
		} else { //共享文件已存在
			if task.Overwrite == 1 {
				err = os.Remove(task.DownLoadHistory.LocalPath)
				if err != nil {
					Progress.Lock()
					Progress.Finish += size
					Progress.Task[task.DownLoadHistory.Share.ShareId].Err = err
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 5
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish = end
					Progress.Unlock()
					wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
					saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
					logrus.Error("下载过程，删除文件错误", err, urlStr)
					continue
				}
				start = 0
				end = size
			} else {
				if cvt.String(fileInfo.ModTime()) == lastModified { //共享文件最后修改时间和本地文件没有变化
					if fileInfo.Size() == size { //共享文件和本地文件大小相同说明已经下载完成，无需重新下载
						Progress.Lock()
						Progress.Finish += size
						Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 2
						Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish = size
						Progress.Unlock()
						wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
						saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
						continue
					} else { //共享文件和本地文件大小不同，需要断点续传
						Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish = task.DownLoadHistory.Finish - 1
						start = fileInfo.Size() - 1
						end = size
					}
				} else { //本地文件已存在且和共享文件最后修改时间不同，说明可能是不同共享主机上重名的文件，将下载的文件重命名后下载
					//重命名本地文件
					for i = 1; i < 255; i++ {
						reTitle = task.DownLoadHistory.Title[:len(task.DownLoadHistory.Title)-len(task.DownLoadHistory.Share.Ext)] + "(" + cvt.String(i) + ")" + task.DownLoadHistory.Share.Ext
						task.DownLoadHistory.LocalPath = task.DownLoadHistory.LocalPath + "/" + reTitle
						fileInfo, err = os.Stat(task.DownLoadHistory.LocalPath)
						if err != nil {
							task.DownLoadHistory.Title = reTitle
							Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Title = reTitle
							reTitleStatus = 1
							break
						} else {
							if cvt.String(fileInfo.ModTime()) == lastModified {
								if fileInfo.Size() == size {
									task.DownLoadHistory.Title = reTitle
									Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Title = reTitle
									reTitleStatus = 2
									break
								} else {
									continue
								}
							}
						}
					}
					if reTitleStatus == 1 {
						start = 0
						end = size
					} else if reTitleStatus == 0 {
						Progress.Lock()
						Progress.Finish += size
						Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 5
						Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish = end
						Progress.Unlock()
						wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
						saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
						logrus.Error("下载过程，文件重命名错误，本地文件保存路径", task.DownLoadHistory.LocalPath, urlStr)
						continue
					} else if reTitleStatus == 2 {
						Progress.Lock()
						Progress.Finish += size
						Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 2
						Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish = size
						Progress.Unlock()
						wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
						saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
						continue
					}
				}
			}
		}
		Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Ext = filepath.Ext(task.DownLoadHistory.LocalPath)
		file, err := os.OpenFile(task.DownLoadHistory.LocalPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			Progress.Lock()
			Progress.Finish += size
			Progress.Task[task.DownLoadHistory.Share.ShareId].Err = err
			Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 5
			Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish = start
			Progress.Unlock()
			_ = file.Close()
			wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
			saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
			logrus.Error("下载过程，打开文件错误", err, urlStr)
			continue
		}

	EndForWrite:
		for i = 0; i < d.Retry; i++ {
			httpSendByte, _ = json.Marshal(RequestFile{task.DownLoadHistory.Share.ShareId})
			reqBytes = bytes.NewBuffer(httpSendByte)
			req, _ = http.NewRequest(http.MethodPost, urlStr, reqBytes)
			header = SignHttpHeader()
			for headerKey, value := range header {
				req.Header.Set(headerKey, value)
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
			req.Header.Set("user-agent", "client version 1.0")
			resp, err = (&http.Client{Timeout: time.Duration(d.TimeOut) * time.Second}).Do(req)
			if err != nil && resp.Status != "200" {
				logrus.Error("下载过程，请求文件错误", err, urlStr, "范围,"+cvt.String(start)+"-"+cvt.String(end))
				continue
			}
			//全局停止下载
			if Progress.DownLoadStop {
				Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 4
				_ = file.Close()
				_ = resp.Body.Close()
				err = os.Chtimes(task.DownLoadHistory.LocalPath, time.Now(), cvt.Time(lastModified))
				if err != nil {
					logrus.Error("下载过程停止下载，修改文件最后修改时间错误", err, urlStr)
				}
				wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
				saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
				break EndFor
			}
			//单个任务停止下载
			if Progress.Task[task.DownLoadHistory.Share.ShareId].Signal == 1 {
				_ = file.Close()
				_ = resp.Body.Close()
				err = os.Chtimes(task.DownLoadHistory.LocalPath, time.Now(), cvt.Time(lastModified))
				if err != nil {
					logrus.Error("下载过程停止单个下载，修改文件最后修改时间错误", err, urlStr)
				}
				Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 4
				wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
				saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
				break EndForWrite
			}
			//减少下载携程
			Progress.Lock()
			if Progress.Signal == 1 {
				if Progress.RunRoutine > Progress.Routine {
					_ = file.Close()
					_ = resp.Body.Close()
					err = os.Chtimes(task.DownLoadHistory.LocalPath, time.Now(), cvt.Time(lastModified))
					if err != nil {
						logrus.Error("下载过程减少下载，修改文件最后修改时间错误", err, urlStr)
					}
					SingleDownLoadQueue.Push(task)
					Progress.Unlock()
					break EndFor
				}
			}
			Progress.Unlock()
			n = 0
			for {
				//全局停止下载
				if Progress.DownLoadStop {
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 4
					_ = file.Close()
					err = os.Chtimes(task.DownLoadHistory.LocalPath, time.Now(), cvt.Time(lastModified))
					if err != nil {
						logrus.Error("下载过程停止下载，修改文件最后修改时间错误", err, urlStr)
					}
					wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
					saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
					break EndFor
				}
				//单个任务停止下载
				if Progress.Task[task.DownLoadHistory.Share.ShareId].Signal == 1 {
					_ = file.Close()
					err = os.Chtimes(task.DownLoadHistory.LocalPath, time.Now(), cvt.Time(lastModified))
					if err != nil {
						logrus.Error("下载过程停止单个下载，修改文件最后修改时间错误", err, urlStr)
					}
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 4
					wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
					saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
					break EndForWrite
				}
				n, err = resp.Body.Read(buf)
				if err != nil && err != io.EOF {
					logrus.Error("下载过程，读取远程文件出错", err, urlStr, "范围,"+cvt.String(start)+"-"+cvt.String(end))
					Progress.Task[task.DownLoadHistory.Share.ShareId].Err = err
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 5
					time.Sleep(time.Duration(d.RetryTime) * time.Millisecond)
					break
				}
				if n > 0 {
					_, err = file.WriteAt(buf[:n], start)
					if err != nil {
						logrus.Error("下载过程，写入文件错误", err, urlStr, "范围,"+cvt.String(start)+"-"+cvt.String(end))
					}
					Progress.Lock()
					Progress.Finish += cvt.Int64(n)
					Progress.Task[task.DownLoadHistory.Share.ShareId].Err = nil
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 3
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish += cvt.Int64(n)
					Progress.Unlock()
					//wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
					start = start + cvt.Int64(n)
				}
				if err == io.EOF {
					Progress.Lock()
					Progress.Finish += cvt.Int64(n)
					Progress.Task[task.DownLoadHistory.Share.ShareId].Err = nil
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 2
					Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Finish += cvt.Int64(n)
					Progress.Unlock()
					//wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
					saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
					err = os.Chtimes(task.DownLoadHistory.LocalPath, time.Now(), cvt.Time(lastModified))
					if err != nil {
						logrus.Error("下载过程，修改文件最后修改时间错误", err, urlStr)
					}
					break EndForWrite
				}
			}
			_ = resp.Body.Close()
		}
		if err != nil {
			Progress.Lock()
			Progress.Finish += size
			Progress.Task[task.DownLoadHistory.Share.ShareId].Err = err
			Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.Status = 5
			Progress.Unlock()
			wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
			saveTask(*Progress.Task[task.DownLoadHistory.Share.ShareId])
			logrus.Error("下载过程出错，已尝试重新下载"+cvt.String(d.RetryTime)+"次，", err, urlStr, "范围,"+cvt.String(start)+"-"+cvt.String(end))
		}
		_ = file.Close()
	}
	Progress.Lock()
	Progress.RunRoutine -= 1
	Progress.FinishRoutine += 1
	Progress.Unlock()
}

func (d *Config) AddDownLoadTask(downLoadHistory []common.DownLoadHistory, overwrite uint8) (downLoadHistoryData []common.DownLoadHistory, err error) {
	var task = Task{}
	var sqlStr string
	for i := 0; i < len(downLoadHistory); i++ {
		_, ok := Progress.Task[downLoadHistory[i].Share.ShareId]
		if !ok {
			Progress.Size += downLoadHistory[i].Share.Size
			task = Task{}
			task.DownLoadHistory = downLoadHistory[i]
			task.Overwrite = overwrite
			_ = SingleDownLoadQueue.Push(task)
			taskProgress := getTaskProgress(task.DownLoadHistory)
			Progress.Lock()
			Progress.Task[downLoadHistory[i].Share.ShareId] = &taskProgress
			Progress.Unlock()
			sqlStr += getSaveTaskSql(taskProgress)
			wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", Progress.Task[task.DownLoadHistory.Share.ShareId])
		}
		downLoadHistoryData = append(downLoadHistoryData, Progress.Task[downLoadHistory[i].Share.ShareId].DownLoadHistory)
	}
	_ = sqlite.Exec(sqlStr)
	if !Progress.DownLoadIsRun {
		Progress.Lock()
		Progress.DownLoadIsRun = true
		Progress.DownLoadStop = false
		queueLen := SingleDownLoadQueue.Length()
		if common.Iprofile.DownloadRoutine > queueLen {
			Progress.Routine = queueLen
		} else {
			Progress.Routine = common.Iprofile.DownloadRoutine
		}
		Progress.RunRoutine = 0
		Progress.FinishRoutine = 0
		Progress.Finish = 0
		Progress.Unlock()
		for i := uint16(0); i < Progress.Routine; i++ {
			go d.HttpDownLoad()
		}
		go Watch()
	}
	return
}

func getDownLoadHistory(share common.Share, localPath string) (downLoadHistory common.DownLoadHistory) {
	if downLoadHistory.HistoryId == 0 {
		downLoadHistory.HistoryId = common.GetSnowFlakeId()
	}
	downLoadHistory.Share.ShareId = share.ShareId
	downLoadHistory.Share.Title = share.Title
	downLoadHistory.LocalPath = localPath
	downLoadHistory.Share.Ext = share.Ext
	downLoadHistory.Share.Size = share.Size
	downLoadHistory.Status = 1
	downLoadHistory.Share.Ip = share.Ip
	return
}

func getTaskProgress(downLoadHistory common.DownLoadHistory) (taskProgress TaskProgress) {
	taskProgress.DownLoadHistory = downLoadHistory
	taskProgress.DownLoadHistory.Finish = 0
	taskProgress.Err = nil
	taskProgress.Signal = 0
	return
}

func getSaveTaskSql(task TaskProgress) (sqlStr string) {
	timeNow := cvt.String(time.Now().Format("2006-01-02 15:04:05"))
	if task.DownLoadHistory.HistoryId == 0 {
		Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.HistoryId = common.GetSnowFlakeId()
		sqlStr = "insert into download_history values (" +
			cvt.String(Progress.Task[task.DownLoadHistory.Share.ShareId].DownLoadHistory.HistoryId) + "," +
			cvt.String(task.DownLoadHistory.Share.ShareId) + "," +
			"'" + task.DownLoadHistory.Title + "'" + "," +
			"'" + task.DownLoadHistory.LocalPath + "'" + "," +
			"'" + task.DownLoadHistory.Ip + "'," +
			"'" + task.DownLoadHistory.Ext + "'" + "," +
			cvt.String(task.DownLoadHistory.Size) + "," +
			cvt.String(task.DownLoadHistory.Finish) + "," +
			cvt.String(task.DownLoadHistory.Status) + "," +
			"'" + timeNow + "'," +
			"'" + timeNow + "'" +
			");"
	} else {
		sqlStr = "update download_history set " +
			"title='" + task.DownLoadHistory.Title + "'," +
			"local_path='" + task.DownLoadHistory.LocalPath + "'," +
			"ip='" + task.DownLoadHistory.Ip + "'," +
			"ext='" + task.DownLoadHistory.Ext + "'," +
			"size=" + cvt.String(task.DownLoadHistory.Size) + "," +
			"finish=" + cvt.String(task.DownLoadHistory.Finish) + "," +
			"status=" + cvt.String(task.DownLoadHistory.Status) + "," +
			"update_time='" + timeNow + "' " +
			" where history_id=" + cvt.String(task.DownLoadHistory.HistoryId) + ";"
	}
	return
}

func saveTask(task TaskProgress) {
	var sqlStr = getSaveTaskSql(task)
	_ = sqlite.Exec(sqlStr)
}

func SaveDownLoadHistory() (err error) {
	if len(Progress.Task) == 0 {
		return
	}
	Progress.Lock()
	var sqlStr = ""
	for _, v := range Progress.Task {
		sqlStr += getSaveTaskSql(*v)
	}
	Progress.Unlock()
	return sqlite.Exec(sqlStr)
}

func Watch() {
	for {
		for _, v := range Progress.Task {
			wailsruntime.EventsEmit(common.AppCtx, "DownLoadTaskEvent", v)
		}
		if Progress.FinishRoutine == Progress.Routine {
			Progress.Lock()
			Progress.Routine = 0
			Progress.Finish = 0
			Progress.RunRoutine = 0
			Progress.FinishRoutine = 0
			Progress.Signal = 0
			Progress.Size = 0
			Progress.Finish = 0
			Progress.DownLoadIsRun = false
			Progress.DownLoadStop = false
			Progress.Task = make(map[int64]*TaskProgress)
			Progress.Unlock()
			SingleDownLoadQueue.Clear()
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}
