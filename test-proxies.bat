@echo off
setlocal enabledelayedexpansion

REM 代理测试脚本 (Windows版本)
REM 使用方法: test-proxies.bat [协议] [选项]

REM 默认参数
set PROTOCOL=http
set TIMEOUT=10
set CONCURRENCY=50
set VERBOSE=false
set OUTPUT_FILE=
set ADVANCED=false

REM 解析命令行参数
:parse_args
if "%1"=="" goto :start_test
if "%1"=="-p" (
    set PROTOCOL=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="--protocol" (
    set PROTOCOL=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="-t" (
    set TIMEOUT=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="--timeout" (
    set TIMEOUT=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="-c" (
    set CONCURRENCY=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="--concurrency" (
    set CONCURRENCY=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="-v" (
    set VERBOSE=true
    shift
    goto :parse_args
)
if "%1"=="--verbose" (
    set VERBOSE=true
    shift
    goto :parse_args
)
if "%1"=="-o" (
    set OUTPUT_FILE=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="--output" (
    set OUTPUT_FILE=%2
    shift
    shift
    goto :parse_args
)
if "%1"=="-a" (
    set ADVANCED=true
    shift
    goto :parse_args
)
if "%1"=="--advanced" (
    set ADVANCED=true
    shift
    goto :parse_args
)
if "%1"=="-h" goto :show_help
if "%1"=="--help" goto :show_help

echo 未知参数: %1
exit /b 1

:show_help
echo 使用方法: %0 [选项]
echo 选项:
echo   -p, --protocol     代理协议 (http, https, socks4, socks5)
echo   -t, --timeout      超时时间(秒) [默认: 10]
echo   -c, --concurrency  并发数 [默认: 50]
echo   -v, --verbose      详细输出
echo   -o, --output       输出文件路径
echo   -a, --advanced     使用高级测试模式
echo   -h, --help         显示帮助信息
exit /b 0

:start_test
REM 检查代理文件是否存在
set PROXY_FILE=list\%PROTOCOL%.txt
if not exist "%PROXY_FILE%" (
    echo 错误: 代理文件不存在: %PROXY_FILE%
    echo 可用的协议文件:
    dir /b list\*.txt
    exit /b 1
)

echo 开始测试代理...
echo 协议: %PROTOCOL%
echo 代理文件: %PROXY_FILE%
echo 超时时间: %TIMEOUT%秒
echo 并发数: %CONCURRENCY%
echo 详细输出: %VERBOSE%
echo 高级模式: %ADVANCED%
echo.

REM 构建命令
if "%ADVANCED%"=="true" (
    REM 高级测试模式
    set CMD=go run cmd/advanced-proxy-test/main.go
    if "%VERBOSE%"=="true" (
        set CMD=!CMD! -verbose
    )
    if not "%OUTPUT_FILE%"=="" (
        set CMD=!CMD! -output-dir %OUTPUT_FILE%
    )
    if not "%TIMEOUT%"=="15" (
        set CMD=!CMD! -timeout %TIMEOUT%
    )
    if not "%CONCURRENCY%"=="100" (
        set CMD=!CMD! -concurrency %CONCURRENCY%
    )
) else (
    REM 简单测试模式
    set CMD=go run cmd/proxy-test/main.go
    set CMD=!CMD! -protocol %PROTOCOL%
    set CMD=!CMD! -timeout %TIMEOUT%
    set CMD=!CMD! -concurrency %CONCURRENCY%
    
    if "%VERBOSE%"=="true" (
        set CMD=!CMD! -verbose
    )
    
    if not "%OUTPUT_FILE%"=="" (
        set CMD=!CMD! -output %OUTPUT_FILE%
    )
)

echo 执行命令: !CMD!
echo ----------------------------------------

REM 执行测试
!CMD!

echo ----------------------------------------
echo 测试完成!
pause