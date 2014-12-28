package util

import (
	"encoding/json"
	"io/ioutil"
)

// Defines the configuration file format
type ServiceConfiguration struct {
	CAKeyFile string
	FeedURL   string
	CWEfile   string
	DBFile    string
	BlackList string
}

// Used to load the configuration from file
func Load_Config(path string) ServiceConfiguration {
	s := ServiceConfiguration{}
	b, err := ioutil.ReadFile(path)
	checkerr(err)

	err = json.Unmarshal(b, &s)
	checkerr(err)

	return s
}
