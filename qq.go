package qq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/OhYee/rainbow/errors"
)

// Connect QQ互联应用 对象
type Connect struct {
	appID       string
	appKey      string
	redirectURI string
}

// New 新建一个 QQ互联应用，需要传入已申请成功的应用参数
func New(appID, appKey, redirectURI string) *Connect {
	return &Connect{
		appID:       appID,
		appKey:      appKey,
		redirectURI: redirectURI,
	}

}

type errMsg struct {
	Error    int64  `json:"error"`
	ErrorMsg string `json:"error_description"`
}

func callbackDecoder(b []byte, res interface{}) (err error) {
	err = json.Unmarshal(b[9:len(b)-3], &res)
	errors.Wrapper(&err)
	return
}

/*
LoginPage 返回跳转的登录页面，需要传入一个随机的 state 用于验证身份

https://wiki.connect.qq.com/%E4%BD%BF%E7%94%A8authorization_code%E8%8E%B7%E5%8F%96access_token

需要将用户跳转到登录页面，而后会被跳转回回调地址
*/
func (conn *Connect) LoginPage(state string) (url string) {
	url = fmt.Sprintf("https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", conn.appID, conn.redirectURI, state)
	return
}

/*
Auth 根据登录回调页面返回的 code 获取用户 token

https://wiki.connect.qq.com/%E4%BD%BF%E7%94%A8authorization_code%E8%8E%B7%E5%8F%96access_token
*/
func (conn *Connect) Auth(code string) (token string, err error) {
	params := make(url.Values)
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", conn.appID)
	params.Add("client_secret", conn.appKey)
	params.Add("code", code)
	params.Add("redirect_uri", conn.redirectURI)
	fmt.Printf("%s\n\n", "https://graph.qq.com/oauth2.0/token?"+params.Encode())

	resp, err := http.Get("https://graph.qq.com/oauth2.0/token?" + params.Encode())
	if err != nil {
		errors.Wrapper(&err)
		return
	}

	var b []byte
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		errors.Wrapper(&err)
		return
	}

	urls, err := url.ParseQuery(string(b))
	if err != nil {
		errors.Wrapper(&err)
		return
	}

	token = urls.Get("access_token")
	if token == "" {
		e := new(errMsg)
		if err = callbackDecoder(b, e); err != nil {
			return
		}
		err = errors.New("Error code %d: %s", e.Error, e.ErrorMsg)
	}
	return
}

/*
OpenID 获取用户的唯一ID

返回用户ClientID、OpenID、UnionID或可能存在的错误信息
https://wiki.connect.qq.com/openapi%E8%B0%83%E7%94%A8%E8%AF%B4%E6%98%8E_oauth2-0

此接口用于获取个人信息。开发者可通过openID来获取用户的基本信息。

特别需要注意的是，如果开发者拥有多个移动应用、网站应用，可通过获取用户的unionID来区分用户的唯一性，因为只要是同一QQ互联平台下的不同应用，unionID是相同的。

换句话说，同一用户，对同一个QQ互联平台下的不同应用，unionID是相同的。
*/
func (conn *Connect) OpenID(token string) (ClientID, OpenID, UnionID string, err error) {
	resp, err := http.Get(fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s&unionid=1", token))

	var b []byte
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		errors.Wrapper(&err)
		return
	}

	l := len(b)

	type response struct {
		ClientID string `json:"client_id"`
		OpenID   string `json:"openid"`
		UnionID  string `json:"unionid"`
	}
	res := new(response)
	if err = json.Unmarshal(b[9:l-3], res); err != nil {
		errors.Wrapper(&err)
		return
	}

	if res.OpenID == "" {
		e := new(errMsg)
		callbackDecoder(b, e)
		err = errors.New("Error code %d: %s", e.Error, e.ErrorMsg)
		return
	}

	ClientID = res.ClientID
	OpenID = res.OpenID
	UnionID = res.UnionID

	return
}

// UserInfo get_user_info api response
type UserInfo struct {
	Ret             int64  `json:"ret"`                // 返回代码
	Msg             string `json:"msg"`                // 错误信息
	Nickname        string `json:"nickname"`           // QQ空间的昵称
	FigType         string `json:"figureurl_type"`     // 头像类型
	Fig             string `json:"figureurl"`          // 30×30 像素空间头像
	Fig1            string `json:"figureurl_1"`        // 50×50 像素空间头像
	Fig2            string `json:"figureurl_2"`        // 100×100 像素空间头像
	FigQQ           string `json:"figureurl_qq"`       // 640×640 QQ头像
	FigQQ1          string `json:"figureurl_qq_1"`     // 40×40 QQ头像
	FigQQ2          string `json:"figureurl_qq_2"`     // 100×100 QQ头像（可能为空）
	Gender          string `json:"gender"`             // 性别（未设置则返回“男”）
	GenderType      int64  `json:"gender_type"`        // 性别类型
	Province        string `json:"province"`           // 省份
	City            string `json:"city"`               // 城市
	Year            string `json:"year"`               // 出生年份
	Constellation   string `json:"constellation"`      // 星座
	IsYellowVIP     string `json:"is_yellow_vip"`      // 是否为黄钻用户
	IsYellowYearVIP string `json:"is_yellow_year_vip"` // 是否年费黄钻用户
	YellowVIPLevel  string `json:"yellow_vip_level"`   // 黄钻登机
	VIP             string `json:"vip"`                // 是非VIP
	Level           string `json:"level"`              // VIP等级
	IsLost          int64  `json:"is_lost"`            // 未知
}

/*
Info 获取用户信息

https://wiki.connect.qq.com/get_user_info

获取登录用户在QQ空间的信息，包括昵称、头像、性别及黄钻信息（包括黄钻等级、是否年费黄钻等）。
*/
func (conn *Connect) Info(token, openID string) (res UserInfo, err error) {
	resp, err := http.Get(fmt.Sprintf("https://graph.qq.com/user/get_user_info?access_token=%s&oauth_consumer_key=%s&openid=%s", token, conn.appID, openID))

	var b []byte
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		errors.Wrapper(&err)
		return
	}

	if err = json.Unmarshal(b, &res); err != nil {
		errors.Wrapper(&err)
		return
	}

	if res.Ret != 0 {
		err = fmt.Errorf("Error code %d: %s", res.Ret, res.Msg)
		errors.Wrapper(&err)
		return
	}

	return
}
