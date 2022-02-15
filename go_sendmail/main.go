package main

import (
	"go_sendmail/handler"
	// "fmt"
	"log"
	// "os"
	// "strings"
	// "io/ioutil"
	// "encoding/base64"

	// "github.com/sendgrid/sendgrid-go"
	// "github.com/sendgrid/sendgrid-go/helpers/mail"
	// "github.com/joho/godotenv"

	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	// "sync"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/post", handler.PostMail),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}
