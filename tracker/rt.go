package tracker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/mikkolehtisalo/cvesync/nvd"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type RT struct {
	BaseURL        string
	CAFile         string
	Username       string
	Password       string
	Queue          string
	HighPriority   string
	MediumPriority string
	LowPriority    string
	TemplateFile   string
	Template       *template.Template
}

func (rt *RT) Init() {
	// Loading RT related settings
	b, err := ioutil.ReadFile("/opt/cvesync/etc/rt.json")
	if err != nil {
		syslog.Errf("Unable to read RT settings file: %v", err)
		panic(err)
	}

	err = json.Unmarshal(b, &rt)
	if err != nil {
		syslog.Errf("Unable to unmarshal RT settings json: %v", err)
		panic(err)
	}

	rt.Template, err = template.New("rt.templ").ParseFiles(rt.TemplateFile)
	if err != nil {
		syslog.Errf("Unable to parse RT template file: %v", err)
		panic(err)
	}
}

func (rt RT) authenticate(jar *cookiejar.Jar) error {

	client := &http.Client{
		Jar: jar,
	}

	data := url.Values{}
	data.Add("user", rt.Username)
	data.Add("pass", rt.Password)

	client.PostForm(rt.BaseURL, data)

	// Check that we got back at least one cookie -> probably successful authentication!
	// Alternatively could check that RT_SID_url.80 exists
	url, err := url.Parse(rt.BaseURL)
	if err != nil {
		syslog.Errf("Unable to parse BaseURL: %v", err)
		panic(err)
	}
	if len(jar.Cookies(url)) < 1 {
		return errors.New("Authentication to RT failed!")
	}

	return nil
}

type RTTicket struct {
	Subject  string
	Queue    string
	Priority string
	Text     string
}

// RT requires that the lines in description are indented.
func indent_text(s string) string {
	lines := strings.Split(s, "\n")
	for x, _ := range lines {
		lines[x] = " " + lines[x]
	}
	return strings.Join(lines, "\n")
}

func (rt RT) build_text(e nvd.Entry) string {
	var result bytes.Buffer

	err := rt.Template.Execute(&result, e)
	if err != nil {
		syslog.Errf("Unable to execute RT template file: %v", err)
		panic(err)
	}

	return result.String()
}

func (rt RT) build_ticket(e nvd.Entry) (RTTicket, error) {
	ticket := RTTicket{}

	subject := fmt.Sprintf("%v: %v", e.Id, e.Summary)
	// Effectively cut the summary at 200 characters (limit is at 255?)
	if len(subject) > 200 {
		subject = subject[:200] + "..."
	}
	ticket.Subject = subject
	ticket.Queue = rt.Queue

	// Priority
	score_float64, err := strconv.ParseFloat(e.CVSS.Score, 64)
	if err != nil {
		// Some CVEs have no CVSS score set yet, this is ok!
		// If err, then score_float64 to 4.0 => medium
		score_float64 = float64(4.0)
	}
	ticket.Priority = rt.LowPriority
	if score_float64 >= 4.0 {
		ticket.Priority = rt.MediumPriority
	}
	if score_float64 >= 7.0 {
		ticket.Priority = rt.HighPriority
	}

	ticket.Text = indent_text(rt.build_text(e))

	return ticket, nil
}

func (rt RT) Add(e nvd.Entry) (string, error) {
	// Authenticate against RT for this operation
	jar, err := cookiejar.New(nil)
	if err != nil {
		syslog.Errf("Unable to create cookie jar: %v", err)
		panic(err)
	}
	err = rt.authenticate(jar)
	if err != nil {
		syslog.Errf("%v", err)
		return "", err
	}

	// Build ticket information...
	ticket, err := rt.build_ticket(e)
	if err != nil {
		return "", err
	}

	// Build the request
	request := fmt.Sprintf("id: ticket/new\nQueue: %v\nSubject: %v\nPriority: %v\nText:%v\n", ticket.Queue, ticket.Subject, ticket.Priority, ticket.Text)

	id, err := rt_request("POST", rt.BaseURL+"/REST/1.0/ticket/new", rt.CAFile, jar, request)
	return id, err
}

func rt_request(reqtype string, path string, cafile string, jar *cookiejar.Jar, ticket string) (string, error) {
	var client *http.Client
	// If https, add CA certificate checking
	if strings.HasPrefix(path, "https://") {
		capool := x509.NewCertPool()
		cacert, err := ioutil.ReadFile(cafile)
		if err != nil {
			syslog.Errf("Unable to read CA file: %v", err)
			return "", err
		}
		capool.AppendCertsFromPEM(cacert)

		// Check server certificate
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: capool},
		}

		client = &http.Client{Transport: tr, Jar: jar}
	} else {
		client = &http.Client{Jar: jar}
	}

	data := url.Values{}
	data.Add("content", ticket)

	req, err := http.NewRequest(reqtype, path, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	// We are handling "form"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make RT's anti-XSS happy
	req.Header.Set("Referer", path)

	// Request!
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ticketid := get_ticket_id(string(body))

	return ticketid, nil

}

// Gets the RT's ticket id
func get_ticket_id(body string) string {
	regexp := regexp.MustCompile("# Ticket (\\d+) created.")
	result := ""
	for _, x := range strings.Split(body, "\n") {
		if regexp.MatchString(x) {
			id := regexp.FindAllStringSubmatch(x, -1)
			result = id[0][1]
		}
	}
	return result
}

func (rt RT) Update(e nvd.Entry, ticketid string) error {
	// Authenticate against RT for this operation
	jar, err := cookiejar.New(nil)
	if err != nil {
		syslog.Errf("Unable to create cookie jar: %v", err)
		panic(err)
	}
	err = rt.authenticate(jar)
	if err != nil {
		syslog.Errf("%v", err)
		return err
	}

	// Build ticket information...
	ticket, err := rt.build_ticket(e)
	if err != nil {
		return err
	}

	// Build the request
	request := fmt.Sprintf("Queue: %v\nSubject: %v\nPriority: %v\nText:%v\n", ticket.Queue, ticket.Subject, ticket.Priority, ticket.Text)

	_, err = rt_request("POST", rt.BaseURL+"/REST/1.0/ticket/"+ticketid+"/edit", rt.CAFile, jar, request)
	return err
}
