package githubcliapp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
)

func CreateIssue() {
	issueOpts := ParseIssueOptionsFromCommandArgs()

	if issueOpts.RepoName == "" {
		log.Fatalf("%s: create-issue: -r flag needs to be set.\n", os.Args[0])
	}
	if issueOpts.Title == "" {
		log.Fatalf("%s: create-issue: -t flag cannot contain empty string.\n", os.Args[0])
	}
	if issueOpts.Desc == "" {
		issueOpts.Desc = getFromEditor()
	}
	fmt.Print("Enter labels to associate with this issue separeted with comma: ")
	labels := getLabels()
	issue := &Issue{Title: issueOpts.Title, Desc: issueOpts.Desc, RepoName: issueOpts.RepoName}
	username := Whoami()
	issue.Assignees.Post = []string{username}
	issue.Labels.Post = labels

	repos := ListRepos()
	if !isRepoExist(issue.RepoName, repos) {
		log.Fatalf("%s: repo with name '%s' does not exist.\n", os.Args[0], issue.RepoName)
	}

	authToken := GetAuthToken()
	paramsJSON, err := json.Marshal(issue)

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", username, issue.RepoName), bytes.NewBuffer(paramsJSON))
	Check(err)

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "CliGithubApp")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("%s: issues was not created: %s\n", os.Args[0], resp.Status)
	} else {
		fmt.Println("Issue was successfully created!")
	}
}

func isRepoExist(repoName string, repos []Repo) bool {
	for _, repo := range repos {
		if repoName == repo.Name {
			return true
		}
	}
	return false
}

func ParseIssueOptionsFromCommandArgs() *IssueOptions {
	fs := flag.NewFlagSet("issue-options", flag.ExitOnError)
	repoName := fs.String("r", "", "the repository")
	title := fs.String("t", "", "title of an issue")
	desc := fs.String("d", "", "description of an issue")
	number := fs.Int("n", -1, "number of an issue")
	fs.Parse(os.Args[2:])

	return &IssueOptions{Title: *title, Desc: *desc, RepoName: *repoName, Number: *number}
}

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

func ListIssues() []Issue {
	issueOpts := ParseIssueOptionsFromCommandArgs()
	if issueOpts.RepoName == "" {
		log.Fatalf("%s: list-issues: -r flag needs to be set.\n", os.Args[0])
	}
	var issues []Issue

	username := Whoami()
	authToken := GetAuthToken()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", username, issueOpts.RepoName), nil)
	Check(err)

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "CliGithubApp")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s: couldn't get issue related to repository '%s': %s\n", os.Args[0], issueOpts.RepoName, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	Check(err)

	err = json.Unmarshal(body, &issues)
	Check(err)

	return issues
}

func GetIssue() {
	issueOpts := ParseIssueOptionsFromCommandArgs()
	if issueOpts.RepoName == "" {
		log.Fatalf("%s: get-issues: -r flag needs to be set.\n", os.Args[0])
	}
	if issueOpts.Number == -1 {
		log.Fatalf("%s: get-issues: -n flag needs to be assigned.\n", os.Args[0])
	}

	issues := ListIssues()
	searchedFor := findIssue(issues, issueOpts.Number)
	if searchedFor == nil {
		fmt.Fprintf(os.Stderr, "There is no issue with number: #%d for '%s' repo.\n", issueOpts.Number, issueOpts.RepoName)
	}
	displayIssue(searchedFor)
}

func displayIssue(issue *Issue) {
	fmt.Printf("Issue #%d\n", issue.IssueNumber)
	fmt.Printf("Title: %s\n", issue.Title)
	fmt.Printf("Description: %s\n", issue.Desc)
	displayStringsWithComma("Assignee", issue.Assignees.Post)
	displayStringsWithComma("Label", issue.Labels.Post)
}

func findIssue(issues []Issue, number int) *Issue {
	for _, issue := range issues {
		if issue.IssueNumber == number {
			return &issue
		}
	}
	return nil
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

func UpdateIssue() {
	options := []string{
		"Change issue's title.",
		"Change issue's description(you will be moved to an editor to type a description).",
		"Add assignee to an issue.",
		"Add labels to an issue",
	}
	issueOpts := ParseIssueOptionsFromCommandArgs()
	if issueOpts.RepoName == "" {
		log.Fatalf("%s: update-issue: -r flag needs to be set.\n", os.Args[0])
	}
	if issueOpts.Number == -1 {
		log.Fatalf("%s: update-issues: -n flag needs to be assigned.\n", os.Args[0])
	}

	issues := ListIssues()
	searchedFor := findIssue(issues, issueOpts.Number)
	if searchedFor == nil {
		log.Fatalf("There is no issue with number: #%d for '%s' repo.\n", issueOpts.Number, issueOpts.RepoName)
	}
	displayIssue(searchedFor)

	hasUpdate := false
	for {
		displayOptions(options, fmt.Sprintf("Pick an action that you want to do with issue(type 0-%d, or -1 to exit): ", len(options)-1))
		picked := getUpdateOption(options)
		if picked == -1 {
			break
		}
		hasUpdate = true
		switch picked {
		case 0:
			fmt.Print("Enter new issue's title: ")
			searchedFor.Title = getString()
		case 1:
			searchedFor.Desc = getFromEditor()
		case 2:
			fmt.Print("Enter new assignee's username: ")
			assignee := getString()
			if !checkUserToExist(assignee) {
				fmt.Fprintf(os.Stderr, "%s: update-issue: user '%s' does not exist.\n", os.Args[0], assignee)
				break
			}
			searchedFor.Assignees.Post = append(searchedFor.Assignees.Post, assignee)
			searchedFor.Assignees.Get = append(searchedFor.Assignees.Get, &User{Login: assignee})
		case 3:
			fmt.Print("Enter labels to associate with this issue separeted with comma: ")
			searchedFor.Labels.Post = getLabels()
		}
	}

	if !hasUpdate {
		return
	}

	username := Whoami()
	authToken := GetAuthToken()
	paramsJSON, err := json.Marshal(searchedFor)

	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", username, issueOpts.RepoName, searchedFor.IssueNumber), bytes.NewBuffer(paramsJSON))
	Check(err)

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "CliGithubApp")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()

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

func checkUserToExist(username string) bool {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/users/%s", username), nil)
	Check(err)

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "CliGithubApp")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)

	return resp.StatusCode == http.StatusOK
}

func LockIssue() {
	options := []string{
		"off-topic",
		"too heated",
		"resolved",
		"spam",
	}
	issueOpts := ParseIssueOptionsFromCommandArgs()
	if issueOpts.RepoName == "" {
		log.Fatalf("%s: create-issue: -r flag needs to be set.\n", os.Args[0])
	}
	if issueOpts.Number == -1 {
		log.Fatalf("%s: get-issues: -n flag needs to be assigned.\n", os.Args[0])
	}

	displayOptions(options, "Pick lock reason for this issue(type 0-3, or -1 to exit): ")
	picked := getUpdateOption(options)
	if picked == -1 {
		return
	}

	username := Whoami()
	lockData := map[string]string{
		"lock_reason": options[picked],
	}
	jsonData, err := json.Marshal(lockData)
	Check(err)

	req, err := http.NewRequest("PUT", fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/lock", username, issueOpts.RepoName, issueOpts.Number), bytes.NewBuffer(jsonData))
	Check(err)

	authToken := GetAuthToken()
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "CliGithubApp")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)

	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("%s: couldn't lock issue: %s\n", os.Args[0], resp.Status)
	}

	fmt.Println("Issue was successfully locked!")
}

func CloseIssue() {
	issueOpts := ParseIssueOptionsFromCommandArgs()
	if issueOpts.RepoName == "" {
		log.Fatalf("%s: update-issue: -r flag needs to be set.\n", os.Args[0])
	}
	if issueOpts.Number == -1 {
		log.Fatalf("%s: update-issues: -n flag needs to be assigned.\n", os.Args[0])
	}

	issues := ListIssues()
	if findIssue(issues, issueOpts.Number) == nil {
		log.Fatalf("There is no issue with number: #%d for '%s' repo.\n", issueOpts.Number, issueOpts.RepoName)
	}

	username := Whoami()
	authToken := GetAuthToken()
	params := map[string]string{
		"state": "closed",
	}
	paramsJSON, err := json.Marshal(params)
	Check(err)

	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", username, issueOpts.RepoName, issueOpts.Number), bytes.NewBuffer(paramsJSON))
	Check(err)

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "CliGithubApp")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s: issues was not closed: %s\n", os.Args[0], resp.Status)
	} else {
		fmt.Println("Issue was successfully closed!")
	}
}
