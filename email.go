package ghreminder

import (
    "bytes"
    "fmt"
    "net/http"
    "net/smtp"

    "github.com/dsoprea/go-logging"
)

const (
    SmtpHostname     = "localhost:25"
    FromEmailAddress = "github-notifications@local"
)

func SendEmailToLocal(toEmail, subject, body string) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    headers := make(http.Header)
    headers.Add("To", toEmail)
    headers.Add("Subject", subject)
    headers.Add("Content-Type", "text/html")

    b := new(bytes.Buffer)

    err = headers.Write(b)
    log.PanicIf(err)

    _, err = fmt.Fprintf(b, "\r\n")
    log.PanicIf(err)

    _, err = fmt.Fprintf(b, body)
    log.PanicIf(err)

    message := b.Bytes()

    toEmailList := []string{toEmail}

    err = smtp.SendMail(SmtpHostname, nil, FromEmailAddress, toEmailList, []byte(message))
    log.PanicIf(err)

    return nil
}
