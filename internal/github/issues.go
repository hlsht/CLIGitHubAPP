package github

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strconv"

	// "flag"
	"fmt"
	// "io"
	"log"
	"net/http"
	"os"
	"os/exec"

	// "strconv"
	"strings"
	"unicode"
)

func (a *App) CreateIssue() {
	var issue Issue

	fmt.Print("Enter the name of the repository you want to create an issue for: ")
	issue.RepoName = getString()
	fmt.Print("Enter the title of the issue you want to create: ")
	issue.Title = getString()
	issue.Desc = getFromEditor()
	fmt.Print("Enter labels to associate with this issue separeted with comma: ")
	labels := getLabels()
	username := a.Whoami()
	issue.Assignees.Post = []string{username}
	issue.Labels.Post = labels

	authToken := a.getAuthToken()
	paramsJSON, err := json.Marshal(issue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error marshaling Go struct: %s\n", err)
		os.Exit(1)
	}

	req, err := a.NewRequest("POST",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", username, issue.RepoName),
		bytes.NewBuffer(paramsJSON),
		map[string]string{
			"Authorization": "Bearer " + authToken,
			"Content-Type":  "application/json",
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("%s: issues was not created: %s\n", os.Args[0], resp.Status)
	} else {
		fmt.Println("Issue was successfully created!")
	}
}

// func (a *App) isRepoExist(repo string) {
// }

//	func isRepoExist(repoName string, repos []Repo) bool {
//		for _, repo := range repos {
//			if repoName == repo.Name {
//				return true
//			}
//		}
//		return false
//	}
//
//	func ParseIssueOptionsFromCommandArgs() *IssueOptions {
//		fs := flag.NewFlagSet("issue-options", flag.ExitOnError)
//		repoName := fs.String("r", "", "the repository")
//		title := fs.String("t", "", "title of an issue")
//		desc := fs.String("d", "", "description of an issue")
//		number := fs.Int("n", -1, "number of an issue")
//		fs.Parse(os.Args[2:])
//
//		return &IssueOptions{Title: *title, Desc: *desc, RepoName: *repoName, Number: *number}
//	}
func getFromEditor() string {
	userMessage := `

# Write description to an issue.
# Use "<Esc>:wq" to quit text Editor and save your desc.
# Lines starting with '#' will be ignored`

	tempFile, err := os.CreateTemp("", "issue-message-*.txt")
	Check(err)
	defer os.Remove(tempFile.Name())

	tempFile.WriteString(userMessage)
	tempFile.Close()

	cmd := exec.Command("vim", tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	Check(err)

	var output bytes.Buffer
	tempFile, err = os.Open(tempFile.Name())
	Check(err)
	scanner := bufio.NewScanner(tempFile)

	for scanner.Scan() {
		if len(scanner.Text()) != 0 && scanner.Text()[0] != '#' {
			output.WriteString(scanner.Text() + "\n")
		}
	}

	return strings.TrimFunc(output.String(), unicode.IsSpace)
}

func (a *App) ListIssues() []Issue {
	var issue Issue

	fmt.Print("Enter the name of the repository you want to create an issue for: ")
	issue.RepoName = getString()
	var issues []Issue

	username := a.Whoami()
	authToken := a.getAuthToken()

	req, err := a.NewRequest("GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", username, issue.RepoName),
		nil,
		map[string]string{
			"Authorization": "Bearer " + authToken,
			"Content-Type":  "application/json",
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "ghcli: couldn't get issue related to repository '%s': %s\n",
			issue.RepoName, resp.Status)
		os.Exit(1)
	}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error unmarshaling json: %s\n", err)
		os.Exit(1)
	}

	return issues
}

func (a *App) GetIssue() *Issue {
	var issue Issue
	var err error
	fmt.Print("Enter the name of the repository you want to create an issue for: ")
	issue.RepoName = getString()
	fmt.Print("Enter the issue number: ")
	issue.Number, err = strconv.Atoi(getString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error parsing input: %s\n", err)
		os.Exit(1)
	}
	username := a.Whoami()
	authToken := a.getAuthToken()

	req, err := a.NewRequest("GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d",
			username, issue.RepoName, issue.Number), nil, map[string]string{
			"Authorization": "Bearer " + authToken,
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "ghcli: couldn't get issue related to repository '%s': %s\n",
			issue.RepoName, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error unmarshaling json: %s\n", err)
		os.Exit(1)
	}

	return &issue
}

func DisplayIssue(issue *Issue) {
	fmt.Printf("Issue #%d\n", issue.Number)
	fmt.Printf("Title: %s\n", issue.Title)
	fmt.Printf("Description: %s\n", issue.Desc)
	displayStringsWithComma("Assignee", issue.Assignees.Post)
	displayStringsWithComma("Label", issue.Labels.Post)
}

func displayStringsWithComma(title string, values []string) {
	if len(values) > 0 {
		fmt.Print(title)
		if len(values) > 1 {
			fmt.Print("s")
		}
		fmt.Print(": ")
		s := ""
		for _, val := range values {
			fmt.Printf("%s%s", s, val)
			s = ", "
		}
		fmt.Println()
	}
}

func (a *App) UpdateIssue() {
	issue := &Issue{}
	var err error
	options := []string{
		"Change issue's title.",
		"Change issue's description(you will be moved to an editor to type a description).",
		"Add assignee to an issue.",
		"Add labels to an issue",
	}
	issue = a.GetIssue()
	DisplayIssue(issue)

	hasUpdate := false
	for {
		displayOptions(options,
			fmt.Sprintf("Pick an action that you want to do with issue(type 0-%d, or -1 to exit): ",
				len(options)-1))
		picked := getUpdateOption(options)
		if picked == -1 {
			break
		}
		hasUpdate = true
		switch picked {
		case 0:
			fmt.Print("Enter new issue's title: ")
			issue.Title = getString()
		case 1:
			issue.Desc = getFromEditor()
		case 2:
			fmt.Print("Enter new assignee's username: ")
			assignee := getString()
			if !a.checkUserToExist(assignee) {
				fmt.Fprintf(os.Stderr, "%s: update-issue: user '%s' does not exist.\n", os.Args[0], assignee)
				break
			}
			issue.Assignees.Post = append(issue.Assignees.Post, assignee)
			issue.Assignees.Get = append(issue.Assignees.Get, &User{Login: assignee})
		case 3:
			fmt.Print("Enter labels to associate with this issue separeted with comma: ")
			issue.Labels.Post = getLabels()
		}
	}

	if !hasUpdate {
		return
	}

	username := a.Whoami()
	authToken := a.getAuthToken()
	paramsJSON, err := json.Marshal(issue)

	req, err := a.NewRequest("PATCH",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d",
			username, issue.RepoName, issue.Number),
		bytes.NewBuffer(paramsJSON), map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + authToken,
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s: issues was not updated: %s\n", os.Args[0], resp.Status)
	} else {
		fmt.Println("Issue was successfully updated!")
	}
}

func displayOptions(options []string, inputMessage string) {
	for i, option := range options {
		fmt.Printf("%d - %s\n", i, option)
	}
	fmt.Print(inputMessage)
}

func getUpdateOption(options []string) int64 {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	Check(err)
	text = text[:len(text)-1]
	picked, err := strconv.ParseInt(text, 10, 0)
	Check(err)
	if picked < -1 || picked >= int64(len(options)) {
		log.Fatalf("%d does not match any option.\n", picked)
	}
	return picked
}

func getString() string {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	Check(err)
	return strings.TrimFunc(text, unicode.IsSpace)
}

func getLabels() []string {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	Check(err)
	output := strings.Split(text, ",")
	for i := range output {
		output[i] = strings.TrimFunc(output[i], unicode.IsSpace)
	}
	return output
}

func (a *App) checkUserToExist(username string) bool {
	req, err := a.NewRequest("GET",
		fmt.Sprintf("https://api.github.com/users/%s", username), nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}

	return resp.StatusCode == http.StatusOK
}

func (a *App) CloseIssue() {
	issue := a.GetIssue()

	username := a.Whoami()
	authToken := a.getAuthToken()
	params := map[string]string{
		"state": "closed",
	}
	paramsJSON, err := json.Marshal(params)
	Check(err)

	req, err := a.NewRequest("PATCH",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d",
			username, issue.RepoName, issue.Number), bytes.NewBuffer(paramsJSON),
		map[string]string{
			"Authorization": "Bearer " + authToken,
			"Content-Type":  "application/json",
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s: issues was not closed: %s\n", os.Args[0], resp.Status)
	} else {
		fmt.Println("Issue was successfully closed!")
	}
}
