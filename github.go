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
func GetIssues(ctx context.Context, gc *github.Client, searchIntervalDuration time.Duration, justAssigned bool) (issues []*github.Issue, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    searchIntervalTimestamp := time.Now().Add(searchIntervalDuration)

    // filter is the criteria used to filter issues. "subscribed" is ideal in
    // concept but doesn't work. Some events that are definitely subscribed-to
    // by the current user with a very recent updated-time and a very old
    // created-time still won't show up.
    var filter string

    if justAssigned == true {
        filter = "assigned"
    } else {
        filter = "all"
    }

    ilo := &github.IssueListOptions{
        Filter:    filter,
        State:     "open",
        Sort:      "updated",
        Direction: "desc",
    }

    // TODO(dustin): !! Add a command-line parameter that restricts the issues we see to only those assigned to us.

    issues = make([]*github.Issue, 0)

ProcessIssues:
    for {
        issuesThis, response, err := gc.Issues.List(ctx, true, ilo)
        log.PanicIf(err)

        for _, issue := range issuesThis {
            // We need to manage when to stop based on the "updated" timestamp
            // because the "since" parameter in the query only applies to the
            // "create" timestamp.
            if issue.UpdatedAt.Before(searchIntervalTimestamp) == true {
                break ProcessIssues
            }

            issues = append(issues, issue)
        }

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

    nowTime := time.Now()
    nearIntervalTimestamp := nowTime.Add(nearIntervalDuration)

    sortBy := "updated"
    sortDirection := "desc"

    ilco := &github.IssueListCommentsOptions{
        Sort:      sortBy,
        Direction: sortDirection,
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

        // If we get here, the latest comment is by us.

        // It's been long-enough that it's time to follow-up.
        if nearIntervalTimestamp.Before(nowTime) == true {
            return false, nil
        }

        // We've very-recently responded.
        return true, nil
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
