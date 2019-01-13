package ghreminder

import (
    "bytes"
    "fmt"
    "net/http"
    "net/smtp"

    "github.com/dsoprea/go-logging"
    "github.com/google/go-github/github"
    "github.com/olekukonko/tablewriter"
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

func GetTextEmail(issues []*github.Issue) (textContent string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    b := new(bytes.Buffer)

    table := tablewriter.NewWriter(b)
    table.SetHeader([]string{"Updated At", "URL", "Repository", "User", "Title"})
    table.SetColWidth(50)

    for _, issue := range issues {
        repositoryName := DistillableRepositoryUrl(*issue.RepositoryURL).Name()

        row := []string{
            issue.UpdatedAt.String(),
            *issue.HTMLURL,
            repositoryName,
            *issue.User.Login,
            *issue.Title,
        }

        table.Append(row)
    }

    table.Render()

    return b.String(), nil
}

func GetHtmlEmail(issues []*github.Issue) (htmlContent string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    b := new(bytes.Buffer)

    _, err = fmt.Fprintf(b, "<table>\n")
    log.PanicIf(err)

    _, err = fmt.Fprintf(b, "<tr><th align=\"left\">Updated At</th><th align=\"left\">URL</th><th align=\"left\">Repository</th><th align=\"left\">User</th><th align=\"left\">Title</th></tr>\n")
    log.PanicIf(err)

    for _, issue := range issues {
        repositoryName := DistillableRepositoryUrl(*issue.RepositoryURL).Name()

        _, err := fmt.Fprintf(b, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
            issue.UpdatedAt.String(),
            *issue.HTMLURL,
            repositoryName,
            *issue.User.Login,
            *issue.Title,
        )
        log.PanicIf(err)
    }

    _, err = fmt.Fprintf(b, "</table>\n")
    log.PanicIf(err)

    return b.String(), nil
}
