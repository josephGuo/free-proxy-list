package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gfpcom/free-proxy-list/internal"
)

var (
	configFile  = flag.String("config", "", "配置文件路径")
	outputDir   = flag.String("output-dir", "test-results", "输出目录")
	timeout     = flag.Int("timeout", 5, "超时时间(秒)")
	concurrency = flag.Int("concurrency", 100, "并发数")
	verbose     = flag.Bool("verbose", false, "详细输出")
	format      = flag.String("format", "json", "输出格式 (json|txt|csv)")
)

// TestConfig 测试配置
type TestConfig struct {
	Protocols []ProtocolConfig `json:"protocols"`
	TestURLs  []string         `json:"testUrls"`
	Timeout   int              `json:"timeout"`
}

// ProtocolConfig 协议配置
type ProtocolConfig struct {
	Name     string                 `json:"name"`
	FilePath string                 `json:"filePath"`
	Enabled  bool                   `json:"enabled"`
	Options  map[string]interface{} `json:"options"`
}

// DetailedTestResult 详细测试结果
type DetailedTestResult struct {
	Proxy        *internal.Proxy `json:"proxy"`
	Success      bool            `json:"success"`
	Latency      time.Duration   `json:"latency"`
	TestTime     time.Time       `json:"testTime"`
	Error        string          `json:"error,omitempty"`
	HTTPStatus   int             `json:"httpStatus,omitempty"`
	ResponseSize int64           `json:"responseSize,omitempty"`
	TestURL      string          `json:"testUrl"`
	IPCheck      string          `json:"ipCheck,omitempty"`
	Country      string          `json:"country,omitempty"`
}

// ProtocolStats 协议统计
type ProtocolStats struct {
	Protocol     string        `json:"protocol"`
	Total        int           `json:"total"`
	Success      int           `json:"success"`
	Failed       int           `json:"failed"`
	SuccessRate  float64       `json:"successRate"`
	AvgLatency   time.Duration `json:"avgLatency"`
	MinLatency   time.Duration `json:"minLatency"`
	MaxLatency   time.Duration `json:"maxLatency"`
	TestDuration time.Duration `json:"testDuration"`
}

// AdvancedProxyTester 高级代理测试器
type AdvancedProxyTester struct {
	config      TestConfig
	timeout     time.Duration
	concurrency int
	verbose     bool
	format      string
	results     map[string][]DetailedTestResult
	stats       map[string]*ProtocolStats
	mu          sync.RWMutex
}

// NewAdvancedProxyTester 创建高级代理测试器
func NewAdvancedProxyTester(config TestConfig, timeout time.Duration, concurrency int, verbose bool, format string) *AdvancedProxyTester {
	return &AdvancedProxyTester{
		config:      config,
		timeout:     timeout,
		concurrency: concurrency,
		verbose:     verbose,
		format:      format,
		results:     make(map[string][]DetailedTestResult),
		stats:       make(map[string]*ProtocolStats),
	}
}

// TestProxyWithURL 使用指定URL测试代理
func (apt *AdvancedProxyTester) TestProxyWithURL(ctx context.Context, proxy *internal.Proxy, testURL string) DetailedTestResult {
	start := time.Now()
	result := DetailedTestResult{
		Proxy:    proxy,
		TestTime: start,
		TestURL:  testURL,
	}

	// 创建代理URL
	proxyURL, err := url.Parse(proxy.String())
	if err != nil {
		result.Error = fmt.Sprintf("解析代理URL失败: %v", err)
		result.Latency = time.Since(start)
		return result
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: apt.timeout,
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		result.Latency = time.Since(start)
		return result
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.HTTPStatus = resp.StatusCode

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Sprintf("状态码错误: %d", resp.StatusCode)
		result.Latency = time.Since(start)
		return result
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		result.Latency = time.Since(start)
		return result
	}

	result.ResponseSize = int64(len(body))
	result.Success = true
	result.Latency = time.Since(start)

	// 如果是IP检查URL，解析IP信息
	if strings.Contains(testURL, "httpbin.org/ip") {
		var ipResp struct {
			Origin string `json:"origin"`
		}
		if err := json.Unmarshal(body, &ipResp); err == nil {
			result.IPCheck = ipResp.Origin
		}
	}

	return result
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (TestConfig, error) {
	var config TestConfig

	// 设置默认配置
	if configPath == "" {
		config = TestConfig{
			Protocols: []ProtocolConfig{
				{Name: "http", FilePath: "list/http.txt", Enabled: true},
				{Name: "https", FilePath: "list/https.txt", Enabled: true},
				{Name: "socks4", FilePath: "list/socks4.txt", Enabled: true},
				{Name: "socks5", FilePath: "list/socks5.txt", Enabled: true},
			},
			TestURLs: []string{
				"http://httpbin.org/ip",
				"http://httpbin.org/get",
				"https://api.ipify.org?format=json",
			},
			Timeout: 15,
		}
		return config, nil
	}

	// 从文件加载配置
	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

// LoadProxies 从文件加载代理列表
func LoadProxies(filePath string) ([]*internal.Proxy, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var proxies []*internal.Proxy
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析代理URL
		proxy, err := internal.ParseProxyURL("", line)
		if err != nil {
			if *verbose {
				log.Printf("⚠️  跳过无效代理: %s - %v", line, err)
			}
			continue
		}

		proxies = append(proxies, proxy)
	}

	return proxies, scanner.Err()
}

// SaveResults 保存测试结果
func (apt *AdvancedProxyTester) SaveResults(outputDir string) error {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// 保存详细结果
	for protocol, results := range apt.results {
		filename := filepath.Join(outputDir, protocol+"_results."+apt.format)
		switch apt.format {
		case "json":
			apt.saveResultsAsJSON(results, filename)
		case "csv":
			apt.saveResultsAsCSV(results, filename)
		case "txt":
			apt.saveResultsAsTXT(results, filename)
		}
	}

	// 保存统计信息
	statsFile := filepath.Join(outputDir, "statistics."+apt.format)
	switch apt.format {
	case "json":
		apt.saveStatsAsJSON(statsFile)
	case "csv":
		apt.saveStatsAsCSV(statsFile)
	case "txt":
		apt.saveStatsAsTXT(statsFile)
	}

	// 保存可用代理列表
	apt.saveWorkingProxies(outputDir)

	return nil
}

// saveResultsAsJSON 保存为JSON格式
func (apt *AdvancedProxyTester) saveResultsAsJSON(results []DetailedTestResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// saveResultsAsCSV 保存为CSV格式
func (apt *AdvancedProxyTester) saveResultsAsCSV(results []DetailedTestResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入CSV头
	file.WriteString("Proxy,Success,Latency(ms),TestTime,Error,HTTPStatus,ResponseSize,TestURL,IPCheck\n")

	// 写入数据
	for _, result := range results {
		line := fmt.Sprintf("%s,%t,%d,%s,%s,%d,%d,%s,%s\n",
			result.Proxy.String(),
			result.Success,
			result.Latency.Milliseconds(),
			result.TestTime.Format("2006-01-02 15:04:05"),
			result.Error,
			result.HTTPStatus,
			result.ResponseSize,
			result.TestURL,
			result.IPCheck)
		file.WriteString(line)
	}

	return nil
}

// saveResultsAsTXT 保存为TXT格式
func (apt *AdvancedProxyTester) saveResultsAsTXT(results []DetailedTestResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("# 代理测试结果\n")
	file.WriteString("# 格式: 代理URL 延迟(ms) 测试时间 状态 IP\n")
	file.WriteString("# 生成时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")

	for _, result := range results {
		if result.Success {
			file.WriteString(fmt.Sprintf("%s %d %s ✅ %s\n",
				result.Proxy.String(),
				result.Latency.Milliseconds(),
				result.TestTime.Format("15:04:05"),
				result.IPCheck))
		}
	}

	return nil
}

// saveStatsAsJSON 保存统计信息为JSON
func (apt *AdvancedProxyTester) saveStatsAsJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(apt.stats)
}

// saveStatsAsCSV 保存统计信息为CSV
func (apt *AdvancedProxyTester) saveStatsAsCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("Protocol,Total,Success,Failed,SuccessRate,AvgLatency(ms),MinLatency(ms),MaxLatency(ms),TestDuration(ms)\n")

	for _, stats := range apt.stats {
		line := fmt.Sprintf("%s,%d,%d,%d,%.2f,%d,%d,%d,%d\n",
			stats.Protocol,
			stats.Total,
			stats.Success,
			stats.Failed,
			stats.SuccessRate,
			stats.AvgLatency.Milliseconds(),
			stats.MinLatency.Milliseconds(),
			stats.MaxLatency.Milliseconds(),
			stats.TestDuration.Milliseconds())
		file.WriteString(line)
	}

	return nil
}

// saveStatsAsTXT 保存统计信息为TXT
func (apt *AdvancedProxyTester) saveStatsAsTXT(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("# 代理测试统计信息\n")
	file.WriteString("# 生成时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")

	for _, stats := range apt.stats {
		file.WriteString(fmt.Sprintf("协议: %s\n", stats.Protocol))
		file.WriteString(fmt.Sprintf("总数: %d\n", stats.Total))
		file.WriteString(fmt.Sprintf("成功: %d\n", stats.Success))
		file.WriteString(fmt.Sprintf("失败: %d\n", stats.Failed))
		file.WriteString(fmt.Sprintf("成功率: %.2f%%\n", stats.SuccessRate))
		file.WriteString(fmt.Sprintf("平均延迟: %v\n", stats.AvgLatency))
		file.WriteString(fmt.Sprintf("最小延迟: %v\n", stats.MinLatency))
		file.WriteString(fmt.Sprintf("最大延迟: %v\n", stats.MaxLatency))
		file.WriteString(fmt.Sprintf("测试用时: %v\n", stats.TestDuration))
		file.WriteString(strings.Repeat("-", 50) + "\n")
	}

	return nil
}

// saveWorkingProxies 保存可用代理列表
func (apt *AdvancedProxyTester) saveWorkingProxies(outputDir string) error {
	workingDir := filepath.Join(outputDir, "working")
	if err := os.MkdirAll(workingDir, 0755); err != nil {
		return err
	}

	for protocol, results := range apt.results {
		var workingProxies []string
		for _, result := range results {
			if result.Success {
				workingProxies = append(workingProxies, result.Proxy.String())
			}
		}

		if len(workingProxies) > 0 {
			filename := filepath.Join(workingDir, protocol+".txt")
			file, err := os.Create(filename)
			if err != nil {
				continue
			}
			defer file.Close()

			for _, proxy := range workingProxies {
				file.WriteString(proxy + "\n")
			}

			log.Printf("保存 %d 个可用 %s 代理到: %s", len(workingProxies), protocol, filename)
		}
	}

	return nil
}

// RunTests 运行所有测试
func (apt *AdvancedProxyTester) RunTests() error {
	ctx := context.Background()
	totalStartTime := time.Now()

	for _, protocolConfig := range apt.config.Protocols {
		if !protocolConfig.Enabled {
			continue
		}

		log.Printf("开始测试协议: %s", protocolConfig.Name)
		protocolStartTime := time.Now()

		// 加载代理
		proxies, err := LoadProxies(protocolConfig.FilePath)
		if err != nil {
			log.Printf("加载 %s 代理失败: %v", protocolConfig.Name, err)
			continue
		}

		if len(proxies) == 0 {
			log.Printf("没有找到 %s 代理", protocolConfig.Name)
			continue
		}

		log.Printf("加载了 %d 个 %s 代理", len(proxies), protocolConfig.Name)

		// 初始化统计
		stats := &ProtocolStats{
			Protocol: protocolConfig.Name,
			Total:    len(proxies),
		}
		apt.stats[protocolConfig.Name] = stats

		// 创建工作池
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, apt.concurrency)
		resultsChan := make(chan DetailedTestResult, apt.concurrency)

		// 启动结果收集
		var results []DetailedTestResult
		var resultsMu sync.Mutex
		go func() {
			for result := range resultsChan {
				resultsMu.Lock()
				results = append(results, result)
				resultsMu.Unlock()
			}
		}()

		// 测试每个代理
		for i, proxy := range proxies {
			wg.Add(1)
			go func(idx int, p *internal.Proxy) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// 测试所有URL
				for _, testURL := range apt.config.TestURLs {
					result := apt.TestProxyWithURL(ctx, p, testURL)
					resultsChan <- result

					// 更新统计
					apt.mu.Lock()
					if result.Success {
						stats.Success++
						if stats.MinLatency == 0 || result.Latency < stats.MinLatency {
							stats.MinLatency = result.Latency
						}
						if result.Latency > stats.MaxLatency {
							stats.MaxLatency = result.Latency
						}
					} else {
						stats.Failed++
					}
					apt.mu.Unlock()

					// 如果一个URL测试成功，就跳过其他URL
					if result.Success {
						break
					}
				}

				// 显示进度
				if (idx+1)%50 == 0 || idx == len(proxies)-1 {
					log.Printf("%s 进度: %d/%d (%.1f%%)",
						protocolConfig.Name,
						idx+1,
						len(proxies),
						float64(idx+1)/float64(len(proxies))*100)
				}
			}(i, proxy)
		}

		wg.Wait()
		close(resultsChan)

		// 计算最终统计
		stats.TestDuration = time.Since(protocolStartTime)
		stats.SuccessRate = float64(stats.Success) / float64(stats.Total) * 100
		if stats.Success > 0 {
			// 重新计算平均延迟（只计算成功的）
			var totalLatency time.Duration
			successCount := 0
			for _, result := range results {
				if result.Success {
					totalLatency += result.Latency
					successCount++
				}
			}
			if successCount > 0 {
				stats.AvgLatency = totalLatency / time.Duration(successCount)
			}
		}

		// 保存结果
		apt.results[protocolConfig.Name] = results

		log.Printf("%s 测试完成: 成功 %d/%d (%.2f%%), 用时 %v",
			protocolConfig.Name,
			stats.Success,
			stats.Total,
			stats.SuccessRate,
			stats.TestDuration)
	}

	// 保存所有结果
	if err := apt.SaveResults(*outputDir); err != nil {
		return fmt.Errorf("保存结果失败: %v", err)
	}

	// 显示总体统计
	apt.printSummary(time.Since(totalStartTime))

	return nil
}

// printSummary 打印总结
func (apt *AdvancedProxyTester) printSummary(totalDuration time.Duration) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("所有测试完成! 总用时: %v\n", totalDuration)
	fmt.Println(strings.Repeat("=", 80))

	var totalProxies, totalSuccess, totalFailed int
	for _, stats := range apt.stats {
		totalProxies += stats.Total
		totalSuccess += stats.Success
		totalFailed += stats.Failed
	}

	fmt.Printf("总体统计:\n")
	fmt.Printf("  总代理数: %d\n", totalProxies)
	fmt.Printf("  成功代理数: %d\n", totalSuccess)
	fmt.Printf("  失败代理数: %d\n", totalFailed)
	fmt.Printf("  总体成功率: %.2f%%\n", float64(totalSuccess)/float64(totalProxies)*100)
	fmt.Println()

	fmt.Println("各协议详细统计:")
	fmt.Printf("%-10s %-8s %-8s %-8s %-10s %-12s %-12s %-12s\n",
		"协议", "总数", "成功", "失败", "成功率", "平均延迟", "最小延迟", "最大延迟")
	fmt.Println(strings.Repeat("-", 80))

	for _, stats := range apt.stats {
		fmt.Printf("%-10s %-8d %-8d %-8d %-9.2f%% %-11v %-11v %-11v\n",
			stats.Protocol,
			stats.Total,
			stats.Success,
			stats.Failed,
			stats.SuccessRate,
			stats.AvgLatency,
			stats.MinLatency,
			stats.MaxLatency)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("结果已保存到: %s\n", *outputDir)
}

func main() {
	flag.Parse()

	// 加载配置
	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建测试器
	tester := NewAdvancedProxyTester(
		config,
		time.Duration(*timeout)*time.Second,
		*concurrency,
		*verbose,
		*format,
	)

	// 运行测试
	if err := tester.RunTests(); err != nil {
		log.Fatalf("运行测试失败: %v", err)
	}
}
