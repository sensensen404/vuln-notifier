# vuln-notifier

#### Switch to [中文版](https://github.com/sensensen404/vuln-notifier/blob/main/README.zh_CN.md)

A simple Go command-line tool that monitors the Openwall security mailing list and sends notifications for new vulnerability posts to Slack and DingTalk. The tool supports keyword filtering to receive only the relevant security notifications.

## Features
- Periodic polling of Openwall security mailing list.
- Send notifications to Slack and DingTalk.
- Filter messages by keywords.
- Keep track of visited vulnerabilities to avoid redundant notifications.
- Configurable polling interval (in minutes).

## Requirements
- Go 1.18 or higher.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/vuln-notifier.git
   cd vuln-notifier
   ```
2. Build the Go application:
   ```bash
   go build vuln-notifier.go
   ```  
3. After building, you can run the tool directly:
   ```bash
   ./vuln-notifier
   ```

## Usage

### Command-line Options

| Flag                | Description                                                                                 |
|---------------------|---------------------------------------------------------------------------------------------|
| `-keywords`          | Comma-separated list of keywords to filter messages (e.g., `apache,critical`)                |
| `-slack-webhook`     | Slack webhook URL for notifications.                                                        |
| `-dingtalk-webhook` | DingTalk webhook URL for notifications.                                                     |
| `-interval`          | Polling interval in minutes. Default is 60 minutes.                                         |

### Example:

To filter vulnerability posts related to **Apache** and send notifications every 30 minutes to both Slack and DingTalk:

```bash
./vuln-notifier -keywords="apache" -slack-webhook="https://hooks.slack.com/services/..." -dingtalk-webhook="https://oapi.dingtalk.com/..." -interval=30
```
This will send notifications for vulnerabilities related to Apache only, filtering out posts unrelated to it.

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/sensensen404/vuln-notifier/blob/main/LICENSE) file for details. 