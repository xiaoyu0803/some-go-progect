package main

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	APPID          = "wxdd426537d721b4bd"
	APPSECRET      = "1141f3d4ba04a0876cadfb9a35493729"
	SentTemplateID = "eJXGLnsdTmzZnEczkQ1rdPItLML1hySZVValLZRUhFY" //每日一句的模板ID，替换成
	WeatherVersion = "v1"
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type sentence struct {
	Content  string `json:"content"`
	birthday string `json:"note"`
}

func main() {
	spec := "0 0 7 * * *" // 每天12:00
	//spec1 := "0 0 7 * * *" // 每天早晨7:00
	c := cron.New()
	//c.AddFunc("@e", everydaysen)
	c.AddFunc(spec, weather)
	c.Start()
	fmt.Println("开启定时任务")
	http.HandleFunc("/tq", func(writer http.ResponseWriter, request *http.Request) {
		weather()

	})
	http.ListenAndServe(":8007", nil)
	select {}
	//weather()
	//everydaysen()

}

//发送天气预报
func weather() {
	fmt.Println("aaaa")
	access_token := getaccesstoken()
	if access_token == "" {
		return
	}

	flist := getflist(access_token)
	if flist == nil {
		return
	}
	fmt.Println("flist", flist)
	for _, v := range flist {
		go sendweather(access_token, "北京", v.Str)

	}
	fmt.Println("weather is ok")
}

//获取微信accesstoken
func getaccesstoken() string {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%v&secret=%v", APPID, APPSECRET)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取微信token失败", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("微信token读取失败", err)
		return ""
	}

	token := token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		fmt.Println("微信token解析json失败", err)
		return ""
	}
	fmt.Println(token.AccessToken, 85)
	return token.AccessToken
}

//获取每日一句
func getsen() (sentence, string) {
	resp, err := http.Get("https://v2.alapi.cn/api/qinghua?token=8X7GFNo89FBX0IMq")
	sent := sentence{}
	if err != nil {
		fmt.Println("获取每日一句失败", err)
		return sent, ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return sent, ""
	}
	fmt.Println(103, string(body))
	sent.Content = gjson.Get(string(body), "data.content").Str
	if err != nil {
		fmt.Println("每日一句解析json失败")
		return sent, ""
	}
	fenxiangurl := gjson.Get(string(body), "fenxiang_img").String()

	sent.birthday = string(getTimesubDay())
	return sent, fenxiangurl
}

//获取关注者列表
func getflist(access_token string) []gjson.Result {
	url := "https://api.weixin.qq.com/cgi-bin/user/get?access_token=" + access_token + "&next_openid="
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取关注列表失败", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return nil
	}
	flist := gjson.Get(string(body), "data.openid").Array()
	return flist
}

//发送模板消息
func templatepost(access_token string, reqdata string, fxurl string, templateid string, openid string) {
	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + access_token

	reqbody := "{\"touser\":\"" + openid + "\", \"template_id\":\"" + templateid + "\", \"url\":\"" + fxurl + "\", \"data\": " + reqdata + "}"

	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(string(reqbody)))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}

//获取
func getweather(city string) (string, string, string, string) {

	url := fmt.Sprintf("https://devapi.qweather.com/v7/weather/3d?location=101010100&key=2f8267ce5c2546c9b8526a811e155ec6")
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取天气失败", err)
		return "", "", "", ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return "", "", "", ""
	}

	data := gjson.Get(string(body), "daily").Array()
	//fmt.Println(data)
	thisday := data[0].String()
	day := gjson.Get(thisday, "fxDate").Str
	wea := gjson.Get(thisday, "textDay").Str
	temax := gjson.Get(thisday, "tempMax").Str
	temin := gjson.Get(thisday, "tempMin").Str
	//tem2 := gjson.Get(thisday, "tem2").Str
	//air_tips := gjson.Get(thisday, "air_tips").Str
	return day, wea, temax, temin
}

//发送天气
func sendweather(access_token, city, openid string) {
	day, wea, temmax, temmin := getweather(city)
	fmt.Println(day, wea, temmax, temmin)
	if day == "" || wea == "" || temmax == "" {
		return
	}
	req, _ := getsen()
	fmt.Println(194, req)
	if req.Content == "" {
		return
	}

	reqdata := "{\"content\":{\"value\":\"" + req.Content + "\", \"color\":\"#0000CD\"}, \"note\":{\"value\":\"我们已经在一起：" + req.birthday + "天了\",\"color\":\"#ff0000\"},\"city\":{\"value\":\"城市：" + city + "\", \"color\":\"#0000CD\"}, \"day\":{\"value\":\"" + day + "\"}, \"wea\":{\"value\":\"天气：" + wea + "\"}, \"tem1\":{\"value\":\"最高温度：" + temmax + "\"},\"tem2\":{\"value\":\"最低温度：" + temmin + "\"}}"

	fmt.Println(reqdata)
	templatepost(access_token, reqdata, "http://43.138.43.191:9999", SentTemplateID, openid)
}
func getTimesubDay() string {
	var day int
	t1, _ := time.Parse("2006-01-02", "2020-10-31")
	c := time.Now().Format("2006-01-02")
	t2, _ := time.Parse("2006-01-02", c)
	swap := false
	fmt.Println(t1, "\n", t2)
	if t1.Unix() > t2.Unix() {
		t1, t2 = t2, t1
		swap = true
	}

	t1_ := t1.Add(time.Duration(t2.Sub(t1).Milliseconds()%86400000) * time.Millisecond)
	day = int(t2.Sub(t1).Hours() / 24)
	// 计算在t1+两个时间的余数之后天数是否有变化
	if t1_.Day() != t1.Day() {
		day += 1
	}

	if swap {
		day = -day
	}
	fmt.Println(224, day)
	day2 := strconv.Itoa(day)
	return day2
}
