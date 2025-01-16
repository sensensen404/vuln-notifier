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
	interval := flag.Int("interval", 30, "Polling interval in minutes")
	flag.Parse()

	if *slackWebhook == "" {
		fmt.Println("Slack webhook URL is required. Provide it using the --slack-webhook flag.")
		os.Exit(1)
	}

	if *keywords == "" {
		fmt.Println("Keywords are required. Provide them using the --keywords flag.")
		os.Exit(1)
	}

	keywordList := strings.Split(*keywords, ",")

	for {
		monitor(keywordList, *slackWebhook)
		time.Sleep(time.Duration(*interval) * time.Minute)
	}
}

func monitor(keywords []string, slackWebhook string) {
	url := generateURL()
	items, err := fetch(url)
	if err != nil {
		fmt.Printf("Failed to fetch and parse content: %v\n", err)
		return
	}

	visitedVuln := loadVisitedVuln()

	for _, item := range items {
		href := item[0]
		title := item[1]

		if _, exists := visitedVuln[title]; exists {
			continue
		}

		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(title), strings.ToLower(strings.TrimSpace(keyword))) {
				message := fmt.Sprintf("Keyword '%s' matched!\nTitle: %s\nURL: %s%s", keyword, title, url, href)
				if err := sendToSlack(slackWebhook, message); err != nil {
					fmt.Printf("Failed to send message to Slack: %v\n", err)
				}
				appendToFile(generateFileName(), title) // 保存匹配到的条目到文件
				break
			}
		}

		visitedVuln[title] = true
	}
}

func appendToFile(fileName, content string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file for appending: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(content + "\n"); err != nil {
		fmt.Printf("Failed to append to file: %v\n", err)
	}
}

func fetch(url string) ([][]string, error) {
	content, err := fetchContent(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	items, err := extractItems(content)
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

func extractItems(content string) ([][]string, error) {
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

func sendToSlack(webhookURL, message string) error {
	payload := fmt.Sprintf(`{"text": "%s"}`, strings.ReplaceAll(message, "\"", "\\\""))
	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("error posting to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, body)
	}

	return nil
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

func saveVisitedVuln(visitedVuln map[string]bool) {
	fileName := generateFileName()
	var titles []string
	for title := range visitedVuln {
		titles = append(titles, title)
	}

	data := strings.Join(titles, "\n")
	err := ioutil.WriteFile(fileName, []byte(data), 0644)
	if err != nil {
		fmt.Printf("Failed to save visited vulnerabilities: %v\n", err)
	}
}

func generateURL() string {
	today := time.Now().UTC()
	return fmt.Sprintf("%s%d/%02d/%02d/", baseURL, today.Year(), int(today.Month()), today.Day())
}

func generateFileName() string {
	today := time.Now().UTC()
	return fmt.Sprintf("vuln-%d-%02d-%02d.txt", today.Year(), int(today.Month()), today.Day())
}
