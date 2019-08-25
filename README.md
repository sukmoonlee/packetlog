# packetlog
네트워크 특정 패킷을 파일로 저장해서 분석하기 위한 프로그램

## dnslog
* DNS 프로토콜을 분석하기 위한 프로그램
* GO + GoPacket + afpacket with zero-copy + Asynchronous Log 기능을 구현
* Client Query(CQ), Client Response(CR), Server Query(SQ), Server Response(SR) 4가지 유형의 로그로 저장
* UDP/TCP 패킷에 대한 로그는 저장하지만, 처리 성능을 고려해서 패킷 Assembly은 지원하지 않음

### 사용법
* /etc/sysconfig/dnslog 혹은 dnslog.conf 파일을 통해서 환경 변수 혹은 파라메터를 이용해서 설정 파일을 조정이 가능
* 로그 파일은 -log_split 파라메터를 이용해서 1개의 파일 혹은 4개의 파일로 저장이 가능
* 포함되어 있는 start_dnslog.sh를 이용해서 root 권한으로 실행이 필요함

### Usage
    
    $ ./dnslog -h
    Usage of ./dnslog:
      -add_vlan
            If true, add VLAN header
      -b int
            Interface buffersize (MB) (default 8)
      -c int
            If >= 0, # of packets to capture before returning (default -1)
      -cpuprofile string
            If non-empty, write CPU profile here
      -d string
            Write directory for log file (default "/log/dnslog")
      -f string
            BPF filter (default "port 53")
      -i string
            Interface to read from (default "bond1")
      -log_every int
            Write a every X packets stat (default 100000)
      -log_split
            If true, write a split log file (CQ, CR, SQ, SR)
      -s int
            Snaplen, if <= 0, use 65535 (default 1560)
      -v    If true, show version


### Build and install
