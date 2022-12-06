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

	"github.com/astaxie/beego"
	//"github.com/astaxie/beego/logs"
)

type CheckerController struct {
	beego.Controller
}

func (c *CheckerController) CheckerController() {
	c.ServeJSON()
}
