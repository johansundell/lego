package e2e

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/certificate"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/e2e/loader"
	"github.com/xenolf/lego/registration"
)

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		HealthCheckURL: "https://localhost:14000/dir",
		Args:           []string{"-strict", "-config", "fixtures/pebble-config.json"},
		Env:            []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=./fixtures/certs/pebble.minica.pem",
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(m))
}

func TestHelp(t *testing.T) {
	output, err := load.RunLego("-h")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}

func TestChallengeHTTP_Run(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "tls-alpn-01",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_Domains(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_CSR(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-csr", "./fixtures/csr.raw",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_CSR_PEM(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-csr", "./fixtures/csr.cert",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_Revoke(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-d", "lego.wtf",
		"-d", "acme.lego.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}

	output, err = load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-d", "lego.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"revoke")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS_Run_Revoke_Non_ASCII(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-d", "légô.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}

	output, err = load.RunLego(
		"-m", "hubert@hubert.com",
		"--accept-tos",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-d", "légô.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"revoke")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeHTTP_Client_Obtain(t *testing.T) {
	os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privKey}
	config := acme.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := acme.NewClient(config)
	require.NoError(t, err)

	client.Challenge.Exclude([]challenge.Type{challenge.DNS01, challenge.TLSALPN01})
	client.Challenge.SetHTTP01Address(":5002")

	reg, err := client.Registration.Register(true)
	require.NoError(t, err)
	user.registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"acme.wtf"},
		Bundle:  true,
	}
	resource, err := client.Certificate.Obtain(request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.Empty(t, resource.CSR)
}

func TestChallengeTLS_Client_Obtain(t *testing.T) {
	os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privKey}
	config := acme.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := acme.NewClient(config)
	require.NoError(t, err)

	client.Challenge.Exclude([]challenge.Type{challenge.DNS01, challenge.HTTP01})
	client.Challenge.SetTLSALPN01Address(":5001")

	reg, err := client.Registration.Register(true)
	require.NoError(t, err)
	user.registration = reg

	request := certificate.ObtainRequest{
		Domains:    []string{"acme.wtf"},
		Bundle:     true,
		PrivateKey: privKey,
	}
	resource, err := client.Certificate.Obtain(request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.Empty(t, resource.CSR)
}

func TestChallengeTLS_Client_ObtainForCSR(t *testing.T) {
	os.Setenv("LEGO_CA_CERTIFICATES", "./fixtures/certs/pebble.minica.pem")
	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privKey}
	config := acme.NewConfig(user)
	config.CADirURL = load.PebbleOptions.HealthCheckURL

	client, err := acme.NewClient(config)
	require.NoError(t, err)

	client.Challenge.Exclude([]challenge.Type{challenge.DNS01, challenge.HTTP01})
	client.Challenge.SetTLSALPN01Address(":5001")

	reg, err := client.Registration.Register(true)
	require.NoError(t, err)
	user.registration = reg

	csrRaw, err := ioutil.ReadFile("./fixtures/csr.raw")
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrRaw)
	require.NoError(t, err)

	resource, err := client.Certificate.ObtainForCSR(*csr, true)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, "acme.wtf", resource.Domain)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:14000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.NotEmpty(t, resource.CSR)
}

type fakeUser struct {
	email        string
	privateKey   crypto.PrivateKey
	registration *registration.Resource
}

func (f *fakeUser) GetEmail() string                        { return f.email }
func (f *fakeUser) GetRegistration() *registration.Resource { return f.registration }
func (f *fakeUser) GetPrivateKey() crypto.PrivateKey        { return f.privateKey }
