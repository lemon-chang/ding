package utils

import (
	"context"
	"ding/global"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"
)

func TypingInviation() (TypingInvitationCode string,err error)  {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", false),// 禁用chrome headless（禁用无窗口模式，那就是开启窗口模式）
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
	var html string
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)

	err = chromedp.Run(ctx,
		chromedp.Navigate("https://dazi.kukuw.com/"),
		//点击“我的打字“按钮
		chromedp.Click(`document.getElementById("globallink").getElementsByTagName("a")[5]`, chromedp.ByJSPath),
		// 锁定用户名框并填写内容
		chromedp.WaitVisible(`document.querySelector("#name")`,chromedp.ByJSPath),
		chromedp.SetValue(`document.querySelector("#name")`, "闫佳鹏", chromedp.ByJSPath),
		//锁定密码框并填写内容
		chromedp.WaitVisible(`document.querySelector("#pass")`,chromedp.ByJSPath),
		chromedp.SetValue(`document.querySelector("#pass")`, "123456", chromedp.ByJSPath),
		//点击登录按钮
		chromedp.WaitVisible(`document.querySelector(".button").firstElementChild`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector(".button").firstElementChild`, chromedp.ByJSPath),
		//点击发布竞赛
		chromedp.WaitVisible(`document.querySelector("a.groupnew")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a.groupnew")`, chromedp.ByJSPath),
		chromedp.Sleep(time.Second),
		////点击所要打字的文章
		chromedp.WaitVisible(`document.querySelector("a#select_b.select_b")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a#select_b.select_b")`,chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelector("a.sys.on")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a.sys.on")`,chromedp.ByJSPath),
		//选择有效期
		chromedp.Evaluate("document.querySelector(\"select#youxiaoqi\").value = document.querySelector(\"select#youxiaoqi\").children[5].value",nil),
		//设置成为不公开
		chromedp.Click(`document.querySelectorAll("input#gongkai")[1]`,chromedp.ByJSPath),
		//点击发布按钮
		chromedp.Click(`document.querySelectorAll(".artnew table tr td input")[7]`,chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelectorAll("#my_main .art_table td")[9].childNodes[0]`,chromedp.ByJSPath),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("打字吗出现了")
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("爬取前:",TypingInvitationCode)
			a := chromedp.OuterHTML(`document.querySelector("body")`, &html, chromedp.ByJSPath)
			err := a.Do(ctx)
			if err != nil {
				zap.L().Error("chromedp获取页面全部数据失败",zap.Error(err))
				return err
			}
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				zap.L().Error("chromedp获取页面全部数据后，转化成dom失败",zap.Error(err))
				return err
			}
			//dom.Find(`#my_main > .art_table > tbody > tr:nth-child(2) > td:nth-child(4) > span`).Each(func(i int, selection *goquery.Selection) {
			//	TypingInvitationCode = TypingInvitationCode + selection.Text()
			//	fmt.Println("爬取后:",TypingInvitationCode)
			//	selection.Next()
			//})
			TypingInvitationCode = dom.Find(`#my_main > .art_table > tbody > tr:nth-child(2) > td:nth-child(4) > span`).First().Text()
			if TypingInvitationCode == "" {
				zap.L().Error("爬取打字邀请码失败")
				return err
			}
			_, err = global.GLOBAL_REDIS.Set(context.Background(), ConstTypingInvitationCode, TypingInvitationCode, time.Second*60*60*11).Result()//11小时过期时间
			if err != nil {
				zap.L().Error("爬取打字邀请码后存入redis失败",zap.Error(err))
			}
			return err
		}),
	)

	if err != nil {
		zap.L().Error("chromedp.Run有误",zap.Error(err))
		return "", err
	}else {
		zap.L().Info("chromedp.Run无误",zap.Error(err))
		return TypingInvitationCode,err
	}

}
