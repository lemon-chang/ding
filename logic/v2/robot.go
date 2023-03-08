package v2

import (
	"context"
	"ding/global"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math"
	"strings"
	"time"
)

func GetChatIdAndTitle(c *gin.Context,Crop string) (chatId,title string ,err error) {
	d :=  data{}
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
	//defer cancel()

	// 创建超时上下文
	ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	//defer cancel()

	// navigate to a page, wait for an element, click

	// capture screenshot of an element
	var buf []byte
	// capture entire browser viewport, returning png with quality=90
	var html string

	if err := chromedp.Run(ctx,
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
			if err := ioutil.WriteFile(fmt.Sprintf("./Screenshot_%s.png",username), buf, 0644); err != nil {
				print(err)
			}
			log.Println("图片写入完成")

			if err != nil {
				return err
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
			data = strings.ReplaceAll(data,"\n","")
			reader := strings.NewReader(data)
			bytearr, err := ioutil.ReadAll(reader)

			err1 := json.Unmarshal(bytearr, &d)
			if err1 != nil {

			}
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	}
	if &d == nil {
		return "","",err
	}
	return d.Result.ChatId,d.Result.Title,err
}
func GetImage(c *gin.Context){ //显示图片的方法
	imageName := c.Query("imageName") //截取get请求参数，也就是图片的路径，可是使用绝对路径，也可使用相对路径
	file, _ := ioutil.ReadFile(imageName) //把要显示的图片读取到变量中
	c.Writer.WriteString(string(file)) //关键一步，写给前端
}
type result struct {

	ChatId string `json:"chatId"`
	Title string `json:"title"`
}
type data struct {
	Result result `json:"result"`
}