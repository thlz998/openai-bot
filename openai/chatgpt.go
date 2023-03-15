package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
)

type ChatGPTResponse struct {
	MessagesId      string `json:"messages_id"`
	ParentMessageId string `json:"parent_message_id"`
	ConversationId  string `json:"conversation_id"`
}

type Message struct {
	ID      string         `json:"id"`
	Content MessageContent `json:"content"`
}
type MessageContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type MessageWrapper struct {
	Message        Message `json:"message"`
	ConversationId string  `json:"conversation_id"`
	Error          string  `json:"error"`
}

type createConversationResponse struct {
	Action          string                         `json:"action"`
	ParentMessageId string                         `json:"parent_message_id"`
	Model           string                         `json:"model"`
	Messages        []messagesConversationResponse `json:"messages"`
}
type updateConversationResponse struct {
	createConversationResponse
	ConversationId string `json:"conversation_id"`
}

type messagesConversationResponse struct {
	Id      string                             `json:"id"`
	Role    string                             `json:"role"`
	Content MessageContent                     `json:"content"`
	Author  messagesAuthorConversationResponse `json:"author"`
}

type messagesAuthorConversationResponse struct {
	Role string `json:"role"`
}

func (c *ChatGPTResponse) ChatGPTCompletions(msg string, count int) (*string, error) {
	if count < 0 {
		v := "异常错误"
		return &v, nil
	}
	if count == 0 {
		// 重置会话
		c.ConversationId = ""
		return c.ChatGPTCompletions(msg, -1)
	}

	var data []byte
	var err error
	createData := createConversationResponse{
		Action:          "next",
		ParentMessageId: c.ParentMessageId,
		Model:           os.Getenv("app_model"),
		Messages: []messagesConversationResponse{messagesConversationResponse{
			Id: c.MessagesId,
			Content: MessageContent{
				ContentType: "text",
				Parts:       []string{msg},
			},
			Author: messagesAuthorConversationResponse{
				Role: "user",
			},
		}},
	}
	if c.ConversationId != "" {
		data, err = json.Marshal(updateConversationResponse{
			createConversationResponse: createData,
			ConversationId:             c.ConversationId,
		})
	} else {
		data, err = json.Marshal(createData)
	}
	if err != nil {
		fmt.Println("请求出现错误", err)
		return nil, err
	}
	ttt := string(data)
	fmt.Println(ttt)
	jsonReader := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", "https://chat.openai.com/backend-api/conversation", jsonReader)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authority", "chat.openai.com")
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Accept-Language", "zh,zh-CN;q=0.9")
	req.Header.Add("Authorization", os.Getenv("chatgpt_authorization"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", os.Getenv("chatgpt_cookie"))
	req.Header.Add("Origin", "https://chat.openai.com")
	req.Header.Add("Referer", "https://chat.openai.com/chat/"+c.ConversationId)
	req.Header.Add("Sec-Ch-Ua", "\"Chromium\";v=\"110\", \"Not A(Brand\";v=\"24\", \"Google Chrome\";v=\"110\"")
	req.Header.Add("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Add("Sec-Ch-Ua-Platform", "\"macOS\"")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", os.Getenv("chatgpt_useragent"))
	req.Header.Add("Accept-Encoding", "gzip")
	// 发送 HTTP 请求并获取响应
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.ChatGPTCompletions(msg, count-1)
	}
	if resp.StatusCode == 429 {
		v := "请求次数过多，请稍后再试"
		return &v, nil
	}
	if resp.StatusCode == 404 {
		c.ConversationId = ""
		fmt.Println("创建新会话", c.ConversationId)
		return c.ChatGPTCompletions(msg, count)
	}
	if resp.StatusCode != 200 {
		v := "遇到了一起奇怪的错误" + resp.Status
		return &v, nil
	}
	defer resp.Body.Close()
	// 解析 SSE 数据流
	scanner := bufio.NewScanner(resp.Body)
	value := ""
	for scanner.Scan() {
		value += scanner.Text()
	}

	re := regexp.MustCompile(`data:(.*?)data:`)
	matches := re.FindAllStringSubmatch(value, -1)
	if len(matches) < 1 {
		v := "遇到了一起奇怪的错误, 找不到匹配项"
		return &v, nil
	}
	if scanner.Err() != nil {
		v := "遇到了一起奇怪的错误: " + scanner.Err().Error()
		return &v, nil
	}
	s := matches[len(matches)-1][1]

	var messageWrapper MessageWrapper
	err = json.Unmarshal([]byte(s), &messageWrapper)
	if err != nil {
		v := "解析结果时候遇到了错误: " + err.Error()
		return &v, nil
	}
	c.ConversationId = messageWrapper.ConversationId
	c.ParentMessageId = messageWrapper.Message.ID
	text := ""
	if len(messageWrapper.Message.Content.Parts) > 0 {
		text = messageWrapper.Message.Content.Parts[0]
	}
	return &text, nil
}
