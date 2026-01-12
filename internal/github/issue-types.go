package github

import (
	"encoding/json"
	"fmt"
)

type IssueOptions struct {
	Title    string
	RepoName string
	Number   int
	Desc     string
}

type Issue struct {
	Title     string    `json:"title"`
	Desc      string    `json:"body"`
	Assignees Assignees `json:"assignees"`
	Labels    Labels    `json:"labels"`
	Number    int       `json:"number"`
	RepoName  string
}

type Assignees struct {
	Post []string
	Get  []*User
}

type User struct {
	Login string `json:"login"`
}

type Labels struct {
	Post []string
	Get  []*Label
}

type Label struct {
	Name string `json:"name"`
}

func (a Assignees) MarshalJSON() ([]byte, error) {
	if len(a.Post) > 0 {
		return json.Marshal(a.Post)
	}

	logins := make([]string, len(a.Get))
	for i, user := range a.Get {
		logins[i] = user.Login
	}
	return json.Marshal(logins)
}

func (a *Assignees) UnmarshalJSON(data []byte) error {
	var users []*User
	if err := json.Unmarshal(data, &users); err == nil {
		a.Get = users
		return nil
	}

	return fmt.Errorf("cannot unmarshal assignees: %s", string(data))
}

func (l Labels) MarshalJSON() ([]byte, error) {
	if len(l.Post) > 0 {
		return json.Marshal(l.Post)
	}

	names := make([]string, len(l.Get))
	for i, label := range l.Get {
		names[i] = label.Name
	}
	return json.Marshal(names)
}

func (l *Labels) UnmarshalJSON(data []byte) error {
	var labels []*Label
	if err := json.Unmarshal(data, &labels); err == nil {
		l.Get = labels
		l.Post = make([]string, len(labels))
		for i, label := range labels {
			l.Post[i] = label.Name
		}
		return nil
	}

	return fmt.Errorf("cannot unmarshal labels: %s", string(data))
}
