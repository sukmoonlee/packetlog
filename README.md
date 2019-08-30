# packetlog
대량의 네트워크 패킷을 Text 형식의 파일로 저장하기 위한 프로그램

![Alt text](/docs/intro.png "Packetlog Introduce")

## dnslog
* DNS 프로토콜을 분석하기 위한 프로그램
* GO(channel, goroutine) + GoPacket + AF_PACKET with zero-copy + Asynchronous Log 기능을 구현
* Client Query(CQ), Client Response(CR), Server Query(SQ), Server Response(SR) 4가지 유형의 로그로 저장
* UDP/TCP 패킷에 대한 로그는 저장하지만, 처리 성능을 고려해서 패킷 Assembly은 지원하지 않음

### 사용법
* /etc/sysconfig/dnslog 혹은 dnslog.conf 파일을 통해서 환경 변수 혹은 파라메터를 이용해서 설정 파일을 조정이 가능
* 로그 파일은 -log_split 파라메터를 이용해서 1개의 파일 혹은 4개의 파일로 저장이 가능
* 포함되어 있는 start_dnslog.sh를 이용해서 root 권한으로 실행이 필요함
* 시간 단위 로그 파일 저장
<pre><code>
$ ls -al /tmp/*.log
-rw-r--r-- 1 root root 3657 Aug 26 16:56 /log/dnslog/dns-20190826-16.log

$ cat /tmp/dns-20190826-16.log
16:56:49.465 172.1.214.14#33974 172.1.214.13#53 CQ FCE9 00 sukmoonlee.com IN SOA
16:56:49.465 172.1.214.13#17057 198.97.190.53#53 SQ BBF7 00  IN NS
16:56:49.465 172.1.214.13#9684 198.97.190.53#53 SQ 39AC 00 sukmoonlee.com IN SOA
16:56:50.297 172.1.214.13#47802 199.7.91.13#53 SQ 90B4 00 sukmoonlee.com IN SOA
16:56:50.297 172.1.214.13#26012 199.7.91.13#53 SQ 728B 00  IN NS
16:56:50.425 199.7.91.13#53 172.1.214.13#26012 SR 728B 00 1/14/0/27  IN NS 518400 a.root-servers.net
16:56:50.425 199.7.91.13#53 172.1.214.13#47802 SR 90B4 00 1/0/15/27 sukmoonlee.com IN SOA
16:56:50.425 172.1.214.13#16481 192.42.93.30#53 SQ 3B1F 00 sukmoonlee.com IN SOA
16:56:50.745 192.42.93.30#53 172.1.214.13#16481 SR 3B1F 00 1/0/9/11 sukmoonlee.com IN SOA
16:56:50.745 172.1.214.13#12174 98.124.243.1#53 SQ C951 00 sukmoonlee.com IN SOA
16:56:50.873 98.124.243.1#53 172.1.214.13#12174 SR C951 00 1/1/0/1 sukmoonlee.com IN SOA 3600 dns1.name-services.com info.name-services.com (1553670517 172800 900 1814400 3600)
16:56:50.873 172.1.214.13#53 172.1.214.14#33974 CR FCE9 00 1/1/5/0 sukmoonlee.com IN SOA 3600 dns1.name-services.com info.name-services.com (1553670517 172800 900 1814400 3600)
16:56:50.873 172.1.214.14#37224 172.1.214.13#53 CQ 6B33 00 sukmoonlee.com IN NS
16:56:50.873 172.1.214.13#53306 98.124.243.1#53 SQ FB41 00 sukmoonlee.com IN NS
16:56:51.001 98.124.243.1#53 172.1.214.13#53306 SR FB41 00 1/5/0/11 sukmoonlee.com IN NS 0 dns4.name-services.com
16:56:51.001 172.1.214.13#53 172.1.214.14#37224 CR 6B33 00 1/5/0/0 sukmoonlee.com IN NS 0 dns4.name-services.com
16:56:51.065 172.1.214.14#53193 172.1.214.13#53 CQ 00F5 00 sukmoonlee.com IN MX
16:56:51.065 172.1.214.13#27985 192.54.112.30#53 SQ E99D 00 sukmoonlee.com IN MX
16:56:51.129 192.54.112.30#53 172.1.214.13#27985 SR E99D 00 1/0/9/11 sukmoonlee.com IN MX
16:56:51.129 172.1.214.13#43477 98.124.243.1#53 SQ 228F 00 sukmoonlee.com IN MX
16:56:51.321 98.124.243.1#53 172.1.214.13#43477 SR 228F 00 1/5/0/1 sukmoonlee.com IN MX 1800 30 aspmx2.googlemail.com
16:56:51.321 172.1.214.13#53 172.1.214.14#53193 CR 00F5 00 1/5/5/0 sukmoonlee.com IN MX 1800 10 aspmx.l.google.com
16:56:51.321 172.1.214.14#13039 172.1.214.13#53 CQ 962A 00 www.sukmoonlee.com IN A
16:56:51.321 172.1.214.13#53 172.1.214.14#13039 CR 962A 00 1/1/0/0 www.sukmoonlee.com IN A 600 172.217.161.83
16:56:51.321 172.1.214.14#9762 172.1.214.13#53 CQ A804 00 www.sukmoonlee.com IN AAAA
16:56:51.321 172.1.214.13#27300 98.124.243.1#53 SQ A189 00 www.sukmoonlee.com IN AAAA
16:56:51.449 98.124.243.1#53 172.1.214.13#27300 SR A189 00 1/1/0/1 www.sukmoonlee.com IN CNAME 1800 ghs.google.com
16:56:51.449 172.1.214.13#48780 192.54.112.30#53 SQ 746D 00 ghs.google.com IN AAAA
16:56:51.577 192.54.112.30#53 172.1.214.13#48780 SR 746D 00 1/0/8/9 ghs.google.com IN AAAA
16:56:51.577 172.1.214.13#10751 216.239.34.10#53 SQ 4F62 00 ghs.google.com IN AAAA
16:56:51.641 216.239.34.10#53 172.1.214.13#10751 SR 4F62 00 1/1/0/1 ghs.google.com IN AAAA 300 2404:6800:4004:806::2013
16:56:51.641 172.1.214.13#53 172.1.214.14#9762 CR A804 00 1/2/4/0 www.sukmoonlee.com IN CNAME 1800 ghs.google.com (ghs.google.com 300 IN AAAA 2404:6800:4004:806::2013)
$
</code></pre>
* 쿼리 유형 단위 로그 파일 저장
<pre><code>
$ ls -al /tmp/*.log
-rw------- 1 root root  534 Aug 26 17:00 /log/dnslog/dns-CQ-20190826-17.log
-rw------- 1 root root  987 Aug 26 17:00 /log/dnslog/dns-CR-20190826-17.log
-rw------- 1 root root 1147 Aug 26 17:00 /log/dnslog/dns-SQ-20190826-17.log
-rw------- 1 root root 1608 Aug 26 17:00 /log/dnslog/dns-SR-20190826-17.log

$ cat /tmp/dns-CQ-20190826-17.log
17:00:07.280 172.1.214.14#15842 172.1.214.13#535EDD 00 sukmoonlee.com IN SOA
17:00:07.664 172.1.214.14#52876 172.1.214.13#537B57 00 sukmoonlee.com IN NS
17:00:07.920 172.1.214.14#13002 172.1.214.13#536B54 00 sukmoonlee.com IN MX
17:00:08.112 172.1.214.14#44725 172.1.214.13#53B861 00 www.sukmoonlee.com IN A
17:00:08.112 172.1.214.14#53801 172.1.214.13#531DBF 00 www.sukmoonlee.com IN AAAA
$ cat /tmp/dns-CR-20190826-17.log
17:00:07.664 172.1.214.13#53 172.1.214.14#158425EDD 00 1/1/5/0 sukmoonlee.com IN SOA 3600 dns1.name-services.com info.name-services.com (1553670517 172800 900 1814400 3600)
17:00:07.920 172.1.214.13#53 172.1.214.14#528767B57 00 1/5/0/0 sukmoonlee.com IN NS 0 dns4.name-services.com
17:00:08.112 172.1.214.13#53 172.1.214.14#130026B54 00 1/5/13/0 sukmoonlee.com IN MX 72 20 alt2.aspmx.l.google.com
17:00:08.112 172.1.214.13#53 172.1.214.14#44725B861 00 1/1/0/0 www.sukmoonlee.com IN A 600 172.217.161.83
17:00:08.816 172.1.214.13#53 172.1.214.14#538011DBF 00 1/2/4/0 www.sukmoonlee.com IN CNAME 917 ghs.google.com (ghs.google.com 300 IN AAAA 2404:6800:4004:808::2013)
$ cat /tmp/dns-SQ-20190826-17.log
17:00:07.280 172.1.214.13#38404 192.35.51.30#536742 00 sukmoonlee.com IN SOA
17:00:07.536 172.1.214.13#39950 98.124.243.1#5387BB 00 sukmoonlee.com IN SOA
17:00:07.664 172.1.214.13#47087 64.98.151.1#534E0E 00 sukmoonlee.com IN NS
17:00:07.920 172.1.214.13#38740 64.98.151.1#534D4B 00 sukmoonlee.com IN MX
17:00:08.112 172.1.214.13#20260 192.31.80.30#536A08 00 www.sukmoonlee.com IN AAAA
17:00:08.240 172.1.214.13#30361 64.98.151.1#53A68C 00 www.sukmoonlee.com IN AAAA
17:00:08.496 172.1.214.13#39275 192.35.51.30#539860 00 ghs.google.com IN AAAA
17:00:08.752 172.1.214.13#27022 216.239.34.10#531A72 00 ghs.google.com IN AAAA
$ cat /tmp/dns-SR-20190826-17.log
17:00:07.536 192.35.51.30#53 172.1.214.13#384046742 00 1/0/9/11 sukmoonlee.com IN SOA
17:00:07.664 98.124.243.1#53 172.1.214.13#3995087BB 00 1/1/0/1 sukmoonlee.com IN SOA 3600 dns1.name-services.com info.name-services.com (1553670517 172800 900 1814400 3600)
17:00:07.920 64.98.151.1#53 172.1.214.13#470874E0E 00 1/5/0/11 sukmoonlee.com IN NS 0 dns3.name-services.com
17:00:08.112 64.98.151.1#53 172.1.214.13#387404D4B 00 1/5/0/1 sukmoonlee.com IN MX 72 10 aspmx.l.google.com
17:00:08.240 192.31.80.30#53 172.1.214.13#202606A08 00 1/0/9/11 www.sukmoonlee.com IN AAAA
17:00:08.496 64.98.151.1#53 172.1.214.13#30361A68C 00 1/1/0/1 www.sukmoonlee.com IN CNAME 917 ghs.google.com
17:00:08.752 192.35.51.30#53 172.1.214.13#392759860 00 1/0/8/9 ghs.google.com IN AAAA
17:00:08.816 216.239.34.10#53 172.1.214.13#270221A72 00 1/1/0/1 ghs.google.com IN AAAA 300 2404:6800:4004:808::2013
$
</code></pre>


### Usage

<pre><code>
$ ./dnslog -h
Usage of ./dnslog:
  -add_vlan
        If true, add VLAN header
  -b int
        Interface buffersize (MB) (default 8)
  -c int
        If >= 0, # of packets to capture before returning (default -1)
  -cpu int
        Number of DNS parsing goroutine (default 6)
  -cpuprofile string
        If non-empty, write CPU profile here
  -d string
        Write directory for log file (default "/log/dnslog")
  -f string
        BPF filter (default "port 53")
  -i string
        Interface to read from (default "eth0")
  -log_every int
        Write a every X packets stat (default 100000)
  -log_split
        If true, write a split log file (CQ, CR, SQ, SR)
  -s int
        Snaplen, if <= 0, use 65535 (default 1560)
  -v    If true, show version
</code></pre>

### Build and install

<pre><code>
# yum install golang libpcap libpcap-devel
# git clone https://github.com/sukmoonlee/packetlog.git
# cd dnslog
# go get
# go build -o dnslog   (or make.sh)

# vi start_dnslog.sh
# ./start_dnslog.sh start
</code></pre>
