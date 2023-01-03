package smtp_test

import (
	// "net/smtp"
	"fmt"
	"testing"

	"github.com/suiguo/hwlib/smtp"
)

var testMail = `<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="utf-8">
    <meta name="theme-color" content="#000000">
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="browsermode" content="application">
    <meta name="full-screen" content="yes" />
    <meta name="x5-fullscreen" content="true" />
    <meta name="x5-page-mode" content="app" />
    <meta name="360-fullscreen" content="true" />
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <meta http-equiv="x-dns-prefetch-control" content="on" />
    <meta name="viewport"
        content="width=320.1,initial-scale=1,minimum-scale=1,maximum-scale=1,user-scalable=no,minimal-ui" />
    <meta name="apple-mobile-web-app-title" content="yeziyuan" />
    <meta content="telephone=no" name="format-detection" />
    <meta name="fullscreen" content="yes">
    <title>HuiOne</title>
    <style>
        * {
            margin: 0;
            padding: 0;
        }

        #root {
            display: flex;
            flex-direction: column;
            flex: 1;
        }

        .head {
            width: 100%;
            height: 4px;
            background-color: #F5222D;
        }

        .log_div {
            margin-top: 50px;
            margin-left: 50px;
            display: flex;
            flex-direction: row;
        }

        .log_img {
            width: 121px;
            height: 38px;
        }
        .page_title {
            margin-top: 36px;
            margin-left: 50px;
            color: #000000;
            font-size: 30px;
            font-weight: 500;
        }
        .welcome_text {
            margin-top: 16px;
            margin-left: 50px;
            color: #000000;
            font-size: 14px;
            font-weight: 400;
        }
        .verification_code_text {
            margin-top: 36px;
            margin-left: 50px;
            color: #000000;
            font-size: 14px;
            font-weight: 400;
        }
        .verification_code_num {
            margin-top: 8px;
            margin-left: 50px;
            color: #F5222D;
            font-size: 30px;
            font-weight: 500;
        }
        .explain_text {
            display: flex;
            flex-direction: row;
            margin-top: 16px;
            margin-left: 50px;
            color: #000000;
            font-size: 14px;
            font-weight: 400;
        }
        .explain_text .explain_text_red {
            color: #F5222D;
        }
        .statement_text {
            margin-top: 64px;
            margin-left: 50px;
            color: #707A8A;
            font-size: 14px;
            font-weight: 400;
        }
    </style>
</head>

<body>
    <div id="root">
        <div class="head" />
        <div class="log_div">
            <img class="log_img" src='https://hwwallet.s3.ap-northeast-1.amazonaws.com/v1/mail/favicon.png' />
        </div>
        <div class="page_title">账户注册</div>
        <div class="welcome_text">欢迎来到HuiOne机构，使用以下验证码确认注册。</div>
        <div class="verification_code_text">您的验证码</div>
        <div class="verification_code_num">#code</div>
        <div class="explain_text">验证码的有效期为30分钟。请不要与任何人分享此代码。</div>
        <div class="explain_text">如非您本人操作，请立即<div class="explain_text_red">联系客服</div>。</div>
        <div class="statement_text">本邮件由系统自动发送，请勿回复。</div>
    </div>
</body>
</html>`

func TestXxx(t *testing.T) {
	cli := smtp.GetClient("smtp.gmail.com", 465, "suiguo3564@gmail.com", "hzdradxhxswufoxs")
	// code := "1380534"
	err := cli.SendMail(
		smtp.WithFrom("suiguo3564@gmail.com"),
		smtp.WithTo("s_nikki@qq.com"),
		smtp.WithTitle("测试邮件"),
		smtp.WithBodyReg("123456"),
	)
	fmt.Println(err)
	err = cli.SendMail(
		smtp.WithFrom("suiguo3564@gmail.com"),
		smtp.WithTo("s_nikki@qq.com"),
		smtp.WithTitle("测试邮件"),
		smtp.WithBodyReg("654321"),
	)
	fmt.Println(err)
}
