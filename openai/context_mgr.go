package openai

import (
	"time"
)

const contextExpireTime = 2 * 60

type Context struct {
	Request  string
	Response string
	Time     int64
}

type ContextMgr struct {
	contextList []*Context
}

func NewContextMgr() *ContextMgr {
	return &ContextMgr{
		contextList: make([]*Context, 0),
	}
}

func (m *ContextMgr) checkExpire() {
	timeNow := time.Now().Unix()
	if len(m.contextList) > 0 {
		startPos := len(m.contextList) - 1
		for i := 0; i < len(m.contextList); i++ {
			if timeNow-m.contextList[i].Time < contextExpireTime {
				startPos = i
				break
			}
		}

		m.contextList = m.contextList[startPos:]
	}
}

func (m *ContextMgr) AppendMsg(request string, response string) {
	m.checkExpire()
	context := &Context{Request: request, Response: response, Time: time.Now().Unix()}
	m.contextList = append(m.contextList, context)
}

func (m *ContextMgr) GetData() []*Context {
	m.checkExpire()
	return m.contextList
}

func (m *ContextMgr) BuildMsg() []ChatMessage {
	messages := make([]ChatMessage, 0)
	list := m.GetData()
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
	return messages
}
