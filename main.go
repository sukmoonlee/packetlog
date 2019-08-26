// Copyright 2019 smlee@sk.com, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// Package dnslog is DNS Packet logging program
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// PGversion is build version - read README/CHANGE
const PGversion = "1.0.0-20190824"

var (
	iface      = flag.String("i", getEnvStr("DNSLOG_DEVICE", "bond1"), "Interface to read from")
	cpuprofile = flag.String("cpuprofile", "", "If non-empty, write CPU profile here")
	snaplen    = flag.Int("s", getEnvInt("DNSLOG_SNAPLEN", 1560), "Snaplen, if <= 0, use 65535")
	bufferSize = flag.Int("b", getEnvInt("DNSLOG_BUFSIZE", 8), "Interface buffersize (MB)")
	filter     = flag.String("f", getEnvStr("DNSLOG_FILTER", "port 53"), "BPF filter")
	count      = flag.Int64("c", -1, "If >= 0, # of packets to capture before returning")
	verbose    = flag.Int64("log_every", 100000, "Write a every X packets stat")
	logSplit   = flag.Bool("log_split", false, "If true, write a split log file (CQ, CR, SQ, SR)")
	addVLAN    = flag.Bool("add_vlan", false, "If true, add VLAN header")
	logdir     = flag.String("d", getEnvStr("DNSLOG_DIR", "/log/dnslog"), "Write directory for log file")
	showVer    = flag.Bool("v", false, "If true, show version")
)

func getEnvStr(name string, def string) string {
	content, found := os.LookupEnv(name)
	if found {
		return content
	}
	return def
}

func getEnvInt(name string, def int) int {
	content, found := os.LookupEnv(name)
	if found {
		parsed, err := strconv.ParseInt(content, 0, 32)
		if err == nil {
			return int(parsed)
		}
		log.Printf("Could not parse the content of %s, %s, as an int", name, content)
		return def
	}
	return def
}

func main() {
	flag.Parse()
	if *showVer {
		fmt.Printf("dnslog version: %s\n", PGversion)
		return
	}

	log.SetFlags(log.Lshortfile)
	if *cpuprofile != "" {
		log.Printf("Writing CPU profile to %q", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	if *snaplen <= 0 {
		*snaplen = 65535
	}

	logger, err := syslog.New(syslog.LOG_INFO, "")
	if err != nil {
		log.Fatalf("failed syslog : %s", err)
	}
	//log.SetOutput(io.MultiWriter(os.Stderr, logger))
	log.SetOutput(io.MultiWriter(logger))
	log.Printf("Starting on dns packetlog %s (filter:'%s', SnapLen:%d)", PGversion, *filter, *snaplen)

	// Afpacket
	szFrame, szBlock, numBlocks, err := afpacketComputeSize(*bufferSize, *snaplen, os.Getpagesize())
	if err != nil {
		log.Fatal(err)
	}
	afpacketHandle, err := newAfpacketHandle(*iface, szFrame, szBlock, numBlocks, *addVLAN, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	err = afpacketHandle.SetBPFFilter(*filter, *snaplen)
	if err != nil {
		log.Fatal(err)
	}
	source := gopacket.ZeroCopyPacketDataSource(afpacketHandle)
	defer afpacketHandle.Close()

	logSetup(*logdir)
	log.Printf("logging directory: %s", *logdir)

	localAddr := make(map[string]string)
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, ifa := range ifaces {
		if ifa.Name == *iface {
			if ifa.Flags&net.FlagUp == 0 {
				break
			}

			addrs, err := ifa.Addrs()
			if err != nil {
				break
			}
			for _, addr := range addrs {
				var ip net.IP

				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
					continue
				}

				ip4 := ip.To4()
				if ip4 != nil {
					localAddr[ip4.String()] = ip4.String()
					log.Printf("listening on %s %s", ifa.Name, ip4.String())
					continue
				}
				ip6 := ip.To16()
				if ip6 != nil {
					localAddr[ip6.String()] = ip6.String()
					log.Printf("listening on %s %s", ifa.Name, ip6.String())
					continue
				}
			}
		}
	}

	bytes := uint64(0)
	packets := uint64(0)
	drops := uint(0)
	for ; *count != 0; *count-- {
		data, _, err := source.ZeroCopyReadPacketData()
		if err != nil {
			log.Fatal(err)
		}
		bytes += uint64(len(data))
		packets++
		if *count%*verbose == 0 {
			_, afpacketStats, err := afpacketHandle.SocketStats()
			if err != nil {
				log.Println(err)
			}
			if drops != afpacketStats.Drops() {
				log.Printf("Read in %d bytes in %d packets", bytes, packets)
				log.Printf("Stats {received dropped queue-freeze}: %d", afpacketStats)
				drops = afpacketStats.Drops()
			}
		}

		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

		var SaveLog logMessage
		var SrcIP string
		var DstIP string
		var SrcPort string
		var DstPort string
		dns := &layers.DNS{}
		TCPFlag := ""

		if dnsLayer := packet.Layer(layers.LayerTypeDNS); dnsLayer != nil {
			dns, _ = dnsLayer.(*layers.DNS)
			//fmt.Printf("DNS: %s\n", gopacket.LayerDump(dns))

			udpLayer := packet.Layer(layers.LayerTypeUDP)
			if udpLayer != nil {
				udp, _ := udpLayer.(*layers.UDP)
				SrcPort = fmt.Sprintf("%d", udp.SrcPort)
				DstPort = fmt.Sprintf("%d", udp.DstPort)
			} else {
				for _, layer := range packet.Layers() {
					log.Println(layer.LayerType())
				}
			}
		} else if payloadLayer := packet.Layer(gopacket.LayerTypePayload); payloadLayer != nil {
			var decoded []gopacket.LayerType

			data := payloadLayer.LayerContents()
			dnsParser := gopacket.NewDecodingLayerParser(layers.LayerTypeDNS, dns)
			err := dnsParser.DecodeLayers(data[2:], &decoded)
			if err != nil {
				log.Printf("decoding error: %v", err)
				log.Println(gopacket.LayerDump(payloadLayer))
				continue
			}

			TCPFlag = " T"
			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)
				SrcPort = fmt.Sprintf("%d", tcp.SrcPort)
				DstPort = fmt.Sprintf("%d", tcp.DstPort)
			} else {
				for _, layer := range packet.Layers() {
					log.Println(layer.LayerType())
				}
				log.Println(gopacket.LayerDump(payloadLayer))
				continue
			}
		} else {
			continue
		}

		ip4Layer := packet.Layer(layers.LayerTypeIPv4)
		if ip4Layer != nil {
			ip4, _ := ip4Layer.(*layers.IPv4)
			SrcIP = fmt.Sprintf("%s", ip4.SrcIP)
			DstIP = fmt.Sprintf("%s", ip4.DstIP)
		} else {
			ip6Layer := packet.Layer(layers.LayerTypeIPv6)
			ip6, _ := ip6Layer.(*layers.IPv6)
			SrcIP = fmt.Sprintf("%s", ip6.SrcIP)
			DstIP = fmt.Sprintf("%s", ip6.DstIP)
		}

		now := time.Now()
		nanos := now.UnixNano()

		if dns.QR == false { // request
			var result2 string
			var qstring []byte
			if localAddr[DstIP] == "" { // Server Request
				if *logSplit == false {
					result2 = " SQ "
				} else {
					result2 = ""
					SaveLog.lType = "SQ"
				}
			} else { // Client Request
				if *logSplit == false {
					result2 = " CQ "
				} else {
					result2 = ""
					SaveLog.lType = "CQ"
				}
			}
			if len(dns.Questions[0].Name) == 0 {
				qstring = []byte(".")
			} else {
				qstring = dns.Questions[0].Name
			}
			SaveLog.msg = fmt.Sprintf("%s %s#%s %s#%s%s%04X %02d %s %s %s%s\n", time.Unix(0, nanos).Format("15:04:05.000"), SrcIP, SrcPort, DstIP, DstPort, result2, dns.ID, dns.OpCode, qstring, dns.Questions[0].Class, dns.Questions[0].Type, TCPFlag)
		} else { // response
			var result, result2 string
			var qstring []byte
			if dns.ANCount != 0 {
				switch dns.Answers[0].Type {
				case 1: //DNSTypeA:
					result = fmt.Sprintf("%s", dns.Answers[0].IP)
				case 28: //DNSTypeAAAA:
					result = fmt.Sprintf("%s", dns.Answers[0].IP)
				case 16, 13: //DNSTypeTXT, DNSTypeHINFO:
					result = fmt.Sprintf("%s", dns.Answers[0].TXT)
				case 2: //DNSTypeNS:
					result = fmt.Sprintf("%s", dns.Answers[0].NS)
				case 5: //DNSTypeCNAME:
					result = fmt.Sprintf("%s", dns.Answers[0].CNAME)
					for i := 1; i < int(dns.ANCount); i++ {
						if dns.Answers[i].Type == 1 || dns.Answers[i].Type == 28 {
							result += fmt.Sprintf(" (%s %d %s %s %s)",
								dns.Answers[i].Name,
								dns.Answers[i].TTL,
								dns.Answers[i].Class,
								dns.Answers[i].Type,
								dns.Answers[i].IP)
							break
						}
					}
				case 12: //DNSTypePTR:
					result = fmt.Sprintf("%s", dns.Answers[0].PTR)
				case 6: //DNSTypeSOA:
					result = fmt.Sprintf("%s %s (%d %d %d %d %d)",
						dns.Answers[0].SOA.MName,
						dns.Answers[0].SOA.RName,
						dns.Answers[0].SOA.Serial,
						dns.Answers[0].SOA.Refresh,
						dns.Answers[0].SOA.Retry,
						dns.Answers[0].SOA.Expire,
						dns.Answers[0].SOA.Minimum)
				case 15: //DNSTypeMX:
					result = fmt.Sprintf("%d %s", dns.Answers[0].MX.Preference, dns.Answers[0].MX.Name)
				case 33: //DNSTypeSRV:
					result = fmt.Sprintf("%d %d %d %s",
						dns.Answers[0].SRV.Priority,
						dns.Answers[0].SRV.Weight,
						dns.Answers[0].SRV.Port,
						dns.Answers[0].SRV.Name)
				case 41: //DNSTypeOPT:
					result = fmt.Sprintf("%s", dns.Answers[0].OPT)
				}

				if len(dns.Answers[0].Name) == 0 {
					qstring = []byte(".")
				} else {
					qstring = dns.Answers[0].Name
				}
				result = fmt.Sprintf("%s %s %s %d ", qstring, dns.Answers[0].Class, dns.Answers[0].Type, dns.Answers[0].TTL) + result
			} else {
				if len(dns.Questions[0].Name) == 0 {
					qstring = []byte(".")
				} else {
					qstring = dns.Questions[0].Name
				}
				result = fmt.Sprintf("%s %s %s", qstring, dns.Questions[0].Class, dns.Questions[0].Type)
			}

			if localAddr[DstIP] == "" { // Client Response
				if *logSplit == false {
					result2 = " CR "
				} else {
					result2 = ""
					SaveLog.lType = "CR"
				}
			} else { // Server Response
				if *logSplit == false {
					result2 = " SR "
				} else {
					result2 = ""
					SaveLog.lType = "SR"
				}
			}
			SaveLog.msg = fmt.Sprintf("%s %s#%s %s#%s%s%04X %02d %d/%d/%d/%d %s%s\n", time.Unix(0, nanos).Format("15:04:05.000"), SrcIP, SrcPort, DstIP, DstPort, result2, dns.ID, dns.ResponseCode, dns.QDCount, dns.ANCount, dns.NSCount, dns.ARCount, result, TCPFlag)
		}

		logChannel <- SaveLog
	}
}
