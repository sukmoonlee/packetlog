// Copyright 2019 smlee@sk.com, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// Package http2log is HTTP Packet logging program
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// PGversion is build version - read README/CHANGE
const PGversion = "1.0.0-20200311"

var localAddr = make(map[string]string)
var filter string
var (
	iface      = flag.String("i", getEnvStr("HTTP2LOG_DEVICE", "eth0"), "Interface to read from")
	cpuprofile = flag.String("cpuprofile", "", "If non-empty, write CPU profile here")
	cpuNo      = flag.Int("cpu", getEnvInt("HTTP2LOG_CPUNO", runtime.NumCPU()), "Number of HTTP parsing goroutine")
	snaplen    = flag.Int("s", getEnvInt("HTTP2LOG_SNAPLEN", 1560), "Snaplen, if <= 0, use 65535")
	bufferSize = flag.Int("b", getEnvInt("HTTP2LOG_BUFSIZE", 8), "Interface buffersize (MB)")
	//filter     = flag.String("f", getEnvStr("HTTP2LOG_FILTER", "port 80 and tcp"), "BPF filter")
	count    = flag.Int64("c", -1, "If >= 0, # of packets to capture before returning")
	verbose  = flag.Int64("log_every", 10000, "Write a every X packets stat")
	logSplit = flag.Bool("log_split", false, "If true, write a split log file (CQ, CR, SQ, SR)")
	addVLAN  = flag.Bool("add_vlan", false, "If true, add VLAN header")
	logdir   = flag.String("d", getEnvStr("HTTP2LOG_DIR", "/log/http2log"), "Write directory for log file")
	httpPort = flag.Int("p", getEnvInt("HTTP2LOG_PORT", 80), "http service port")
	showVer  = flag.Bool("v", false, "If true, show version")
)

func getEnvStr(name string, def string) string {
	content, found := os.LookupEnv(name)
	if found && content != "" {
		return content
	}
	return def
}

func getEnvInt(name string, def int) int {
	content, found := os.LookupEnv(name)
	if found && content != "" {
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
	if runtime.NumCPU() == 1 {
		*cpuNo = 1
	} else {
		*cpuNo = runtime.NumCPU() / 2
	}
	flag.Parse()

	if *showVer {
		fmt.Printf("http2log version: %s\n", PGversion)
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
	filter = fmt.Sprintf("port %d and tcp", *httpPort)
	log.Printf("Starting on http packetlog %s (filter:'%s', SnapLen:%d)", PGversion, filter, *snaplen)

	// Afpacket
	szFrame, szBlock, numBlocks, err := afpacketComputeSize(*bufferSize, *snaplen, os.Getpagesize())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("afpacket pageSize:%d bufferSize:%d szFrame:%d szBlock:%d numBlocks:%d", os.Getpagesize(), *bufferSize, szFrame, szBlock, numBlocks)
	afpacketHandle, err := newAfpacketHandle(*iface, szFrame, szBlock, numBlocks, *addVLAN, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	err = afpacketHandle.SetBPFFilter(filter, *snaplen)
	if err != nil {
		log.Fatal(err)
	}
	source := gopacket.ZeroCopyPacketDataSource(afpacketHandle)
	defer afpacketHandle.Close()

	logSetup(*logdir)
	log.Printf("logging directory: %s", *logdir)
	packetSetup()

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
		data, ci, err := source.ZeroCopyReadPacketData()
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

		var capturePacket packetMessage
		capturePacket.data = make([]byte, *snaplen)
		copy(capturePacket.data, data[:])
		capturePacket.ci = ci

		packetChannel <- capturePacket
	}
}
