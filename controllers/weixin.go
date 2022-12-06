package controllers

import (
	"PrometheusAlert/models"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/ysicing/workwxbot"
)

type Mark struct {
	Content string `json:"content"`
}
type WXMessage struct {
	Msgtype  string `json:"msgtype"`
	Markdown Mark   `json:"markdown"`
}

type WXTxtMessage struct {
	Msgtype string            `json:"msgtype"`
	BotText workwxbot.BotText `json:"text"`
}

func PostToWeiXin(text, WXurl, atuserid, logsign string) string {
	open := beego.AppConfig.String("open-weixin")
	if open != "1" {
		logs.Info(logsign, "[weixin]", "企业微信接口未配置未开启状态,请先配置open-weixin为1")
		return "企业微信接口未配置未开启状态,请先配置open-weixin为1"
	}

	mode := beego.AppConfig.String("wx-mode")
	b := new(bytes.Buffer)
	SendContent := text
	if mode == "0" {
		if atuserid != "" {
			userid := strings.Split(atuserid, ",")
			idtext := ""
			for _, id := range userid {
				idtext += "<@" + id + ">"
			}
			SendContent += idtext
		}

		u := WXMessage{
			Msgtype:  "markdown",
			Markdown: Mark{Content: SendContent},
		}

		json.NewEncoder(b).Encode(u)
	} else if mode == "1" {
		var userlist []string
		if atuserid != "" {
			userid := strings.Split(atuserid, ",")
			for _, id := range userid {
				userlist = append(userlist, "@"+id)
			}
			if len(userlist) <= 0 {
				userlist = []string{"@all"}
			}
		}

		u := WXTxtMessage{
			Msgtype: "text",
			BotText: workwxbot.BotText{
				Content:       SendContent,
				MentionedList: userlist,
			},
		}

		json.NewEncoder(b).Encode(u)
	} else {
		logs.Info(logsign, "[weixin]", "企业微信模式不正确,wx-mode只能为0或1")
		return "企业微信模式不正确"
	}

	logs.Info(logsign, "[weixin]", b)
	var tr *http.Transport
	if proxyUrl := beego.AppConfig.String("proxy"); proxyUrl != "" {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyUrl)
		}
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           proxy,
		}
	} else {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client := &http.Client{Transport: tr}
	logs.Info(logsign, "[weixin]", &b)
	res, err := client.Post(WXurl, "application/json", b)
	if err != nil {
		logs.Error(logsign, "[weixin]", err.Error())
	}
	defer res.Body.Close()
	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logs.Error(logsign, "[weixin]", err.Error())
	}
	models.AlertToCounter.WithLabelValues("weixin").Add(1)
	ChartsJson.Weixin += 1
	logs.Info(logsign, "[weixin]", string(result))
	return string(result)
}
