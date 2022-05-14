package auth

import (
	"encoding/json"
	"goweb/author-admin/server/models"
	"goweb/author-admin/server/pkg/e"
	"goweb/author-admin/server/pkg/util"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
)

type auth struct {
	Username string `valid:"Required; MaxSize(50)" json:"username"`
	Password string `valid:"Required; MaxSize(50)" json:"password"`
}

func validate(username, password string) bool {
	valid := validation.Validation{}
	a := auth{Username: username, Password: password}
	ok, _ := valid.Valid(&a)

	for _, err := range valid.Errors {
		log.Println(err.Key, err.Message)
	}

	// 防止注入：也许beego的validation已经考虑了此情况
	semicolon := ";"
	var usernameFlag bool
	if strings.Index(username, semicolon) < 0 {
		usernameFlag = true
	}
	var passwordFlag bool
	if strings.Index(password, semicolon) < 0 {
		passwordFlag = true
	}

	return ok && usernameFlag && passwordFlag
}

func Login(c *gin.Context) {
	// username := c.PostForm("username")
	// password := c.PostForm("password")

	decoder := json.NewDecoder(c.Request.Body)
	var au auth
	decoder.Decode(&au)
	username := au.Username
	password := au.Password
	log.Println(username, ":", password)

	ok := validate(username, password)

	data := make(map[string]interface{})
	code := e.INVALID_PARAMS
	if ok {
		isExist := models.CheckAuth(username, password)
		if isExist {
			token, err := util.GenerateToken(username, password)
			if err != nil {
				code = e.ERROR_AUTH_TOKEN_FAIL
			} else {
				data["token"] = token
				code = e.SUCCESS
			}
		} else {
			code = e.ERROR_AUTH
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": data,
	})
}

func Logout(c *gin.Context) {
	code := e.SUCCESS
	data := ""
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": data,
	})
}

func Info(c *gin.Context) {
	token := c.Query("token")
	code := e.SUCCESS

	var claims *util.Claims
	var err error
	if token == "" {
		code = e.ERROR_AUTH_TOKEN_ILLEGAL
	} else {
		claims, err = util.ParseToken(token)
		if err != nil {
			code = e.ERROR_AUTH_TOKEN_ILLEGAL
		} else if time.Now().Unix() > claims.ExpiresAt {
			code = e.ERROR_AUTH_TOKEN_EXPIRED
		}
	}

	name := claims.Username
	roles := []string{"admin"}
	data := make(map[string]interface{})
	data["roles"] = roles
	data["introduction"] = ""
	data["avatar"] = ""
	data["name"] = name
	var statusCode int
	if code == e.SUCCESS {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusBadRequest
	}
	c.JSON(statusCode, gin.H{
		"code": code,
		"data": data,
	})
}

// func Regist(c *gin.Context) {
// 	username := c.PostForm("username")
// 	password := c.PostForm("password")

// 	ok := validate(username, password)
// }
