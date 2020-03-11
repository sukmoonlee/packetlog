// Copyright 2019 smlee@sk.com, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// Package http2log is HTTP Packet logging program
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

type packetMessage struct {
	data []byte
	ci   gopacket.CaptureInfo
}

var packetChannel chan packetMessage

const frameHeaderLen = 9

func packetSetup() {
	packetChannel = make(chan packetMessage, 10240)

	log.Printf("create %d http2 parsing goroutine", *cpuNo)
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
				var loopFlag = 0

				if payloadLayer := packet.Layer(gopacket.LayerTypePayload); payloadLayer != nil {
					data := payloadLayer.LayerContents()
					datalen := len(data)

					startP := 0
					if bytes.Equal(data[0:24], []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")) {
						startP = 24
					}

					for startP < datalen {
						//log.Printf("datalen(%d) startP(%d)", datalen, startP)
						//log.Println(data[startP : startP+frameHeaderLen])

						buf := data[startP : startP+frameHeaderLen]
						fh := http2.FrameHeader{
							Length:   (uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])),
							Type:     http2.FrameType(buf[3]),
							Flags:    http2.Flags(buf[4]),
							StreamID: binary.BigEndian.Uint32(buf[5:]) & (1<<31 - 1),
						}
						// verify frame information
						//log.Printf("frame: length(%d), type(%d/%s), flags(%d), StreamID(%d)", fh.Length, fh.Type, fh.Type, fh.Flags, fh.StreamID)

						if fh.Type == 0x1 {
							//log.Println(data[startP+frameHeaderLen : startP+frameHeaderLen+int(fh.Length)])

							d := hpack.NewDecoder(4096, nil)
							buf := data[startP+frameHeaderLen : startP+frameHeaderLen+int(fh.Length)]
							hf, err := d.DecodeFull(buf)
							if err != nil {
								//log.Printf("%v", err)
								result = fmt.Sprintf("%d DecodeError(%d/%d)-%s", fh.StreamID, err, int(fh.Length), data[startP+frameHeaderLen:startP+frameHeaderLen+int(fh.Length)])
							} else {
								//log.Println(hf)

								var hfPath string
								var hfMethod string
								var hfStatus string
								var hfLength string

								for _, v := range hf {
									//log.Printf("key(%s)=value(%s)", v.Name, v.Value)

									if v.Name == ":status" {
										hfStatus = v.Value
									} else if v.Name == ":path" {
										hfPath = v.Value
									} else if v.Name == ":method" {
										hfMethod = v.Value
									} else if v.Name == "content-length" {
										hfLength = v.Value
									}
								}

								if hfStatus == "" {
									result = fmt.Sprintf("%d %s %s", fh.StreamID, hfMethod, hfPath)
								} else {
									result = fmt.Sprintf("%d %s %s", fh.StreamID, hfStatus, hfLength)
								}
							}
							loopFlag = 1
							break
						}
						startP += int(fh.Length) + frameHeaderLen
					}
					if loopFlag == 0 {
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

					if localAddr[DstIP] == "" {
						if int(tcp.SrcPort) == *httpPort {
							SaveLog.lType = "CR"
						} else {
							SaveLog.lType = "SQ"
						}
					} else {
						if int(tcp.DstPort) == *httpPort {
							SaveLog.lType = "CQ"
						} else {
							SaveLog.lType = "SR"
						}
					}
				} else {
					continue
				}

				SaveLog.msg = fmt.Sprintf("%s %s#%s %s#%s %s %s\n", pktMsg.ci.Timestamp.Format("2006-01-02 15:04:05.000"), SrcIP, SrcPort, DstIP, DstPort, SaveLog.lType, result)
				logChannel <- SaveLog
			}
		}()
	}
}
