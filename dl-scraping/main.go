package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/net/html"
)

// 指定したURLからHTMLを取得
func fetchHTML(url string) (*html.Node, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch URL: %s", resp.Status)
    }

    doc, err := html.Parse(resp.Body)
    if err != nil {
        return nil, err
    }

	return doc, nil
}

// テキストを抽出して出力
func extractTextToFile(n *html.Node, file *os.File) {
    fmt.Fprintf(file, "Type: %v, Data: %s\n", n.Type, n.Data)
    for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractTextToFile(c, file)
    }
}


func main() {
    godotenv.Load()
    url := os.Getenv("TEST_URL")
    doc, err := fetchHTML(url)
    if err != nil {
        log.Fatalf("Error fetching HTML: %v", err)
    }

    file, err := os.Create("bin/output.txt")
    if err != nil {
        log.Fatalf("Error creating file: %v", err)
    }
    defer file.Close()

    extractTextToFile(doc, file)
}


