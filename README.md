# QQ互联基本接口封装

[![Sync to Gitee](https://github.com/OhYee/qqconnect/workflows/Sync%20to%20Gitee/badge.svg)](https://gitee.com/OhYee/qqconnect)

[QQ互联](https://connect.qq.com/)

首先在前端跳转至 `LoginPage()` 返回的页面，进行登录
接着在回调页面按顺序调用下面的函数，即可成功获取用户信息

其中，`openID`每个用户重新登录后，在本应用内不变，而`unionID`则在不同的QQ互联应用中也不会改变
从`res`可以得到用户的基本资料(不过实际上也就头像和昵称有意义)

```go
conn := qq.New("Your app id", "Your app key", "Your redirect uri")

token, err := conn.Auth(code)
if err != nil {
    return
}
output.Debug("%+v", token)

_, openID, unionID, err := conn.OpenID(token)
if err != nil {
    return
}
res, err := conn.Info(token, openID)
if err != nil {
    return
}
```