package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

// These are our command-line options
var (
	mailGunAPIKey     string
	slackWebhookURL   string
	listenPort        int
	serverSSLCertFile string
	serverSSLKeyFile  string
	showVersion       bool
)

// These are set by the build command using -ldflags (see Makefile)
var version = "not set"
var githash = "not set"
var buildstamp = "not set"

func init() {
	flag.StringVar(&mailGunAPIKey, "mailgun", os.Getenv("MAGSLI_MAILGUN_API_KEY"), "MailGun API Key")
	flag.StringVar(&slackWebhookURL, "slack", os.Getenv("MAGSLI_SLACK_WEBHOOK_URL"), "Slack webhook to post to")

	//
	flag.IntVar(&listenPort, "port", 8080, "Port to lisen on")

	// optional for SSL listening
	flag.StringVar(&serverSSLCertFile, "sslCert", os.Getenv("MAGSLI_SERVER_SSL_CERT_FILE"), "Path to server certificate file")
	flag.StringVar(&serverSSLKeyFile, "sslKey", os.Getenv("MAGSLI_SERVER_SSL_KEY_FILE"), "Path to server private key file")

	flag.BoolVar(&showVersion, "v", false, "Show version information and exit")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
}

func main() {

	var err error

	flag.Parse()

	if showVersion {
		printVersionInfo()
		os.Exit(0)
	}

	if mailGunAPIKey == "" ||
		slackWebhookURL == "" {
		printUsage()
		os.Exit(1)
	}

	// ensure both ssl options passed
	if (serverSSLCertFile != "" && serverSSLKeyFile == "") ||
		(serverSSLCertFile == "" && serverSSLKeyFile != "") {
		fmt.Println("Invalid SSL flag combination")
		os.Exit(1)
	}

	http.HandleFunc("/", handler)

	printVersionInfo()
	fmt.Printf("magsli listening on: %v\n", listenPort)

	if serverSSLCertFile != "" && serverSSLKeyFile != "" {
		err = http.ListenAndServeTLS(fmt.Sprintf(":%d", listenPort), serverSSLCertFile, serverSSLKeyFile, logRequest(http.DefaultServeMux))
	} else {
		err = http.ListenAndServe(fmt.Sprintf(":%d", listenPort), logRequest(http.DefaultServeMux))
	}

	if err != nil {
		log.Fatal(err)
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("INFO: %s %s %s [%s]\n", r.Method, r.URL, r.RemoteAddr, r.Header)
		handler.ServeHTTP(w, r)
	})
}

func printVersionInfo() {
	fmt.Printf("magsli %v\n", version)
	fmt.Printf("  git hash: %v\n", githash)
	fmt.Printf("  built (UTC): %v\n", buildstamp)
}

func printUsage() {
	fmt.Printf("Usage: magsli [options]\n")
	fmt.Println("Options:")

	flag.PrintDefaults()
}
