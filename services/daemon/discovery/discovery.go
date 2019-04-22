// @APIVersion 1.0.0
// @Title UDP广播服务
// @Description 提供UDP广播服务，为客户端发现服务端的IP地址提供服务的具体实现
// @Author xuchuangxin@icanmake.cn
// @Date 2018-04-26

package discovery

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/ipv4"
)

// Discovered 发现对等广播主机，保存了它们的本地地址（端口被移除）以及有效载荷。
type Discovered struct {
	// 发现广播主机的IP地址
	Address string
	// 广播的信息
	Payload []byte
}

//Settings 配置信息
type Settings struct {
	// 限制是要发现的广播主机的数量，使用<1表示无限制。
	Limit int
	// 广播的端口，默认为9999
	Port string
	// 指定广播的网段，规定的广播地址为（224.0.0.0 ~ 239.255.255.255)。默认为 (239.255.255.250).
	MulticastAddress string
	// 广播的消息
	Payload []byte
	// 广播之间的延迟时间。 默认延迟是1秒。
	Delay time.Duration
	// 广播的时间，默认为10秒
	//TimeLimit time.Duration

	portNum                 int
	multicastAddressNumbers []uint8
}

// 局域网的广播主机
type peerDiscovery struct {
	settings Settings

	received map[string][]byte
	sync.RWMutex
}

// 初始化广播实例
func initialize(settings Settings) (p *peerDiscovery, err error) {
	p = new(peerDiscovery)
	p.Lock()
	defer p.Unlock()

	// 初始化
	p.settings = settings

	if p.settings.Port == "" {
		p.settings.Port = "9999"
	}
	if p.settings.MulticastAddress == "" {
		p.settings.MulticastAddress = "239.255.255.250"
	}
	if len(p.settings.Payload) == 0 {
		p.settings.Payload = []byte("hi")
	}
	if p.settings.Delay == time.Duration(0) {
		p.settings.Delay = 1 * time.Second
	}
	// if p.settings.TimeLimit == time.Duration(0) {
	// 	p.settings.TimeLimit = 10 * time.Second
	// }
	p.received = make(map[string][]byte)
	p.settings.multicastAddressNumbers = []uint8{0, 0, 0, 0}
	for i, num := range strings.Split(p.settings.MulticastAddress, ".") {
		var nInt int
		nInt, err = strconv.Atoi(num)
		if err != nil {
			return
		}
		p.settings.multicastAddressNumbers[i] = uint8(nInt)
	}
	p.settings.portNum, err = strconv.Atoi(p.settings.Port)
	if err != nil {
		return
	}
	return
}

// Discover 扫描LAN广播主机。
// 返回一个发现的对等体及其相关负载的数组， 不会返回发送给自己的广播。
func Discover(settings ...Settings) (discoveries []Discovered, err error) {
	s := Settings{}
	if len(settings) > 0 {
		s = settings[0]
	}
	p, err := initialize(s)
	if err != nil {
		return
	}

	p.RLock()
	address := p.settings.MulticastAddress + ":" + p.settings.Port
	portNum := p.settings.portNum
	multicastAddressNumbers := p.settings.multicastAddressNumbers
	payload := p.settings.Payload
	tickerDuration := p.settings.Delay
	//timeLimit := p.settings.TimeLimit
	p.RUnlock()

	// 获取网络接口
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	c, err := net.ListenPacket("udp4", address)
	if err != nil {
		return
	}
	defer c.Close()

	group := net.IPv4(multicastAddressNumbers[0], multicastAddressNumbers[1], multicastAddressNumbers[2], multicastAddressNumbers[3])
	p2 := ipv4.NewPacketConn(c)

	for i := range ifaces {
		if errJoinGroup := p2.JoinGroup(&ifaces[i], &net.UDPAddr{IP: group, Port: portNum}); errJoinGroup != nil {
			continue
		}
	}

	go p.listen()
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()
	//start := time.Now()
	for _ = range ticker.C {
		// fmt.Println(t)
		exit := false
		p.Lock()
		if len(p.received) >= p.settings.Limit && p.settings.Limit > 0 {
			exit = true
		}
		p.Unlock()

		// 发送多播
		dst := &net.UDPAddr{IP: group, Port: portNum}
		for i := range ifaces {
			if errMulticast := p2.SetMulticastInterface(&ifaces[i]); errMulticast != nil {
				continue
			}
			p2.SetMulticastTTL(2)
			if _, errMulticast := p2.WriteTo([]byte(payload), nil, dst); errMulticast != nil {
				continue
			}
		}

		// if exit || t.Sub(start) > timeLimit {
		// 	break
		// }

		if exit {
			break
		}
	}

	// 广播发送完毕
	dst := &net.UDPAddr{IP: group, Port: portNum}
	for i := range ifaces {
		if errMulticast := p2.SetMulticastInterface(&ifaces[i]); errMulticast != nil {
			continue
		}
		p2.SetMulticastTTL(2)
		if _, errMulticast := p2.WriteTo([]byte(payload), nil, dst); errMulticast != nil {
			continue
		}
	}

	p.Lock()
	discoveries = make([]Discovered, len(p.received))
	i := 0
	for ip := range p.received {
		discoveries[i] = Discovered{
			Address: ip,
			Payload: p.received[ip],
		}
		i++
	}
	p.Unlock()
	return
}

const (
	// 报文规范
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	maxDatagramSize = 66507
)

// 监听绑定到给定的UDP地址和端口，并将从该地址收到的数据包写入传递给处理器的缓冲区
func (p *peerDiscovery) listen() (recievedBytes []byte, err error) {
	p.RLock()
	address := p.settings.MulticastAddress + ":" + p.settings.Port
	portNum := p.settings.portNum
	multicastAddressNumbers := p.settings.multicastAddressNumbers
	p.RUnlock()
	localIPs := getLocalIPs()

	// 获取网络接口
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	// 打开udp连接
	c, err := net.ListenPacket("udp4", address)
	if err != nil {
		return
	}
	defer c.Close()

	group := net.IPv4(multicastAddressNumbers[0], multicastAddressNumbers[1], multicastAddressNumbers[2], multicastAddressNumbers[3])
	p2 := ipv4.NewPacketConn(c)
	for i := range ifaces {
		if errJoinGroup := p2.JoinGroup(&ifaces[i], &net.UDPAddr{IP: group, Port: portNum}); errJoinGroup != nil {
			continue
		}
	}

	// 循环读取套接字
	for {
		buffer := make([]byte, maxDatagramSize)
		n, _, src, errRead := p2.ReadFrom(buffer)
		if errRead != nil {
			err = errRead
			return
		}

		if _, ok := localIPs[strings.Split(src.String(), ":")[0]]; ok {
			continue
		}

		p.Lock()
		if _, ok := p.received[strings.Split(src.String(), ":")[0]]; !ok {
			p.received[strings.Split(src.String(), ":")[0]] = buffer[:n]
		}
		if len(p.received) >= p.settings.Limit && p.settings.Limit > 0 {
			p.Unlock()
			break
		}
		p.Unlock()
	}

	return
}

// 返回本地IP地址
func getLocalIPs() (ips map[string]struct{}) {
	ips = make(map[string]struct{})
	ips["localhost"] = struct{}{}
	ips["127.0.0.1"] = struct{}{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, address := range addrs {
		ips[strings.Split(address.String(), "/")[0]] = struct{}{}
	}
	return
}
