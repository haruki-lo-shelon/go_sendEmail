package handler

import (
	"fmt"
	"log"
	"os"
	"strings"
	// "io/ioutil"
	// "encoding/base64"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/joho/godotenv"

	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	"sync"
)

type Mail struct {
	Name    string
	Subject string
	Text    string
	Email   string
}

var store = map[string]*Mail{}

var lock = sync.RWMutex{}

func PostMail(w rest.ResponseWriter, r *rest.Request) {
	send := Mail{}
	err := r.DecodeJsonPayload(&send)
	if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	if send.Email == "" {
			rest.Error(w, "mail required", 400)
			return
	}
	lock.Lock()
	store[send.Name] = &send
	lock.Unlock()

	err_read := godotenv.Load()
	if err_read != nil {
		log.Fatalf("error: %v", err_read)
	}

	// .envから環境変数読み込み
	API_KEY := os.Getenv("API_KEY")
	TOS := strings.Split(os.Getenv("TOS"), ",")
	fr := send.Email
	//FROM := os.Getenv(send.Email)

	// メッセージの構築
	message := mail.NewV3Mail()
	// 送信元を設定
	from := mail.NewEmail("", fr)
	message.SetFrom(from)

	// 1つ目の宛先と、対応するSubstitutionタグを指定
	p := mail.NewPersonalization()
	to := mail.NewEmail("", TOS[0])
	p.AddTos(to)
	name := send.Name
	subject := send.Subject
	text := send.Text

	p.SetSubstitution("%name%", name)
	p.SetSubstitution("%m_subject%", subject)
	p.SetSubstitution("%m_text%", text)
	message.AddPersonalizations(p)

	// 2つ目の宛先と、対応するSubstitutionタグを指定
	// p2 := mail.NewPersonalization()
	// to2 := mail.NewEmail("", TOS[1])
	// p2.AddTos(to2)
	// p2.SetSubstitution("%fullname%", "佐藤 次郎")
	// p2.SetSubstitution("%familyname%", "佐藤")
	// p2.SetSubstitution("%place%", "目黒")
	// message.AddPersonalizations(p2)

	// 3つ目の宛先と、対応するSubstitutionタグを指定
	// p3 := mail.NewPersonalization()
	// to3 := mail.NewEmail("", TOS[2])
	// p3.AddTos(to3)
	// p3.SetSubstitution("%fullname%", "鈴木 三郎")
	// p3.SetSubstitution("%familyname%", "鈴木")
	// p3.SetSubstitution("%place%", "中野")
	// message.AddPersonalizations(p3)

	// 件名を設定
	message.Subject = "%m_subject%"
	// テキストパートを設定
	c := mail.NewContent("text/plain", "%m_text%\r\n")
	message.AddContent(c)
	// HTMLパートを設定
	//c = mail.NewContent("text/html", "<strong> %familyname% さんは何をしていますか？</strong><br>彼は%place%にいます。")
	// message.AddContent(c)

	// カテゴリ情報を付加
	message.AddCategories("category1")
	// カスタムヘッダを指定
	message.SetHeader("X-Sent-Using", "SendGrid-API")
	// 画像ファイルを添付
	// a := mail.NewAttachment()
	// file, _ := os.OpenFile("./gif.gif", os.O_RDONLY, 0600)
	// defer file.Close()
	// data, _ := ioutil.ReadAll(file)
	// data_enc := base64.StdEncoding.EncodeToString(data)
	// a.SetContent(data_enc)
	// a.SetType("image/gif")
	// a.SetFilename("owl.gif")
	// a.SetDisposition("attachment")
	// message.AddAttachment(a)

	// メール送信を行い、レスポンスを表示
	client := sendgrid.NewSendClient(API_KEY)
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	w.WriteJson(&send)
}
