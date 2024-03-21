package wechat

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
	"time"
)

const (
	cacheRevocationTime = 60 * 2
	cacheRevokeInterval = 10 * time.Second
)

type MessageCache struct {
	m *sync.Map
}

func NewMessageCache() *MessageCache {
	return &MessageCache{
		m: &sync.Map{},
	}
}

func (i *MessageCache) Store(key string, msgs ...*Message) error {
	list, ok := i.m.Load(key)
	if !ok {
		list = make([]*Message, 0)
	}

	for _, msg := range msgs {
		if msg.IsPicture() {
			log.Info("Received Image Msg, saving to cache")
			resp, err := msg.GetFile()
			if err != nil {
				return errors.Wrap(err, "failed to get file")
			}
			file, _ := os.CreateTemp("", "wechat_handle.image.*")
			_, err = io.Copy(file, resp.Body)
			resp.Body.Close()
			file.Close()
			if err != nil {
				return errors.Wrap(err, "failed to copy file")
			}
			msg.imagePathIfPicture = file.Name()
		}
		list = append(list.([]*Message), msg)
		log.Debugf("Store %s Cached Vision Message: %v, len: %d", key, list, len(list.([]*Message)))
	}
	i.m.Store(key, list)
	return nil
}

func (i *MessageCache) Load(key string) ([]*Message, bool) {
	v, ok := i.m.Load(key)
	if !ok {
		return nil, false
	}
	return v.([]*Message), true
}

func (i *MessageCache) Delete(key string) {
	i.m.Delete(key)
}

func (i *MessageCache) Range(f func(key, msg any) bool) {
	i.m.Range(f)
}

var messageCache = NewMessageCache()

func init() {
	go revokeImageCacheDaemon()
}

func revokeImageCacheDaemon() {
	for {
		messageCache.Range(func(key, value any) bool {
			list := value.([]*Message)
			for i := len(list) - 1; i >= 0; i-- {
				// remove expired message
				// cacheRevocationTime is 2 minutes, convert to int64
				if list[i].createTime+cacheRevocationTime < time.Now().Unix() {
					log.Debugf("Revoke %s Cached Vision Message: %v, len: %d", key, list[i], len(list))
					newList := make([]*Message, len(list)-i-1)
					copy(newList, list[i+1:])
					messageCache.m.Store(key, newList)
					break
				}
			}
			return true
		})
		time.Sleep(cacheRevokeInterval)
	}
}
