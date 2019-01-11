package ghreminder

import (
    "context"
    "strings"
    "time"

    "github.com/dsoprea/go-logging"
    "github.com/google/go-github/github"
)

// getIssues returns a list of recently-updated, open issues that we're
// subscribed to.
func GetIssues(ctx context.Context, gc *github.Client, searchIntervalDuration time.Duration) (issues []*github.Issue, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    searchIntervalTimestamp := time.Now().Add(searchIntervalDuration)

    ilo := &github.IssueListOptions{
        Filter:    "subscribed",
        State:     "open",
        Sort:      "updated",
        Direction: "desc",
        Since:     searchIntervalTimestamp,
    }

    issues = make([]*github.Issue, 0)
    for {
        issuesThis, response, err := gc.Issues.List(ctx, true, ilo)
        log.PanicIf(err)

        issues = append(issues, issuesThis...)

        if response.NextPage == 0 {
            break
        }

        ilo.Page = response.NextPage
    }

    return issues, nil
}

// hasRecentlyPosted returns whether the current user has posted to the given
// issue very recently.
func HasRecentlyPosted(ctx context.Context, gc *github.Client, username string, nearIntervalDuration time.Duration, issue *github.Issue) (recentlyPosted bool, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    r := DistillableRepositoryUrl(*issue.RepositoryURL)
    owner, repository := r.OwnerAndRepository()

    nearIntervalTimestamp := time.Now().Add(nearIntervalDuration)

    ilco := &github.IssueListCommentsOptions{
        Sort:      "created",
        Direction: "desc",
        Since:     nearIntervalTimestamp,
    }

    for {
        commentsThis, response, err := gc.Issues.ListComments(ctx, owner, repository, *issue.Number, ilco)
        log.PanicIf(err)

        for _, comment := range commentsThis {
            if comment.UpdatedAt.After(nearIntervalTimestamp) == true {
                break
            } else if *comment.User.Login == username {
                return true, nil
            }
        }

        if response.NextPage == 0 {
            break
        }

        ilco.Page = response.NextPage
    }

    return false, nil
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

func (r DistillableRepositoryUrl) OwnerAndRepository() (owner string, repository string) {
    parts := strings.Split(r.Name(), "/")

    owner, repository = parts[0], parts[1]
    return owner, repository
}
