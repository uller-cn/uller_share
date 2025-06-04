package lanNet

import (
	"crypto/md5"
	"errors"
	"github.com/uller_share/common"
	"time"
)

const (
	HeaderLen             = 33   //包头长度
	MulticastJoin         = 0x01 //加入共享网络
	MulticastLeave        = 0x02 //离开共享网络
	MulticastNewShare     = 0x03 //新建共享文件
	MulticastDelShare     = 0x04 //删除共享文件
	MulticastEditHostName = 0x05 //修改主机名称
	UdpMtuSize            = 1472
	MulticastSecret       = "Xyhd34&*^D&S%F123421GSs1"
)

type Packet struct {
	Secret     string
	Type       byte
	Time       [8]byte
	Sign       [16]byte
	Len        [8]byte
	ShareId    [8]byte
	Ext        [32]byte
	Size       [8]byte
	ExpireTime [8]byte
	Title      []byte
}

func NewMulticastPacket() Packet {
	packet := Packet{}
	packet.Secret = MulticastSecret
	return packet
}

func (p *Packet) Encoder(pkgType byte, lenMsg int64) (header []byte, err error) {
	if lenMsg > UdpMtuSize-HeaderLen {
		return header, errors.New("数据包内容超过限制")
	}
	p.Type = pkgType
	p.Time = common.Int64ToBytes(time.Now().Unix())
	p.Len = common.Int64ToBytes(lenMsg)
	var signBytes []byte
	signBytes = append(signBytes, p.Type)
	signBytes = append(signBytes, p.Time[:]...)
	signBytes = append(signBytes, p.Len[:]...)
	signBytes = append(signBytes, []byte(p.Secret)...)
	p.Sign = md5.Sum(signBytes)
	header = append(header, p.Type)
	header = append(header, p.Time[:]...)
	header = append(header, p.Sign[:]...)
	header = append(header, p.Len[:]...)
	return
}

func (p *Packet) Decoder(msgByte []byte) (signRet bool, share common.Share) {
	signRet = false
	if len(msgByte) < 33 {
		return
	}
	p.Type = msgByte[0]
	for i := 1; i < 8; i++ {
		p.Time[i-1] = msgByte[i]
	}
	packetTime := common.BytesToInt64(p.Time)
	t := time.Unix(packetTime, 0)
	minTime := time.Now().Add(-30 * time.Minute)
	maxTime := time.Now().Add(30 * time.Minute)
	if t.Before(minTime) && t.After(maxTime) {
		return
	}
	if time.Now().Unix()-packetTime >= int64(30*time.Second) {
		return
	}
	for i := 25; i < 33; i++ {
		p.Len[i-25] = msgByte[i]
	}
	titleLen := common.BytesToInt64(p.Len)
	signRet = true
	var signBytes []byte
	signBytes = append(signBytes, p.Type)
	signBytes = append(signBytes, p.Time[:]...)
	signBytes = append(signBytes, p.Len[:]...)
	signBytes = append(signBytes, []byte(p.Secret)...)
	sign := md5.Sum(signBytes)
	for i := 9; i < 25; i++ {
		p.Sign[i-9] = msgByte[i]
		if sign[i-9] != p.Sign[i-9] {
			signRet = false
			break
		}
	}
	if !signRet {
		return
	}
	switch p.Type {
	case MulticastNewShare:
		for i := 33; i < 41; i++ {
			if msgByte[i] == 0x00 {
				break
			}
			p.ShareId[i-33] = msgByte[i]
		}
		share.ShareId = common.BytesToInt64(p.ShareId)
		extLen := 0
		for i := 41; i < 73; i++ {
			if msgByte[i] == 0x00 {
				break
			}
			p.Ext[i-41] = msgByte[i]
			extLen++
		}
		share.Ext = string(p.Ext[:extLen])
		for i := 73; i < 81; i++ {
			p.Size[i-73] = msgByte[i]
		}
		share.Size = common.BytesToInt64(p.Size)
		for i := 81; i < 89; i++ {
			p.ExpireTime[i-81] = msgByte[i]
		}
		share.ExpireTime = common.BytesToInt64(p.ExpireTime)
		for i := int64(89); i < 89+titleLen; i++ {
			p.Title = append(p.Title, msgByte[i])
		}
		share.Title = string(p.Title)
	case MulticastDelShare:
		for i := 33; i < 41; i++ {
			if msgByte[i] == 0x00 {
				break
			}
			p.ShareId[i-33] = msgByte[i]
		}
		share.ShareId = common.BytesToInt64(p.ShareId)
	case MulticastEditHostName:
		for i := int64(33); i < 33+titleLen; i++ {
			p.Title = append(p.Title, msgByte[i])
		}
		share.Title = string(p.Title)
	}
	return
}
