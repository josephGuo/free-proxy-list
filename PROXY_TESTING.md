# 代理可用性批量测试工具

本项目提供了多种工具来批量测试代理地址的可用性，包括简单测试和高级测试两种模式。

## 🚀 快速开始

### 1. 简单测试模式

测试HTTP代理的可用性：

```bash
# Linux/Mac
./test-proxies.sh -p http -v

# Windows
test-proxies.bat -p http -v
```

### 2. 高级测试模式

使用配置文件测试多种协议：

```bash
# Linux/Mac
./test-proxies.sh -a -v

# Windows
test-proxies.bat -a -v
```

## 📁 工具文件说明

### 核心测试工具

- **`cmd/proxy-test/main.go`** - 简单代理测试器
- **`cmd/advanced-proxy-test/main.go`** - 高级代理测试器
- **`cmd/quick-test/main.go`** - 快速代理测试器
- **`test-config.json`** - 高级测试配置文件

### 脚本工具

- **`test-proxies.sh`** - Linux/Mac 测试脚本
- **`test-proxies.bat`** - Windows 测试脚本

## 🔧 使用方法

### 简单测试器 (proxy-tester.go)

适用于快速测试单一协议的代理可用性。

#### 基本用法

```bash
go run cmd/proxy-test/main.go -protocol http -timeout 10 -concurrency 50 -verbose
```

#### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-file` | 代理文件路径 | `list/{protocol}.txt` |
| `-protocol` | 代理协议 | `http` |
| `-timeout` | 超时时间(秒) | `10` |
| `-concurrency` | 并发数 | `50` |
| `-test-url` | 测试URL | `http://httpbin.org/ip` |
| `-output` | 输出文件路径 | 空 |
| `-verbose` | 详细输出 | `false` |

#### 示例

```bash
# 测试HTTP代理，超时5秒，并发100
go run cmd/proxy-test/main.go -protocol http -timeout 5 -concurrency 100 -verbose

# 测试SOCKS5代理并保存结果
go run cmd/proxy-test/main.go -protocol socks5 -output working-socks5.txt -verbose

# 使用自定义代理文件
go run cmd/proxy-test/main.go -file custom-proxies.txt -timeout 15 -verbose
```

### 快速测试器 (quick-test/main.go)

适用于快速验证少量代理的可用性，提供快速的采样测试。

#### 基本用法

```bash
go run cmd/quick-test/main.go -file sample-proxies.txt -sample 10 -timeout 5 -verbose
```

#### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-file` | 代理文件路径 | 必需 |
| `-sample` | 采样数量 | `10` |
| `-timeout` | 超时时间(秒) | `5` |
| `-concurrency` | 并发数 | `3` |
| `-verbose` | 详细输出 | `false` |

#### 示例

```bash
# 快速测试10个代理
go run cmd/quick-test/main.go -file sample-proxies.txt -sample 10 -verbose

# 测试更多代理，增加超时时间
go run cmd/quick-test/main.go -file list/http.txt -sample 50 -timeout 10 -concurrency 5
```

### 高级测试器 (advanced-proxy-test/main.go)

支持多协议、多URL、详细统计和多种输出格式。

#### 基本用法

```bash
go run cmd/advanced-proxy-test/main.go -config test-config.json -output-dir results -verbose
```

#### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-config` | 配置文件路径 | 空(使用默认配置) |
| `-output-dir` | 输出目录 | `test-results` |
| `-timeout` | 超时时间(秒) | `15` |
| `-concurrency` | 并发数 | `100` |
| `-format` | 输出格式 | `json` |
| `-verbose` | 详细输出 | `false` |

#### 配置文件格式

```json
{
  "protocols": [
    {
      "name": "http",
      "filePath": "list/http.txt",
      "enabled": true,
      "options": {}
    },
    {
      "name": "socks5",
      "filePath": "list/socks5.txt",
      "enabled": true,
      "options": {}
    }
  ],
  "testUrls": [
    "http://httpbin.org/ip",
    "http://httpbin.org/get",
    "https://api.ipify.org?format=json"
  ],
  "timeout": 15
}
```

### 脚本工具使用

#### Linux/Mac (test-proxies.sh)

```bash
# 基本用法
./test-proxies.sh -p http -v

# 高级模式
./test-proxies.sh -a -v

# 自定义参数
./test-proxies.sh -p socks5 -t 15 -c 100 -o results.txt -v

# 显示帮助
./test-proxies.sh -h
```

#### Windows (test-proxies.bat)

```cmd
# 基本用法
test-proxies.bat -p http -v

# 高级模式
test-proxies.bat -a -v

# 自定义参数
test-proxies.bat -p socks5 -t 15 -c 100 -o results.txt -v

# 显示帮助
test-proxies.bat -h
```

## 📊 输出结果

### 简单测试器输出

- 控制台显示测试进度和统计信息
- 可选保存可用代理到文件
- 显示最快的10个代理

### 高级测试器输出

生成以下文件：

```
test-results/
├── http_results.json          # HTTP代理详细结果
├── https_results.json         # HTTPS代理详细结果
├── socks4_results.json        # SOCKS4代理详细结果
├── socks5_results.json        # SOCKS5代理详细结果
├── statistics.json            # 统计信息
└── working/                   # 可用代理目录
    ├── http.txt              # 可用HTTP代理
    ├── https.txt             # 可用HTTPS代理
    ├── socks4.txt            # 可用SOCKS4代理
    └── socks5.txt            # 可用SOCKS5代理
```

## 🎯 测试策略

### 1. 连接性测试

- 测试代理是否能成功建立连接
- 检查HTTP响应状态码
- 测量连接延迟

### 2. 功能性测试

- 通过代理访问测试URL
- 验证返回的IP地址
- 检查响应内容完整性

### 3. 性能测试

- 测量连接建立时间
- 记录数据传输速度
- 统计成功率

## ⚡ 性能优化建议

### 1. 并发控制

- 根据网络带宽调整并发数
- 建议起始值：50-100
- 网络较差时降低并发数

### 2. 超时设置

- 快速测试：5-10秒
- 标准测试：10-15秒
- 严格测试：20-30秒

### 3. 测试URL选择

- **快速测试**：`http://httpbin.org/ip`
- **标准测试**：`http://httpbin.org/get`
- **严格测试**：多个URL轮询

## 🔍 故障排除

### 常见问题

1. **代理文件不存在**
   ```
   错误: 代理文件不存在: list/xxx.txt
   解决: 检查文件路径或先运行 go run cmd/main.go 生成代理列表
   ```

2. **网络连接问题**
   ```
   错误: 请求失败: dial tcp: i/o timeout
   解决: 检查网络连接或增加超时时间
   ```

3. **并发过高**
   ```
   错误: too many open files
   解决: 降低并发数或调整系统文件描述符限制
   ```

### 调试技巧

1. **使用详细输出**
   ```bash
   go run cmd/proxy-tester.go -verbose
   ```

2. **降低并发数**
   ```bash
   go run cmd/proxy-tester.go -concurrency 10
   ```

3. **增加超时时间**
   ```bash
   go run cmd/proxy-tester.go -timeout 30
   ```

## 📈 统计指标说明

### 基本指标

- **总数**：测试的代理总数
- **成功数**：测试成功的代理数量
- **失败数**：测试失败的代理数量
- **成功率**：成功代理占总数的百分比

### 性能指标

- **平均延迟**：所有成功代理的平均响应时间
- **最小延迟**：最快代理的响应时间
- **最大延迟**：最慢代理的响应时间
- **测试用时**：完成所有测试的总时间

## 🛠️ 自定义扩展

### 添加新的测试URL

在配置文件中添加更多测试URL：

```json
{
  "testUrls": [
    "http://httpbin.org/ip",
    "http://httpbin.org/get",
    "https://api.ipify.org?format=json",
    "https://httpbin.org/ip",
    "http://icanhazip.com"
  ]
}
```

### 自定义输出格式

修改 `advanced-proxy-tester.go` 中的保存函数来支持更多格式。

### 添加新的代理协议

1. 在 `internal/parser.go` 中添加解析器
2. 在配置文件中启用新协议
3. 运行测试

## 📝 最佳实践

1. **定期测试**：建议每小时测试一次
2. **分层测试**：先快速筛选，再详细验证
3. **结果备份**：保存历史测试结果
4. **性能监控**：关注测试性能变化
5. **协议选择**：根据需求选择合适的代理协议

## 🤝 贡献指南

欢迎提交改进建议和代码贡献：

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 发起 Pull Request

## 📄 许可证

本项目遵循原项目的许可证条款。