package functions

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

type Response struct {
	DlSecKey string       `json:"dlSecKey"`
	Effects  []EffectInfo `json:"effects"`
}

type EffectInfo struct {
	Name   string
	Id     string
	HashId string
}

func init() {
	functions.HTTP("GetEffectList", GetEffectList)
}

func extractHashId(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return ""
	}
	return u.Query().Get("ti")
}

func extractIdFromImgSrc(imgSrc string) string {
	re := regexp.MustCompile(`theme_(\d+)\.jpg`)
	matches := re.FindStringSubmatch(imgSrc)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
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

	var effects []EffectInfo
	imgUrl := os.Getenv("EFFECT_IMAGE_URL")
	var dlSecKey string
	var dlSecKeyOnce sync.Once

	c.OnHTML("li.item", func(e *colly.HTMLElement) {
		info := EffectInfo{
			Name:   e.ChildText("div.name"),
			Id:     extractIdFromImgSrc(e.ChildAttr("img", "src")),
			HashId: extractHashId(e.ChildAttr("a", "href")),
		}
		effects = append(effects, info)

		dlSecKeyOnce.Do(func() {
			link := e.ChildAttr("a", "href")
			u, _ := url.Parse(link)
			dlSecKey = u.Query().Get("__DL__SEC__KEY__")
		})

		imagePath := fmt.Sprintf("bin/images/%s.jpg", info.Name)
		err := downloadImage(fmt.Sprintf(imgUrl, info.Id), imagePath)
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
	for _, info := range effects {
		fmt.Fprintf(file, "HashId: %s\nId: %s\nName: %s\n\n", info.HashId, info.Id, info.Name)
	}

	response := Response{
		DlSecKey: dlSecKey,
		Effects:  effects,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
