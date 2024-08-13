package main

import (
	"context"
	"log"
	"os"

	"asa-o.net/dl-scraping/functions/functions"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	ctx := context.Background()

	// 関数を登録
	funcframework.RegisterHTTPFunctionContext(ctx, "/", functions.GetEffectList)
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	log.Printf("Serving on port %s", port)
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
