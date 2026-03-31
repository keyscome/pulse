// main.go
package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/keyscome/pulse/checker"
	"github.com/keyscome/pulse/config"
	"github.com/keyscome/pulse/logger"
)

// ReportData 用于模板渲染，记录每个服务检测的成功和失败结果
type ReportData struct {
	Timestamp string
	Results   map[string]ServiceResult
}

// ServiceResult 保存某个服务下成功和失败的地址列表
type ServiceResult struct {
	Success []string
	Failure []string
}

func main() {
	// 初始化日志记录器
	successLogger, failureLogger, reportLogger, cleanup, err := logger.NewLoggers()
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// 加载配置文件 config.yml
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		failureLogger.Fatalf("加载配置文件失败: %v", err)
	}

	// 准备存储检测结果，按服务分类
	results := make(map[string]ServiceResult)

	// 设置检测超时时间
	timeout := 3 * time.Second

	// 遍历配置中的每个服务类型及其地址列表
	for service, addresses := range cfg {
		// 初始化结果记录
		results[service] = ServiceResult{
			Success: []string{},
			Failure: []string{},
		}

		// 跳过空列表（或只有空字符串的列表）
		if len(addresses) == 0 || (len(addresses) == 1 && addresses[0] == "") {
			continue
		}

		for _, addr := range addresses {
			// 检测连接：zookeeper 使用专用检测器（ruok/imok 四字命令），其余服务使用 TCP 检测
			var err error
			if service == "zookeeper" {
				err = checker.CheckZookeeperConnection(addr, timeout)
			} else {
				err = checker.CheckConnection(addr, timeout)
			}
			if err != nil {
				failureLogger.Printf("[%s] 连接 %s 失败: %v", service, addr, err)
				tmp := results[service]
				tmp.Failure = append(tmp.Failure, addr)
				results[service] = tmp
			} else {
				successLogger.Printf("[%s] 连接 %s 成功", service, addr)
				tmp := results[service]
				tmp.Success = append(tmp.Success, addr)
				results[service] = tmp
			}
		}
	}

	// 使用 report.tpl 模板生成检测报告
	reportTpl, err := template.ParseFiles("report.tpl")
	if err != nil {
		failureLogger.Fatalf("解析模板文件失败: %v", err)
	}

	reportData := ReportData{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Results:   results,
	}

	// 将报告先渲染到内存缓冲区
	var reportBuffer bytes.Buffer
	err = reportTpl.Execute(&reportBuffer, reportData)
	if err != nil {
		failureLogger.Fatalf("生成报告失败: %v", err)
	}

	// 将检测报告输出到 stdout（同时 reportLogger 也配置了 stdout）
	fmt.Println("\n===== 检测报告 =====")
	fmt.Println(reportBuffer.String())

	// 同时记录报告到日志文件
	reportLogger.Println(reportBuffer.String())
}
