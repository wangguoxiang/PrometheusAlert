package controllers

import (
	//"PrometheusAlert/models"
	//"PrometheusAlert/models/elastic"
	//"bytes"
	//"encoding/json"
	//tmplhtml "html/template"
	//"regexp"
	//"strconv"
	//"strings"
	//"text/template"
	//"time"
	"crypto/x509"
	"encoding/pem"
	"net"

	//"net/http"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/spacemonkeygo/openssl"
)

type CheckerController struct {
	beego.Controller
}

type Domaininfo struct {
	Host        string `json:"host"`
	Domain      []byte `json:"domain"`
	Before_time int64  `json:"before_time"`
	After_time  int64  `json:"after_time"`
}

func (c *CheckerController) CheckerDomain() {
	logsign := "[" + LogsSign() + "]"
	domain := c.GetString("domain")
	logs.Info(logsign, string(c.Ctx.Input.RequestBody))
	logs.Info(logsign, c.Data["json"])

	domaininfo := Domaininfo{
		Host:        "0.0.0.0",
		Domain:      []byte(domain),
		Before_time: 0,
		After_time:  0,
	}

	if domain != "" {
		domaininfo.Host = GetDnsIp(domain)
		ctx, err := openssl.NewCtx()
		if err != nil {
			//body.SetStatus(http.StatusInternalServerError, err.Error(), nil, nil, 0)
			panic(err)
			logs.Error(logsign, err.Error())
			c.Data["json"] = domaininfo
			return
		}
		conn, errs := openssl.Dial("tcp", domaininfo.DnsName+":443", ctx, 0)

		if errs != nil {
			//log.Fatal(errs)
			//body.SetStatus(http.StatusInternalServerError, "网络连接错误:"+errs.Error(), nil, nil, 0)
			panic(errs)
			logs.Error(logsign, err.Error())
			c.Data["json"] = domaininfo
			return
		}
		defer conn.Close()
		pem, error := conn.PeerCertificate()
		if error != nil {
			//log.Fatal(error)
			//.SetStatus(http.StatusInternalServerError, "请求DNS"+Dns.DnsName+"失败！", nil, nil, 0)
			logs.Error(logsign, err.Error())
		} else {
			d, _ := pem.MarshalPEM()

			at, bt, err := parsePemData(d)
			if err == nil {
				domaininfo.Domain = d
				domaininfo.Before_time = bt
				domaininfo.After_time = at
				logs.Error(logsign, "success")
				c.Data["json"] = domaininfo
				//body.SetStatus(http.StatusOK, "请求成功！", domaininfo, nil, 0)
			} else {
				//body.SetStatus(http.StatusOK, "错误:"+err.Error(), nil, nil, 0)
				logs.Error(logsign, err)
			}
		}

	}

	c.ServeJSON()
}

func (c *CheckerController) WXSend() {
	c.ServeJSON()
}

func parsePemData(pemdate []byte) (After_time, Before_time int64, err error) {
	logsign := "[" + LogsSign() + "]"
	certDERBlock, _ := pem.Decode(pemdate)
	if certDERBlock == nil {
		logs.Warn(logsign, "value is empty")
		return 0, 0, nil
	}
	x509Cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		logs.Warn(logsign, err.Error())
		return 0, 0, err
	}
	//log.Printf("validation time %s ~ %s",
	//	x509Cert.NotBefore.Format("2006-01-02 15:04"), x509Cert.NotAfter.Format("2006-01-02 15:04"))
	logs.Info(logsign, "validation time %s ~ %s", x509Cert.NotBefore.Format("2006-01-02 15:04"), x509Cert.NotAfter.Format("2006-01-02 15:04"))
	return x509Cert.NotBefore.Unix(), x509Cert.NotAfter.Unix(), nil
}

func GetDnsIp(domain string) (ip string) {
	logsign := "[" + LogsSign() + "]"
	if domain == "" {
		logs.Warn(logsign, "domain is empty")
		return ""
	}

	ns, err := net.LookupHost(domain)
	if err != nil {
		logs.Warn(logsign, err.Error())
		return ""
	}
	//iplist := make(string, len(ns))
	var groupip string
	for i, n := range ns {
		//fmt.Fprintf(os.Stdout, "--%s\n", n)
		if i > 0 {
			groupip = groupip + "," + n
		} else {
			groupip = n
		}
	}

	return groupip
}
