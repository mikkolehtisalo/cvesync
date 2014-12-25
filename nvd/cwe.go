package nvd

import (
	"encoding/xml"
	"github.com/blackjack/syslog"
	"io/ioutil"
)

type CWE struct {
	Weaknesses []Weakness `xml:"Weaknesses>Weakness"`
}

type Weakness struct {
	ID          string `xml:"ID,attr"`
	Description string `xml:"Description>Description_Summary"`
}

func Unmarshal_CWE(data []byte) CWE {
	var c CWE
	err := xml.Unmarshal(data, &c)
	if err != nil {
		syslog.Errf("Unable to parse CWEs: %v", err)
		panic(err)
	}

	return c
}

func Get_CWEs(filename string) CWE {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		syslog.Errf("Unable to read CWE file: %v", err)
		panic(err)
	}

	cwes := Unmarshal_CWE(b)
	return cwes
}
