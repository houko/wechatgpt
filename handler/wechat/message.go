package wechat

import (
	"github.com/eatmoreapple/openwechat"
	"github.com/pkg/errors"
	"os"
	"runtime"
	"strings"
	"time"
)

type MessageType uint8

const (
	TextMessage MessageType = iota
	ImageMessage
	TextMessageVisionChat
	TextMessageImageGen
	TextMessageImageEdit
	TextMessageImageVariate
)

type Message struct {
	*openwechat.Message
	typ                MessageType
	related            []*Message
	imagePathIfPicture string
	createTime         int64
}

func WrapMessage(raw *openwechat.Message) (*Message, error) {
	sender, err := raw.Sender()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sender")
	}

	wrapped := &Message{Message: raw, createTime: time.Now().Unix()}
	runtime.SetFinalizer(wrapped, func(m *Message) {
		if m.imagePathIfPicture != "" {
			_ = os.Remove(m.imagePathIfPicture)
		}
	})

	if raw.IsPicture() {
		wrapped.typ = ImageMessage
		err = messageCache.Store(sender.UserName, wrapped)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store message")
		}
		return wrapped, nil
	}

	if raw.IsText() {
		if cached, ok := messageCache.Load(sender.UserName); ok {
			wrapped.related = cached

			if strings.Contains(raw.Content, "编辑图片") ||
				strings.Contains(raw.Content, "edit image") ||
				strings.Contains(raw.Content, "修改图片") ||
				strings.Contains(raw.Content, "change image") ||
				strings.Contains(raw.Content, "图片编辑") ||
				strings.Contains(raw.Content, "图片修改") {
				wrapped.typ = TextMessageImageEdit
				return wrapped, nil
			}

			if strings.Contains(raw.Content, "变异图片") ||
				strings.Contains(raw.Content, "variate image") ||
				strings.Contains(raw.Content, "图片变异") ||
				strings.Contains(raw.Content, "图片变化") {
				wrapped.typ = TextMessageImageVariate
				return wrapped, nil
			}

			wrapped.typ = TextMessageVisionChat
			wrapped.related = cached
			return wrapped, nil
		}

		if strings.Contains(raw.Content, "生成图片") ||
			strings.Contains(raw.Content, "generate image") ||
			strings.Contains(raw.Content, "图片生成") {
			wrapped.typ = TextMessageImageGen
			return wrapped, nil
		}

		wrapped.typ = TextMessage
		return wrapped, nil
	}

	return nil, errors.New("unsupported message type")
}
