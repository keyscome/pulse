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

	// 检测 Elasticsearch 集群（HTTP API + 可选 Basic Auth）
	if cfg.Elasticsearch != nil {
		esResult := ServiceResult{Success: []string{}, Failure: []string{}}
		for _, addr := range cfg.Elasticsearch.Addresses {
			err := checker.CheckElasticsearch(addr, cfg.Elasticsearch.Username, cfg.Elasticsearch.Password, timeout)
			if err != nil {
				failureLogger.Printf("[elasticsearch] 连接 %s 失败: %v", addr, err)
				esResult.Failure = append(esResult.Failure, addr)
			} else {
				successLogger.Printf("[elasticsearch] 连接 %s 成功", addr)
				esResult.Success = append(esResult.Success, addr)
			}
		}
		results["elasticsearch"] = esResult
	}

	// 遍历配置中的每个 TCP 服务类型及其地址列表
	for service, addresses := range cfg.Network {
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
			// 检测 TCP 连接
			err := checker.CheckConnection(addr, timeout)
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
