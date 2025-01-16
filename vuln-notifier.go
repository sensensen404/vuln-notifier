package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
)

const (
	baseURL = "https://www.openwall.com/lists/oss-security/"
)

func main() {
	keywords := flag.String("keywords", "", "Comma-separated list of keywords to filter messages")
	slackWebhook := flag.String("slack-webhook", "", "Slack webhook URL for notifications")
	dingtalkWebhook := flag.String("dingtalk-webhook", "", "DingTalk webhook URL for notifications")
	interval := flag.Int("interval", 60, "Polling interval in minutes")
	flag.Parse()

	keywordList := strings.Split(*keywords, ",")

	for {
		monitor(keywordList, *slackWebhook, *dingtalkWebhook)
		time.Sleep(time.Duration(*interval) * time.Minute)
	}
}

func monitor(keywords []string, slackWebhook string, dingtalkWebhook string) {
	today := time.Now().UTC()
	url := fmt.Sprintf("%s%d/%02d/%02d/", baseURL, today.Year(), int(today.Month()), today.Day())

	items, err := fetch(url)
	if err != nil {
		fmt.Printf("Failed to fetch and parse content: %v\n", err)
		return
	}

	visitedVuln := loadVisitedVuln()
	fileName := generateFileName()

	for _, item := range items {
		href := item[0]
		title := strings.ReplaceAll(item[1], "\n", " ")
		detailUrl := fmt.Sprintf("%s%s", url, href)

		if strings.HasPrefix(title, "Re:") {
			continue
		}

		if _, exists := visitedVuln[title]; exists {
			continue
		}
		if len(keywords) == 0 {
			send(title, detailUrl, slackWebhook, dingtalkWebhook)
			appendToFile(fileName, title, detailUrl)
		} else {
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(title), strings.ToLower(strings.TrimSpace(keyword))) {
					send(title, detailUrl, slackWebhook, dingtalkWebhook)
					appendToFile(fileName, title, detailUrl)
					break
				}
			}
		}

		visitedVuln[title] = true
	}
}

func send(title string, url string, slackWebhook string, dingtalkWebhook string) {
	if slackWebhook != "" {
		message := fmt.Sprintf("%s(%s)", title, url)
		sendToSlack(slackWebhook, message)
	}
	if dingtalkWebhook != "" {
		message := fmt.Sprintf("%s %s", title, url)
		sendToDingTalk(dingtalkWebhook, message)
	}
}

func appendToFile(fileName, content string, url string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file for appending: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(content + "\n" + url + "\n\n"); err != nil {
		fmt.Printf("Failed to append to file: %v\n", err)
	}
}

func fetch(url string) ([][]string, error) {
	content, err := fetchContent(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	items, err := parseItems(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract items: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no list items found")
	}

	return items, nil
}

func fetchContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func parseItems(content string) ([][]string, error) {
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	listItems := htmlquery.Find(doc, "/html/body/ul/li/a")
	var items [][]string
	for _, item := range listItems {
		href := htmlquery.SelectAttr(item, "href")
		title := htmlquery.InnerText(item)
		items = append(items, []string{href, title})
	}

	return items, nil
}

func sendToSlack(webhookURL, message string) {
	payload := fmt.Sprintf(`{"text": "%s"}`, strings.ReplaceAll(message, "\"", "\\\""))
	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(payload))
	if err != nil {
		fmt.Errorf("error posting to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, body)
	}

}

func sendToDingTalk(webhookURL, message string) {
	payload := fmt.Sprintf(`{"msgtype": "text", "text": {"content": "%s"}}`, message)
	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(payload))
	if err != nil {
		fmt.Errorf("error posting to DingTalk: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, body)
	}
}

func loadVisitedVuln() map[string]bool {
	visitedVuln := make(map[string]bool)
	fileName := generateFileName()
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return visitedVuln
		}
		fmt.Printf("Failed to read visited vulnerabilities file: %v\n", err)
		return visitedVuln
	}

	titles := strings.Split(string(data), "\n")
	for _, title := range titles {
		if title != "" {
			visitedVuln[title] = true
		}
	}

	return visitedVuln
}

func generateFileName() string {
	today := time.Now().UTC()
	return fmt.Sprintf("vuln-%d-%02d-%02d.txt", today.Year(), int(today.Month()), today.Day())
}
