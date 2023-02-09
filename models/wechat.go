package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/astaxie/beego"
)

const (
	SENDFAILED  = -1
	SENDSUCCESS = 0
)

//-----------------------------

//返回的响应
type Resp struct {
	State State `json:"state"`
}

type State struct {
	Rc  int    `json:"rc"`
	Msg string `json:"msg"`
}

var token requestToken
var locker sync.RWMutex
var QiYeReqMsg RequestMesage

const (
	send_msg_url          = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s"
	gettokenurl    string = `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=`
	sendmessageurl        = `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=`
)

type responseErr struct {
	Errcode int    `json:"errcode`
	Errmsg  string `json:"errmsg"`
}

type requestToken struct {
	responseErr
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

type RequestMesage struct {
	CorpId     string            `json:"corpid"`
	CorpSecret string            `json:"corpsecret"`
	Touser     string            `json:"touser"`  //成员ID列表（消息接收者，多个接收者用‘|’分隔，最多支持1000个）。特殊情况：指定为@all，则向该企业应用的全部成员发送
	Toparty    string            `json:"toparty"` //部门ID列表，多个接收者用‘|’分隔，最多支持100个。当touser为@all时忽略本参数
	Totag      string            `json:"totag"`   //标签ID列表，多个接收者用‘|’分隔，最多支持100个。当touser为@all时忽略本参数
	Msgtype    string            `json:"msgtype"` //表示是否是保密消息，0表示否，1表示是，默认0
	AgentId    int64             `json:"agentid"` //表示是否是保密消息，0表示否，1表示是，默认0
	Text       map[string]string `json:"text"`
	Safe       int64             `json:"safe"` //表示是否是保密消息，0表示否，1表示是，默认0
}

type QiYiMessage struct {
	Message RequestMesage
}

type CommonMessage struct {
	Touser  string `json:"touser"`
	Toparty string `json:"toparty"`
	AgentId int64  `json:"agentid"`
	Text    string `json:"msg"`
}

func WxInit() {
	initQiYeWechat()
	getToken()
	//定时更新token
	go func() {
		tick := time.NewTicker(time.Minute * 7200)
		for {
			select {
			case <-tick.C:
				{
					getToken()
				}
			}
		}
	}()
}

func initQiYeWechat() {
	QiYeReqMsg = RequestMesage{
		CorpId:     "",
		CorpSecret: "",
		Touser:     "",
		Toparty:    "",
		Totag:      "",
		Msgtype:    "",
		AgentId:    0,
		Text:       map[string]string{},
		Safe:       0,
	}

}

func MessageNotify(sendmsg *CommonMessage) (resp Resp) {
	var msg RequestMesage
	//发送消息
	msg.Touser = sendmsg.Touser //QiYeReqMsg.Touser
	msg.Totag = QiYeReqMsg.Totag
	msg.Toparty = sendmsg.Toparty //QiYeReqMsg.Toparty
	msg.Safe = QiYeReqMsg.Safe
	msg.Msgtype = QiYeReqMsg.Msgtype
	msg.Text = map[string]string{"content": sendmsg.Text}
	msg.AgentId = sendmsg.AgentId //QiYeReqMsg.AgentId

	msgbuf, err := json.Marshal(msg)
	if err != nil {
		resp.State.Rc = SENDFAILED
		resp.State.Msg = "json msg failed," + err.Error()
		return
	}
	err = sendMessage(msgbuf)
	if err != nil {

		resp.State.Rc = SENDFAILED
		resp.State.Msg = "sendMessage failed," + err.Error()

		return
	}
	resp.State.Rc = SENDSUCCESS
	resp.State.Msg = "sendMessage success"

	return
}

func getToken() (err error) {

	corpid := beego.AppConfig.String("CorpId")
	corpsecret := beego.AppConfig.String("CorpSecret")

	var url string = gettokenurl + corpid + "&corpsecret=" + corpsecret
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("get token failed:", err)
		//initbase.Error("get token failed:", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("get token failed:", err)
		//initbase.Error("get token failed:", err)
	}
	body, _ := ioutil.ReadAll(resp.Body)

	locker.Lock()
	defer locker.Unlock()
	err = json.Unmarshal([]byte(body), &token)
	if err != nil {
		fmt.Println("json token failed:", err)
		//initbase.Error("json token failed:", err)
	}
	//initbase.Info("get token info", token)
	fmt.Println("get token info:", token)

	return
}

func sendMessage(msg []byte) error {

	body := bytes.NewBuffer(msg)
	locker.RLock()
	accesstoken := token.Access_token
	locker.RUnlock()

	resp, err := http.Post(sendmessageurl+accesstoken, "application/json", body)
	if resp.StatusCode != 200 {
		return errors.New("request error")
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var e responseErr
	err = json.Unmarshal(buf, &e)
	if err != nil {
		return err
	}

	if e.Errcode == 40014 || e.Errcode == 42001 {
		getToken(&QiYeReqMsg)
		return sendMessage(msg)
	}

	if e.Errcode != 0 && e.Errmsg != "ok" {
		return errors.New(string(buf))
	}
	return nil
}

//func NewCommonMessage() *CommonMessage {
//	return &CommonMessage{
//		Touser:  "",
//		Toparty: "",
//		Msgtype: "Text",
//		Agentid: 0,
//		Text: {
//			Content: "",
//		},
//		Safe: 0,
//	}
//}

func SetJson(touser, toparty, content string, agentid int64) string {
	msg := CommonMessage{
		Touser:  touser,
		Toparty: toparty,
		//Msgtype: "text",
		AgentId: agentid,
		//Safe:    0,
		Text: content,
		// struct {
		// 	//Subject string `json:"subject"`
		// 	Content string `json:"content"`
		// }{Content: content},
	}

	sed_msg, _ := json.Marshal(msg)
	//fmt.Printf("%s",string(sed_msg))
	return string(sed_msg)
}

func SendMessage(access_token, sendmsg string) {
	send_url := fmt.Sprintf(send_msg_url, access_token)
	print(send_url)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", send_url, bytes.NewBuffer([]byte(sendmsg)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("charset", "UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	//fmt.Println("response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
