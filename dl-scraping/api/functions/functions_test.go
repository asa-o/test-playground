package functions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
)

func Test_downloadImage(t *testing.T) {
	type args struct {
		ctx        context.Context
		client     *storage.Client
		bucketName string
		url        string
		objectName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := downloadImage(tt.args.ctx, tt.args.client, tt.args.bucketName, tt.args.url, tt.args.objectName); (err != nil) != tt.wantErr {
				t.Errorf("downloadImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEffectList(t *testing.T) {
	godotenv.Load()
	mail := os.Getenv("TEST_MAIL_ADDRESS")
	pass := os.Getenv("TEST_PASSWORD")

	bodyJSON := `{"sessionId":"","page":1,"mailAddress":"` + mail + `","password":"` + pass + `"}`
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(bodyJSON))

	GetEffectList(response, req)
	if response.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", response.Code)
	}

	var res Response
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return
	}
}

func TestGetEffectList_2page(t *testing.T) {
	godotenv.Load()
	mail := os.Getenv("TEST_MAIL_ADDRESS")
	pass := os.Getenv("TEST_PASSWORD")

	bodyJSON := `{"sessionId":"","page":1,"mailAddress":"` + mail + `","password":"` + pass + `"}`
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(bodyJSON))

	GetEffectList(response, req)
	if response.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", response.Code)
	}

	var res Response
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return
	}

	nextJSON := `{"sessionId":"` + res.SessionId + `","page":2,"mailAddress":"","password":""}`
	nextReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(nextJSON))

	GetEffectList(response, nextReq)
	if response.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", response.Code)
	}

	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return
	}

}

func TestChangeEffect(t *testing.T) {
	godotenv.Load()

	bodyJSON := `{"sessionId":"","hashId":"1","dlSecKey":"1"}`
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(bodyJSON))

	ChangeEffect(response, req)
	if response.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", response.Code)
	}

	var res Response
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return
	}
}
