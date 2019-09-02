# packetlog
대량의 네트워크 패킷을 Text 형식의 파일로 저장하기 위한 프로그램

![Alt text](/docs/intro.png "Packetlog Introduce")

## dnslog ([바로가기](/dnslog/))
* DNS 프로토콜을 분석
* GO(channel, goroutine) + GoPacket + AF_PACKET with zero-copy + Asynchronous Log 기능을 구현
* Client Query(CQ), Client Response(CR), Server Query(SQ), Server Response(SR) 4가지 유형의 로그로 저장
* UDP/TCP 패킷에 대한 로그는 저장하지만, 처리 성능을 고려해서 패킷 Assembly은 지원하지 않음
* [https://github.com/sukmoonlee/packetlog/tree/master/dnslog/](/dnslog/) [README.md](/dnslog/README.md)

## httplog ([바로가기](/httplog/))
* HTTP 프로토콜을 분석
* GO(channel, goroutine) + GoPacket + AF_PACKET with zero-copy + Asynchronous Log 기능을 구현
* Client Query(CQ), Client Response(CR), Server Query(SQ), Server Response(SR) 4가지 유형의 로그로 저장
* TCP 패킷에 대한 로그는 저장하지만, 처리 성능을 고려해서 패킷 Assembly을 하지 않고 로그 기능 처리
* [https://github.com/sukmoonlee/packetlog/tree/master/httplog/](/httplog/) [README.md](/httplog/README.md)
