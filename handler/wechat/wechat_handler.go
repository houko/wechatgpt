package wechat

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"wechatbot/config"
	"wechatbot/openai"
	"wechatbot/utils"

	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
)

var _ MessageHandlerInterface = (*GroupMessageHandler)(nil)

type GroupMessageHandler struct {
}

func (gmh *GroupMessageHandler) handle(msg *openwechat.Message) error {
	if msg.IsText() {
		return gmh.ReplyText(msg)
	}

	if msg.IsPicture() {
		return gmh.ReplyPicture(msg)
	}

	return nil
}

func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{}
}

var imageQuestionQueue = &sync.Map{}

func init() {
	go RevokeImageQuestionQueueDaemon()
}

func RevokeImageQuestionQueueDaemon() {
	for {
		imageQuestionQueue.Range(func(key, value interface{}) bool {
			filePath := value.(string)
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				imageQuestionQueue.Delete(key)
				return true
			}

			if fileInfo.ModTime().Unix()+60*2 < fileInfo.ModTime().Unix() {
				imageQuestionQueue.Delete(key)
				_ = os.Remove(filePath)
				return true
			}

			return true
		})
		time.Sleep(10 * time.Second)
	}
}

func (gmh *GroupMessageHandler) ReplyPicture(msg *openwechat.Message) error {
	sender, err := msg.Sender()
	group := openwechat.Group{User: sender}
	log.Printf("Received Group %v Text Msg : %v", group.NickName, msg.Content)

	resp, err := msg.GetFile()
	if err != nil {
		// handle err here
		return err
	}
	defer resp.Body.Close()
	file, _ := os.CreateTemp("", "wechat_handle.image.*")
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	log.Println("接收图片：", file.Name())
	_, err = msg.ReplyText("请问您需要了解关于这张图片的什么问题？")
	if err != nil {
		return err
	}
	imageQuestionQueue.Store(sender.UserName, file.Name())

	return nil
}

func (gmh *GroupMessageHandler) ReplyText(msg *openwechat.Message) error {
	sender, err := msg.Sender()
	group := openwechat.Group{User: sender}
	log.Printf("Received Group %v Text Msg : %v", group.NickName, msg.Content)

	wechat := config.GetWechatKeyword()
	requestText := msg.Content
	if wechat != "" {
		content, key := utils.ContainsI(requestText, wechat)
		if len(key) == 0 {
			return nil
		}

		splitItems := strings.Split(content, key)
		if len(splitItems) < 2 {
			return nil
		}

		requestText = strings.TrimSpace(splitItems[1])
	}

	log.Println("问题：", requestText)
	imagePath := ""
	if cache, ok := imageQuestionQueue.Load(sender.UserName); ok {
		imagePath = cache.(string)
		defer imageQuestionQueue.Delete(sender.UserName)
		defer os.Remove(imagePath)
	}

	reply, err := openai.Completions(requestText, imagePath)
	if err != nil {
		log.Println(err)
		if reply != "" {
			// 如果文字超过4000个字会回错，截取前4000个文字进行回复
			if len(reply) > 4000 {
				_, err = msg.ReplyText(reply[:4000])
				if err != nil {
					log.Println("回复出错：", err.Error())
					return err
				}
			}
		}

		text, err := msg.ReplyText(fmt.Sprintf("bot error: %s", err.Error()))
		log.Println(text)
		return err
	}

	// 如果在提问的时候没有包含？,AI会自动在开头补充个？看起来很奇怪
	if strings.HasPrefix(reply, "?") {
		reply = strings.Replace(reply, "?", "", -1)
	}

	if strings.HasPrefix(reply, "？") {
		reply = strings.Replace(reply, "？", "", -1)
	}

	// 微信不支持markdown格式，所以把反引号直接去掉
	if strings.Contains(reply, "`") {
		reply = strings.Replace(reply, "`", "", -1)
	}

	if reply != "" {
		_, err = msg.ReplyText(reply)
		if err != nil {
			log.Println(err)
		}
		return err
	}

	return nil
}
