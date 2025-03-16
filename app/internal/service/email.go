package service

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

var (
	SMTPEmail  = flag.String("SMTPEmail", "", "SMTP email address")
	SMTPPass   = flag.String("SMTPPass", "", "SMTP password")
	SMTPServer = flag.String("SMTPServer", "", "SMTP server address")
	SMTPPort   = flag.String("SMTPPort", "", "SMTP server port")
)

func sendVerificationEmail(toEmail, code, subject, bodyMessage string) error {
	from := *SMTPEmail
	to := []string{toEmail}
	body := fmt.Sprintf("Your code: %s\n\n%s", code, bodyMessage)

	message := []byte("From: " + from + "\r\n" +
		"To: " + strings.Join(to, ",") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", from, *SMTPPass, *SMTPServer)

	err := smtp.SendMail(*SMTPServer+":"+*SMTPPort, auth, from, to, message)
	if err != nil {
		log.Println("Ошибка отправки email:", err)
		return err
	}
	return nil
}
