package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hlsht/CLIGitHubApp/internal/github"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <command>\n", os.Args[0])
		os.Exit(1)
	}

	app := github.NewApp()
	switch os.Args[1] {
	case "help":
		help()
	case "close-issue":
		app.CloseIssue()
	case "create-issue":
		app.CreateIssue()
	case "get-issue":
		github.DisplayIssue(app.GetIssue())
	case "list-issues":
		issues := app.ListIssues()
		for _, issue := range issues {
			fmt.Printf("#%d - %s\n", issue.Number, issue.Title)
		}
	case "list-repos":
		isLogin := app.CheckAuthorization()
		if !isLogin {
			fmt.Fprintf(os.Stderr, "ghcli: you are not authorized run `login` command.\n")
			os.Exit(1)
		}
		for _, repo := range app.ListRepos() {
			fmt.Println(repo.Name)
		}
	case "login":
		app.Login()
	case "update-issue":
		app.UpdateIssue()
	case "whoami":
		fmt.Printf("You are @%s\n", app.Whoami())
	default:
		log.Fatalf("%s: unknown command %s\n", os.Args[0], os.Args[1])
	}
}

func help() {
	helpMessage := `CLIGitHubApp is a utility that helps you work with issue through the command line interface.

Usage: ./ghcli <command>

Where command can be:
	login         use this command to login to the app before any other commands
	create-issue  create an issue for the specified repository
	get-issue     allows you to get information on the specified issue
	list-issue    allows you to get all issues related to the specified repository
	list-repos    shows all repositories that the application has access to
	update-issue  allows you to update information related to specified issue
	whoami        shows the username of the authorized user
	help          shows this message`
	fmt.Println(helpMessage)
}
