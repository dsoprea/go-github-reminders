package ghreminder

import (
    "fmt"

    "github.com/google/go-github/github"
)

func DumpIssue(issue *github.Issue) {
    fmt.Printf("Updated at: %s\n", issue.UpdatedAt.String())
    fmt.Printf("Title: %s\n", *issue.Title)
    fmt.Printf("Number: %d\n", *issue.Number)
    fmt.Printf("UI URL: %s\n", *issue.HTMLURL)
    fmt.Printf("API URL: %s\n", *issue.URL)

    fmt.Printf("\n")
}
