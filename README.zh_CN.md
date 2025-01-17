# vuln-notifier

### 切换到 [English Version](#vuln-notifier)

一个简单的 Go 命令行工具，用于监控 Openwall 安全邮件列表，并将新的漏洞帖子通知到 Slack 和 DingTalk。该工具支持关键词过滤，确保只接收相关的安全通知。

## 特性
- 定期轮询 Openwall 安全邮件列表。
- 将通知发送到 Slack 和 DingTalk。
- 根据关键词过滤消息。
- 跟踪已访问的漏洞，避免重复通知。
- 可配置轮询间隔（以分钟为单位）。

## 要求
- Go 1.18 或更高版本。

## 安装

1. 克隆代码库：

   ```bash
   git clone https://github.com/yourusername/vuln-notifier.git
   cd vuln-notifier
   ```

2. 编译 Go 应用程序：
   ```bash
   go build vuln-notifier.go
   ```  
3. 编译完成后，可以直接运行该工具：
   ```bash
   ./vuln-notifier
   ```

## 使用方法

### 命令行选项

| 标志                | 描述                                                                                |
|---------------------|--------------------------------------------------------------------------------------------|
| `-keywords`          | 用逗号分隔的关键词列表，用于过滤消息（例如：apache）                |
| `-slack-webhook`     | 用于通知的 Slack Webhook URL。                                                      |
| `-dingtalk-webhook` | 用于通知的 DingTalk Webhook URL。                                                   |
| `-interval`          | 轮询间隔（以分钟为单位）。默认为 60 分钟。                                       |

### 示例：

过滤与 Apache 相关的漏洞帖子，并每 30 分钟向 Slack 和 DingTalk 发送通知：

```bash
./vuln-notifier -keywords="apache" -slack-webhook="https://hooks.slack.com/services/..." -dingtalk-webhook="https://oapi.dingtalk.com/..." -interval=30
```
这将仅发送与 Apache 相关的漏洞通知，过滤掉与其无关的帖子。