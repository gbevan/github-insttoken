package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
	jwtOnly := flag.Bool("jwt-only", false, "Only generate the JWT (workaround for proxy issue)")
	flag.Parse()

	if *privateKey == "" {
		panic("private-key-file is required")
	}

	if *appIDStr == "" {
		panic("app-id is required")
	}

	if !*jwtOnly && *repo == "" {
		panic("repo is required")
	}

	if !*jwtOnly && *gitURL == "" {
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

	if *jwtOnly {
		fmt.Println("jwt:", jwt)
		return
	}

	// fmt.Println(tokenString)

	client := &http.Client{}

	// Get Installation ID
	insResp, err := reqGithub(
		client,
		"GET",
		fmt.Sprintf("%v/repos/%v/installation", *gitURL, *repo),
		nil,
		jwt,
	)
	if err != nil {
		panic(err)
	}
	// fmt.Println("insResp ID:", insResp["id"])

	// Get Installation Token
	insTokResp, err := reqGithub(
		client,
		"POST",
		fmt.Sprintf("%v/app/installations/%v/access_tokens", *gitURL, insResp["id"]),
		nil,
		jwt,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("token:", insTokResp["token"])
}

func reqGithub(client *http.Client, method string, url string, postBody io.Reader, jwt string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, postBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
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
