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

// recordResult 根据检测结果将地址记录到对应服务的成功或失败列表中，并写入日志
func recordResult(results map[string]ServiceResult, service, addr string, err error, successLogger, failureLogger interface {
	Printf(string, ...interface{})
}) {
	tmp := results[service]
	if err != nil {
		failureLogger.Printf("[%s] 连接 %s 失败: %v", service, addr, err)
		tmp.Failure = append(tmp.Failure, addr)
	} else {
		successLogger.Printf("[%s] 连接 %s 成功", service, addr)
		tmp.Success = append(tmp.Success, addr)
	}
	results[service] = tmp
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

	// ── Redis（支持密码认证）──────────────────────────────────────────────
	if len(cfg.Redis.Addresses) > 0 && !(len(cfg.Redis.Addresses) == 1 && cfg.Redis.Addresses[0] == "") {
		results["redis"] = ServiceResult{Success: []string{}, Failure: []string{}}
		for _, addr := range cfg.Redis.Addresses {
			err := checker.CheckRedisConnection(addr, cfg.Redis.Password, timeout)
			recordResult(results, "redis", addr, err, successLogger, failureLogger)
		}
	}

	// ── 通用 TCP 服务检测（zookeeper 使用专用检测器）────────────────────
	for service, addresses := range cfg.Services {
		results[service] = ServiceResult{Success: []string{}, Failure: []string{}}

		// 跳过空列表（或只有空字符串的列表）
		if len(addresses) == 0 || (len(addresses) == 1 && addresses[0] == "") {
			continue
		}

		for _, addr := range addresses {
			var connErr error
			if service == "zookeeper" {
				connErr = checker.CheckZookeeperConnection(addr, timeout)
			} else {
				connErr = checker.CheckConnection(addr, timeout)
			}
			recordResult(results, service, addr, connErr, successLogger, failureLogger)
		}
	}

	// ── MinIO 认证连接检测 ────────────────────────────────────────────────
	if cfg.Minio != nil {
		results["minio"] = ServiceResult{Success: []string{}, Failure: []string{}}
		for _, addr := range cfg.Minio.Addresses {
			if addr == "" {
				continue
			}
			err := checker.CheckMinioConnection(addr, cfg.Minio.Username, cfg.Minio.Password, cfg.Minio.UseSSL, timeout)
			recordResult(results, "minio", addr, err, successLogger, failureLogger)
		}
	}

	// ── Kibana（HTTP 基础认证）────────────────────────────────────────────
	if len(cfg.Kibana.Addresses) > 0 {
		results["kibana"] = ServiceResult{Success: []string{}, Failure: []string{}}
		for _, addr := range cfg.Kibana.Addresses {
			err := checker.CheckKibanaConnection(addr, cfg.Kibana.Username, cfg.Kibana.Password, timeout)
			recordResult(results, "kibana", addr, err, successLogger, failureLogger)
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
