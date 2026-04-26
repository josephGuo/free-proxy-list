package main

import (
	"context"
	"net/url"
	"sync/atomic"
	"time"
)

func (l *ProxyTester) Allow() bool {
	return atomic.LoadInt32(&l.current) > 0
}

func (l *ProxyTester) Acquire() {
	atomic.AddInt32(&l.current, -1)
}

func (l *ProxyTester) Release() {
	atomic.AddInt32(&l.current, 1)
}

func (l *ProxyTester) Report(success bool) {
	if success {
		atomic.AddInt32(&l.successCount, 1)
	} else {
		atomic.AddInt32(&l.failCount, 1)
	}
}

func (l *ProxyTester) Adjust() {
	s := atomic.LoadInt32(&l.successCount)
	f := atomic.LoadInt32(&l.failCount)

	total := s + f
	if total < 100 {
		return
	}

	rate := float64(s) / float64(total)

	cur := atomic.LoadInt32(&l.current)

	if rate > 0.3 && cur < int32(l.max) {
		atomic.AddInt32(&l.current, 100) // 扩容
	} else if rate < 0.1 && cur > int32(l.min) {
		atomic.AddInt32(&l.current, -100) // 收缩
	}

	atomic.StoreInt32(&l.successCount, 0)
	atomic.StoreInt32(&l.failCount, 0)
}

func startAutoAdjust(limiter *ProxyTester) {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			limiter.Adjust()
		}
	}()
}

func startWorkers(ctx context.Context, pt *ProxyTester, proxyCh <-chan *url.URL, results chan<- *ProxyTestResult) {

	for proxy := range proxyCh {

		// 等待可用并发
		for !pt.Allow() {
			time.Sleep(10 * time.Millisecond)
		}

		pt.Acquire()

		go func(p *url.URL) {
			defer pt.Release()
			client := createSharedClient(3 * time.Second)
			res := pt.TestProxy(ctx, p, client)
			results <- &res
			pt.Report(res.Success)

		}(proxy)
	}
}
