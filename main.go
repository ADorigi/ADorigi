package main

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"
)

type RepoInfo struct {
	Name    string
	HTMLURL string
	Updated time.Time
	Owner   string
	Date    string
}

type Data struct {
	ActiveRepos []RepoInfo
	Date        string
}

func main() {
	// Set up GitHub API client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	repos, _, err := client.Repositories.ListByAuthenticatedUser(ctx, &github.RepositoryListByAuthenticatedUserOptions{
		Sort: "updated",
		Type: "public",
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	})

	if err != nil {
		fmt.Println("Error fetching repositories:", err)
		return
	}

	var activeRepos []RepoInfo

	for _, repo := range repos {

		repository, _, _ := client.Repositories.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName())
		// fmt.Println(repository.GetParent())
		var repoInfo RepoInfo

		if repository.GetParent() == nil {

			repoInfo = RepoInfo{
				Name:    repo.GetName(),
				HTMLURL: repo.GetHTMLURL(),
				Updated: repo.GetUpdatedAt().Time,
				Owner:   repo.GetOwner().GetLogin(),
			}
		} else {
			repoInfo = RepoInfo{
				Name:    repo.GetName(),
				HTMLURL: repo.GetHTMLURL(),
				Updated: repo.GetUpdatedAt().Time,
				Owner:   repository.GetParent().GetOwner().GetLogin(),
			}

		}
		activeRepos = append(activeRepos, repoInfo)
		if len(activeRepos) >= 10 {
			break
		}
	}

	data := Data{
		ActiveRepos: activeRepos[:min(len(activeRepos), 10)],
		Date:        time.Now().Format(time.DateOnly),
	}

	// Read the template file
	templateBytes, err := os.ReadFile("template.md")
	if err != nil {
		fmt.Println("Error reading template file:", err)
		return
	}
	readmeTemplate := string(templateBytes)

	// Generate README.md
	tmpl, err := template.New("readme").Parse(readmeTemplate)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	file, err := os.Create("README.md")
	if err != nil {
		fmt.Println("Error creating README.md file:", err)
		return
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
