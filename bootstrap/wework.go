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

	"strings"

	"github.com/gin-gonic/gin"

	"gitee.com/cloudtrek/chatgptbot/openai"
	"github.com/sbzhu/weworkapi_golang/wxbizmsgcrypt"
	log "github.com/sirupsen/logrus"
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
	fmt.Println("weworkSafeController")
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
	reply := *weWorkReply(requestText)
	if reply == "" {
		reply = "我出了一下问题，你可以试试其他的"
	}
	accessToken := getAccessToken()
	sendMsg(reply, accessToken, msgContent.FromUsername)
}

func sendMsg(msg string, accessToken string, user string) {
	fmt.Println("sendMsg", msg)
	fmt.Println("accessToken", accessToken)
	fmt.Println("user", user)
	json := []byte(`{"touser": "` + user + `","agentid": 1000107,"text": {"content": "` + msg + `"},"msgtype": "text"}`)
	body := bytes.NewBuffer(json)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://qyapi.weixin.qq.com/cgi-bin/linkedcorp/message/send?access_token="+accessToken, body)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	parseFormErr := req.ParseForm()
	if parseFormErr != nil {
		fmt.Println(parseFormErr)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failure : ", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body : ", string(respBody))
}

func getAccessToken() string {
	client := &http.Client{}
	receiverId := os.Getenv("wework_corpid")
	corpsecret := os.Getenv("wework_secret")
	req, _ := http.NewRequest("GET", "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid="+receiverId+"&corpsecret="+corpsecret, nil)
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
	return d.AccessToken
}

func weWorkReply(msg string) *string {
	requestText := strings.TrimSpace(msg)
	reply, err := openai.Completions(requestText)
	if err != nil {
		log.Println(err)
	}
	return reply
}
