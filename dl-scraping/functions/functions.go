package functions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Response struct {
	SessionId string       `json:"sessionId"`
	DlSecKey  string       `json:"dlSecKey"`
	Effects   []EffectInfo `json:"effects"`
	IsNext    bool         `json:"isNext"`
}

type EffectInfo struct {
	Name   string
	Id     string
	HashId string
}

type RequestInfo struct {
	SessionId   string `json:"sessionId"`
	Page        int    `json:"page"`
	MailAddress string `json:"mailAddress"`
	Password    string `json:"password"`
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

func downloadFileFromStorage(ctx context.Context, client *storage.Client, bucketName, objectName, localFilePath string) error {
	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	file, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}

func downloadImageLocal(url, filepath string) error {
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

func downloadImage(ctx context.Context, client *storage.Client, bucketName, url, objectName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	writer := object.NewWriter(ctx)
	defer writer.Close()

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func buildLoginUrl(mailAddress string, password string) string {
	return fmt.Sprintf(os.Getenv("LOGIN_URL"), mailAddress, password)
}

func GetEffectList(w http.ResponseWriter, r *http.Request) {
	// リクエストメソッドのチェック
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JSONデコード
	var request RequestInfo
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	godotenv.Load()

	// 秘密鍵は環境変数にbase64エンコードして格納
	encodedServiceAccountKey := os.Getenv("SERVICE_ACCOUNT_KEY")
	if encodedServiceAccountKey == "" {
		http.Error(w, "Service account key is not set", http.StatusInternalServerError)
		return
	}

	// 秘密鍵のデコード
	serviceAccountKey, err := base64.StdEncoding.DecodeString(encodedServiceAccountKey)
	if err != nil {
		http.Error(w, "Failed to decode service account key", http.StatusInternalServerError)
		return
	}

	// Firestoreクライアントの初期化
	ctx := context.Background()
	sa := option.WithCredentialsJSON(serviceAccountKey)
	client, err := firestore.NewClient(ctx, "asa-o-experiment", sa)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	// Storageクライアントの初期化
	storageClient, err := storage.NewClient(ctx, sa)
	if err != nil {
		log.Fatalf("Failed to create Storage client: %v", err)
	}
	defer storageClient.Close()

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
	)

	var sessionId string
	if request.SessionId == "" {
		c.Visit(buildLoginUrl(request.MailAddress, request.Password))
		cookies := c.Cookies(os.Getenv("TOP_URL"))
		for _, cookie := range cookies {
			fmt.Printf("Cookie: %s = %s\n", cookie.Name, cookie.Value)
			if cookie.Name == "JSESSIONID" {
				sessionId = cookie.Value
			}
		}
	} else {
		sessionId = request.SessionId
		c.OnRequest(func(r *colly.Request) {
			r.Ctx.Put("cookie", "JSESSIONID="+sessionId)
			r.Headers.Set("Cookie", r.Ctx.Get("cookie"))
		})
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
		fmt.Println(info.Name)

		dlSecKeyOnce.Do(func() {
			link := e.ChildAttr("a", "href")
			u, _ := url.Parse(link)
			dlSecKey = u.Query().Get("__DL__SEC__KEY__")
		})

		if false {
			isEnableStorage := true
			if isEnableStorage {
				objectName := fmt.Sprintf("images/%s.jpg", info.Name)
				err := downloadImage(ctx, storageClient, "asa-o-experiment.appspot.com", fmt.Sprintf(imgUrl, info.Id), objectName)
				if err != nil {
					log.Printf("Error downloading image: %v", err)
				} else {
					fmt.Printf("Image saved to Firebase Storage: %s\n", objectName)
				}
			} else {
				imagePath := fmt.Sprintf("bin/images/%s.jpg", info.Name)
				err := downloadImageLocal(fmt.Sprintf(imgUrl, info.Id), imagePath)
				if err != nil {
					log.Printf("Error downloading image: %v", err)
				} else {
					fmt.Printf("Image saved to %s\n", imagePath)
				}
			}
		}
	})
	var pagerNextExists bool
	c.OnHTML("li.pagerNext", func(e *colly.HTMLElement) {
		pagerNextExists = true
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Fatalf("Error fetching URL: %v", err)
	})

	c.Visit(os.Getenv("EFFECT_LIST_URL") + strconv.Itoa(request.Page))

	response := Response{
		SessionId: sessionId,
		DlSecKey:  dlSecKey,
		Effects:   effects,
		IsNext:    pagerNextExists,
	}

	// firestoreへの書き込み
	_, _, err = client.Collection("effects").Add(ctx, map[string]interface{}{
		"seccionId": response.SessionId,
		"dlSecKey":  response.DlSecKey,
		"effects":   response.Effects,
	})
	if err != nil {
		log.Fatalf("Failed adding data to Firestore: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
