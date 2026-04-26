package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//////////////////// CONFIG ////////////////////

const (
	WorkerCount = 2000
	TCPTimeout  = 1 * time.Second
	HTTPTimeout = 2 * time.Second
	TestURL     = "https://www.baidu.com"
	InputFile   = "proxies.txt"
)

//////////////////// METRICS ////////////////////

var total int64
var alive int64

//////////////////// TASK ////////////////////

type Task struct {
	Addr string
}

//////////////////// TCP CHECK ////////////////////

func tcpCheck(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, TCPTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

//////////////////// HTTP CHECK ////////////////////

func httpCheck(addr string) bool {
	proxyURL, err := url.Parse("http://" + addr)
	if err != nil {
		return false
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout: TCPTimeout,
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   HTTPTimeout,
	}

	req, _ := http.NewRequest("GET", TestURL, nil)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200 || resp.StatusCode == 204
}

//////////////////// WORKER ////////////////////

func worker(jobs <-chan Task, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range jobs {

		atomic.AddInt64(&total, 1)

		// Phase 1: TCP
		if !tcpCheck(task.Addr) {
			continue
		}

		// Phase 2: HTTP
		if httpCheck(task.Addr) {
			atomic.AddInt64(&alive, 1)
			results <- task.Addr
		}
	}
}

//////////////////// MAIN ////////////////////

func main() {

	start := time.Now()
	var inputFile string
	if inputFile == "" {
		// 使用默认路径
		inputFile = filepath.Join("list", "http.txt")
	}
	file, err := os.Open(InputFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	jobs := make(chan Task, 10000)
	results := make(chan string, 1000)

	var wg sync.WaitGroup

	// 启动 worker
	for i := 0; i < WorkerCount; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// 结果收集
	go func() {
		f, _ := os.Create("alive.txt")
		defer f.Close()

		for r := range results {
			fmt.Fprintln(f, r)
		}
	}()

	// 进度打印
	go func() {
		for {
			time.Sleep(2 * time.Second)
			fmt.Printf(
				"Processed: %d | Alive: %d | Speed: %.0f/s\n",
				atomic.LoadInt64(&total),
				atomic.LoadInt64(&alive),
				float64(atomic.LoadInt64(&total))/time.Since(start).Seconds(),
			)
		}
	}()

	// 读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		jobs <- Task{Addr: line}
	}

	close(jobs)
	wg.Wait()
	close(results)

	fmt.Println("Done. Time:", time.Since(start))
}
