package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/go-ping/ping"
	"github.com/hbollon/go-edlib"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shockerli/cvt"
	"github.com/yitter/idgenerator-go/idgen"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type ReqConfig struct {
	TimeOut time.Duration
	Encrypt bool
}

// 字符串数组查找函数
func FindStringInArray(strArray []string, searchStr string) int {
	for index, value := range strArray {
		if value == searchStr {
			return index
		}
	}
	return -1
}

/*
* 获取当前运行程序的运行路径
* auth guolei at 20191128
* return 程序文件运行路径
 */
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err.Error())
	}
	return strings.Replace(dir, "\\", "/", -1)
}

/*
* 判断文件是否存在
* auth guolei at 20191128
* param file 文件全路径
* return true文件存在，false不存在
 */
func FileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

/*
* 获取中文字符在字符串中的位置
* auth guolei at 20191101
* param str 要检索的字符串
* param substr 包含的字符串
* return int 包含的字符串出现的位置，-1未检索到
 */
func UnicodeIndex(str, substr string) int {
	result := strings.Index(str, substr)
	if result != -1 {
		prefix := []byte(str)[0:result]
		rs := []rune(string(prefix))
		result = len(rs)
	}
	return result
}

/*
* http 协程请求
* auth guolei at 20191101
* param url 请求url全路径带http、https
* param method 请求方式get、post
* param ch 通道，请求结果通知给chan
* param header 自定义请求头
* param data 发送请求数据
 */
func HttpRequest(url string, method string, ch chan<- []byte, chErr chan<- error, header map[string]string, data []byte, config ...ReqConfig) {
	c := ReqConfig{}
	if len(config) == 0 {
		c.TimeOut = 30 * time.Second
		c.Encrypt = false
	} else {
		c.TimeOut = config[0].TimeOut
		c.Encrypt = config[0].Encrypt
	}
	var err = errors.New("")
	if c.Encrypt {
		data, err = Rc4Encrypt(HttpSecret, data)
		if err != nil {
			ch <- nil
			chErr <- err
			return
		}
	}
	reqBytes := &bytes.Buffer{}
	if data != nil {
		reqBytes = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(strings.ToUpper(method), url, reqBytes)
	if err != nil {
		ch <- nil
		chErr <- err
		return
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	resp, err := (&http.Client{Timeout: c.TimeOut}).Do(req)
	if err != nil {
		ch <- nil
		chErr <- err
		return
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		//fmt.Println(string(body))
		if c.Encrypt {
			body, err = Rc4Decrypt(HttpSecret, body)
			if err != nil {
				ch <- nil
				chErr <- err
				return
			}
		}
		ch <- body
		chErr <- nil
	}
}

func HttpSingleRequest(url string, method string, header map[string]string, data []byte, config ...ReqConfig) (resonse []byte, err error) {
	c := ReqConfig{}
	if len(config) == 0 {
		c.TimeOut = 30 * time.Second
		c.Encrypt = false
	} else {
		c.TimeOut = config[0].TimeOut
		c.Encrypt = config[0].Encrypt
	}
	if c.Encrypt {
		data, err = Rc4Encrypt(HttpSecret, data)
		if err != nil {
			return nil, err
		}
	}
	reqBytes := &bytes.Buffer{}
	if data != nil {
		reqBytes = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(strings.ToUpper(method), url, reqBytes)
	if err != nil {
		return nil, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	resp, err := (&http.Client{Timeout: c.TimeOut}).Do(req)
	if err != nil {
		return resonse, err
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		err = nil
		if resp.StatusCode == 200 {
			if c.Encrypt {
				body, err = Rc4Decrypt(HttpSecret, body)
				if err != nil {
					return nil, err
				}
			}

		} else {
			return nil, errors.New("服务器错误状态码为" + resp.Status)
		}
		return body, err
	}
}

/*
* 数字数组转换为英文逗号分隔的字符串
* auth guolei at 20191105
* param obj 数字数组
* return 以逗号分隔的字符串
 */
func IntArrToString(slice []int64) (ret string) {
	var builder strings.Builder
	for i, num := range slice {
		builder.WriteString(strconv.FormatInt(num, 10)) // 转换为字符串
		if i < len(slice)-1 {
			builder.WriteString(",")
		}
	}
	return builder.String()
}

/*
* 字符串数组转换为英文逗号分隔的字符串
* auth guolei at 20191105
* param obj 字符串数组
* return 以逗号分隔的字符串
 */
func StringArrToString(obj []string) (ret string) {
	for i := 0; i < len(obj); i++ {
		ret += "'" + obj[i] + "',"
	}
	return ret[0 : len(ret)-1]
}

/*
* 判断url是否为合法url
 */
func IsURL(str string) bool {
	_, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ftp" || u.Scheme == "ftps"
}

/*
* 生成机器码
 */
func GetMachineId() (id string) {
	id, err := machineid.ID()
	if err != nil {
		fmt.Println(err)
	}
	return
}

/*
* 获取操作系统信息
 */
func GetOS() (os string, err error) {
	os, _, _, err = host.PlatformInformation()
	return
}

/*
* 获取cpu使用率
 */
func GetCpuPercent() (cpuPercent float64, err error) {
	percent, err := cpu.Percent(100*time.Millisecond, true)
	if err != nil {
		return
	}
	for _, v := range percent {
		if v != 0 {
			cpuPercent = v
			break
		}
	}
	cpuPercent = cpuPercent / 100
	cpuPercent = cvt.Float64(fmt.Sprintf("%.2f", cpuPercent))
	return
}

/*
* 获取cpu信息
 */
func GetCpu() (cpuInfo cpu.InfoStat, err error) {
	cpus, err := cpu.Info()
	if err != nil {
		return
	}
	if len(cpus) > 0 {
		cpuInfo = cpus[0]
	}
	return
}

/*
* 获取磁盘信息
 */
func GetDiskInfo(dir string) (diskInfo *disk.UsageStat, err error) {
	if strings.Contains(runtime.GOOS, "win") {
		drive := filepath.VolumeName(dir)
		diskInfo, err = disk.Usage(drive)
	} else {
		err = errors.New("不是windows系统，暂不支持")
	}
	return
}

/*
* 获取内存使用率
 */
func GetMemPercent() (memPercent float64, err error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	memPercent = memInfo.UsedPercent
	memPercent = memPercent / 100
	memPercent = cvt.Float64(fmt.Sprintf("%.2f", memInfo.UsedPercent))
	return
}

/*
* 获取内存信息
 */
func GetMem() (memInfo *mem.VirtualMemoryStat, err error) {
	memInfo, err = mem.VirtualMemory()
	if err != nil {
		return
	}
	return
}

/*
* 检测操作系统大小端
* auth guolei at 20210730
 */
func IsLittleEndian() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	p := (*[4]byte)(u)
	return (*p)[0] == 0x04
}

// 获取雪花编号
func GetSnowFlakeId() int64 {
	idgen.SetIdGenerator(idgen.NewIdGeneratorOptions(1))
	return idgen.NextId()
}

func Ping(host string) (ret bool, err error) {
	ret = false
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return
	}
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}
	pinger.Interval = 50 * time.Millisecond
	pinger.Count = 4
	err = pinger.Run()
	if err != nil {
		return
	}
	stats := pinger.Statistics()
	if stats.PacketLoss >= 2 {
		return false, nil
	}
	return true, nil
}

/*
* int64转[]byte
* auth guolei at 20210730
 */
func Int64ToBytes(num int64) [8]byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	if IsLittleEndian() {
		binary.Write(bytesBuffer, binary.LittleEndian, num)
	} else {
		binary.Write(bytesBuffer, binary.BigEndian, num)
	}
	var ret [8]byte
	for i := 0; i < len(ret); i++ {
		ret[i] = bytesBuffer.Bytes()[i]
	}
	return ret
}

/*
* 字节转换成int64
* auth guolei at 20210730
 */
func BytesToInt64(b [8]byte) int64 {
	if IsLittleEndian() {
		return int64(binary.LittleEndian.Uint64(b[0:8]))
	} else {
		return int64(binary.BigEndian.Uint64(b[0:8]))
	}
}

/*
* 验证字符串是否为数字
* auth guolei at 20191105
* param string 字符串
* return bool
 */
func IsNum(s string) bool {
	_, err := strconv.ParseInt(s, 0, 64)
	return err == nil
}

func Like(shareList []Share, targetStr string) (ret []Share) {
	var sim float32
	var err error
	for _, share := range shareList {
		sim, err = edlib.StringsSimilarity(strings.ToLower(share.Title), strings.ToLower(targetStr), edlib.JaroWinkler)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if sim >= 0.7 {
			share.Sim = sim
			ret = append(ret, share)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Sim < ret[j].Sim
	})
	return ret
}

func GetPartitions() (partitions []disk.PartitionStat, err error) {
	return disk.Partitions(false)
}

func GetSysDefaultIP() (ip string, err error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	err = syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return "", err
	}
	ipByte := []byte{}
	for adapter := a; adapter != nil; adapter = adapter.Next {
		if adapter.GatewayList.IpAddress.String[0] != 0 {
			for i := 0; i < len(adapter.IpAddressList.IpAddress.String); i++ {
				if adapter.IpAddressList.IpAddress.String[i] != 0 {
					ipByte = append(ipByte, adapter.IpAddressList.IpAddress.String[i])
				}
			}
			ip = string(ipByte)
			break
		}
	}
	return
}

func KeepWinSysActive() {
	const (
		EsSystemRequired = 0x00000001
		EsContinuous     = 0x80000000
	)
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setThreadExecStateProc := kernel32.NewProc("SetThreadExecutionState")

	pulse := time.NewTicker(10 * time.Second)
	log.Println("Starting keep alive poll...")

	for {
		select {
		case <-pulse.C:
			setThreadExecStateProc.Call(uintptr(EsSystemRequired | EsContinuous))
		}
		time.Sleep(1 * time.Minute)
	}
}

func IsValidTimestamp(timestamp int64) bool {
	t := time.Unix(timestamp, 0)
	minTime := time.Now().Add(-24 * time.Hour)
	maxTime := time.Now().Add(24 * time.Hour)
	return !t.Before(minTime) && !t.After(maxTime)
}
