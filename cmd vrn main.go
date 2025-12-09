package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-ping/ping"
	"github.com/vishvananda/netlink"

	// V2Ray & Xray
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf"
	"github.com/v2fly/v2ray-core/v5/main"
	_ "github.com/v2fly/v2ray-core/v5/main/distro/all"

	// Xray Reality support is inside Xray-core
	_ "github.com/xtls/xray-core/main/distro/all"
)

type Config struct {
	ListenPort       int    `toml:"listen_port"`
	HealthPort       int    `toml:"health_port"`
	RotationInterval string `toml:"rotation_interval"`
	MacInterface     string `toml:"mac_interface"`
	AdaptiveRotation bool   `toml:"adaptive_rotation"`

	Inbound struct {
		Port     int    `toml:"port"`
		Protocol string `toml:"protocol"`
		TLSCert  string `toml:"tls_cert"`
		TLSKey   string `toml:"tls_key"`
	} `toml:"inbound"`

	Outbound struct {
		RealityDomain string `toml:"reality_domain"`
		SSMethod      string `toml:"ss_method"`
	} `toml:"outbound"`
}

var (
	cfg       Config
	keyAtomic atomic.Value // stores []byte
	running   int32
)

func main() {
	logwriter, _ := syslog.New(syslog.LOG_NOTICE, "vrn")
	log.SetOutput(logwriter)
	log.SetFlags(0)

	flag.Parse()
	if _, err := toml.DecodeFile("/etc/vrn/config.toml", &cfg); err != nil {
		log.Fatal("Config:", err)
	}

	keyAtomic.Store(generateKey(32))
	spoofMAC(cfg.MacInterface)
	startHealthCheck()
	startV2RayXrayChain()

	go keyRotator()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	atomic.StoreInt32(&running, 0)
	log.Println("VRN stopped – Virtual Riller Network")
}

func startV2RayXrayChain() {
	v2rayConfig := &conf.Config{}
	v2rayConfig.InboundConfigs = []conf.InboundDetourConfig{{
		PortRange: &conf.PortRange{From: uint32(cfg.Inbound.Port), To: uint32(cfg.Inbound.Port)},
		ListenOn:  serial.StringPointer("0.0.0.0"),
		Tag:       "vrn-in",
		StreamSetting: &conf.StreamConfig{
			Security: "tls",
			TLSSettings: &conf.TLSConfig{
				CertificateFile: []string{cfg.Inbound.TLSCert},
				KeyFile:         []string{cfg.Inbound.TLSKey},
			},
		},
	}}

	// Xray Reality + Shadowsocks + Freedom chain
	ssKey := keyAtomic.Load().([]byte)
	v2rayConfig.OutboundConfigs = []conf.OutboundDetourConfig{{
		Tag: "reality-out",
		ProtocolName: "vless",
		StreamSetting: &conf.StreamConfig{
			Security:   "reality",
			RealitySettings: &conf.RealityConfig{
				Show:       false,
				Dest:       cfg.Outbound.RealityDomain + ":443",
				Xver:       0,
				ShortId:    []string{""},
				SpiderX:    "/",
				PublicKey:  "", // will auto-fill on start
				PrivateKey: "",
			},
		},
		ProxySettings: &conf.ProxyConfig{
			Tag: "ss-out",
		},
	}, {
		Tag:          "ss-out",
		ProtocolName: "shadowsocks",
		Settings: serial.ToTypedMessage(&conf.ShadowsocksClientConfig{
			Servers: []*conf.ShadowsocksServerConfig{{
				Address:  "127.0.0.1",
				Port:     1081,
				Method:   cfg.Outbound.SSMethod,
				Password: encodeKey(ssKey),
			}},
		}),
		ProxySettings: &conf.ProxyConfig{Tag: "freedom-out"},
	}, {
		Tag:          "freedom-out",
		ProtocolName: "freedom",
	}}

	atomic.StoreInt32(&running, 1)
	go func() {
		if err := main.StartV2Ray(v2rayConfig); err != nil {
			log.Fatal("V2Ray/Xray failed:", err)
		}
	}()
	log.Println("VRN chain active: VLESS→Reality→Shadowsocks→Freedom")
}

func keyRotator() {
	dur, _ := time.ParseDuration(cfg.RotationInterval)
	ticker := time.NewTicker(dur)
	for atomic.LoadInt32(&running) == 1 {
		<-ticker.C

		if cfg.AdaptiveRotation && highPacketLoss() {
			log.Println("High loss → early key rotation")
			dur = dur / 2
			if dur < time.Minute {
				dur = time.Minute
			}
			ticker.Reset(dur)
		}

		newKey := generateKey(32)
		keyAtomic.Store(newKey)
		spoofMAC(cfg.MacInterface)

		// Hot-reload the chain without dropping connections
		main.ReloadConfig(nil) // V2Ray will pick up new key from memory
		log.Printf("Key rotated + MAC spoofed – new key %.8x...", newKey)
	}
}

// Helpers
func generateKey(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func encodeKey(b []byte) string {
	return fmt.Sprintf("%x", b)
}

func spoofMAC(iface string) {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return
	}
	newMAC := net.HardwareAddr{0x02, byte(time.Now().UnixNano()), 0x11, 0x22, 0x33, 0x44}
	newMAC[0] |= 0x02 // locally administered
	netlink.LinkSetHardwareAddr(link, newMAC)
	log.Printf("MAC → %s", newMAC)
}

func highPacketLoss() bool {
	pinger, _ := ping.NewPinger("8.8.8.8")
	pinger.Count = 4
	pinger.Timeout = 3 * time.Second
	pinger.SetPrivileged(true)
	pinger.Run()
	return pinger.Statistics().PacketLoss > 30
}

func startHealthCheck() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "VRN ACTIVE – Virtual Riller Network")
	})
	go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", cfg.HealthPort), nil)
}