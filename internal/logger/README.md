# OSS Manager Logger

这个日志系统为OSS Manager应用程序提供了强大而灵活的日志记录功能。

## 主要特点

- 支持多个日志级别（DEBUG, INFO, WARN, ERROR）
- 支持多种日志格式（TEXT, JSON）
- 日志可同时输出到文件和控制台
- 自动日志轮转功能，防止日志文件过大
- 支持上下文相关的日志记录
- 线程安全
- 可在运行时动态修改日志级别

## 基本用法

### 初始化日志系统

```go
// 使用默认配置初始化
if err := logger.InitializeWithDefaults(); err != nil {
    fmt.Printf("日志初始化失败: %v\n", err)
    os.Exit(1)
}
defer logger.Shutdown() // 确保日志正确关闭

// 或者使用自定义配置
config := &logger.LogConfig{
    Level:       logger.DEBUG,        // 日志级别
    Format:      logger.TEXT,         // 日志格式
    LogDir:      "logs",              // 日志目录
    LogFile:     "myapp.log",         // 日志文件名
    MaxSize:     10,                  // 单个日志文件最大尺寸(MB)
    MaxBackups:  5,                   // 保留的旧日志文件数量
    MaxAge:      30,                  // 保留日志的最大天数
    Compress:    true,                // 是否压缩旧日志
    ToConsole:   true,                // 是否同时输出到控制台
    TimestampFormat: time.RFC3339,    // 时间戳格式
}

if err := logger.Initialize(config); err != nil {
    fmt.Printf("日志初始化失败: %v\n", err)
    os.Exit(1)
}
```

### 记录不同级别的日志

```go
// 基本日志记录
logger.Debug("调试信息: %s", "详细的调试数据")
logger.Info("信息: %s", "普通的信息")
logger.Warn("警告: %s", "警告消息")
logger.Error("错误: %s", "错误消息")

// 条件日志
if somethingImportant {
    logger.Info("重要事件发生了")
}

// 复杂对象日志
user := User{ID: 123, Name: "张三"}
logger.Info("用户操作: %+v", user)
```

### 使用上下文日志

```go
// 创建带上下文的日志记录器
requestLogger := logger.WithContext(map[string]interface{}{
    "request_id": "req-123456",
    "user_id": 1001,
    "client_ip": "192.168.1.100",
})

// 使用上下文记录器记录日志
requestLogger.Info("接收到请求")
requestLogger.Info("处理请求")
requestLogger.Info("请求完成") 

// 可以添加更多上下文信息
operationLogger := requestLogger.WithContext(map[string]interface{}{
    "operation": "upload_file",
    "file_size": 1024 * 1024,
})

operationLogger.Info("开始上传文件")
```

### 动态调整日志级别

```go
// 获取当前日志级别
currentLevel := logger.GetLevel()
fmt.Printf("当前日志级别: %s\n", currentLevel)

// 临时调整日志级别为DEBUG
logger.SetLevel(logger.DEBUG)

// 执行需要更详细日志的操作...

// 恢复到之前的日志级别
logger.SetLevel(currentLevel)
```

## 最佳实践

1. **使用适当的日志级别**:
   - DEBUG: 详细的调试信息，仅在开发和问题排查时启用
   - INFO: 常规操作信息，表示程序正常运行
   - WARN: 潜在问题的警告，不影响正常运行但需要注意
   - ERROR: 错误信息，表示操作失败或异常

2. **使用上下文日志追踪请求**:
   为每个请求或操作创建带有唯一标识符的上下文日志器，便于问题追踪和性能分析。

3. **敏感信息处理**:
   不要在日志中记录密码、访问密钥等敏感信息。如果必须包含部分信息，可以进行脱敏处理。

4. **结构化日志**:
   在生产环境中使用JSON格式的日志，便于日志聚合和分析工具处理。

5. **异常情况下的详细日志**:
   错误发生时，记录尽可能多的上下文信息，包括错误类型、错误信息、堆栈跟踪等。

## 性能考虑

日志系统对性能的影响主要来自以下几点：

1. **磁盘I/O**:
   频繁的日志写入会增加磁盘I/O，可以通过调整日志级别减少写入量。

2. **字符串格式化**:
   复杂的日志格式化会消耗CPU资源，特别是当日志不会被记录时（如在INFO级别下的DEBUG日志）。
   可以使用条件判断减少不必要的字符串格式化:

   ```go
   // 不好的方式 - 即使不记录也会进行字符串格式化
   logger.Debug("复杂计算结果: %v", calculateExpensiveResult())

   // 好的方式 - 仅在需要时进行计算和格式化
   if logger.GetLevel() == logger.DEBUG {
       logger.Debug("复杂计算结果: %v", calculateExpensiveResult())
   }
   ```

3. **并发性能**:
   日志系统使用互斥锁保证线程安全，在高并发场景下可能成为瓶颈。

## 常见问题

1. **日志文件权限问题**:
   确保应用程序有权限读写日志目录。初始化时会尝试创建日志目录。

2. **日志轮转不工作**:
   检查日志配置的MaxSize、MaxBackups和MaxAge设置是否合理。

3. **日志没有输出**:
   - 检查日志级别设置
   - 确认日志配置正确初始化
   - 验证日志目录是否可写

4. **JSON日志格式问题**:
   使用JSON格式时，确保日志消息不包含可能破坏JSON结构的非转义特殊字符。

## 扩展与定制

日志系统通过接口设计支持扩展。如果需要将日志输出到其他目标（如ELK、Prometheus等），可以创建自定义的Logger实现。 