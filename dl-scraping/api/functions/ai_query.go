package functions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/vertexai/genai"
)

type modelTypes int

const (
	GPT_4o modelTypes = iota
	GPT_4o_mini
	Gemini_1_5_flash
	Gemini_1_5_pro
)

type AiResponse struct {
	Message string `json:"message"`
}

func queryGpt(aiResponse *AiResponse, ctx context.Context, storeClient *firestore.Client, model modelTypes, prompt string, imageData string, systemInstructions string, temperature float64, responseFormat map[string]interface{}) error {
	openAiApiKey := os.Getenv("OPEN_AI_API_KEY")

	url := "https://api.openai.com/v1/chat/completions"
	modelName := ""
	if model == GPT_4o {
		modelName = "gpt-4o-2024-08-06"
	} else if model == GPT_4o_mini {
		modelName = "gpt-4o-mini-2024-07-18"
	}

	userContent := []map[string]interface{}{
		{"type": "text", "text": prompt},
	}
	if imageData != "" {
		userContent = append(userContent, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": imageData,
			},
		})
	}

	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": userContent,
		},
	}

	if systemInstructions != "" {
		messages = append([]map[string]interface{}{
			{
				"role": "system",
				"content": []map[string]interface{}{
					{"type": "text", "text": systemInstructions},
				},
			},
		}, messages...)
	}

	reqBody := map[string]interface{}{
		"model":       modelName,
		"messages":    messages,
		"temperature": temperature,
	}
	if responseFormat != nil {
		reqBody["response_format"] = responseFormat
	}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(reqBodyJson)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openAiApiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	var resBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return err
	}

	choices := resBody["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	content := message["content"].(string)

	usage := resBody["usage"].(map[string]interface{})
	completionTokens := int(usage["completion_tokens"].(float64))
	promptTokens := int(usage["prompt_tokens"].(float64))
	println("completion_tokens: ", completionTokens)
	println("prompt_tokens: ", promptTokens)

	aiResponse.Message = content

	return nil
}

func getSchemaType(schemaType string) genai.Type {
	switch schemaType {
	case "object":
		return genai.TypeObject
	case "array":
		return genai.TypeArray
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "boolean":
		return genai.TypeBoolean
	case "integer":
		return genai.TypeInteger
	default:
		return genai.TypeUnspecified
	}
}

func convertJsonSchemaToGeminiSchema(responseFormat map[string]interface{}) (*genai.Schema, error) {
	thisType, ok := responseFormat["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid schema format")
	}

	schema := &genai.Schema{
		Type:       getSchemaType(thisType),
		Properties: make(map[string]*genai.Schema),
	}

	properties, ok := responseFormat["properties"].(map[string]interface{})
	if ok {
		for key, value := range properties {
			propMap, ok := value.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid property format for key: %s", key)
			}

			propSchema := &genai.Schema{}
			if propType, ok := propMap["type"].(string); ok {
				fmt.Printf("propType: %s\n", propType)
				propSchema.Type = getSchemaType(propType)
			}

			if description, ok := propMap["description"].(string); ok {
				propSchema.Description = description
			}

			if propSchema.Type == genai.TypeArray {
				items, ok := propMap["items"].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid items format for key: %s", key)
				}
				itemSchema, err := convertJsonSchemaToGeminiSchema(items)
				if err != nil {
					return nil, err
				}
				propSchema.Items = itemSchema
			}

			if propSchema.Type == genai.TypeObject {
				subSchema, err := convertJsonSchemaToGeminiSchema(propMap)
				if err != nil {
					return nil, err
				}
				propSchema.Properties = subSchema.Properties
			}

			schema.Properties[key] = propSchema
		}
	}

	if required, ok := responseFormat["required"].([]interface{}); ok {
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				schema.Required = append(schema.Required, reqStr)
			}
		}
	}

	return schema, nil
}

func queryGemini(aiResponse *AiResponse, ctx context.Context, prompt string, imageData string, systemInstructions string, temperature float64, responseFormat map[string]interface{}) error {

	client, err := genai.NewClient(ctx, "asa-o-experiment", "asia-northeast1")
	gemini := client.GenerativeModel("gemini-1.5-flash")
	if systemInstructions != "" {
		gemini.SystemInstruction = &genai.Content{
			Role:  "user",
			Parts: []genai.Part{genai.Text(systemInstructions)},
		}
	}
	gemini.SetTemperature(float32(temperature))

	promptParts := []genai.Part{
		genai.Text(prompt),
	}
	if imageData != "" {
		parts := strings.Split(imageData, ";base64,")
		imageType := strings.Split(parts[0], "/")[1]
		imageBinary, _ := base64.StdEncoding.DecodeString(parts[1])

		img := genai.ImageData(imageType, imageBinary)

		promptParts = append([]genai.Part{img}, promptParts...)
	}

	// スキーマが指定されている場合は、geminiのスキーマに変換して設定
	if responseFormat != nil {
		if json_schema, ok := responseFormat["json_schema"].(map[string]interface{}); ok {
			if schema, ok := json_schema["schema"].(map[string]interface{}); ok {
				geminiSchema, err := convertJsonSchemaToGeminiSchema(schema)
				if err != nil {
					return err
				}
				gemini.GenerationConfig.ResponseMIMEType = "application/json"
				gemini.GenerationConfig.ResponseSchema = geminiSchema
			}
		}
	}

	resp, err := gemini.GenerateContent(ctx, promptParts...)
	if err != nil {
		return fmt.Errorf("error generating content: %w", err)
	}

	inputTokens := int(resp.UsageMetadata.PromptTokenCount)
	outputTokens := int(resp.UsageMetadata.CandidatesTokenCount)
	println("input_tokens: ", inputTokens)
	println("output_tokens: ", outputTokens)

	aiResponse.Message = fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	return nil
}
