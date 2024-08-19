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
	funcframework.RegisterHTTPFunctionContext(ctx, "/get-effect-list", functions.GetEffectList)
	funcframework.RegisterHTTPFunctionContext(ctx, "/change-effect", functions.ChangeEffect)
	funcframework.RegisterHTTPFunctionContext(ctx, "/get-effect-image", functions.GetEffectImage)
	port := "8081"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	log.Printf("Serving on port %s", port)
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
