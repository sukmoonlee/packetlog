# packetlog
대량의 네트워크 패킷을 Text 형식의 파일로 저장하기 위한 프로그램

![Alt text](/docs/intro.png "Packetlog Introduce")

## httplog
* HTTP 프로토콜을 분석하기 위한 프로그램
* GO(channel, goroutine) + GoPacket + AF_PACKET with zero-copy + Asynchronous Log 기능을 구현
* Client Query(CQ), Client Response(CR), Server Query(SQ), Server Response(SR) 4가지 유형의 로그로 저장
* TCP 패킷에 대한 로그는 저장하지만, 처리 성능을 고려해서 패킷 Assembly을 하지 않고 로그 기능 처리

### 사용법
* /etc/sysconfig/httplog 혹은 httplog.conf 파일을 통해서 환경 변수 혹은 파라메터를 이용해서 설정 파일을 조정이 가능
* 로그 파일은 -log_split 파라메터를 이용해서 1개의 파일 혹은 4개의 파일로 저장이 가능
* 포함되어 있는 start_httplog.sh를 이용해서 root 권한으로 실행이 필요함
* 날짜 단위 로그 파일 저장
<pre><code>
$ ls -al /tmp/*.log
-rw-r--r-- 1 root root 84003 Sep  1 14:35 /tmp/http-20190901.log

$ cat /tmp/http-20190901.log
2019-09-01 14:16:51.002 223.39.188.71#50905 10.140.0.2#80 CQ GET / HTTP/1.1 cloud.sukmoonlee.com Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36
2019-09-01 14:16:51.003 10.140.0.2#80 223.39.188.71#50905 CR HTTP/1.1 302 Found 0 text/html; charset=UTF-8
2019-09-01 14:17:00.573 10.140.0.2#57154 169.254.169.254#80 SQ GET /computeMetadata/v1/instance/virtual-clock/drift-token?timeout_sec=60&last_etag=c6a7bbfe995acc98&alt=json&recursive=False&wait_for_change=True HTTP/1.1 etadata.google.internal Python-urllib/2.7
2019-09-01 14:17:54.337 169.254.169.254#80 10.140.0.2#57152 SR HTTP/1.1 200 OK 3274 application/json
$
</code></pre>

### Usage

<pre><code>
$ ./httplog -h
Usage of ./httplog:
  -add_vlan
        If true, add VLAN header
  -b int
        Interface buffersize (MB) (default 8)
  -c int
        If >= 0, # of packets to capture before returning (default -1)
  -cpu int
        Number of HTTP parsing goroutine (default 1)
  -cpuprofile string
        If non-empty, write CPU profile here
  -d string
        Write directory for log file (default "/log/httplog")
  -i string
        Interface to read from (default "eth0")
  -log_every int
        Write a every X packets stat (default 10000)
  -log_split
        If true, write a split log file (CQ, CR, SQ, SR)
  -p int
        http service port (default 80)
  -s int
        Snaplen, if <= 0, use 65535 (default 1560)
  -v    If true, show version
</code></pre>

### Build and install

<pre><code>
# yum install golang libpcap libpcap-devel
# git clone https://github.com/sukmoonlee/packetlog.git
# cd httplog
# go get
# go build -o httplog   (or make.sh)

# vi start_httplog.sh
# ./start_httplog.sh start
</code></pre>
