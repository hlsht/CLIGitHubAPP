package main

import (
	"fmt"
	"githubcliapp"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s [help|close-issue|create-issue|get-issue|list-issue|list-repos|lock-issue|login|update-issue|whoami]", os.Args[0])
	}

	switch os.Args[1] {
	case "help":
		help()
	case "close-issue":
		githubcliapp.CloseIssue()
	case "create-issue":
		githubcliapp.CreateIssue()
	case "get-issue":
		githubcliapp.GetIssue()
	case "list-issues":
		issues := githubcliapp.ListIssues()
		for _, issue := range issues {
			fmt.Printf("#%d - %s\n", issue.IssueNumber, issue.Title)
		}
	case "list-repos":
		githubcliapp.ListRepos()
	case "lock-issue":
		githubcliapp.LockIssue()
	case "login":
		githubcliapp.Login()
	case "update-issue":
		githubcliapp.UpdateIssue()
	case "whoami":
		fmt.Printf("You are @%s\n", githubcliapp.Whoami())
	default:
		log.Fatalf("%s: unknown command\n", os.Args[0])
	}
}

func help() {
	fmt.Printf("usage: %s <create-issue | get-issue | list-issues| list-repos | login | update-issue | whoaim | help>\n", os.Args[0])
}
