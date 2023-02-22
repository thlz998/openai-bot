package bootstrap

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"strings"

	"github.com/gin-gonic/gin"

	uuid "github.com/satori/go.uuid"
	"github.com/sbzhu/weworkapi_golang/wxbizmsgcrypt"
	"github.com/thlz998/openai-bot/db"
	"github.com/thlz998/openai-bot/openai"
)

type MsgContent struct {
	ToUsername   string `xml:"ToUserName"`
	FromUsername string `xml:"FromUserName"`
	CreateTime   uint32 `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	Msgid        string `xml:"MsgId"`
	Agentid      uint32 `xml:"AgentId"`
}

func StartWeWorkBot() {
	r := gin.Default()
	if os.Getenv("gin_mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	ping := r.Group("/ping")
	ping.GET("", pingController)
	api := r.Group("/api")
	api.GET("/wework", weworkController)
	api.POST("/wework", weworkSafeController)
	r.Run(os.Getenv("app_port"))
}

func pingController(c *gin.Context) {
	c.JSON(200, struct {
		Data string `json:"ping"`
	}{
		Data: "pong",
	})
}

type callBackMsg struct {
	MsgSignature string `json:"msg_signature" form:"msg_signature" comment:"msg_signature"`
	Timestamp    string `json:"timestamp" form:"timestamp" comment:"timestamp"`
	Nonce        string `json:"nonce" form:"nonce" comment:"nonce"`
	Echostr      string `json:"echostr" form:"echostr" comment:"echostr"`
}

func weworkController(c *gin.Context) {

	token := os.Getenv("wework_token")
	receiverId := os.Getenv("wework_corpid")
	encodingAeskey := os.Getenv("wework_encodingAeskey")
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(token, encodingAeskey, receiverId, wxbizmsgcrypt.XmlType)
	var s callBackMsg
	if err := c.ShouldBind(&s); err == nil {
		verifyMsgSign := s.MsgSignature
		verifyTimestamp := s.Timestamp
		verifyNonce := s.Nonce
		verifyEchoStr := s.Echostr
		echoStr, cryptErr := wxcpt.VerifyURL(verifyMsgSign, verifyTimestamp, verifyNonce, verifyEchoStr)
		if nil != cryptErr {
			c.JSON(500, cryptErr)
			fmt.Println(cryptErr)
			return
		}
		i, err := strconv.Atoi(string(echoStr))
		if err == nil {
			c.JSON(200, i)
		} else {
			c.JSON(200, string(echoStr))
		}
	} else {
		c.JSON(500, err)
	}
}

func weworkSafeController(c *gin.Context) {
	token := os.Getenv("wework_token")
	receiverId := os.Getenv("wework_corpid")
	encodingAeskey := os.Getenv("wework_encodingAeskey")
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(token, encodingAeskey, receiverId, wxbizmsgcrypt.XmlType)
	reqMsgSign := c.Query("msg_signature")
	reqTimestamp := c.Query("timestamp")
	reqNonce := c.Query("nonce")
	reqData, _ := c.GetRawData()
	msg, cryptErr := wxcpt.DecryptMsg(reqMsgSign, reqTimestamp, reqNonce, reqData)
	if nil != cryptErr {
		fmt.Println("DecryptMsg fail", cryptErr)
	}
	var msgContent MsgContent
	err := xml.Unmarshal(msg, &msgContent)
	if nil != err {
		c.JSON(500, "Unmarshal fail")
	} else {
		if msgContent.MsgType == "text" {
			go func() {
				reply(msgContent)
			}()
		}
		c.JSON(200, "success")
	}
}

func reply(msgContent MsgContent) {
	requestText := msgContent.Content
	reply := weWorkReply(requestText, msgContent)
	if reply == "" {
		reply = "我出了一些问题，你可以试试其他的"
	}
	// 从db中获取accessToken
	accessToken := ""
	accessTokenModel := db.WeWorkAccessToken{}
	dbAccessToken := db.DB.First(&accessTokenModel)
	if dbAccessToken.RowsAffected == 0 {
		accessToken = getAccessToken()
		accessTokenModel = db.WeWorkAccessToken{
			AccessToken: accessToken,
		}
		db.DB.Create(&accessTokenModel)
	}
	accessToken = accessTokenModel.AccessToken
	if accessToken == "" || accessTokenModel.UpdatedAt.Add(time.Hour*2).Before(time.Now()) {
		accessToken = getAccessToken()
		db.DB.Model(&accessTokenModel).Update("access_token", accessToken)
	}
	sendMsg(reply, accessToken, msgContent.FromUsername)
}

type weWorkMsgStruct struct {
	Touser  string        `json:"touser"`
	Agentid string        `json:"agentid"`
	Text    weWorkMsgText `json:"text"`
	Msgtype string        `json:"msgtype"`
}
type weWorkMsgText struct {
	Content string `json:"content"`
}

func sendMsg(msg string, accessToken string, user string) {
	dataStruct := weWorkMsgStruct{
		Touser:  user,
		Agentid: os.Getenv("wework_agentid"),
		Text: weWorkMsgText{
			Content: msg,
		},
		Msgtype: "text",
	}

	data, err := json.Marshal(dataStruct)
	if err != nil {
		fmt.Println("请求出现错误", err)
		return
	}
	body := bytes.NewBuffer(data)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://qyapi.weixin.qq.com/cgi-bin/linkedcorp/message/send?&access_token="+accessToken, body)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	parseFormErr := req.ParseForm()
	if parseFormErr != nil {
		fmt.Println(parseFormErr)
	}
	_, err = client.Do(req)
	if err != nil {
		fmt.Println("Failure : ", err)
	}
}

func getAccessToken() string {
	client := &http.Client{}
	receiverId := os.Getenv("wework_corpid")
	corpsecret := os.Getenv("wework_secret")
	req, _ := http.NewRequest("GET", "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid="+receiverId+"&debug=1&corpsecret="+corpsecret, nil)
	parseFormErr := req.ParseForm()
	if parseFormErr != nil {
		fmt.Println(parseFormErr)
	}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	type data struct {
		AccessToken string `json:"access_token"`
	}
	var d data
	json.Unmarshal(respBody, &d)
	// 存入db
	return d.AccessToken
}

func weWorkReply(msg string, msgContent MsgContent) string {
	requestText := strings.TrimSpace(msg)
	apiType := os.Getenv("api_type")
	if apiType == "openai" {
		reply, err := openai.OpenAICompletions(requestText)
		if err != nil {
			fmt.Println(err)
		}
		return *reply
	}
	user := db.User{}
	users := db.DB.Where("username = ?", msgContent.FromUsername).First(&user)
	if users.RowsAffected == 0 {
		user.Username = msgContent.FromUsername
		user.Status = "enabled"
		user.IdType = "wework"
		user.ParentMessageId = uuid.NewV4().String() //
		db.DB.Create(&user)
	}
	messagesId := uuid.NewV4().String()
	chatGPTResponse := openai.ChatGPTResponse{
		MessagesId:      messagesId,
		ConversationId:  user.ConversationId,
		ParentMessageId: user.ParentMessageId,
	}
	reply, err := chatGPTResponse.ChatGPTCompletions(requestText, 2)
	user.ConversationId = chatGPTResponse.ConversationId
	user.ParentMessageId = chatGPTResponse.ParentMessageId
	db.DB.Save(&user)

	if err != nil {
		fmt.Println(err)
	}

	return *reply
}
