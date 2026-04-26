#!/bin/bash

# 代理测试脚本
# 使用方法: ./test-proxies.sh [协议] [选项]

set -e

# 默认参数
PROTOCOL="http"
TIMEOUT=10
CONCURRENCY=50
VERBOSE=false
OUTPUT_FILE=""
ADVANCED=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--protocol)
            PROTOCOL="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -c|--concurrency)
            CONCURRENCY="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -a|--advanced)
            ADVANCED=true
            shift
            ;;
        -h|--help)
            echo "使用方法: $0 [选项]"
            echo "选项:"
            echo "  -p, --protocol     代理协议 (http, https, socks4, socks5)"
            echo "  -t, --timeout      超时时间(秒) [默认: 10]"
            echo "  -c, --concurrency  并发数 [默认: 50]"
            echo "  -v, --verbose      详细输出"
            echo "  -o, --output       输出文件路径"
            echo "  -a, --advanced     使用高级测试模式"
            echo "  -h, --help         显示帮助信息"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            exit 1
            ;;
    esac
done

# 检查代理文件是否存在
PROXY_FILE="list/${PROTOCOL}.txt"
if [ ! -f "$PROXY_FILE" ]; then
    echo "错误: 代理文件不存在: $PROXY_FILE"
    echo "可用的协议文件:"
    ls -1 list/*.txt | sed 's/list\///g' | sed 's/\.txt//g'
    exit 1
fi

echo "开始测试代理..."
echo "协议: $PROTOCOL"
echo "代理文件: $PROXY_FILE"
echo "超时时间: ${TIMEOUT}秒"
echo "并发数: $CONCURRENCY"
echo "详细输出: $VERBOSE"
echo "高级模式: $ADVANCED"

# 构建命令
if [ "$ADVANCED" = true ]; then
    # 高级测试模式
    CMD="go run cmd/advanced-proxy-test/main.go"
    if [ "$VERBOSE" = true ]; then
        CMD="$CMD -verbose"
    fi
    if [ -n "$OUTPUT_FILE" ]; then
        CMD="$CMD -output-dir $OUTPUT_FILE"
    fi
    if [ "$TIMEOUT" != "15" ]; then
        CMD="$CMD -timeout $TIMEOUT"
    fi
    if [ "$CONCURRENCY" != "100" ]; then
        CMD="$CMD -concurrency $CONCURRENCY"
    fi
else
    # 简单测试模式
    CMD="go run cmd/proxy-test/main.go"
    CMD="$CMD -protocol $PROTOCOL"
    CMD="$CMD -timeout $TIMEOUT"
    CMD="$CMD -concurrency $CONCURRENCY"
    
    if [ "$VERBOSE" = true ]; then
        CMD="$CMD -verbose"
    fi
    
    if [ -n "$OUTPUT_FILE" ]; then
        CMD="$CMD -output $OUTPUT_FILE"
    fi
fi

echo "执行命令: $CMD"
echo "----------------------------------------"

# 执行测试
eval $CMD

echo "----------------------------------------"
echo "测试完成!"