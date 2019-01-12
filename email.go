package ghreminder

import (
    "bytes"
    "net/smtp"
    "text/template"

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

    messageTemplate := "To: {{.email}}\r\nSubject: {{.subject}}\r\n\r\n{{.body}}\r\n"
    t, err := template.New("").Parse(messageTemplate)
    log.PanicIf(err)

    b := new(bytes.Buffer)

    replacements := map[string]string{
        "email":   toEmail,
        "subject": subject,
        "body":    body,
    }

    err = t.Execute(b, replacements)
    log.PanicIf(err)

    message := b.Bytes()
    toEmailList := []string{toEmail}

    err = smtp.SendMail(SmtpHostname, nil, FromEmailAddress, toEmailList, []byte(message))
    log.PanicIf(err)

    return nil
}
