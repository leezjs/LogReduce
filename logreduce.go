package main

import (
	"github.com/ActiveState/tail"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"os"
	"strconv"
	"strings"
	"time"
)

// 全局变量
var mResult map[string]int
var mFiles map[string]*tail.Tail
var log = logs.NewLogger(10000)

// 线程一 统计数据
func DoStatistics(pathTpl, strDate string) {
	arrDate := strings.Split(strDate, "|")
	logfilePathTmp := strings.Replace(pathTpl, "YYYYMMDD", arrDate[0], 1)
	logfilePath := strings.Replace(logfilePathTmp, "YYYYMMDDHH", arrDate[1], -1)

	// 检查文件是否存在，如果不存在等待
	timer := time.Tick(1 * time.Second)
	i := 0
	for _ = range timer {
		// check file exist
		if _, err := os.Stat(logfilePath); os.IsNotExist(err) {
			continue
		} else {
			break
		}

		i++
		// 一个半小时超时
		if i >= 5400 {
			return
		}
	}

	// tail 这个文件
	t, err := tail.TailFile(logfilePath, tail.Config{Follow: true})
	mFiles[arrDate[1]] = t
	if err != nil {
		log.Error("Error: " + err.Error())
	} else {
		for line := range mFiles[arrDate[1]].Lines {
			arrResult := strings.Split(line.Text, "|")
			mResult[arrResult[1]] += 1
		}
	}

}

// 线程，更新时间
// 每小时更新到chan里
func UpdateTime(c chan string) {
	timer := time.Tick(1 * time.Second)
	for _ = range timer {
		t := time.Now()
		sec := t.Second()
		//min := t.Minute()

		//if sec == 0 && min == 29 {
		if sec == 0 {
			toSendTmp := t.Format("20060102")
			toSend := toSendTmp + "|" + toSendTmp + t.Format("15")
			//toSend := toSendTmp + "|" + toSendTmp + t.Format("1504")
			c <- toSend

			log.Debug("send date info to channel " + toSend)
		}
	}
}

// 打印统计信息
func LogStatistics(interval int) {
	timer := time.Tick(time.Duration(interval) * time.Second)
	for _ = range timer {
		log.Debug("statistic info for one interval ...")
		for k, v := range mResult {
			log.Debug("code " + k + ":" + strconv.Itoa(v))
		}
		log.Debug("===================================")
	}
}

// 线程 关闭tail
func CloseTail() {
	timer := time.Tick(time.Duration(5) * time.Second)
	for _ = range timer {
		log.Debug("in close tail...")
		for k, v := range mFiles {
			log.Debug("key is " + k)

			tailTime, _ := time.Parse("2006010215", k)

			// 删除过期tail
			t := time.Now()
			_, offset := t.Zone()
			tsCurrent := t.Unix() + int64(offset)
			// 一个半小时超时
			if tsCurrent-tailTime.Unix() >= 5400 {
				log.Debug("####################################")
				log.Debug("delete tail :" + k)
				v.Stop()
				delete(mFiles, k)
			}
		}
	}
}

// 线程三 读取DB 发出告警
// TODO

func main() {
	// parse config information
	configInfo, _ := config.NewConfig("ini", "./conf/app.conf")
	logfilePath := configInfo.String("logfile_path")
	statisticInterval, _ := strconv.Atoi(configInfo.String("statistic_interval"))

	//初始化全局变量
	mResult = make(map[string]int)
	mFiles = make(map[string]*tail.Tail)
	cUpdateTime := make(chan string)
	log.SetLogger("file", `{"filename":"log/logreduce.log"}`)

	go LogStatistics(statisticInterval)
	go UpdateTime(cUpdateTime)
	go CloseTail()

	// 主循环
	for {
		select {
		case strDate := <-cUpdateTime:
			go DoStatistics(logfilePath, strDate)
		}
	}

}
