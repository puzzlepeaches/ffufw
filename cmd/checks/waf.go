package cmd

import (
	"errors"
	"net"
	"net/url"

	valid "github.com/asaskevich/govalidator"
	cdn "github.com/projectdiscovery/cdncheck"
	"github.com/sirupsen/logrus"
)

func extractHostname(host string) (string, error) {
	// Extract hostname or IP address from URL
	hostname, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	return hostname.Hostname(), nil
}

func resolveDomain(host string) (string, error) {

	// Resolve domain to IP address
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", errors.New("No IPs found for the host")
	}
	return ips[0].String(), nil
}

func CheckWaf(url string) (string, error) {
	cdncheck := cdn.New()

	hostname, err := extractHostname(url)
	if err != nil {
		return "", err
	}

	// Check if the hostname is an IP address
	if !valid.IsIP(hostname) {
		hostname, err = resolveDomain(hostname)
		if err != nil {
			return "", err
		}
	}

	// Check if WAF
	matched, val, err := cdncheck.CheckWAF(net.ParseIP(hostname))
	if err != nil {
		return "", err
	}
	if matched {
		logrus.Infof("WAF detected for URL: %s", url)
		logrus.Debugf("WAF %s found for URL %s", val, url)
		return val, nil
	}
	logrus.Debugf("No WAF detected for URL: %s", url)
	return "", nil
}
