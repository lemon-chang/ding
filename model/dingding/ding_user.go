package dingding

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"database/sql/driver"
	"ding/global"
	"ding/initialize/jwt"
	"ding/model/params/ding"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const secret = "liwenzhou.com"

var wg = sync.WaitGroup{}
var hrefs []string
var blogs []string
var start int64 = 1676822400
var end int64 = 1677427200
var jin *DingUser

type Strs []string

type DingUser struct {
	UserId              string ` gorm:"primaryKey;foreignKey:UserId" json:"userid"`
	DingRobots          []DingRobot
	Deleted             gorm.DeletedAt
	Name                string     `json:"name"`
	Mobile              string     `json:"mobile"`
	Password            string     `json:"password"`
	DeptIdList          []int      `json:"dept_id_list" gorm:"-"` //所属部门
	DeptList            []DingDept `json:"dept_list" gorm:"many2many:user_dept"`
	Title               string     `json:"title"` //职位
	JianShuAddr         string     `json:"jianshu_addr"`
	BlogAddr            string     `json:"blog_addr"`
	AuthToken           string     `json:"auth_token" gorm:"-"`
	DingToken           `json:"ding_token" gorm:"-"`
	JianShuArticleURL   Strs `gorm:"type:longtext" json:"jian_shu_article_url"`
	BlogArticleURL      Strs `gorm:"type:longtext" json:"blog_article_url"`
	IsExcellentJianBlog bool `json:"is_excellentBlogJian" `
}

func (d *DingUser) SendFrequencyLeave(start int) error {
	fmt.Println("推送个人请假频率")
	return nil
}
func (d *DingUser) CountFrequencyLeave(startWeek int, result map[string][]DingAttendance) error {
	fmt.Println("存储个人请假频率")
	return nil
}

type JinAndBlog struct {
	UserId            string `gorm:"primary_key" json:"id"`
	Name              string `json:"name"`
	JianShuArticleURL Strs   `gorm:"type:longtext" json:"jian_shu_article_url"`
	BlogArticleURL    Strs   `gorm:"type:longtext" json:"blog_article_url"`
	IsExcellent       bool   `json:"is_excellent"`
}

func (d *DingUser) GetUserByUserId() (user DingUser, err error) {
	err = global.GLOAB_DB.Where("user_id = ?", d.UserId).First(&user).Error
	return
}

func (d *DingUser) Login() (user *DingUser, err error) {
	user = &DingUser{
		Mobile:   d.Mobile,
		Password: d.Password,
	}
	//此处的Login函数传递的是一个指针类型的数据
	if err := Login(user); err != nil {
		return nil, err
	}
	// 生成JWT
	token, err := jwt.GenToken(user.UserId, user.Name)
	if err != nil {
		zap.L().Debug("JWT生成错误")
		return
	}
	user.AuthToken = token
	return user, err
}
func encryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}
func Login(user *DingUser) (err error) {
	opassword := user.Password //此处是用户输入的密码，不一定是对的
	result := global.GLOAB_DB.Where(&DingUser{Mobile: user.Mobile}).First(user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return ErrorUserNotExist
	}
	if result.Error != nil {
		return result.Error
	}
	//如果到了这里还没有结束的话，那就说明该用户至少是存在的，于是我们解析一下密码
	//password := encryptPassword(opassword)
	password := opassword
	//拿到解析后的密码，我们看看是否正确
	if password != user.Password {
		return ErrorInvalidPassword
	}
	//如果能到这里的话，那就登录成功了
	return nil

}
func (d *DingUser) GetUserDetailByUserId() (user DingUser, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/get?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		UserId string `json:"userid"`
	}{UserId: d.UserId}

	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		User DingUser `json:"result"` //必须大写，不然的话，会被忽略，从而反序列化不上
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return DingUser{}, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可

	return r.User, nil

}

// ImportUserToMysql 把钉钉用户信息导入到数据库中
func (d *DingUser) ImportUserToMysql() error {
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		token, err := (&DingToken{}).GetAccessToken()
		if err != nil {
			zap.L().Error("从redis中取出token失败", zap.Error(err))
			return err
		}
		var deptIdList []int
		err = tx.Model(&DingDept{}).Select("dept_id").Find(&deptIdList).Error
		if err != nil {
			zap.L().Error("从数据库中取出所有部门id失败", zap.Error(err))
			return err
		}
		for i := 0; i < len(deptIdList); i++ {
			Dept := &DingDept{DeptId: deptIdList[i], DingToken: DingToken{Token: token}}
			DeptUserList, _, err := Dept.GetUserListByDepartmentID(0, 100)

			if err != nil {
				zap.L().Error("获取部门成员失败", zap.Error(err))
				continue
			}
			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "title"}),
			}).Create(&DeptUserList)
			if err != nil {
				zap.L().Error(fmt.Sprintf("存储部门:%s成员到数据库失败", Dept.Name), zap.Error(err))
				continue
			}
		}
		return err
	})

}

func (d *DingUser) FindDingUsers() (us []DingUser, err error) {
	//keys, err := global.GLOBAL_REDIS.Keys(context.Background(), "user*").Result()
	err = global.GLOAB_DB.Model(&DingUser{}).Select("user_id", "name", "mobile").Find(&us).Error
	//往redis中做一份缓存
	//for i := 0; i < len(us); i++ {
	//	batchData := make(map[string]interface{})
	//	batchData["name"] = us[i].Name
	//	_, err := global.GLOBAL_REDIS.HMSet(context.Background(), "user:"+us[i].UserId, batchData).Result()
	//	if err != nil {
	//		zap.L().Error("把数据缓存到redis中失败", zap.Error(err))
	//	}
	//}
	return
}

// UpdateDingUserAddr 根据用户id修改其简书和博客地址
func (d *DingUser) UpdateDingUserAddr(userParam ding.UserAndAddrParam) error {
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Model(&DingUser{}).Where("user_id = ?", userParam.UserId).Updates(DingUser{BlogAddr: userParam.BlogAddr, JianShuAddr: userParam.JianShuAddr}).Error; err != nil {
			zap.L().Error("更新用户博客和简书地址失败", zap.Error(err))
			return
		}
		return
	})
}

func (d *DingUser) GoCrawlerDingUserJinAndBlog() (err error) {
	//spec = "00 03,33,45 08,14,21 * * ?"
	//spec := "50 30 21 * * 1"
	//task := func() {
	err = global.GLOAB_DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&DingUser{}).UpdateColumn("is_excellent_jian_blog", false).Error
	if err != nil {
		zap.L().Error("爬取文章之前清空之前的优秀简书博客人员", zap.Error(err))
	}
	wg.Add(2)
	//go UpdateDingUserHref()
	//UpdateDingUserJianShu()
	//go UpdateDingUserBlog()
	UpdateDingUserBlog()
	wg.Wait()
	zap.L().Info("爬取完毕，已成功存入数据库")
	//}
	//taskId, err := global.GLOAB_CORN.AddFunc(spec, task)

	if err != nil {
		zap.L().Error("启动爬虫爬取文文章地址失败", zap.Error(err))
	}
	//zap.L().Info(fmt.Sprintf("启动爬虫爬取文文章地址定时任务成功（非爬虫成功），定时任务id:%v", taskId))
	return err
}
func UpdateDingUserJianShu() {
	//获取所有人的博客和简书主页链接
	urls, err := (&DingUser{}).FindDingUserAddr()
	if err != nil {
		zap.L().Error("获取简书链接错误", zap.Error(err))
		return
	}

	for _, v := range urls {
		falg := true
		if v.JianShuAddr == "" {
			//为了避免偶然因素
			err := v.UpdateDingUserHref(hrefs, v.UserId)
			if err != nil {
				fmt.Println("空简书链接未清空")
			}
			continue
		}
		client := http.Client{}
		v.JianShuAddr = strings.ReplaceAll(v.JianShuAddr, "\n", "")
		//strings.Replace(v.JianShuAddr, "\n", "", -1)
		req, err := http.NewRequest("GET", v.JianShuAddr, nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Content-Type", "text/html; charset=utf-8")
		req.Header.Set("Keep-Alive", "timeout=30")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		//解析网页
		docDetail, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		docDetail.Find("#list-container > ul > li ").Each(func(i int, selection *goquery.Selection) {
			if falg == false {
				return
			}
			//获取文章链接
			attr, exists := selection.Find(".content > a").Attr("href")
			//获取时间，格式为utc时间
			text, exits := selection.Find(".content > .meta > span.time").Attr("data-shared-at")
			if exits == true {
				//拿到当前时刻的时间戳和最近一篇简书的时间戳
				timeUnix := time.Now().Unix()
				timeUnix = 1678022322
				t1 := UTCTransLocal(text)
				unix := switchTime(t1)
				//先把之前的文章进行一下清空
				err = v.UpdateDingUserHref(hrefs, v.UserId)
				if err != nil {
					zap.L().Error("清空记录失败", zap.Error(err))
				}
				for end < timeUnix {
					start += 604800
					end += 604800
				}
				if unix >= start && unix < end {
					zap.L().Info(fmt.Sprintf("%v简书文章爬取成功，文章链接：%v,文章发布时间：%v", v.Name, attr, text))
					if exists == true {
						str := "https://www.jianshu.com" + attr
						hrefs = append(hrefs, str)
					}
				} else {
					zap.L().Info(fmt.Sprintf("%v简书文章爬取成功，旦不满足时间要求，结束爬取，文章链接：%v,文章发布时间：%v", v.Name, attr, text))
					falg = false
					return
				}

			}

		})

		err = v.UpdateDingUserHref(hrefs, v.UserId)
		if err != nil {
			zap.L().Error("更新简书数组到数据库出错", zap.Error(err))
			return
		}
		hrefs = []string{}
	}
	wg.Done()
}
func UpdateDingUserBlog() {

	urls, err := jin.FindDingUserAddr()
	if err != nil {
		zap.L().Error("获取简书链接错误", zap.Error(err))
		return
	}
	for _, v := range urls {
		if v.Name != "闫佳鹏" {
			continue
		}
		if v.BlogAddr == "" {
			err = v.UpdateDingUserBlog(blogs, v.UserId)
			if err != nil {
				fmt.Println("空博客链接未清空")
			}
			zap.L().Info(fmt.Sprintf("%v博客链接是空，直接跳过", v.Name))
			continue
		}
		target := v.BlogAddr + "/article/list"
		htmls, err := GetHtml(target)
		if err != nil {
			fmt.Println(err)
			panic("Get target ERROR!!!")
		}

		var html string
		html = strings.Replace(string(htmls), "\n", "", -1)
		html = strings.Replace(string(htmls), " ", "", -1)

		//fmt.Println(html)
		reBlog := regexp.MustCompile(`<div class="article-item-box csdn-tracking-statistics(.*?)</div>`)
		reLink := regexp.MustCompile(`href="(.*?)"`)
		reTime := regexp.MustCompile(`<span class="date">(.*?)</span>`)

		articles := reBlog.FindAllString(html, -1)
		if articles == nil || len(articles) == 0 {
			zap.L().Info(fmt.Sprintf("%s本周未写博客", v.Name))
			continue
		}
		for _, value := range articles {
			BlogLink := reLink.FindAllStringSubmatch(value, -1)

			BlogTime := reTime.FindAllStringSubmatch(value, -1)

			timeUnix := time.Now().Unix()
			timeUnix = 1678024664
			t1 := UTCTransLocal(BlogTime[0][1])
			unix := switchTime(t1)

			err = v.UpdateDingUserBlog(blogs, v.UserId)
			if err != nil {
				zap.L().Error("清空博客数据失败", zap.Error(err))
				return
			}
			for end < timeUnix {
				start += 604800
				end += 604800
			}
			if unix >= start && unix < end {
				hrefs = append(hrefs, BlogLink[0][1])
				zap.L().Info(fmt.Sprintf("%v博客文章爬取成功，文章链接：%v,文章发布时间：%v", v.Name, BlogLink, BlogTime))
			} else {
				zap.L().Info(fmt.Sprintf("%v博客文章爬取成功，旦不满足时间要求，结束爬取，文章链接：%v,文章发布时间：%v", v.Name, BlogLink, BlogTime))
				break
			}
		}
		err = v.UpdateDingUserBlog(blogs, v.UserId)
		if err != nil {
			zap.L().Error("更新博客数组到数据库出错", zap.Error(err))
			return
		}
		blogs = []string{}
	}
	wg.Done()
}

func switchTime(ans string) (unix int64) {
	loc, _ := time.LoadLocation("Local")
	theTime, err := time.ParseInLocation("2006-01-02 15:04:05", ans, loc)
	if err == nil {
		unix = theTime.Unix()
		return unix
	}
	return
}

func UTCTransLocal(utcTime string) string {
	loc, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation("2006-01-02T15:04:05+08:00", utcTime, loc)
	return t.Local().Format("2006-01-02 15:04:05")
}

func GetHtml(URL string) (html []byte, err error) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: true,
		// Proxy:              http.ProxyURL(proxyUrl),
	}

	req, err := http.NewRequest("GET", URL, nil)
	req.Header.Add("UserAgent", " Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.41")

	client := &http.Client{
		Transport: tr, /*使用transport参数*/
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	html, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return html, err
}

func (d *DingUser) FindDingUserAddr() (addrs []DingUser, err error) {
	var address []DingUser
	err = global.GLOAB_DB.Model(&DingUser{}).Select("jian_shu_addr", "blog_addr", "user_id", "name").Find(&address).Error
	if err != nil {
		zap.L().Error("获取钉钉用户的简书或博客链接失败", zap.Error(err))
		return
	}
	return address, nil
}

func (d *DingUser) UpdateDingUserHref(jins Strs, id string) (err error) {
	err = global.GLOAB_DB.Model(&DingUser{}).Where("user_id = ?", id).UpdateColumns(map[string]interface{}{
		"jian_shu_article_url": jins,
	}).Error
	if err != nil {
		zap.L().Error("在mysql中更新这周简书链接失败", zap.Error(err))
		return
	}
	return nil
}

func (d *DingUser) UpdateDingUserBlog(blogs Strs, id string) (err error) {
	err = global.GLOAB_DB.Model(&DingUser{}).Where("user_id = ?", id).UpdateColumns(map[string]interface{}{
		"blog_article_url": blogs,
	}).Error
	if err != nil {
		zap.L().Error("在mysql中更新这周博客链接失败", zap.Error(err))
		return
	}
	return nil
}

func (m *Strs) Scan(val interface{}) error {
	s := val.([]uint8)
	ss := strings.Split(string(s), "|")
	*m = ss
	return nil
}

func (m Strs) Value() (driver.Value, error) {
	str := strings.Join(m, "|")
	return str, nil
}

//获取二维码buf，chatId, title
func (u *DingUser) GetQRCode(c *gin.Context) (buf []byte, chatId, title string, err error) {
	zap.L().Info("进入到了chromedp")
	d := data{}
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		chromedp.NoFirstRun, //设置网站不是首次运行
		chromedp.WindowSize(1921, 1024),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	//defer cancel()
	print(cancel)

	// 创建上下文实例
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// 创建超时上下文
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	// navigate to a page, wait for an element, click

	// capture screenshot of an element

	// capture entire browser viewport, returning png with quality=90
	var html string
	fmt.Println("开始运行chromedp")
	err = chromedp.Run(ctx,
		//打开网页
		chromedp.Navigate(`https://open-dev.dingtalk.com/apiExplorer?spm=ding_open_doc.document.0.0.20bf4063FEGqWg#/jsapi?api=biz.chat.chooseConversationByCorpId`),
		//定位登录按钮
		chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//等二维码出现
		chromedp.WaitVisible(`document.querySelector(".ant-modal")`, chromedp.ByJSPath),
		//截图
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, _, _, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			buf, err = page.CaptureScreenshot().
				WithQuality(90).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			username, _ := c.Get(global.CtxUserNameKey)
			fmt.Println(username)
			err = ioutil.WriteFile(fmt.Sprintf("./Screenshot_%s.png", username), buf, 0644)
			if err != nil {
				zap.L().Error("二维码写入失败", zap.Error(err))
			}
			zap.L().Info("写入二维码成功", zap.Error(err))
			return nil
		}),
		//等待用户扫码连接成功
		chromedp.WaitVisible(`document.querySelector(".connect-info")`, chromedp.ByJSPath),
		//chromedp.SendKeys(`document.querySelector("#corpId")`, "caonima",chromedp.ByJSPath),
		//设置输入框中的值为空
		chromedp.SetValue(`document.querySelector("#corpId")`, "", chromedp.ByJSPath),
		//chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//chromedp.Clear(`#corpId`,chromedp.ByID),
		//输入正确的值
		chromedp.SendKeys(`document.querySelector("#corpId")`, "ding7625646e1d05915a35c2f4657eb6378f", chromedp.ByJSPath),
		//点击发起调用按钮
		chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),

		chromedp.WaitVisible(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div:nth-child(2) > pre > span > span.cm-tab")`, chromedp.ByJSPath),
		//自定义函数进行爬虫
		chromedp.ActionFunc(func(ctx context.Context) error {
			//b := chromedp.WaitEnabled(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre")`, chromedp.ByJSPath)
			//b.Do(ctx)
			a := chromedp.OuterHTML(`document.querySelector("body")`, &html, chromedp.ByJSPath)
			a.Do(ctx)
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Println("123", err.Error())
				return err
			}
			var data string
			dom.Find("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre").Each(func(i int, selection *goquery.Selection) {
				data = data + selection.Text()
				selection.Next()
			})
			data = strings.ReplaceAll(data, " ", "")
			data = strings.ReplaceAll(data, "\n", "")
			reader := strings.NewReader(data)
			bytearr, err := ioutil.ReadAll(reader)

			err1 := json.Unmarshal(bytearr, &d)
			if err1 != nil {

			}
			return nil
		}),
	)
	if err != nil {
		zap.L().Error("使用chromedp失败", zap.Error(err))
		return nil, "", "", err
	}
	if &d == nil {
		return nil, "", "", err
	}
	return buf, d.Result.ChatId, d.Result.Title, err
}

var ChromeCtx context.Context

func GetChromeCtx(focus bool) context.Context {
	if ChromeCtx == nil || focus {
		allocOpts := chromedp.DefaultExecAllocatorOptions[:]
		allocOpts = append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			//chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
			//chromedp.Flag("headless", false),
			chromedp.Flag("blink-settings", "imagesEnabled=false"), //开启图像界面,重点是开启这个
			chromedp.Flag("ignore-certificate-errors", true),       //忽略错误
			chromedp.Flag("disable-web-security", true),            //禁用网络安全标志
			chromedp.Flag("disable-extensions", true),              //开启插件支持
			chromedp.Flag("accept-language", `zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,zh-TW;q=0.6`),
			//chromedp.Flag("disable-default-apps", true),
			//chromedp.NoFirstRun, //设置网站不是首次运行
			chromedp.WindowSize(1921, 1024),
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
		)

		if checkChromePort() {
			// 不知道为何，不能直接使用 NewExecAllocator ，因此增加 使用 ws://127.0.0.1:9222/ 来调用
			c, _ := chromedp.NewRemoteAllocator(context.Background(), "ws://172.17.0.8:9222/")
			ChromeCtx, _ = chromedp.NewContext(c)
		} else {
			c, _ := chromedp.NewExecAllocator(context.Background(), allocOpts...)
			ChromeCtx, _ = chromedp.NewContext(c)
		}
	}
	return ChromeCtx
}
func (u *DingUser) GetQRCode1(c *gin.Context) (buf []byte, chatId, title string, err error) {

	timeCtx, cancel := context.WithTimeout(GetChromeCtx(false), 5*time.Minute)
	defer cancel()
	d := data{}
	var html string
	fmt.Println("开始运行chromedp")
	err = chromedp.Run(timeCtx,
		//打开网页
		chromedp.Navigate(`https://open-dev.dingtalk.com/apiExplorer?spm=ding_open_doc.document.0.0.20bf4063FEGqWg#/jsapi?api=biz.chat.chooseConversationByCorpId`),
		//定位登录按钮
		chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//等二维码出现
		chromedp.WaitVisible(`document.querySelector(".ant-modal")`, chromedp.ByJSPath),
		//截图
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, _, _, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			buf, err = page.CaptureScreenshot().
				WithQuality(90).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			username, _ := c.Get(global.CtxUserNameKey)
			err = ioutil.WriteFile(fmt.Sprintf("./Screenshot_%s.png", username), buf, 0644)
			if err != nil {
				zap.L().Error("二维码写入失败", zap.Error(err))
			}
			return nil
		}),
		//等待用户扫码连接成功

		chromedp.WaitVisible(`document.querySelector(".connect-info")`, chromedp.ByJSPath),
		//chromedp.SendKeys(`document.querySelector("#corpId")`, "caonima",chromedp.ByJSPath),
		//设置输入框中的值为空
		chromedp.SetValue(`document.querySelector("#corpId")`, "", chromedp.ByJSPath),
		//chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//chromedp.Clear(`#corpId`,chromedp.ByID),
		//输入正确的值
		chromedp.SendKeys(`document.querySelector("#corpId")`, "ding7625646e1d05915a35c2f4657eb6378f", chromedp.ByJSPath),
		//点击发起调用按钮
		chromedp.Click(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.param-list > div > div.api-param-footer > button")`, chromedp.ByJSPath),
		//chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div:nth-child(2) > pre > span > span.cm-tab")`, chromedp.ByJSPath),
		//自定义函数进行爬虫
		chromedp.ActionFunc(func(ctx context.Context) error {
			//b := chromedp.WaitEnabled(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre")`, chromedp.ByJSPath)
			//b.Do(ctx)
			a := chromedp.OuterHTML(`document.querySelector("body")`, &html, chromedp.ByJSPath)
			a.Do(ctx)
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Println("123", err.Error())
				return err
			}
			var data string
			dom.Find("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre").Each(func(i int, selection *goquery.Selection) {
				data = data + selection.Text()
				selection.Next()
			})
			data = strings.ReplaceAll(data, " ", "")
			data = strings.ReplaceAll(data, "\n", "")
			reader := strings.NewReader(data)
			bytearr, err := ioutil.ReadAll(reader)

			err1 := json.Unmarshal(bytearr, &d)
			if err1 != nil {

			}
			return nil
		}),
	)
	if err != nil {
		zap.L().Error("使用chromedp失败", zap.Error(err))
		return nil, "", "", err
	}
	if &d == nil {
		return nil, "", "", err
	}
	return buf, d.Result.ChatId, d.Result.Title, err
}
func checkChromePort() bool {
	addr := net.JoinHostPort("", "9222")
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
func (u *DingUser) GetRobotList() (RobotList []DingRobot, err error) {
	//err = global.GLOAB_DB.Where("ding_user_id = ?", u.UserId).Find(&RobotList).Error
	err = global.GLOAB_DB.Model(u).Association("DingRobots").Find(&RobotList)
	return
}

func (u *DingRobot) GetOpenConversationId(access_token, chatId string) (openConversationId string, _err error) {
	client, _err := createClient()
	if _err != nil {
		return
	}

	chatIdToOpenConversationIdHeaders := &dingtalkim_1_0.ChatIdToOpenConversationIdHeaders{}
	chatIdToOpenConversationIdHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.ChatIdToOpenConversationIdWithOptions(tea.String(chatId), chatIdToOpenConversationIdHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		openConversationId = *(result.Body.OpenConversationId)
		return nil
	}()

	if tryErr != nil {
		var err = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			err = _t
		} else {
			err.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
			// err 中含有 code 和 message 属性，可帮助开发定位问题
		}

	}
	return
}
