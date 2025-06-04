package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/beevik/ntp"
	"github.com/ncruces/zenity"
	"github.com/shockerli/cvt"
	"github.com/sirupsen/logrus"
	"github.com/uller_share/common"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"uller_share/lanHttp"
	"uller_share/lanNet"
	"uller_share/sqlite"
)

// App struct
type App struct {
	ctx             context.Context
	RunPath         string
	LastShareFolder string
	HostNick        string
	DeviceId        string
	userId          string
	lanNet          *lanNet.LanNet
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	_ = a.lanNet.LeaveGroup()
	_ = a.lanNet.Shutdown()
	if lanHttp.Progress.DownLoadIsRun {
		_ = lanHttp.StopDownLoad()
	}
	_ = lanHttp.StopHttp()
	os.Exit(0)
	return false
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	common.AppCtx = ctx
}

// 删除分享
func (a *App) DelShare(shareId []string) (err error) {
	if len(shareId) == 0 {
		return errors.New("参数错误")
	}
	if err != nil {
		return
	}
	sqlStr := "delete from `share` where share_id in (" + strings.Join(shareId, ",") + ")"
	err = sqlite.Exec(sqlStr)
	if err != nil {
		return
	}
	var shareIdInt int64
	for i := 0; i < len(shareId); i++ {
		shareIdInt, err = cvt.Int64E(shareId[i])
		if err != nil {
			continue
		}
		err = a.lanNet.DelShare(cvt.Int64(shareId[i]))
		if err != nil {
			logrus.Error("发送删除共享文件失败：", err)
		}
		err = common.NetShare.Del(shareIdInt, common.LocalIp.String())
		if err != nil {
			logrus.Error("删除共享文件缓存失败：", err)
		}
	}
	wailsruntime.EventsEmit(common.AppCtx, "LocalShareCount", len(common.NetShare.Host[common.LocalIp.String()].Share))
	return
}

// 获取共享主机
func (a *App) GetHostList(nick string) (ret []map[string]string, err error) {
	hostList := common.NetShare.LikeNick(nick)
	hostMap := make(map[string]string)
	for _, v := range hostList {
		hostMap["ip"] = v.Ip
		hostMap["nick"] = v.Nick
		hostMap["shareCount"] = cvt.String(v.ShareCount)
		ret = append(ret, hostMap)
	}
	return
}

func (a *App) GetShareExt(ip string) (ret []string, err error) {
	if ip != "" {
		if _, ok := common.NetShare.Host[ip]; ok {
			for i := 0; i < len(common.NetShare.Host[ip].Share); i++ {
				if !strings.Contains(strings.Join(ret, ","), strings.ToLower(common.NetShare.Host[ip].Share[i].Ext)) {
					ret = append(ret, strings.ToLower(common.NetShare.Host[ip].Share[i].Ext))
				}
			}
		}
	} else {
		for key, _ := range common.NetShare.Host {
			for i := 0; i < len(common.NetShare.Host[key].Share); i++ {
				if !strings.Contains(strings.Join(ret, ","), strings.ToLower(common.NetShare.Host[key].Share[i].Ext)) {
					ret = append(ret, strings.ToLower(common.NetShare.Host[key].Share[i].Ext))
				}
			}
		}
	}
	return
}

// 全局搜索
func (a *App) GetHostShareList(title string, ext []string, pageSize int64, pageNumber int64) (ret common.DownLoadHistoryList, err error) {
	share := []common.Share{}
	if len(ext) > 0 {
		for key, value := range common.NetShare.Host {
			if key == common.LocalIp.String() {
				continue
			}
			for i := 0; i < len(value.Share); i++ {
				if strings.Contains(strings.Join(ext, ","), value.Share[i].Ext) {
					share = append(share, value.Share[i])
				}
			}
		}
	} else {
		for key, value := range common.NetShare.Host {
			if key == common.LocalIp.String() {
				continue
			}
			for i := 0; i < len(value.Share); i++ {
				share = append(share, value.Share[i])
			}
		}
	}
	if strings.Trim(title, " ") != "" {
		share = common.Like(share, title)
	}
	end := pageSize * (pageNumber + 1)
	if pageSize*(pageNumber+1) > ret.Total {
		end = cvt.Int64(len(share))
	}
	share = share[pageSize*pageNumber : end]
	ret.Total = cvt.Int64(len(share))
	shareIds := []int64{}
	downLoadHistory := common.DownLoadHistory{}
	for _, v := range share {
		downLoadHistory.Share = v
		ret.DownLoadHistory = append(ret.DownLoadHistory, downLoadHistory)
		shareIds = append(shareIds, v.ShareId)
	}
	downLoadHistoryList := sqlite.QueryDownloadHistoryIds(common.IntArrToString(shareIds))
	for k, v := range ret.DownLoadHistory {
		for _, dv := range downLoadHistoryList {
			if v.Share.ShareId == dv.Share.ShareId {
				ret.DownLoadHistory[k].HistoryId = dv.HistoryId
				ret.DownLoadHistory[k].Title = dv.Title
				ret.DownLoadHistory[k].LocalPath = dv.LocalPath
				ret.DownLoadHistory[k].Ext = dv.Ext
				ret.DownLoadHistory[k].Finish = dv.Finish
				ret.DownLoadHistory[k].Size = dv.Size
				ret.DownLoadHistory[k].Status = dv.Status
				ret.DownLoadHistory[k].CreateTime = dv.CreateTime
				ret.DownLoadHistory[k].UpdateTime = dv.UpdateTime
			}
		}
	}
	return
}

func (a *App) GetShareList(ip string, title string, ext []string, pageSize int64, pageNumber int64) (ret common.DownLoadHistoryList, err error) {
	if _, ok := common.NetShare.Host[ip]; ok {
		shareList := []common.Share{}
		if len(ext) > 0 {
			for i := 0; i < len(common.NetShare.Host[ip].Share); i++ {
				if strings.Contains(strings.Join(ext, ","), common.NetShare.Host[ip].Share[i].Ext) {
					shareList = append(shareList, common.NetShare.Host[ip].Share[i])
				}
			}
		} else {
			shareList = common.NetShare.Host[ip].Share
		}
		if title != "" {
			shareList = common.Like(shareList, title)
		}
		ret.Total = cvt.Int64(len(shareList))
		end := pageSize * (pageNumber + 1)
		if pageSize*(pageNumber+1) > ret.Total {
			end = cvt.Int64(len(shareList))
		}
		shareList = shareList[pageSize*pageNumber : end]
		shareIds := []int64{}
		downLoadHistory := common.DownLoadHistory{}
		for _, v := range shareList {
			downLoadHistory.Share = v
			ret.DownLoadHistory = append(ret.DownLoadHistory, downLoadHistory)
			shareIds = append(shareIds, v.ShareId)
		}
		downLoadHistoryList := sqlite.QueryDownloadHistoryIds(common.IntArrToString(shareIds))
		for k, v := range ret.DownLoadHistory {
			for _, dv := range downLoadHistoryList {
				if v.Share.ShareId == dv.Share.ShareId {
					ret.DownLoadHistory[k].HistoryId = dv.HistoryId
					ret.DownLoadHistory[k].Title = dv.Title
					ret.DownLoadHistory[k].LocalPath = dv.LocalPath
					ret.DownLoadHistory[k].Ext = dv.Ext
					ret.DownLoadHistory[k].Finish = dv.Finish
					ret.DownLoadHistory[k].Size = dv.Size
					ret.DownLoadHistory[k].Status = dv.Status
					ret.DownLoadHistory[k].CreateTime = dv.CreateTime
					ret.DownLoadHistory[k].UpdateTime = dv.UpdateTime
				}
			}
		}
	}
	return
}

func (a *App) OpenDir(historyId int64, file string) (err error) {
	_, err = os.Stat(file)
	if err != nil {
		_ = sqlite.UpdateDownloadHistoryNoFinish(cvt.String(historyId))
		return errors.New("本地文件已被删除，请重新下载。")
	}
	cmd := exec.Command("explorer", "/select,", file)
	return cmd.Start()
}

func (a *App) OpenFile(historyId int64, file string) (err error) {
	_, err = os.Stat(file)
	if err != nil {
		_ = sqlite.UpdateDownloadHistoryNoFinish(cvt.String(historyId))
		return errors.New("本地文件已被删除，请重新下载。")
	}
	cmd := exec.Command("cmd.exe", "/C", "start", file)
	return cmd.Start()
}

func (a *App) Init() (ret map[string]string, err error) {
	for {
		if sqlite.IsRun && lanHttp.IsRun && lanNet.IsRun {
			break
		}
	}
	ret = make(map[string]string)
	ret["ip"] = common.LocalIp.String()
	shareList := sqlite.QueryShareList("")
	common.Iprofile = sqlite.QueryProfile()
	if common.Iprofile.Nick == "" {
		common.Iprofile.Nick, err = os.Hostname()
	}
	if common.Iprofile.DownloadRoutine == 0 {
		common.Iprofile.DownloadRoutine = 5
	}
	hostData := &common.HostData{common.Iprofile.Nick, common.LocalIp.String(), shareList}
	common.NetShare.Host[common.LocalIp.String()] = hostData
	ret["nick"] = common.Iprofile.Nick
	ret["shareCount"] = cvt.String(len(shareList))
	ret["timeErr"] = "0"
	ret["version"] = common.Version
	ntpTime, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err == nil {
		if cvt.Int64(math.Abs(float64(ntpTime.Unix()-time.Now().Unix()))) > cvt.Int64(30*time.Second) {
			ret["timeErr"] = "1"
		}
	}
	a.LastShareFolder = common.Iprofile.LastShareFolder
	err = a.lanNet.JoinGroup()
	if err != nil {
		logrus.Error("发送入组消息失败：", err)
	}
	return
}

// 下载文件
func (a *App) DownLoad(downLoadHistory common.DownLoadHistory, ip string, mod string) (downLoadHistoryRecord common.DownLoadHistory, err error) {
	share := common.Share{}
	share, err = common.NetShare.Get(downLoadHistory.Share.ShareId, ip)
	if share.ShareId == 0 {
		share, err = common.NetShare.GetSame(downLoadHistory.Title, downLoadHistory.Size)
		if share.ShareId == 0 {
			return downLoadHistoryRecord, errors.New("共享文件：" + cvt.String(downLoadHistory.Share.Title) + "已移除共享或共享文件已到期，无法下载。")
		}
	}
	var dir string
	overwrite := uint8(0)
	if mod == "continue" {
		dir = downLoadHistory.LocalPath
	} else {
		dir, err = zenity.SelectFileSave(
			zenity.Filename(a.LastShareFolder+"\\"+share.Title),
			zenity.FileFilters{
				{"", []string{"*.*"}, true},
			})
		if err != nil {
			return
		}
		overwrite = 1
		dir = filepath.Dir(dir)
	}
	tmpDownLoadHistory := []common.DownLoadHistory{}
	downLoadHistory.Status = 1
	downLoadHistory.LocalPath = dir + "\\" + downLoadHistory.Share.Title
	downLoadHistory.Share = share
	if downLoadHistory.HistoryId == 0 {
		downLoadHistory.Title = downLoadHistory.Share.Title
		downLoadHistory.Ext = downLoadHistory.Share.Ext
		downLoadHistory.Size = downLoadHistory.Share.Size
	}
	tmpDownLoadHistory = append(tmpDownLoadHistory, downLoadHistory)
	config := lanHttp.NewHttpDownLoad()
	var downLoadHistoryData = []common.DownLoadHistory{}
	downLoadHistoryData, err = config.AddDownLoadTask(tmpDownLoadHistory, overwrite)
	if len(downLoadHistoryData) > 0 {
		downLoadHistoryRecord = downLoadHistoryData[0]
	}
	return
}

// 停止下载文件
func (a *App) DownLoadTaskStop(shareId int64) (err error) {
	_ = lanHttp.StopTask(shareId)
	return
}

func (a *App) AddShare() (err error) {
	file, err := zenity.SelectFileMultiple(
		zenity.Filename(a.LastShareFolder),
		zenity.FileFilters{
			{"", []string{"*.*"}, false},
		})
	if err != nil {
		return err
	}
	var fileLen = len(file)
	if fileLen == 0 {
		return
	}
	var files []string
	if _, ok := common.NetShare.Host[common.LocalIp.String()]; ok {
		var hasFile = false
		for j := 0; j < fileLen; j++ {
			hasFile = false
			for i := 0; i < len(common.NetShare.Host[common.LocalIp.String()].Share); i++ {
				if file[j] == common.NetShare.Host[common.LocalIp.String()].Share[i].LocalPath {
					hasFile = true
				}
			}
			if !hasFile {
				files = append(files, file[j])
			}
		}
	} else {
		return errors.New("未找到本地数据缓存。")
	}
	if len(files) == 0 {
		return
	}
	var dir = filepath.Dir(file[0])
	var fileInfo os.FileInfo
	var share = common.Share{}
	var sqlStr string
	for _, v := range files {
		fileInfo, err = os.Stat(v)
		if err != nil {
			continue
		}
		share.ShareId = common.GetSnowFlakeId()
		share.Title = fileInfo.Name()
		share.Ext = filepath.Ext(fileInfo.Name())
		share.LocalPath = v
		share.Size = fileInfo.Size()
		share.ExpireTime = 0
		sqlStr += "insert into `share` values(" + cvt.String(share.ShareId) + ",'" + share.Title + "','" + v + "','" + filepath.Ext(v) + "'," + cvt.String(share.Size) + "," + cvt.String(share.ExpireTime) + ",'" + cvt.String(time.Now().Format("2006-01-02 15:04:05")) + "','" + cvt.String(time.Now().Format("2006-01-02 15:04:05")) + "');"
		common.NetShare.Host[common.LocalIp.String()].Share = append(common.NetShare.Host[common.LocalIp.String()].Share, share)
		err = a.lanNet.NewShare(share)
	}
	err = sqlite.Exec(sqlStr)
	a.LastShareFolder = dir
	wailsruntime.EventsEmit(common.AppCtx, "LocalShareCount", len(common.NetShare.Host[common.LocalIp.String()].Share))
	return
}

func (a *App) EditHostName(hostNick string) (err error) {
	if hostNick == "" {
		return errors.New("电脑名称不能为空。")
	}
	if hostNick != common.NetShare.Host[common.LocalIp.String()].Nick {
		common.NetShare.Host[common.LocalIp.String()].Nick = hostNick
		common.Iprofile.Nick = hostNick
		sqlStr := "update profile set nick='" + hostNick + "'"
		err = sqlite.Exec(sqlStr)
		if err != nil {
			logrus.Error("发送修改主机名称通知失败：", err)
		}
		err = a.lanNet.EditHostName(hostNick)
	}
	return
}

func (a *App) SyncHost() (err error) {
	err = a.lanNet.JoinGroup()
	time.Sleep(1 * time.Second)
	return
}

func (a *App) SyncHostShare(ip string) (err error) {
	if ip == common.LocalIp.String() {

	} else {
		var ret = make(map[string]string)
		ret, err = a.lanNet.SyncNewHost(ip)
		if err != nil {
			logrus.Error("同步主机信息错误：", err)
			return
		}
		wailsruntime.EventsEmit(common.AppCtx, "HostJoinGroup", ret)
	}
	return
}

func (a *App) EditShare(shareId string, expireTimeUnit string, expireTimeNum string) (err error) {
	if shareId == "" || expireTimeUnit == "" || expireTimeNum == "" {
		return errors.New("参数错误。")
	}
	var unit, expireTime int64
	switch expireTimeUnit {
	case "分钟":
		unit = 60
	case "小时":
		unit = 60 * 60
	case "天":
		unit = 60 * 60 * 24
	case "永久":
		unit = 0
	}
	expireTime = (unit * cvt.Int64(expireTimeNum)) + time.Now().Unix()
	sql := "update share set expire_time=" + cvt.String(expireTime) + " where share_id=" + shareId + ""
	err = sqlite.Exec(sql)
	if err != nil {
		return
	}
	if _, ok := common.NetShare.Host[common.LocalIp.String()]; ok {
		var shareIdInt int64
		shareIdInt, err = cvt.Int64E(shareId)
		if err != nil {
			return
		}
		for i := 0; i < len(common.NetShare.Host[common.LocalIp.String()].Share); i++ {
			if common.NetShare.Host[common.LocalIp.String()].Share[i].ShareId == shareIdInt {
				common.NetShare.Host[common.LocalIp.String()].Share[i].ExpireTime = expireTime
				break
			}
		}
	}
	return
}

func (a *App) GetHistoryList(title string, ext []string, sType uint8, pageSize int, pageNumber int) (ret common.DownLoadHistoryList, err error) {
	return sqlite.QueryDownloadHistoryList(title, ext, sType, pageSize, pageNumber), nil
}

func (a *App) DelHistory(historyId []string) (err error) {
	sqlStr := "delete from download_history where history_id in (" + strings.Join(historyId, ",") + ")"
	return sqlite.Exec(sqlStr)
}

func (a *App) GetHttpSign() (header map[string]string, err error) {
	header = lanHttp.SignHttpHeader()
	return
}

func (a *App) CodeLogin(code string) (token string, err error) {
	loginCode := common.LoginCode{}
	loginCode.Code = code
	loginCode.DeviceId = a.DeviceId
	loginCode.Origin = common.Origin
	os, err := common.GetOS()
	if err == nil {
		loginCode.Os = os
	}
	cpu, err := common.GetCpu()
	if err == nil {
		loginCode.CpuPhysicalId = cpu.PhysicalID
		loginCode.CpuModelName = cpu.ModelName
		loginCode.CpuLogicalCnt = cpu.Cores
	}
	mem, err := common.GetMem()
	if err == nil {
		loginCode.MemTotal = cvt.Float64(strconv.FormatFloat(cvt.Float64(mem.Total/1024/1024/1024), 'f', 2, 64))
		loginCode.CpuModelName = cpu.ModelName
		loginCode.CpuLogicalCnt = cpu.Cores
	}
	disk, err := common.GetDiskInfo(a.RunPath)
	if err == nil {
		loginCode.DiskDevice = disk.Path
		loginCode.DiskFstype = disk.Fstype
		loginCode.DiskTotal = cvt.Float64(strconv.FormatFloat(cvt.Float64(disk.Total/1024/1024/1024), 'f', 2, 64))
	}
	loginCode.Version = common.Version
	reqConfig := common.ReqConfig{5 * time.Second * time.Second, true}
	sendData, _ := json.Marshal(loginCode)
	ret, err := common.HttpSingleRequest(common.ApiServer+"/private/wechat/officialAccount/login/code", "post", nil, sendData, reqConfig)
	if err != nil {
		return
	}
	retJson := common.JsonResultData{}
	err = json.Unmarshal(ret, &retJson)
	if err != nil {
		return
	}
	if retJson.Code != 0 {
		return "", errors.New(retJson.Msg)
	}
	ret, _ = common.Rc4Decrypt(common.TokenSecret, []byte(cvt.String(retJson.Data)))
	token = string(ret)
	a.userId = token
	return token, nil
}

func (a *App) GetLocalInfo() (ret map[string]string, err error) {
	ret = make(map[string]string)
	ret["nick"] = common.Iprofile.Nick
	ret["ip"] = common.LocalIp.String()
	ret["shareCount"] = cvt.String(len(common.NetShare.Host[common.LocalIp.String()].Share))
	return
}

func (a *App) GetLocalIp() (ip string, err error) {

	ip = common.LocalIp.String()
	return
}
