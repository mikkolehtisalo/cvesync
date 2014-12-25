package tracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/mikkolehtisalo/cvesync/nvd"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type Jira struct {
	BaseURL        string
	Username       string
	Password       string
	Project        string
	Issuetype      string
	TemplateFile   string
	HighPriority   string
	MediumPriority string
	LowPriority    string
	Template       *template.Template
}

func (j *Jira) Init() {
	// Loading Jira related settings
	b, err := ioutil.ReadFile("/opt/cvesync/etc/jira.json")
	if err != nil {
		syslog.Errf("Unable to read Jira settings file: %v", err)
		panic(err)
	}

	err = json.Unmarshal(b, &j)
	if err != nil {
		syslog.Errf("Unable to unmarshal Jira settings json: %v", err)
		panic(err)
	}

	funcMap := template.FuncMap{
		"escape_text": escape_text,
	}

	j.Template, err = template.New("jira.templ").Funcs(funcMap).ParseFiles(j.TemplateFile)
	if err != nil {
		syslog.Errf("Unable to parse Jira template file: %v", err)
		panic(err)
	}

}

// A few CVEs contain characters that break Jira's text formatting
func escape_text(s string) string {
	result := strings.Replace(s, "[", "\\[", -1)
	result = strings.Replace(result, "]", "\\]", -1)
	result = strings.Replace(result, "~", "\\~", -1)

	return result
}

func (j Jira) build_description(e nvd.Entry) string {
	var result bytes.Buffer

	err := j.Template.Execute(&result, e)
	if err != nil {
		syslog.Errf("Unable to execute Jira template file: %v", err)
		panic(err)
	}

	return result.String()
}

// Populates struct for JSON request
func (j Jira) build_ticket(e nvd.Entry) (JiraTicket, error) {
	ticket := JiraTicket{}
	summary := fmt.Sprintf("%v: %v", e.Id, e.Summary)
	// Effectively cut the summary at 200 characters (Jira supports <255 by default)
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}
	ticket.Fields.Summary = summary
	ticket.Fields.Issuetype.Id = j.Issuetype
	ticket.Fields.Project.Id = j.Project
	ticket.Fields.Description = j.build_description(e)

	// Priority
	score_float64, err := strconv.ParseFloat(e.CVSS.Score, 64)
	if err != nil {
		// Some CVEs have no CVSS score set yet, this is ok!
		// If err, then score_float64 to 4.0 => medium
		score_float64 = float64(4.0)
	}
	ticket.Fields.Priority.Id = j.LowPriority
	if score_float64 >= 4.0 {
		ticket.Fields.Priority.Id = j.MediumPriority
	}
	if score_float64 >= 7.0 {
		ticket.Fields.Priority.Id = j.HighPriority
	}

	return ticket, nil

}

// Add new ticket, return the Jira's ticket id
func (j Jira) Add(e nvd.Entry) (string, error) {
	ticket, err := j.build_ticket(e)
	if err != nil {
		return "", err
	}

	json, err := json.Marshal(ticket)
	if err != nil {
		return "", err
	}

	id, err := Request("POST", j.BaseURL+"/rest/api/2/issue", j.Username, j.Password, string(json))
	return id, err
}

// Modify existing ticket, ticketid is Jira's ticket id
func (j Jira) Update(e nvd.Entry, ticketid string) error {
	ticket, err := j.build_ticket(e)
	if err != nil {
		return err
	}

	json, err := json.Marshal(ticket)
	if err != nil {
		return err
	}

	_, err = Request("PUT", j.BaseURL+"/rest/api/2/issue/"+ticketid, j.Username, j.Password, string(json))
	return err
}

func Request(reqtype string, path string, username string, password string, jsonstr string) (string, error) {
	client := &http.Client{}

	jsonreader := strings.NewReader(jsonstr)
	req, err := http.NewRequest(reqtype, path, jsonreader)
	if err != nil {
		return "", err
	}

	// Without application/json Jira returns 415
	req.Header.Set("Content-Type", "application/json")
	// Basic Authentication
	req.SetBasicAuth(username, password)

	// Request!
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// If not successful, return with statuscode
	if (resp.StatusCode < 200) || (resp.StatusCode > 299) {
		return "", errors.New(fmt.Sprintf("Response contained %v", resp.StatusCode))
	}

	ticketid := ""
	// Only POST returns something meaningful
	if reqtype == "POST" {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		jira_response := Jira_Response{}
		err = json.Unmarshal(body, &jira_response)
		if err != nil {
			return "", err
		}
		ticketid = jira_response.Id
	}

	return ticketid, nil
}

type JiraTicket struct {
	Fields Field `json:"fields"`
}

type Field struct {
	Project     Project_field   `json:"project"`
	Summary     string          `json:"summary"`
	Issuetype   Issuetype_field `json:"issuetype"`
	Priority    Priority_field  `json:"priority"`
	Description string          `json:"description"`
}

type Project_field struct {
	Id string `json:"id"`
}

type Issuetype_field struct {
	Id string `json:"id"`
}

type Priority_field struct {
	Id string `json:"id"`
}

// Jira responds with basic information about the created/modified ticket
type Jira_Response struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}
