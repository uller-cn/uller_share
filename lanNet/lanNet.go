package lanNet

import (
	"encoding/json"
	"errors"
	"github.com/shockerli/cvt"
	"github.com/sirupsen/logrus"
	"github.com/uller_share/common"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/net/ipv4"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
	"uller_share/lanHttp"
)

var (
	multicastAddrV4 = &net.UDPAddr{
		IP:   net.ParseIP("224.0.0.254"),
		Port: 35285,
	}
	multicastAddrV6 = &net.UDPAddr{
		IP: net.ParseIP("ff02::1"),
		// IP:   net.ParseIP("fd00::12d3:26e7:48db:e7d"),
		Port: 35285,
	}
	udpPort = 35285
	IsRun   bool
)

type LanNet struct {
	shouldShutdown chan struct{}
	shutdownLock   sync.Mutex
	shutdownEnd    sync.WaitGroup
	ipv4conn       *ipv4.PacketConn
	ifaces         []net.Interface
	localIp        []string
}

func NewLanNet() (l LanNet, err error) {
	l.shouldShutdown = make(chan struct{})
	address := net.JoinHostPort(multicastAddrV4.IP.String(), cvt.String(multicastAddrV4.Port))
	udpConn, err := net.ListenPacket("udp4", address)
	if err != nil {
		return
	}

	l.ipv4conn = ipv4.NewPacketConn(udpConn)
	l.ipv4conn.SetControlMessage(ipv4.FlagInterface, true)
	_ = l.ipv4conn.SetMulticastTTL(255)

	var failedJoins int
	var addrs []net.Addr
	var ipNet *net.IPNet
	var ok bool
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, iface := range ifaces {
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}
		if (iface.Flags&net.FlagMulticast) > 0 && iface.HardwareAddr != nil && !strings.Contains(strings.ToLower(iface.Name), "vethernet") {
			l.ifaces = append(l.ifaces, iface)
			addrs, err = iface.Addrs()
			if err != nil {
				logrus.Error("网卡", iface.Name, "获取ip地址错误：", err)
				continue
			}
			//获取本地所有ipv4地址
			for _, addr := range addrs {
				if ipNet, ok = addr.(*net.IPNet); ok {
					if ipNet.IP.To4() != nil {
						l.localIp = append(l.localIp, ipNet.IP.String())
					}
				}
			}
			if err = l.ipv4conn.JoinGroup(&iface, multicastAddrV4); err != nil {
				logrus.Error("网卡", iface.Name, "加入群错误：", err)
				failedJoins++
				continue
			}
		}
	}
	if failedJoins == len(l.ifaces) {
		l.ipv4conn.Close()
		logrus.Error("启动群消息错误：", err)
	}
	return
}

func (l *LanNet) Shutdown() (err error) {
	l.shutdownLock.Lock()
	defer l.shutdownLock.Unlock()
	close(l.shouldShutdown)
	if l.ipv4conn != nil {
		_ = l.ipv4conn.Close()
	}
	l.shutdownEnd.Wait()
	return err
}

func (l *LanNet) ListenUdp4Multicast() {
	IsRun = true
	go l.recv4()
	return
}

func (l *LanNet) recv4() {
	if l.ipv4conn == nil {
		return
	}
	buf := make([]byte, UdpMtuSize)
	l.shutdownEnd.Add(1)
	defer l.shutdownEnd.Done()
	var udpAddr *net.UDPAddr
	for {
		select {
		case <-l.shouldShutdown:
			return
		default:
			n, _, from, err := l.ipv4conn.ReadFrom(buf)
			if err != nil {
				continue
			}
			udpAddr = from.(*net.UDPAddr)
			go l.unPacket(udpAddr.IP, buf[:n])
		}
	}
}

func (l *LanNet) SendUdp4Multicast(sendData []byte) (err error) {
	var wcm ipv4.ControlMessage
	for ifi := range l.ifaces {
		switch runtime.GOOS {
		case "darwin", "ios", "linux":
			wcm.IfIndex = l.ifaces[ifi].Index
		default:
			if err = l.ipv4conn.SetMulticastInterface(&l.ifaces[ifi]); err != nil {
				logrus.Error("发送群消息失败：", err)
			}
		}
		_, err = l.ipv4conn.WriteTo(sendData, &wcm, multicastAddrV4)
	}
	return
}

func (l *LanNet) ListenMulticast() {
	conn, err := net.ListenMulticastUDP("udp4", nil, multicastAddrV4)
	if err != nil {
		logrus.Error("创建接受群消息错误：", err)
	}
	l.shouldShutdown = make(chan struct{})
	var buffer = make([]byte, UdpMtuSize)
	IsRun = true
	var src *net.UDPAddr
	var n int
	for {
		select {
		case <-l.shouldShutdown:
			break
		default:
			n, src, err = conn.ReadFromUDP(buffer)
			if err != nil {
				logrus.Error("接受群消息错误：", err)
			}
			go l.unPacket(src.IP, buffer[:n])
		}
	}
	_ = conn.Close()
}

func (l *LanNet) unPacket(ip net.IP, buffer []byte) {
	for _, selfIp := range l.localIp {
		if ip.String() == selfIp {
			return
		}
	}
	if common.Iprofile.Nick == "" {
		return
	}
	packet := NewMulticastPacket()
	signRet, share := packet.Decoder(buffer)
	share.Ip = ip.String()
	//fmt.Println("接收到广播消息", buffer[0], signRet, share, ip.String())
	if !signRet {
		return
	}
	switch packet.Type {
	case MulticastJoin:
		//join
		ret, err := l.SyncNewHost(ip.String())
		if err != nil {
			logrus.Error("同步数据错误", err)
			return
		}
		wailsruntime.EventsEmit(common.AppCtx, "HostJoinGroup", ret)
		break
	case MulticastLeave:
		//leave
		common.NetShare.Lock()
		delete(common.NetShare.Host, ip.String())
		common.NetShare.Unlock()
		wailsruntime.EventsEmit(common.AppCtx, "HostLeaveGroup", ip.String())
		break
	case MulticastNewShare:
		//newShare
		downLoadHistory := common.DownLoadHistory{}
		downLoadHistory.Share = share
		downLoadHistory.Title = share.Title
		downLoadHistory.Ip = share.Ip
		downLoadHistory.Ext = share.Ext
		downLoadHistory.Size = share.Size
		downLoadHistory.Status = 0
		common.NetShare.Lock()
		common.NetShare.Host[ip.String()].Share = append(common.NetShare.Host[ip.String()].Share, share)
		common.NetShare.Unlock()
		wailsruntime.EventsEmit(common.AppCtx, "NewShare", downLoadHistory)
		ret := make(map[string]string)
		ret["ip"] = ip.String()
		ret["nick"] = common.NetShare.Host[ip.String()].Nick
		ret["shareCount"] = cvt.String(len(common.NetShare.Host[ip.String()].Share))
		wailsruntime.EventsEmit(common.AppCtx, "NewShareCount", ret)
		break
	case MulticastDelShare:
		//delShare
		err := common.NetShare.Del(share.ShareId, ip.String())
		if err != nil {
			logrus.Error("删除共享文件缓存失败：", err)
			return
		}
		var delShare = common.DelShare{ip.String(), share.ShareId}
		wailsruntime.EventsEmit(common.AppCtx, "DelShare", delShare)
		break
	case MulticastEditHostName:
		//editHostName
		if _, ok := common.NetShare.Host[ip.String()]; ok {
			common.NetShare.Host[ip.String()].Nick = share.Title
			wailsruntime.EventsEmit(common.AppCtx, "EditHostName", common.EditHostName{ip.String(), share.Title})
		}
		break
	}
}

// 加入网络
func (l *LanNet) JoinGroup() (err error) {
	packet := NewMulticastPacket()
	sendData, err := packet.Encoder(MulticastJoin, 0)
	if err != nil {
		return
	}
	err = l.SendUdp4Multicast(sendData)
	return
}

// 应用关闭离开网络
func (l *LanNet) LeaveGroup() (err error) {
	packet := NewMulticastPacket()
	sendData, err := packet.Encoder(MulticastLeave, 0)
	if err != nil {
		return
	}
	err = l.SendUdp4Multicast(sendData)
	return
}

// 新增共享文件
func (l *LanNet) NewShare(share common.Share) (err error) {
	var msgByte []byte
	shareId := common.Int64ToBytes(share.ShareId)
	msgByte = append(msgByte, shareId[:]...)
	ext := []byte(share.Ext)
	for i := len(ext); i < 32; i++ {
		ext = append(ext, 0)
	}
	msgByte = append(msgByte, ext[:]...)
	size := common.Int64ToBytes(share.Size)
	msgByte = append(msgByte, size[:]...)
	expireTime := common.Int64ToBytes(share.ExpireTime)
	msgByte = append(msgByte, expireTime[:]...)
	msgByte = append(msgByte, []byte(share.Title)[:]...)
	packet := NewMulticastPacket()
	sendData, err := packet.Encoder(MulticastNewShare, cvt.Int64(len(share.Title)))
	if err != nil {
		return
	}
	sendData = append(sendData, msgByte[:]...)
	err = l.SendUdp4Multicast(sendData)
	return
}

// 删除共享文件
func (l *LanNet) DelShare(shareId int64) (err error) {
	packet := NewMulticastPacket()
	var msgByte []byte
	shareByte := common.Int64ToBytes(shareId)
	msgByte = append(msgByte, shareByte[:]...)
	sendData, err := packet.Encoder(MulticastDelShare, cvt.Int64(len(msgByte)))
	if err != nil {
		return
	}
	sendData = append(sendData, msgByte[:]...)
	err = l.SendUdp4Multicast(sendData)
	return
}

// 新主机上线同步共享数据
func (l *LanNet) SyncNewHost(ip string) (ret map[string]string, err error) {
	send, _ := json.Marshal(common.NetShare.Host[common.LocalIp.String()])
	reqConfig := common.ReqConfig{5 * time.Second, false}
	url := "http://" + ip + cvt.String(lanHttp.LanHttpPort) + lanHttp.RouterPath["sync"]
	header := lanHttp.SignHttpHeader()
	retByte, err := common.HttpSingleRequest(url, "post", header, send, reqConfig)
	if err != nil {
		return
	}
	retJson := common.JsonShareList{}
	err = json.Unmarshal(retByte, &retJson)
	if err != nil {
		return
	}
	if retJson.Code != 0 {
		return ret, errors.New("共享主机" + ip + "返回错误信息：" + retJson.Msg)
	}
	common.NetShare.Lock()
	common.NetShare.Host[ip] = &retJson.Data
	common.NetShare.Unlock()
	ret = make(map[string]string)
	ret["ip"] = ip
	ret["nick"] = retJson.Data.Nick
	ret["shareCount"] = cvt.String(len(retJson.Data.Share))
	return
}

// 修改主机名称
func (l *LanNet) EditHostName(hostNick string) (err error) {
	packet := NewMulticastPacket()
	var msgByte []byte
	msgByte = []byte(hostNick)
	sendData, err := packet.Encoder(MulticastEditHostName, cvt.Int64(len(msgByte)))
	if err != nil {
		return
	}
	sendData = append(sendData, msgByte[:]...)
	err = l.SendUdp4Multicast(sendData)
	return
}

func GetLocalIP() (ip net.IP, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = localAddr.IP
	return
}

func GetLocalIfi() (netInterface net.Interface, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	var addrs []net.Addr
	for _, ifi := range ifaces {
		if (ifi.Flags & net.FlagUp) == 0 {
			continue
		}
		if (ifi.Flags&net.FlagMulticast) > 0 && ifi.HardwareAddr != nil && !strings.Contains(strings.ToLower(ifi.Name), "vethernet") {
			addrs, err = ifi.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}
				if ipNet.String() == common.LocalIp.String() {
					netInterface = ifi
				}
			}
		}
	}
	return
}

func listMulticastInterfaces() []net.Interface {
	var interfaces []net.Interface
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ifaces {
		if (ifi.Flags & net.FlagUp) == 0 {
			continue
		}
		if (ifi.Flags&net.FlagMulticast) > 0 && ifi.HardwareAddr != nil && !strings.Contains(strings.ToLower(ifi.Name), "vethernet") {
			interfaces = append(interfaces, ifi)
		}
	}
	return interfaces
}

//func sendUdp(sendData []byte) {
//	interfaces := listMulticastInterfaces()
//	var ladder []net.Addr
//	var ip net.IP
//	var err error
//	for _, iface := range interfaces {
//		ladder, err = iface.Addrs()
//		if err != nil {
//			logrus.Error("发送群消息，本地网卡数据获取失败：", err)
//			continue
//		}
//		for _, addr := range ladder {
//			ip = addr.(*net.IPNet).IP
//			if ip.To4() != nil {
//				conn, udpErr := net.DialUDP("udp4", nil, multicastAddrV4)
//				if udpErr != nil {
//					continue
//				}
//				_, udpErr = conn.Write(sendData)
//				if udpErr != nil {
//					logrus.Error("发送v4群消息失败：", udpErr)
//				}
//				_ = conn.Close()
//			} //else if ip.To16() != nil {
//			//	conn, udpErr := net.DialUDP("udp6", nil, multicastAddrV6)
//			//	if udpErr != nil {
//			//		continue
//			//	}
//			//	_, udpErr = conn.Write(sendData)
//			//	if udpErr != nil {
//			//		logrus.Error("发送v6群消息失败：", udpErr)
//			//	}
//			//	_ = conn.Close()
//			//}
//		}
//	}
//	return
//}
