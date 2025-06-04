package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/hbollon/go-edlib"
	"github.com/shockerli/cvt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

type JsonResultData struct {
	Code  int         `json:"code"`            //返回0、正常 1前端弹出服务器错误提示 2前端需要提示用户确认操作 3凭证过期需要重新登录 4凭证过期需要重新刷新凭证
	Msg   string      `json:"msg"`             //返回信息
	Data  interface{} `json:"data"`            //返回数据
	Total *int64      `json:"total,omitempty"` //返回数据行数
}

type JsonShareList struct {
	Code  int      `json:"code"`            //返回0、正常 1前端弹出服务器错误提示 2前端需要提示用户确认操作 3凭证过期需要重新登录 4凭证过期需要重新刷新凭证
	Msg   string   `json:"msg"`             //返回信息
	Data  HostData `json:"data"`            //返回数据
	Total *int64   `json:"total,omitempty"` //返回数据行数
}

type LoginCode struct {
	Code          string  `json:"code"`
	Version       string  `json:"version"`
	DeviceId      string  `json:"deviceId"`
	UserId        string  `json:"userId"`
	GroupId       uint64  `json:"groupId"`
	Origin        string  `json:"origin"`
	Soft          uint8   `json:"soft"`
	Os            string  `json:"os"`
	CpuPhysicalId string  `json:"cpuPhysicalId"`
	CpuModelName  string  `json:"cpuModelName"`
	CpuLogicalCnt int32   `json:"cpuLogicalCnt"`
	CpuMhz        float64 `json:"cpuMhz"`
	MemTotal      float64 `json:"memTotal"`
	DiskDevice    string  `json:"diskDevice"`
	DiskFstype    string  `json:"diskFstype"`
	DiskTotal     float64 `json:"diskTotal"`
}

type Page struct {
	PageSize   int64
	PageNumber int64
}

type ShareDownLoad struct {
	ShareId   []int64 `json:"shareId"`
	HistoryId []int64 `json:"historyId"`
}

type ShareList struct {
	Share []Share `json:"share"`
	Total int64   `json:"total,omitempty"` //返回数据行数
}

type Share struct {
	ShareId    int64     `json:"shareId"`
	Title      string    `json:"title"`
	LocalPath  string    `json:"localPath"`
	Ext        string    `json:"ext"`
	Size       int64     `json:"size"`
	ExpireTime int64     `json:"expireTime"`
	Sim        float32   `json:"-"`
	Ip         string    `json:"ip"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

type DownLoadHistoryList struct {
	DownLoadHistory []DownLoadHistory `json:"downLoadHistory"`
	Total           int64             `json:"total,omitempty"` //返回数据行数
}

type DownLoadHistory struct {
	HistoryId  int64     `json:"historyId"`
	Title      string    `json:"title"`
	LocalPath  string    `json:"localPath"`
	Ip         string    `json:"ip"`
	Share      Share     `json:"share"`
	Ext        string    `json:"ext"`
	Size       int64     `json:"size"`
	Finish     int64     `json:"finish"`
	Status     uint8     `json:"status"` //0没有加入下载队列，1未开始下载，2已完成下载，3下载中，4手动停止下载，5下载出错，6共享文件已过期，7共享文件已删除
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

var AppCtx context.Context
var LocalIp net.IP
var Iprofile Profile

type Profile struct {
	Nick            string `json:"nick"`
	LastShareFolder string `json:"lastShareFolder"`
	QrTime          int64  `json:"qrTime"`
	DownloadRoutine uint16 `json:"downloadRoutine"`
	UpdateTime      string `json:"updateTime"`
}

var NetShare netShare

type netShare struct {
	sync.Mutex
	Host map[string]*HostData `json:"host"`
}

type DelShare struct {
	Ip      string `json:"ip"`
	ShareId int64  `json:"shareId"`
}

type EditHostName struct {
	Ip   string `json:"ip"`
	Nick string `json:"nick"`
}

type HostData struct {
	Nick  string  `json:"nick"`
	Ip    string  `json:"ip"`
	Share []Share `json:"share"`
}

type HostList struct {
	Nick       string `json:"nick"`
	Ip         string `json:"ip"`
	ShareCount int    `json:"shareCount"`
	Sim        float32
}

func (n *netShare) GetHostList(nick string) (ret []map[string]string) {
	item := make(map[string]string)
	if nick == "" {
		for key, v := range n.Host {
			item["host"] = key
			item["nick"] = v.Nick
			item["fileCount"] = cvt.String(len(n.Host[key].Share))
			ret = append(ret, item)
		}
	} else {
		for key, v := range n.Host {
			if v.Nick == nick {
				item["host"] = key
				item["nick"] = v.Nick
				item["fileCount"] = cvt.String(len(n.Host[key].Share))
				ret = append(ret, item)
			}
		}
	}
	return
}

func (n *netShare) Del(shareId int64, ip string) (err error) {
	if _, ok := n.Host[ip]; !ok {
		return errors.New("ip不存在")
	}
	n.Lock()
	for i := 0; i < len(n.Host[ip].Share); i++ {
		if n.Host[ip].Share[i].ShareId == shareId {
			n.Host[ip].Share = append(n.Host[ip].Share[:i], n.Host[ip].Share[i+1:]...)
			break
		}
	}
	n.Unlock()
	return
}

func (n *netShare) DelIds(shareId []int64, ip string) (err error) {
	if _, ok := n.Host[ip]; !ok {
		return errors.New("ip不存在")
	}
	n.Lock()
	for i := 0; i < len(n.Host[ip].Share); i++ {
		for j := 0; j < len(shareId); j++ {
			if n.Host[ip].Share[i].ShareId == shareId[j] {
				n.Host[ip].Share = append(n.Host[ip].Share[:i], n.Host[ip].Share[i+1:]...)
			}
		}
	}
	n.Unlock()
	return
}

func (n *netShare) GetSame(title string, size int64) (share Share, err error) {
	n.Lock()
EndFor:
	for _, value := range n.Host {
		for i := 0; i < len(value.Share); i++ {
			if value.Share[i].Title == title && value.Share[i].Size == size {
				share = value.Share[i]
				break EndFor
			}
		}
	}
	n.Unlock()
	return
}

func (n *netShare) Get(shareId int64, ip string) (share Share, err error) {
	if _, ok := n.Host[ip]; !ok {
		return share, errors.New("ip不存在")
	}
	n.Lock()
	for i := 0; i < len(n.Host[ip].Share); i++ {
		if n.Host[ip].Share[i].ShareId == shareId {
			share = n.Host[ip].Share[i]
			break
		}
	}
	n.Unlock()
	return
}

func (n *netShare) GetIds(shareId []int64, ip string) (share []Share, err error) {
	if ip == "" {
		for _, value := range n.Host {
			for i := 0; i < len(value.Share); i++ {
				for j := 0; j < len(shareId); j++ {
					if value.Share[i].ShareId == shareId[j] {
						share = append(share, value.Share[i])
					}
				}
			}
		}
	} else {
		if _, ok := n.Host[ip]; !ok {
			return share, errors.New("ip不存在")
		}
		n.Lock()
		for i := 0; i < len(n.Host[ip].Share); i++ {
			for j := 0; j < len(shareId); j++ {
				if n.Host[ip].Share[i].ShareId == shareId[j] {
					share = append(share, n.Host[ip].Share[i])
				}
			}
		}
		n.Unlock()
	}
	return
}

func (n *netShare) LikeNick(targetStr string) (ret []HostList) {
	if strings.Trim(targetStr, " ") == "" {
		for _, host := range n.Host {
			ret = append(ret, HostList{host.Nick, host.Ip, len(host.Share), 1})
		}
	} else {
		var sim float32
		var err error
		for _, host := range n.Host {
			sim, err = edlib.StringsSimilarity(strings.ToLower(targetStr), strings.ToLower(host.Nick), edlib.JaroWinkler)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if sim >= 0.7 {
				ret = append(ret, HostList{host.Nick, host.Ip, len(host.Share), sim})
			}
		}
		sort.Slice(ret, func(i, j int) bool {
			return ret[i].Sim < ret[j].Sim
		})
	}
	return ret
}
