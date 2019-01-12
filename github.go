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
// issue very recently. This is used to determine which of the issues we have
// been involved in should be ignored for now.
func HasVeryRecentlyPosted(ctx context.Context, gc *github.Client, username string, nearIntervalDuration time.Duration, issue *github.Issue) (recentlyPosted bool, err error) {
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

    commentsThis, _, err := gc.Issues.ListComments(ctx, owner, repository, *issue.Number, ilco)
    log.PanicIf(err)

    // If there were any comments recently.
    if len(commentsThis) > 0 {
        comment := commentsThis[0]
        postedByCurrentUser := (*comment.User.Login == username)

        // If the latest comment returned wasn't posted by us, return false.
        // We should respond.
        if postedByCurrentUser == false {
            return false, nil
        }

        // If the latest comment was posted by us, return true. No response is
        // currently required.
        if postedByCurrentUser == true {
            return true, nil
        }
    }

    // We haven't posted any very-recent messages.
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
