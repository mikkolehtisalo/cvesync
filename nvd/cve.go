package nvd

import (
	"encoding/xml"
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/mikkolehtisalo/cvesync/util"
	"strings"
	"time"
)

type CVE struct {
	Entries []Entry `xml:"entry"`
}

//Ignored elements: vuln:vulnerable-configuration, most often just repeats vuln:vulnerable-software-list
type Entry struct {
	Id            string      `xml:"cve-id"`
	Products      []string    `xml:"vulnerable-software-list>product"`
	Published     time.Time   `xml:"published-datetime"`
	Last_Modified time.Time   `xml:"last-modified-datetime"`
	CVSS          Cvss        `xml:"cvss"`
	CWE           Cwe         `xml:"cwe"`
	References    []Reference `xml:"references"`
	Summary       string      `xml:"summary"`
}

type Cvss struct {
	Score                  string    `xml:"base_metrics>score"`
	Access_Vector          string    `xml:"base_metrics>access-vector"`
	Access_Complexity      string    `xml:"base_metrics>access-complexity"`
	Authentication         string    `xml:"base_metrics>authentication"`
	Confidentiality_Impact string    `xml:"base_metrics>confidentiality-impact"`
	Integrity_Impact       string    `xml:"base_metrics>integrity-impact"`
	Availability_Impact    string    `xml:"base_metrics>availability-impact"`
	Source                 string    `xml:"base_metrics>source"`
	Generated_On           time.Time `xml:"base_metrics>generated-on-datetime"`
}

// To use a>b,attr directly in Entry would have been cleaner, but Unmarshal doesn't support that
type Cwe struct {
	Id         string `xml:"id,attr"`
	CWECatalog *CWE
}

// Links CWE to mitre.org
func (c Cwe) Definition_Link() string {
	link := ""
	split := strings.Split(c.Id, "-")
	if len(split) == 2 {
		link = fmt.Sprintf("http://cwe.mitre.org/data/definitions/%v.html", split[1])
	}
	return link
}

// Description for the CWE
func (c Cwe) CWE_Definition() string {
	definition := ""
	split := strings.Split(c.Id, "-")
	if len(split) == 2 {
		for x, _ := range c.CWECatalog.Weaknesses {
			if c.CWECatalog.Weaknesses[x].ID == split[1] {
				definition = c.CWECatalog.Weaknesses[x].Description
				// Remove line feeds, carriage returns and tabs
				definition = strings.Replace(definition, "\n", "", -1)
				definition = strings.Replace(definition, "\r", "", -1)
				definition = strings.Replace(definition, "\t", "", -1)
			}
		}
	}
	return definition
}

type Reference struct {
	Type   string           `xml:"reference_type,attr"`
	Source string           `xml:"source"`
	Target Reference_Target `xml:"reference"`
}

type Reference_Target struct {
	URL  string `xml:"href,attr"`
	Text string `xml:",chardata"`
}

func Unmarshal_CVE(data []byte) CVE {
	var c CVE
	err := xml.Unmarshal(data, &c)
	if err != nil {
		syslog.Errf("Unable to parse feed: %v", err)
		panic(err)
	}

	return c
}

func Get_CVE_feed(feedurl string, cakeyfile string) CVE {
	data := util.Download_File(feedurl, cakeyfile)

	var feed CVE
	if strings.HasSuffix(feedurl, ".gz") {
		unzipped := util.Gunzip(data)
		feed = Unmarshal_CVE(unzipped)
	} else {
		feed = Unmarshal_CVE(data)
	}

	return feed
}
