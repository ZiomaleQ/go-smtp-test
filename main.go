package main

import (
	"encoding/base64"
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
		return fmt.Errorf("error reading message: %v", err)
	}

	// Extract headers
	headers := msg.Header
	from := headers.Get("From")
	to := headers.Get("To")
	subject := headers.Get("Subject")

	fmt.Printf("Got new email: %s -> %s\n", from, to)

	if strings.HasPrefix(subject, "=?") && strings.HasSuffix(subject, "?=") {
		rawEncoding := strings.Split(subject, "?")[3]

		decoded := base64.NewDecoder(base64.StdEncoding, strings.NewReader(rawEncoding))
		decodedBytes, err := io.ReadAll(decoded)

		if err != nil {
			return fmt.Errorf("error decoding base64: %v", err)
		}

		subject = string(decodedBytes)
	}

	fmt.Printf("Subject: %s\n", subject)

	encoding := headers.Get("Content-Transfer-Encoding")

	raw, err := io.ReadAll(msg.Body)

	if err != nil {
		return fmt.Errorf("error reading message body: %v", err)
	}

	body := string(raw)

	if encoding == "base64" {
		decoded := base64.NewDecoder(base64.StdEncoding, strings.NewReader(body))
		decodedBytes, err := io.ReadAll(decoded)

		if err != nil {
			return fmt.Errorf("error decoding base64: %v", err)
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

	s.Addr = "127.0.0.1:1025"
	s.Domain = "localhost"
	s.AllowInsecureAuth = false

	log.Println("Starting SMTP server at", "127.0.0.1:1025")
	log.Fatal(s.ListenAndServe())
}

func main() {
	go newServer()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
