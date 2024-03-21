package wechat

import (
	"github.com/eatmoreapple/openwechat"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
	"time"
)

type MessageType uint8

const (
	TextMessage MessageType = iota
	VisionMessageImage
	VisionMessageText
	ImageGenMessage
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
		wrapped.typ = VisionMessageImage
		err = messageCache.Store(sender.UserName, wrapped)
		if err != nil {
			return nil, errors.Wrap(err, "failed to store message")
		}
		return wrapped, nil
	}

	if raw.IsText() {
		if cached, ok := messageCache.Load(sender.UserName); ok {
			wrapped.typ = VisionMessageText
			log.Debugf("Load %s Cached Vision Message: %v, len: %d", sender.UserName, cached, len(cached))
			wrapped.related = cached
			return wrapped, nil
		}

		if strings.Contains(raw.Content, "生成图片") || strings.Contains(raw.Content, "generate image") {
			wrapped.typ = ImageGenMessage
			return wrapped, nil
		}

		wrapped.typ = TextMessage
		return wrapped, nil
	}

	return nil, errors.New("unsupported message type")
}
