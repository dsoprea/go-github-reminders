package main

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "text/template"

    "github.com/dsoprea/go-logging"
    "github.com/dsoprea/go-time-index"
    "github.com/google/go-github/github"
    "github.com/jessevdk/go-flags"
)

const (
    // TODO(dustin): !! Does "is:issue" exclude PRs?
    DefaultQuery = "involves:{{.username}} is:issue is:open"
)

type AuthenticationMixinParameters struct {
    Username string `long:"username" description:"Username" required:"true"`
    Password string `long:"password" description:"Password" required:"true"`
}

type issueRemindersParameters struct {
    *AuthenticationMixinParameters

    Queries []string `long:"query" description:"Zero or more queries used to report activity. If not provided, the default is used: 'involves:{{.username}} is:issue is:open'"`
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

func getIssueResults(issueRemindersArguments issueRemindersParameters) (ts timeindex.TimeSlice, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    gc, err := getClient(issueRemindersArguments.AuthenticationMixinParameters)
    log.PanicIf(err)

    ctx := context.Background()

    queries := issueRemindersArguments.Queries
    if queries == nil {
        queries = []string{DefaultQuery}
    }

    issues := make(map[string]struct{})

    ts = make(timeindex.TimeSlice, 0)
    for _, queryTemplate := range queries {
        // Build query.

        tmpl, err := template.New("").Parse(queryTemplate)
        log.PanicIf(err)

        replacements := map[string]string{
            "username": issueRemindersArguments.AuthenticationMixinParameters.Username,
        }

        b := new(bytes.Buffer)

        err = tmpl.Execute(b, replacements)
        log.PanicIf(err)

        query := b.String()

        // Execute.

        searchOptions := new(github.SearchOptions)
        for {
            isr, response, err := gc.Search.Issues(ctx, query, searchOptions)
            log.PanicIf(err)

            for _, issue := range isr.Issues {
                if _, found := issues[*issue.URL]; found == true {
                    continue
                }

                issues[*issue.URL] = struct{}{}
                ts = ts.Add(*issue.UpdatedAt, issue)
            }

            if response.NextPage == 0 {
                break
            }

            searchOptions.Page = response.NextPage
        }
    }

    // TODO(dustin): !! We still need to filter by:
    // - restricting only to issues that we have responded to in the last six months but have not recently responded to
    // - filtering-out any issues we're no longer following

    // TODO(dustin): !! We might prefer multiple queries where we're the owner or were mentioned or are following or have posted, where each might get a timestamp, but a different one: updated:>=2013-02-01

    return ts, nil
}

type DistillableRepositoryUrl string

const (
    RepositoryUrlToNameStrippablePrefix = "https://api.github.com/repos/"
)

func (r DistillableRepositoryUrl) Name() string {
    url := string(r)

    len_ := len(RepositoryUrlToNameStrippablePrefix)
    if url[:len_] == RepositoryUrlToNameStrippablePrefix {
        return url[len_:]
    }

    return url
}

func handleIssueReminders(issueRemindersArguments issueRemindersParameters) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ts, err := getIssueResults(issueRemindersArguments)
    log.PanicIf(err)

    for i := len(ts) - 1; i >= 0; i-- {
        ti := ts[i]
        for _, item := range ti.Items {
            issue := item.(github.Issue)

            repositoryName := DistillableRepositoryUrl(*issue.RepositoryURL).Name()
            fmt.Printf("%s %s %s %s %s\n", *issue.UpdatedAt, *issue.URL, repositoryName, *issue.User.Login, *issue.Title)
        }
    }

    return nil
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
