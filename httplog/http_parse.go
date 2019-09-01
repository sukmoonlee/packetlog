// Copyright 2019 smlee@sk.com, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// Package httplog is HTTP Packet logging program
package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type packetMessage struct {
	data []byte
	ci   gopacket.CaptureInfo
}

var packetChannel chan packetMessage

func packetSetup() {
	packetChannel = make(chan packetMessage, 10240)

	log.Printf("create %d http parsing goroutine", *cpuNo)
	httpPortStr := string(*httpPort)
	for i := 1; i <= *cpuNo; i++ {
		go func() {
			for pktMsg := range packetChannel {

				packet := gopacket.NewPacket(pktMsg.data, layers.LayerTypeEthernet, gopacket.Default)

				var SaveLog logMessage
				var SrcIP string
				var DstIP string
				var SrcPort string
				var DstPort string
				var result string

				if payloadLayer := packet.Layer(gopacket.LayerTypePayload); payloadLayer != nil {
					//var decoded []gopacket.LayerType

					data := payloadLayer.LayerContents()
					if bytes.Equal(data[0:5], []byte("HTTP/")) {
						headers := strings.Split(string(data), "\n")
						statusLine := strings.TrimSpace(headers[0])
						var contentLength string
						var contentType string

						for i := 1; i < len(headers); i++ {
							//log.Printf("%d %s", i, headers[i])
							if len(headers[i]) > 16 && strings.EqualFold(headers[i][0:16], "Content-Length: ") {
								contentLength = strings.TrimSpace(headers[i][16:])
							} else if len(headers[i]) > 14 && strings.EqualFold(headers[i][0:14], "Content-Type: ") {
								contentType = strings.TrimSpace(headers[i][14:])
							} else if len(headers[i]) < 2 && strings.TrimSpace(headers[i]) == "" {
								break
							}
						}

						result = fmt.Sprintf("%s %s %s", statusLine, contentLength, contentType)
					} else if bytes.Equal(data[0:4], []byte("GET ")) ||
						bytes.Equal(data[0:5], []byte("POST ")) ||
						bytes.Equal(data[0:5], []byte("HEAD ")) ||
						bytes.Equal(data[0:4], []byte("PUT ")) ||
						bytes.Equal(data[0:7], []byte("DELETE ")) ||
						bytes.Equal(data[0:8], []byte("OPTIONS ")) ||
						bytes.Equal(data[0:6], []byte("TRACE ")) ||
						bytes.Equal(data[0:6], []byte("PATCH ")) ||
						bytes.Equal(data[0:8], []byte("CONNECT ")) {

						headers := strings.Split(string(data), "\n")
						requestLine := strings.TrimSpace(headers[0])
						var hostData string
						var userAgent string

						for i := 1; i < len(headers); i++ {
							//log.Printf("%d %s", i, headers[i])
							if len(headers[i]) > 6 && strings.EqualFold(headers[i][0:6], "Host: ") {
								hostData = strings.TrimSpace(headers[i][7:])
							} else if len(headers[i]) > 12 && strings.EqualFold(headers[i][0:12], "User-Agent: ") {
								userAgent = strings.TrimSpace(headers[i][12:])
							} else if len(headers[i]) < 2 && strings.TrimSpace(headers[i]) == "" {
								break
							}
						}

						/*ch := bytes.IndexByte(data, byte('\r'))
						if ch == -1 {
							ch = bytes.IndexByte(data, byte('\n'))
							if ch == -1 {
								continue
							}
						}*/
						result = fmt.Sprintf("%s %s %s", requestLine, hostData, userAgent)
					} else {
						continue
					}
				} else {
					/*for _, layer := range packet.Layers() {
						log.Println(layer.LayerType())
					}*/
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

				tcpLayer := packet.Layer(layers.LayerTypeTCP)
				if tcpLayer != nil {
					tcp, _ := tcpLayer.(*layers.TCP)
					SrcPort = fmt.Sprintf("%d", tcp.SrcPort)
					DstPort = fmt.Sprintf("%d", tcp.DstPort)
				} else {
					continue
				}

				if localAddr[DstIP] == "" {
					if SrcPort == httpPortStr {
						SaveLog.lType = "CR"
					} else {
						SaveLog.lType = "SQ"
					}
				} else {
					if DstPort == httpPortStr {
						SaveLog.lType = "CQ"
					} else {
						SaveLog.lType = "SR"
					}
				}

				SaveLog.msg = fmt.Sprintf("%s %s#%s %s#%s %s %s\n", pktMsg.ci.Timestamp.Format("2006-01-02 15:04:05.000"), SrcIP, SrcPort, DstIP, DstPort, SaveLog.lType, result)
				logChannel <- SaveLog
			}
		}()
	}
}
