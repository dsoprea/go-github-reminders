package main

import (
    "bytes"
    "context"
    "fmt"
    "os"

    "github.com/dsoprea/go-github-reminders"
    "github.com/dsoprea/go-logging"
    "github.com/dsoprea/go-time-parse"
    "github.com/google/go-github/github"
    "github.com/jessevdk/go-flags"
    "github.com/olekukonko/tablewriter"
)

const (
    EmailSubject = "Some Github issues need attention"
)

type AuthenticationMixinParameters struct {
    Username string `long:"username" description:"Username" required:"true"`
    Password string `long:"password" description:"Password" required:"true"`
}

type issueRemindersParameters struct {
    *AuthenticationMixinParameters

    SearchIntervalPhrase string `long:"search-distance" description:"Time range to look for activity" default:"6 months ago"`
    NearIntervalPhrase   string `long:"near-distance" description:"Time range to consider updates too recent to remind" default:"3 days ago"`
    EmailAddress         string `long:"email" description:"Send to the given email using an unauthenticated local SMTP server. In most cases it might be preferable to just capture the output and send the email using a more-advanced configuration elsewhere."`
    PrintAsHtml          bool   `long:"html" description:"Print HTML-formatted table"`
}

type subcommands struct {
    IssueReminders issueRemindersParameters `command:"issue-reminders" description:"Issue reminders"`
}

var (
    rootArguments = new(subcommands)
)

func getClient(authenticationArguments *AuthenticationMixinParameters) (gc *github.Client, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    bat := &github.BasicAuthTransport{
        Username: authenticationArguments.Username,
        Password: authenticationArguments.Password,
    }

    bc := bat.Client()
    gc = github.NewClient(bc)

    return gc, nil
}

func getIssues(issueRemindersArguments issueRemindersParameters) (issues []*github.Issue, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    searchIntervalDuration, phraseType, err := timeparse.ParseDuration(issueRemindersArguments.SearchIntervalPhrase)
    log.PanicIf(err)

    if phraseType != timeparse.PhraseTypeTime {
        log.Panicf("please use a 'search' interval time-phrase that describes an interval: [%s]", issueRemindersArguments.SearchIntervalPhrase)
    } else if searchIntervalDuration >= 0 {
        log.Panicf("Please provide a 'search' interval in the past: [%s]", issueRemindersArguments.SearchIntervalPhrase)
    }

    nearIntervalDuration, phraseType, err := timeparse.ParseDuration(issueRemindersArguments.NearIntervalPhrase)
    log.PanicIf(err)

    if phraseType != timeparse.PhraseTypeTime {
        log.Panicf("please use a 'near' interval time-phrase that describes an interval: [%s]", issueRemindersArguments.NearIntervalPhrase)
    } else if nearIntervalDuration >= 0 {
        log.Panicf("Please use a 'near' interval in the past: [%s]", issueRemindersArguments.NearIntervalPhrase)
    }

    ctx := context.Background()

    gc, err := getClient(issueRemindersArguments.AuthenticationMixinParameters)
    log.PanicIf(err)

    issues, err = ghreminder.GetIssues(ctx, gc, searchIntervalDuration)
    log.PanicIf(err)

    filtered := make([]*github.Issue, 0)
    for _, issue := range issues {
        hasRecentlyUpdated, err := ghreminder.HasVeryRecentlyPosted(ctx, gc, issueRemindersArguments.Username, nearIntervalDuration, issue)
        log.PanicIf(err)

        if hasRecentlyUpdated == true {
            continue
        }

        filtered = append(filtered, issue)
    }

    return filtered, nil
}

func handleIssueReminders(issueRemindersArguments issueRemindersParameters) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    issues, err := getIssues(issueRemindersArguments)
    log.PanicIf(err)

    if len(issues) == 0 {
        fmt.Printf("You currently do not have or need any reminders.\n")
        return nil
    }

    var content string
    if issueRemindersArguments.PrintAsHtml == true {
        content, err = GetHtmlEmail(issues)
        log.PanicIf(err)
    } else {
        content, err = GetTextEmail(issues)
        log.PanicIf(err)
    }

    fmt.Printf(content)

    if issueRemindersArguments.EmailAddress != "" {
        err = ghreminder.SendEmailToLocal(issueRemindersArguments.EmailAddress, EmailSubject, content)
        log.PanicIf(err)
    }

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
        repositoryName := ghreminder.DistillableRepositoryUrl(*issue.RepositoryURL).Name()

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

    _, err = fmt.Fprintf(b, "<tr><th>Updated At</th><th>URL</th><th>Repository</th><th>User</th><th>Title</th></tr>\n")
    log.PanicIf(err)

    for _, issue := range issues {
        repositoryName := ghreminder.DistillableRepositoryUrl(*issue.RepositoryURL).Name()

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

func main() {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintError(err)
            os.Exit(-1)
        }
    }()

    p := flags.NewParser(rootArguments, flags.Default)

    _, err := p.Parse()
    if err != nil {
        os.Exit(1)
    }

    switch p.Active.Name {
    case "issue-reminders":
        err := handleIssueReminders(rootArguments.IssueReminders)
        log.PanicIf(err)
    default:
        fmt.Printf("Subcommand not handled: [%s]\n", p.Active.Name)
        os.Exit(2)
    }
}
