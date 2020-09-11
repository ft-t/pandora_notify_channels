package web

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"strings"
)

type telegramUpdateHook struct {
	UpdateId int64           `json:"update_id"`
	Message  telegramMessage `json:"message"`
}

type telegramChat struct {
	Id int `json:"id"`
}

type telegramFrom struct {
	Id    int  `json:"id"`
	IsBot bool `json:"is_bot"`
}

type telegramMessage struct {
	Id   int          `json:"message_id"`
	Text string       `json:"text"`
	Chat telegramChat `json:"chat"`
	From telegramFrom `json:"from"`
}

type telegramSendMessageRequest struct {
	ChatId           int    `json:"chat_id"`
	Text             string `json:"text"`
	ReplyToMessageId int    `json:"reply_to_message_id"`
}

func handleWebhookRequest(ctx *fasthttp.RequestCtx) {
	var hookData telegramUpdateHook

	err := json.Unmarshal(ctx.Request.Body(), &hookData)

	if err != nil { // todo log
		ctx.Error(errors.WithStack(err).Error(), 500)
		return
	}
}

func processWebhook(hook telegramUpdateHook) telegramSendMessageRequest {
	command := strings.TrimSpace(hook.Message.Text)

	switch strings.ToLower(command) {
	case "info":
		return handleInfo(hook)
	case "status":

	}
}

func handleInfo(hook telegramUpdateHook) telegramSendMessageRequest {
	return telegramSendMessageRequest{
		ChatId:           hook.Message.Chat.Id,
		Text:             fmt.Sprintf("ChatId %v", hook.Message.Chat.Id),
		ReplyToMessageId: hook.Message.Id,
	}
}

func handleStatus(hook telegramUpdateHook) telegramSendMessageRequest {

}
