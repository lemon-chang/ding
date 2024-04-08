package dingding

import (
	"context"
	"crypto/tls"
	"ding/global"
	"ding/model/common/request"
	"ding/model/system"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const secret = "liwenzhou.com"

type Strs []string

type DingUser struct {
	UserId           string      `gorm:"primaryKey;foreignKey:UserId" json:"userid"`
	DingRobots       []DingRobot `json:"ding_robots,omitempty"`
	Deleted          gorm.DeletedAt
	Name             string                `json:"name"`
	Mobile           string                `json:"mobile"`
	Password         string                `json:"password,omitempty" `
	DeptIdList       []int                 `json:"dept_id_list,omitempty" gorm:"-"` //所属部门
	DeptId           int                   `json:"dept_id"`
	DeptList         []DingDept            `json:"dept_list,omitempty" gorm:"many2many:user_dept"`
	AuthorityId      uint                  `json:"authorityId,omitempty" gorm:"default:888;comment:用户角色ID"`
	Authority        system.SysAuthority   `json:"authority,omitempty" gorm:"foreignKey:AuthorityId;references:AuthorityId;comment:用户角色"`
	Authorities      []system.SysAuthority `json:"authorities,omitempty" gorm:"many2many:sys_user_authority;"`
	IsStudyWeekPaper int                   `json:"is_study_week_paper"` // 是否学习周报
	IsLeetCode       int                   `json:"is_leet_code"`
	IsJianShuOrBlog  int                   `json:"is_jianshu_or_blog" gorm:"column:is_jianshu_or_blog"`
	Title            string                `json:"title,omitempty"` // 职位
	JianshuAddr      string                `json:"jianshu_addr"`
	BlogAddr         string                `json:"blog_addr"`
	LeetcodeAddr     string                `json:"leetcode_addr"`
	AuthToken        string                `json:"auth_token" gorm:"-"`
	DingToken        `json:"ding_token,omitempty" gorm:"-"`
	Admin            bool `json:"admin,omitempty" gorm:"-"`
	ExtAttrs         []struct {
		Code  string `json:"code"`
		Name  string `json:"name"`
		Value struct {
			Text string `json:"text"`
		} `json:"value"`
	} `json:"ext_attrs" gorm:"-"`
}

// 通过userid查询部门id
func GetDeptByUserId(UserId string) (user *DingUser) {
	err := global.GLOAB_DB.Where("user_id = ?", UserId).Preload("DeptList").First(&user).Error
	if err != nil {
		zap.L().Error("通过userid查询用户失败", zap.Error(err))
		return
	}
	return
}

func (d *DingUser) GetUserInfo() (user DingUser, err error) {
	err = global.GLOAB_DB.Where("user_id = ?", d.UserId).First(&user).Error
	return
}
func (d *DingUser) GetUserInfoDetailByUserId() (err error) {
	err = global.GLOAB_DB.Where("user_id = ?", d.UserId).Preload("Authority").Preload("Authorities").Preload("DeptList").First(&d).Error
	return
}
func (d *DingUser) Delete() (err error) {
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Unscoped().Delete(d).Error
		if err != nil {
			return err
		}
		// 添加部门信息
		err = tx.Model(d).Association("DeptList").Clear()
		if err != nil {
			return err
		}
		return err
	})
	return
}
func (d *DingUser) Add() (err error) {
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		err = d.GetUserDetailByUserId()
		d.DeptList = make([]DingDept, len(d.DeptIdList))
		for i := 0; i < len(d.DeptList); i++ {
			d.DeptList[i].DeptId = d.DeptIdList[i]
		}
		err = tx.Create(d).Error
		if err != nil {
			return err
		}
		// 添加部门信息
		err = tx.Model(d).Association("DeptList").Replace(d.DeptList)
		if err != nil {
			return err
		}
		return err
	})
	return
}
func (d *DingUser) UpdateByDingEvent() (err error) {
	d.DeptList = make([]DingDept, len(d.DeptIdList))
	for i := 0; i < len(d.DeptList); i++ {
		d.DeptList[i].DeptId = d.DeptIdList[i]
	}
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Select("name", "title", "jianshu_addr", "blog_addr", "mobile", "leetcode_addr").Updates(d).Error
		if err != nil {
			return err
		}
		err = tx.Model(d).Association("DeptList").Replace(d.DeptList)
		return err
	})
	return
}
func (d *DingUser) Update() (err error) {
	err = global.GLOAB_DB.Select("name", "title", "jianshu_addr", "blog_addr", "mobile", "leetcode_addr", "is_study_week_paper", "is_leet_code", "is_jianshu_or_blog").Updates(d).Error
	return
}

// 初始化用户表部门字段--获取全部用户
func (d *DingUser) InitGetUserId() (users []DingUser, err error) {
	err = global.GLOAB_DB.Select("user_id").Find(&users).Error
	return
}

// 初始化用户表部门字段--更新（插入）用户的数据
func (d *DingUser) InitInsertDeptId(deptId int) (err error) {
	var user DingUser
	err = global.GLOAB_DB.Where("user_id = ? and dept_id=?", d.UserId, deptId).Find(&user).Limit(1).Error
	if user.UserId == "" {
		err = global.GLOAB_DB.Where("user_id = ?", d.UserId).Find(&user).Limit(1).Error
		if user.DeptId == 0 {
			err = global.GLOAB_DB.Model(&user).Update("dept_id", deptId).Limit(1).Error
		}
		if user.DeptId != deptId {
			user.DeptId = deptId
			err = global.GLOAB_DB.Model(&user).Create(&user).Error
		}
	}
	return
}
func (d *DingUser) GetIsWeekPaperUsersByDeptId(deptId, flg int) (users []DingUser, err error) {
	userIds := global.GLOAB_DB.Table("user_dept").Select("ding_user_user_id").Where("ding_dept_dept_id=? ", deptId)
	err = global.GLOAB_DB.Where("is_week_paper = ? and user_id IN (?)", flg, userIds).Find(&users).Error
	return
}
func (d *DingUser) GetWeekPaperUsersStatusByDeptId(deptId int) (users []DingUser, err error) {
	userIds := global.GLOAB_DB.Table("user_dept").Select("ding_user_user_id").Where("ding_dept_dept_id=? ", deptId)
	err = global.GLOAB_DB.Where("user_id IN (?)", userIds).Find(&users).Error
	return
}

func (d *DingUser) GetUserDeptIdByUserId() (dept DingDept, err error) {
	userId := global.GLOAB_DB.Table("user_dept").Select("ding_dept_dept_id").Where("ding_user_user_id=? ", d.UserId).Limit(1)
	err = global.GLOAB_DB.Select("dept_id").Where("user_id  =  ?", userId).Find(&dept).Error
	return
}

// 更新部门周报检测状态
func (d *DingUser) UpdateUserWeekCheckStatus() (err error) {
	err = global.GLOAB_DB.Model(d).Where("user_id", d.UserId).Update("is_week_paper", d.IsStudyWeekPaper).Error
	return
}
func (d *DingUser) InitUserWeekPaper(num int) (err error) {
	err = global.GLOAB_DB.Model(d).Update("is_week_paper", num).Error
	return nil
}

func (m *DingUser) UserAuthorityDefaultRouter(user *DingUser) {
	var menuIds []string
	err := global.GLOAB_DB.Model(&system.SysAuthorityMenu{}).Where("sys_authority_authority_id = ?", user.AuthorityId).Pluck("sys_base_menu_id", &menuIds).Error
	if err != nil {
		return
	}
	var am system.SysBaseMenu
	err = global.GLOAB_DB.First(&am, "name = ? and id in (?)", user.Authority.DefaultRouter, menuIds).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user.Authority.DefaultRouter = "404"
	}
}

// https://open.dingtalk.com/document/isvapp/query-user-details
func (d *DingUser) GetUserDetailByUserId() (err error) {
	token, err := (&DingToken{}).GetAccessToken()
	d.Token = token
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
		return errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	*d = r.User
	for j := 0; j < len(d.ExtAttrs); j++ {
		if d.ExtAttrs[j].Code == "1263467522" {
			d.JianshuAddr = d.ExtAttrs[j].Value.Text
		} else if d.ExtAttrs[j].Code == "1263534303" {
			d.BlogAddr = d.ExtAttrs[j].Value.Text
		} else if d.ExtAttrs[j].Code == "1263581295" {
			d.LeetcodeAddr = d.ExtAttrs[j].Value.Text
		}
	}
	return
}

func (d *DingUser) FindDingUsersInfo(name, mobile string, deptId int, authorityId int, p request.PageInfo, c *gin.Context) (us []DingUser, total int64, err error) {
	limit := p.PageSize
	offset := p.PageSize * (p.Page - 1)
	db := global.GLOAB_DB
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if mobile != "" {
		db = db.Where("mobile like ?", "%"+mobile+"%")
	}
	if deptId != 0 {
		total = db.Model(&DingDept{DeptId: deptId}).Association("UserList").Count()
		err = db.Limit(limit).Offset(offset).Model(&DingDept{DeptId: deptId}).Association("UserList").Find(&us)
		return
	}
	// 声明一个引用类型的变量，UserIds这个变量被分配内存了，但是其结构体底层指向数组的指针没有初始化,是nil
	// 当你声明一个切片变量并将其初始化为 var UserIds []string 时，它的零值是 nil，表示切片不引用任何底层数组。当你打印一个 nil 切片时，输出的结果是 []，这是 Go 语言的约定
	// UserIds这个变量被分配内存了，println(len(s))是0
	var UserIds []string
	if authorityId != 0 {
		err = db.Table("sys_user_authority").Where("sys_authority_authority_id = ?", authorityId).Count(&total).Error
		err = db.Limit(limit).Offset(offset).Table("sys_user_authority").Where("sys_authority_authority_id = ?", authorityId).Select("ding_user_user_id").Scan(&UserIds).Error
		err = global.GLOAB_DB.Preload("DeptList").Preload("Authorities").Find(&us, UserIds).Error
		return
	}
	if strings.Split(c.Request.URL.Path, "/")[len(strings.Split(c.Request.URL.Path, "/"))-1] == "FindDingUsersInfoBase" {
		err = db.Model(&DingUser{}).Count(&total).Error
		// Limit 和 Offset方法一定要放在最后一行查询的代码中执行，Count方法要单独起一行来绑定total
		err = db.Limit(limit).Offset(offset).Select("user_id", "name", "mobile").Find(&us).Error
	} else if strings.Split(c.Request.URL.Path, "/")[len(strings.Split(c.Request.URL.Path, "/"))-1] == "FindDingUsersInfoDetail" {
		err = db.Model(&DingUser{}).Count(&total).Error
		err = db.Limit(limit).Offset(offset).Omit("password").Preload(clause.Associations).Find(&us).Error
	}
	return
}

// 获取二维码buf，chatId, title
func (u *DingUser) GetQRCodeInWindows(c *gin.Context) (buf []byte, chatId, title string, err error) {
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
			// 不知道为何，不能直接使用 NewExecAllocator ，因此增加 使用 ws://172.17.0.7:9222/ 来调用
			c, _ := chromedp.NewRemoteAllocator(context.Background(), "ws://172.17.0.7:9222/")
			ChromeCtx, _ = chromedp.NewContext(c)
		} else {
			c, _ := chromedp.NewExecAllocator(context.Background(), allocOpts...)
			ChromeCtx, _ = chromedp.NewContext(c)
		}
	}
	return ChromeCtx
}
func (u *DingUser) GetQRCodeInLinux(c *gin.Context) (buf []byte, chatId, title string, err error) {
	timeCtx, cancel := context.WithTimeout(GetChromeCtx(false), 5*time.Minute)
	defer cancel()
	d := data{}
	var html string
	zap.L().Info("开始运行chromedp")
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
func (u *DingUser) GetRobotList(p *ParamGetRobotList) (RobotList []DingRobot, count int64, err error) {

	limit := p.PageSize
	offset := p.PageSize * (p.Page - 1)
	db := global.GLOAB_DB
	err = db.Model(&DingRobot{}).Where("ding_user_id = ?", u.UserId).Count(&count).Error
	if err != nil {
		return
	}
	if p.Name != "" {
		db = db.Where("name = ?", p.Name)
	}
	err = db.Limit(limit).Offset(offset).Model(u).Association("DingRobots").Find(&RobotList)
	return
}
