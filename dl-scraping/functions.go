package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

type Item struct {
	Link   string
	ImgSrc string
	Name   string
}

func downloadImage(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetEffectList(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
	)

	c.Visit(os.Getenv("LOGIN_URL"))

	cookies := c.Cookies(os.Getenv("TOP_URL"))
	for _, cookie := range cookies {
		fmt.Printf("Cookie: %s = %s\n", cookie.Name, cookie.Value)
	}

	var items []Item

	c.OnHTML("li.item", func(e *colly.HTMLElement) {
		item := Item{
			Link:   e.ChildAttr("a", "href"),
			ImgSrc: e.ChildAttr("img", "src"),
			Name:   e.ChildText("div.name"),
		}
		items = append(items, item)

		imagePath := fmt.Sprintf("bin/images/%s.jpg", item.Name)
		err := downloadImage(item.ImgSrc, imagePath)
		if err != nil {
			log.Printf("Error downloading image: %v", err)
		} else {
			fmt.Printf("Image saved to %s\n", imagePath)
		}
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Fatalf("Error fetching URL: %v", err)
	})

	c.Visit(os.Getenv("EFFECT_LIST_URL"))

	file, err := os.Create("bin/output.txt")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()
	for _, item := range items {
		fmt.Fprintf(file, "Link: %s\nImage: %s\nName: %s\n\n", item.Link, item.ImgSrc, item.Name)
	}
}
