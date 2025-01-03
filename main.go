package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/mail"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/emersion/go-smtp"
)

type backend struct{}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{}, nil
}

type session struct{}

func (s *session) AuthPlain(username, password string) error {
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	return nil
}

func (s *session) Data(r io.Reader) error {
	msg, err := mail.ReadMessage(r)

	if err != nil {
		return errors.Join(errors.New("error reading message"), err)
	}

	fmt.Printf("Got new email: %s -> %s\n", msg.Header.Get("From"), msg.Header.Get("To"))

	subject := msg.Header.Get("Subject")

	if strings.HasPrefix(subject, "=?") && strings.HasSuffix(subject, "?=") {
		rawEncoding := strings.Split(subject, "?")[3]

		decoded := base64.NewDecoder(base64.StdEncoding, strings.NewReader(rawEncoding))
		decodedBytes, err := io.ReadAll(decoded)

		if err != nil {
			return errors.Join(errors.New("error decoding base64"), err)
		}

		subject = string(decodedBytes)
	}

	fmt.Printf("Subject: %s\n", subject)

	encoding := msg.Header.Get("Content-Transfer-Encoding")

	raw, err := io.ReadAll(msg.Body)

	if err != nil {
		return errors.Join(errors.New("error reading message body"), err)
	}

	body := string(raw)

	if encoding == "base64" {
		decoded := base64.NewDecoder(base64.StdEncoding, strings.NewReader(body))
		decodedBytes, err := io.ReadAll(decoded)

		if err != nil {
			return errors.Join(errors.New("error decoding base64"), err)
		}

		body = string(decodedBytes)
	}

	fmt.Printf("Body: %s\n", body)

	return nil
}

func (s *session) Reset() {}

func (s *session) Logout() error {
	return nil
}

func newServer() {
	s := smtp.NewServer(&backend{})

	s.Addr = addr
	s.AllowInsecureAuth = false

	log.Println("Starting SMTP server at", addr)
	log.Fatal(s.ListenAndServe())
}

var addr string

func main() {
	flag.StringVar(&addr, "addr", ":1025", "SMTP server address")

	flag.Parse()

	go newServer()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
