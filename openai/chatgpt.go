package openai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"wechatbot/config"

	log "github.com/sirupsen/logrus"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type VisionMessage struct {
	Role    string           `json:"role"`
	Content []*VisionContent `json:"content"`
}

type VisionContent struct {
	Typ      string                     `json:"type"`
	Text     string                     `json:"text,omitempty"`
	ImageUrl VisionImageContentImageUrl `json:"image_url,omitempty"`
}

type VisionImageContentImageUrl struct {
	Url string `json:"url"`
}

// ChatGPTRequestBody 请求体
type ChatGPTRequestBody struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatGPTVisionRequestBody struct {
	Model    string          `json:"model"`
	Messages []VisionMessage `json:"messages"`
}

type ImageGenRequestBody struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	ResponseFormat string `json:"response_format"`
}

type ImageGenResponseBody struct {
	Created int `json:"created"`
	Data    []struct {
		B64Json string `json:"b64_json,omitempty"`
		Url     string `json:"url,omitempty"`
	} `json:"data"`
}

type ResponseChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatGPTResponseBody 响应体
type ChatGPTResponseBody struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int              `json:"created"`
	Choices []ResponseChoice `json:"choices"`
	Usage   ResponseUsage    `json:"usage"`
}

type ChatGPTErrorBody struct {
	Error map[string]interface{} `json:"error"`
}

/*
curl https://api.openai.com/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_API_KEY' \
  -d '{
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "Hello!"}]
}'

{
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "Hello!"}]
}

{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "\n\nHello there, how may I assist you today?",
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21
  }
}

*/

var contextMgr ContextMgr

// Completions sendMsg
func Completions(msg string, imagePath string) (string, error) {
	apiKey := config.GetOpenAiApiKey()
	if apiKey == "" {
		return "", errors.New("未配置apiKey")
	}

	imageGeneration := strings.HasPrefix(msg, "生成图片")

	var messages []ChatMessage
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: "你是一个乐于助人的人工智能猫猫，名字叫来福福，品种是奶牛猫。",
	})

	list := contextMgr.GetData()
	for i := 0; i < len(list); i++ {
		messages = append(messages, ChatMessage{
			Role:    "user",
			Content: list[i].Request,
		})

		messages = append(messages, ChatMessage{
			Role:    "assistant",
			Content: list[i].Response,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: msg,
	})

	var requestData []byte
	var err error

	if imagePath != "" {
		imageRawData, err := os.ReadFile(imagePath)
		if err != nil {
			return "", err
		}
		base64Image := make([]byte, base64.StdEncoding.EncodedLen(len(imageRawData)))
		base64.StdEncoding.Encode(base64Image, imageRawData)

		requestBody := ChatGPTVisionRequestBody{
			Model: config.GetOpenAiVisionModel(),
			Messages: []VisionMessage{
				{
					Role: "user",
					Content: []*VisionContent{
						{
							Typ:  "text",
							Text: msg,
						},
						{
							Typ: "image_url",
							ImageUrl: VisionImageContentImageUrl{
								Url: "data:image/jpeg;base64," + string(base64Image),
							},
						},
					},
				},
			},
		}
		requestData, err = json.Marshal(requestBody)
	} else {
		requestBody := ChatGPTRequestBody{
			Model:    config.GetOpenAiTextModel(),
			Messages: messages,
		}
		requestData, err = json.Marshal(requestBody)
	}

	if err != nil {
		log.Error(err)
		return "", err
	}

	var req *http.Request

	if imageGeneration {
		requestBody := ImageGenRequestBody{
			Model:          config.GetOpenAiImageGenModel(),
			Prompt:         strings.TrimPrefix(msg, "生成图片："),
			N:              1,
			Size:           "1024x1024",
			ResponseFormat: "url",
		}
		requestData, err = json.Marshal(requestBody)
		if err != nil {
			log.Error(err)
			return "", err
		}
		req, err = http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(requestData))
		if err != nil {
			log.Error(err)
			return "", err
		}
		log.Infof("request openai image gen json string : %v", string(requestData))
	} else {
		log.Debugf("request openai json string : %v", string(requestData))
		req, err = http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestData))
		if err != nil {
			log.Error(err)
			return "", err
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New("openai response status code is not 200")
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	log.Infof("openai response body: %v", string(body))

	if imageGeneration {
		imageGenResponseBody := &ImageGenResponseBody{}
		err = json.Unmarshal(body, imageGenResponseBody)
		if err != nil {
			log.Error(err)
			return "", err
		}

		if len(imageGenResponseBody.Data) > 0 {
			return imageGenResponseBody.Data[0].Url, nil
		}
		return "", nil
	}

	gptResponseBody := &ChatGPTResponseBody{}
	log.Debug(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		log.Error(err)
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		for _, v := range gptResponseBody.Choices {
			reply += "\n"
			reply += v.Message.Content
		}

		contextMgr.AppendMsg(msg, reply)
	}

	if len(reply) == 0 {
		gptErrorBody := &ChatGPTErrorBody{}
		err = json.Unmarshal(body, gptErrorBody)
		if err != nil {
			log.Error(err)
			return "", err
		}

		reply += "Error: "
		reply += gptErrorBody.Error["message"].(string)
	}

	return reply, nil
}
