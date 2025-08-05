package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)


func SendVerificationEmail(toEmail, token string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	sender := os.Getenv("EMAIL_SENDER")

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	subject := "Subject: SocialSync Email Verification Code\r\n"
	from := fmt.Sprintf("From: SocialSync <%s>\r\n", sender)
	body := fmt.Sprintf("Your verification code is: %s\r\n\r\nIf you did not request this, please ignore.\r\n", token)
	msg := []byte(from + subject + "\r\n" + body)


	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{toEmail}, msg)
	if err != nil {
		log.Printf("Error sending email to %s: %v", toEmail, err)
		return err
	}
	return nil
}
