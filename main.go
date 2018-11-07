package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func main() {
	// user := flag.String("user", "", "User name to embed in the JWT")
	privateKey := flag.String("private-key-file", "", "Path to file containing GitHub app private key PEM")
	appIDStr := flag.String("app-id", "", "GitHub Application ID")
	gitURL := flag.String("git-url", "https://api.github.com", "Github api base usr (ent: https://github.example.com/api/v3/)")
	repo := flag.String("repo", "", "GitHub repository (owner/project)")
	flag.Parse()

	if *privateKey == "" {
		panic("private-key-file is required")
	}

	if *appIDStr == "" {
		panic("app-id is required")
	}

	if *repo == "" {
		panic("repo is required")
	}

	if *gitURL == "" {
		panic("git-url is required")
	}

	appID := -1
	if i, err := strconv.Atoi(*appIDStr); err == nil {
		appID = i
	} else {
		panic(err)
	}

	// Get RSA Private key
	pkBytes, err := ioutil.ReadFile(*privateKey)
	if err != nil {
		panic(err)
	}
	pk, err := jwt.ParseRSAPrivateKeyFromPEM(pkBytes)
	if err != nil {
		panic(err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		// issued at time
		"iat": time.Now().Unix(),
		// expiration time - 10mins max
		"exp": time.Now().Add(time.Minute * 10).Unix(),
		// GitHub App ID
		"iss": appID,
	})

	// Sign and get the complete encoded token as a string using the private key
	jwt, err := token.SignedString(pk)
	if err != nil {
		panic(err)
	}

	// // fmt.Println(tokenString)
	// tr := &http.Transport{
	// 	Proxy: http.ProxyFromEnvironment,
	// 	DialContext: (&net.Dialer{
	// 		Timeout:   30 * time.Second,
	// 		KeepAlive: 30 * time.Second,
	// 		DualStack: true,
	// 	}).DialContext,
	// 	MaxIdleConns:          100,
	// 	IdleConnTimeout:       90 * time.Second,
	// 	TLSHandshakeTimeout:   10 * time.Second,
	// 	ExpectContinueTimeout: 1 * time.Second,
	// }
	// tr.ProxyConnectHeader = http.Header{}
	// client := &http.Client{Transport: tr}

	// Get Installation ID
	insResp, err := reqGithub(
		// client,
		"GET",
		fmt.Sprintf("%v/repos/%v/installation", *gitURL, *repo),
		nil,
		jwt,
		// tr,
	)
	if err != nil {
		panic(err)
	}
	// fmt.Println("insResp ID:", insResp["id"])

	// Get Installation Token
	insTokResp, err := reqGithub(
		// client,
		"POST",
		fmt.Sprintf("%v/app/installations/%v/access_tokens", *gitURL, insResp["id"]),
		nil,
		jwt,
		// tr,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("token:", insTokResp["token"])
}

// func reqGithub(client *http.Client, method string, url string, postBody io.Reader, jwt string, tr *http.Transport) (map[string]interface{}, error) {
func reqGithub(method string, url string, postBody io.Reader, jwt string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, postBody)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// DisableCompression:    true,
	}
	// tr.ProxyConnectHeader = http.Header{}

	req.Header.Set("Host", "github.dxc.com")
	// req.Header.Set("Host", "GitHub.com")
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))

	// proxy
	// req.Header.Set("Proxy-Connection", "Keep-Alive")
	// tr.ProxyConnectHeader.Set("Proxy-Connection", "Keep-Alive")

	// force user-agent for proxy
	// req.Header.Set("User-Agent", "Wget/1.9.1")
	// tr.ProxyConnectHeader = req.Header

	client := &http.Client{Transport: tr}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > 15 {
			return fmt.Errorf("%d consecutive redirects", len(via))
		}
		if len(via) == 0 {
			return nil
		}
		for key, val := range via[0].Header {
			req.Header[key] = val
		}
		fmt.Printf("Redirect header: %v\n", req.Header)
		return nil
	}
	// fmt.Printf("req: %v\n", req)
	fmt.Printf("req.Header: %v\n", req.Header)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("Error: Invalid response from github, resp: %v", resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	subResp := make(map[string]interface{})
	err = json.Unmarshal(body, &subResp)
	if err != nil {
		return nil, err
	}
	return subResp, nil
}
