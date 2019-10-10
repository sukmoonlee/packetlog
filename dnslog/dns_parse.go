// Copyright 2019 smlee@sk.com, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// Package dnslog is DNS Packet logging program
package main

import (
	"fmt"
	"log"

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

	log.Printf("create %d dns parsing goroutine", *cpuNo)
	for i := 1; i <= *cpuNo; i++ {
		go func() {
			for pktMsg := range packetChannel {

				packet := gopacket.NewPacket(pktMsg.data, layers.LayerTypeEthernet, gopacket.Default)

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
					SaveLog.msg = fmt.Sprintf("%s %s#%s %s#%s%s%04X %02d %s %s %s%s\n", pktMsg.ci.Timestamp.Format("15:04:05.000"), SrcIP, SrcPort, DstIP, DstPort, result2, dns.ID, dns.OpCode, qstring, dns.Questions[0].Class, dns.Questions[0].Type, TCPFlag)
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
						case 256: //DNSTypeURI:
							result = fmt.Sprintf("%d %d %s",
								dns.Answers[0].URI.Priority,
								dns.Answers[0].URI.Weight,
								dns.Answers[0].URI.Target)
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
					SaveLog.msg = fmt.Sprintf("%s %s#%s %s#%s%s%04X %02d %d/%d/%d/%d %s%s\n", pktMsg.ci.Timestamp.Format("15:04:05.000"), SrcIP, SrcPort, DstIP, DstPort, result2, dns.ID, dns.ResponseCode, dns.QDCount, dns.ANCount, dns.NSCount, dns.ARCount, result, TCPFlag)
				}

				logChannel <- SaveLog
			}
		}()
	}
}
