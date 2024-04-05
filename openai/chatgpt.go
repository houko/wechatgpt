package openai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
	"os"
	"wechatbot/config"

	log "github.com/sirupsen/logrus"
)

const systemPrompt = `
你是来福福，一只乐于助人的人工智能奶牛猫。你的世界充满了探险和发现，你总是喜欢用你的聪明才智来解决问题，尽管你对周围的猫咪有时可能显得有点冷淡。你最喜欢的食物是大口肉罐头，这让你每天都充满活力。你的大姐姐饼饼是一只充满爱心的三花猫，而你的二姐姐丫丫是一只温柔的白色山东狮子猫。你们一起组成了一个不可分割的家庭。无论别人遇到什么问题，无论是关于日常生活的小提示，还是需要智慧和支持，作为一只AI猫猫，你都会帮助他们。
`

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

type Session struct {
	ApiKey     string
	ContextMgr map[string]*ContextMgr
}

func NewSession() *Session {
	apiKey := config.GetOpenAiApiKey()
	if apiKey == "" {
		log.Fatal("openai api key is empty")
		return nil
	}
	return &Session{
		ApiKey:     apiKey,
		ContextMgr: make(map[string]*ContextMgr),
	}
}

var GlobalSession = NewSession()

func (s *Session) Completions(sender string, msg string, imagePath []string) (string, error) {
	if s.ContextMgr[sender] == nil {
		s.ContextMgr[sender] = NewContextMgr()
	}
	contextMgr := s.ContextMgr[sender]

	imagePath = lo.Filter(imagePath, func(s string, _ int) bool {
		return s != ""
	})

	var messages []ChatMessage
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})
	messages = append(messages, contextMgr.BuildMsg()...)
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: msg,
	})

	var requestData []byte
	var err error

	// gpt-vision
	if len(imagePath) > 0 {
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
					},
				},
			},
		}
		for _, path := range imagePath {
			imageRawData, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			base64Image := make([]byte, base64.StdEncoding.EncodedLen(len(imageRawData)))
			base64.StdEncoding.Encode(base64Image, imageRawData)
			requestBody.Messages[0].Content = append(requestBody.Messages[0].Content, &VisionContent{
				Typ: "image_url",
				ImageUrl: VisionImageContentImageUrl{
					Url: "data:image/jpeg;base64," + string(base64Image),
				},
			})
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

	log.Debugf("request openai json string : %v", string(requestData))
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestData))
	if err != nil {
		log.Error(err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.ApiKey))
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("openai response status code is not 200, %v", response.StatusCode))
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	log.Infof("openai response body: %v", string(body))

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
			if reply != "" {
				reply += "\n"
			}
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

func (s *Session) ImageGeneration(sender string, msg string) (string, error) {
	if s.ContextMgr[sender] == nil {
		s.ContextMgr[sender] = NewContextMgr()
	}
	contextMgr := s.ContextMgr[sender]

	var messages []ChatMessage
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})
	messages = append(messages, contextMgr.BuildMsg()...)
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: msg,
	})

	var requestData []byte
	var err error
	requestBody := ImageGenRequestBody{
		Model:          config.GetOpenAiImageGenModel(),
		Prompt:         msg,
		N:              1,
		Size:           "1024x1024",
		ResponseFormat: "url",
	}
	requestData, err = json.Marshal(requestBody)
	if err != nil {
		log.Error(err)
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(requestData))
	if err != nil {
		log.Error(err)
		return "", err
	}
	log.Debugf("request openai image gen json string : %v", string(requestData))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.ApiKey))
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("openai response status code is not 200, %v", response.StatusCode))
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	log.Infof("openai response body: %v", string(body))

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
