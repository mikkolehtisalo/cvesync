package util

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/blackjack/syslog"
	"io/ioutil"
	"net/http"
)

func checkerr(err error) {
	if err != nil {
		syslog.Errf("Error: %v", err)
		panic(err)
	}
}

func Download_File(url string, cafile string) []byte {

	// Load the CA certificate for server certificate validation
	capool := x509.NewCertPool()
	cacert, err := ioutil.ReadFile(cafile)
	checkerr(err)
	capool.AppendCertsFromPEM(cacert)

	// Check server certificate
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: capool},
	}

	// Get!
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	checkerr(err)
	// 500s and such
	if resp.StatusCode != 200 {
		errr := errors.New(fmt.Sprintf("File download failed with status code %v", resp.StatusCode))
		syslog.Errf("%v", errr)
		panic(errr)
	}

	// Read the body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // body is []byte
	checkerr(err)

	return body
}

func Gunzip(in []byte) []byte {
	br := bytes.NewReader(in)
	r, err := gzip.NewReader(br)
	checkerr(err)
	defer r.Close()

	out, err := ioutil.ReadAll(r)
	checkerr(err)

	return out
}
