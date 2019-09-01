// Copyright 2019 smlee@sk.com, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// packet logging program
package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type logMessage struct {
	lType string
	msg   string
}

var logChannel chan logMessage

func logSetup(logDir string) {
	now := time.Now()
	nanos := now.UnixNano()
	var logFile string
	if *logSplit == true {
		logFile = filepath.Join(logDir, "http-CQ-"+time.Unix(0, nanos).Format("20060102")+".log")
	} else {
		logFile = filepath.Join(logDir, "http-"+time.Unix(0, nanos).Format("20060102")+".log")
	}
	//logFile := fmt.Sprintf("%s/http-%s.log", logDir, time.Unix(0, nanos).Format("20060102"))

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		f, e := os.Create(logFile)
		if e != nil {
			log.Fatal(e)
		}
		f.Close()
	}

	// 로그 채널을 만든다
	logChannel = make(chan logMessage, 10240)

	// 채널을 통한 비동기 로깅
	go func() {
		// 채널이 닫힐 때까지 메시지 받으면 로깅
		last := "24"
		var f [5]*os.File
		if *logSplit == false {
			f[0], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
		} else {
			now := time.Now()
			nanos := now.UnixNano()
			logFile = filepath.Join(logDir, "http-CQ-"+time.Unix(0, nanos).Format("20060102")+".log")
			f[1], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
			logFile = filepath.Join(logDir, "http-CR-"+time.Unix(0, nanos).Format("20060102")+".log")
			f[2], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
			logFile = filepath.Join(logDir, "http-SQ-"+time.Unix(0, nanos).Format("20060102")+".log")
			f[3], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
			logFile = filepath.Join(logDir, "http-SR-"+time.Unix(0, nanos).Format("20060102")+".log")
			f[4], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
		}
		for logMsg := range logChannel {
			if logMsg.msg[0:10] != last {
				now := time.Now()
				nanos := now.UnixNano()

				if *logSplit == false {
					f[0].Close()

					logFile = filepath.Join(logDir, "http-"+time.Unix(0, nanos).Format("20060102")+".log")
					f[0], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
				} else {
					f[1].Close()
					f[2].Close()

					logFile = filepath.Join(logDir, "http-CQ-"+time.Unix(0, nanos).Format("20060102")+".log")
					f[1], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
					logFile = filepath.Join(logDir, "http-CR-"+time.Unix(0, nanos).Format("20060102")+".log")
					f[2], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
					logFile = filepath.Join(logDir, "http-SQ-"+time.Unix(0, nanos).Format("20060102")+".log")
					f[3], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
					logFile = filepath.Join(logDir, "http-SR-"+time.Unix(0, nanos).Format("20060102")+".log")
					f[4], _ = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0600)
				}

				last = logMsg.msg[0:10]
			}

			switch {
			case *logSplit == false:
				f[0].WriteString(logMsg.msg)
			case logMsg.lType == "CQ":
				f[1].WriteString(logMsg.msg)
			case logMsg.lType == "CR":
				f[2].WriteString(logMsg.msg)
			case logMsg.lType == "SQ":
				f[3].WriteString(logMsg.msg)
			case logMsg.lType == "SR":
				f[4].WriteString(logMsg.msg)
			}
		}
		if *logSplit == false {
			f[0].Close()
		} else {
			f[1].Close()
			f[2].Close()
			f[3].Close()
			f[4].Close()
		}
	}()
}
