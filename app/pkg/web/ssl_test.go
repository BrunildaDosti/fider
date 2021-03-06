package web

import (
	"crypto/tls"
	"testing"

	"github.com/getfider/fider/app/pkg/dbx"
	"github.com/getfider/fider/app/pkg/env"

	. "github.com/getfider/fider/app/pkg/assert"
)

func Test_GetCertificate(t *testing.T) {
	RegisterT(t)
	db := dbx.New()
	defer db.Close()

	var testCases = []struct {
		mode       string
		cert       string
		serverName string
		valid      bool
	}{
		{"multi", "all-test-fider-io", "", true},
		{"multi", "all-test-fider-io", "fider", true},
		{"multi", "all-test-fider-io", "feedback.test.fider.io", true},
		{"multi", "all-test-fider-io", "FEEDBACK.test.fider.io", true},
		{"multi", "all-test-fider-io", "app.feedback.test.fider.io", false},
		{"multi", "all-test-fider-io", "my.app.feedback.test.fider.io", false},
		{"single", "test-fider-io", "test.fider.io", true},
		{"single", "test-fider-io", "fider.io", true},
	}

	for _, testCase := range testCases {
		env.Config.HostMode = testCase.mode
		certFile := env.Path("/app/pkg/web/testdata/" + testCase.cert + ".crt")
		keyFile := env.Path("/app/pkg/web/testdata/" + testCase.cert + ".key")
		wildcardCert, _ := tls.LoadX509KeyPair(certFile, keyFile)

		manager, err := NewCertificateManager(certFile, keyFile, db)
		Expect(err).IsNil()
		cert, err := manager.GetCertificate(&tls.ClientHelloInfo{
			ServerName: testCase.serverName,
		})

		if testCase.valid {
			Expect(err).IsNil()
			Expect(cert.Certificate).Equals(wildcardCert.Certificate)
		} else {
			Expect(cert).IsNil()
			Expect(err.Error()).ContainsSubstring(`ssl: invalid server name "` + testCase.serverName + `"`)
		}
	}
}

func Test_UseAutoCert(t *testing.T) {
	RegisterT(t)
	db := dbx.New()
	defer db.Close()

	manager, err := NewCertificateManager("", "", db)
	Expect(err).IsNil()

	invalidServerNames := []string{"ideas.app.com", "feedback.mysite.com"}

	for _, serverName := range invalidServerNames {
		cert, err := manager.GetCertificate(&tls.ClientHelloInfo{
			ServerName: serverName,
		})
		Expect(err.Error()).ContainsSubstring(`acme/autocert: unable to authorize "` + serverName + `"; challenge "tls-alpn-01" failed with error: acme: authorization error`)
		Expect(cert).IsNil()
	}
}
