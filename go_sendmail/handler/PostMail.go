package handler

import (
	"os"
	"fmt"
	"log"
	"sync"
	"strings"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail" // 上記でダウンロードされてるから二重にダウンロードされてない？
	"github.com/ant0ine/go-json-rest/rest"
)

//メールの構造体
type Mail struct {
	Name    string
	Subject string
	Text    string
	Email   string
}

//送信する関数
func PostMail(w rest.ResponseWriter, r *rest.Request) {
	store := map[string]*Mail{}

	// トクメモ管理者メールに送信する内容
	sendMailContents := Mail{}
	// リクエストをMailの構造体形式にパース
	err := r.DecodeJsonPayload(&sendMailContents)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError) // サーバー応答エラー
		return
	}

	// 送信先のメールアドレスが空白の場合
	if sendMailContents.Email == "" {
			rest.Error(w, "mail required", 400) // クライアントのリクエストエラー
			return
	}

	// store[sendMailContents.Name]の書き込みが終了するまでロックする
	lock := sync.RWMutex{}
	lock.Lock()
	store[sendMailContents.Name] = &sendMailContents
	lock.Unlock()


	mailContents := writeEmail(sendMailContents)

	sendEmail(mailContents)

	w.WriteJson(&sendMailContents)
}

// private
func loadEnv(key string) string {
	// .envから環境変数読み込み
	err_read := godotenv.Load()
	if err_read != nil {
		log.Fatalf("error: %v", err_read)
	}
	return os.Getenv(key)
}


func writeEmail(sendMailContents Mail) SGMailV3 {
	// メッセージの構築
	mailContentsMessage := mail.NewV3Mail()

	// 送信元を設定
	mailContentsFrom := mail.NewEmail("", sendMailContents.Email)
	mailContentsMessage.SetFrom(mailContentsFrom)

	// 生存期間が長すぎるから区切る
	{
		// 宛先と対応するSubstitutionタグを指定(宛先は複数指定可能)
		p := mail.NewPersonalization()
		TOS := strings.Split(loadEnv("TOS"), ",")
		to := mail.NewEmail("", TOS[0])
		p.AddTos(to)

		//残りのpostされた値を変数に格納
		name := sendMailContents.Name
		subject := sendMailContents.Subject
		text := sendMailContents.Text

		p.SetSubstitution("%name%", name)
		p.SetSubstitution("%m_subject%", subject)
		p.SetSubstitution("%m_text%", text)
		mailContentsMessage.AddPersonalizations(p)
	}

	// 件名を設定
	mailContentsMessage.Subject = "%m_subject%"
	// テキストパートを設定
	{
		c := mail.NewContent("text/plain", "%m_text%\r\n")
		mailContentsMessage.AddContent(c)
	}
	// HTMLパートを設定
	//c = mail.NewContent("text/html", "<strong> %name% さんは何をしていますか？</strong><br>　文章ー－－。")
	// mailContentsMessage.AddContent(c)

	// カテゴリ情報を付加
	mailContentsMessage.AddCategories("category1")
	// カスタムヘッダを指定
	mailContentsMessage.SetHeader("X-Sent-Using", "SendGrid-API")

	return mailContentsMessage
}


// https://github.com/sendgrid/sendgrid-go/blob/1101132fabbaac513f12beedfdd4bc32ec22ec97/helpers/mail/mail_v3.go#L22
func sendEmail(mailContentsMessage SGMailV3) {
	// メール送信を行い、レスポンスを表示
	API_KEY := loadEnv("API_KEY")
	client := sendgrid.NewSendClient(API_KEY)
	response, err := client.Send(mailContentsMessage)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
