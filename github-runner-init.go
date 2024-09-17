package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Configuration
var (
	githubToken        = os.Getenv("GITHUB_TOKEN")
	githubOrganization = os.Getenv("GITHUB_ORGANIZATION")
	githubURL          = os.Getenv("GITHUB_URL")
	serviceAccountName = os.Getenv("GITHUB_RUNNER_SERVICE_ACCOUNT")
	githubRunnerTokenDest    = os.Getenv("GITHUB_RUNNER_TOKEN_DEST")
)

func init() {
	if githubURL == "" {
		githubURL = "https://api.github.com"
		fmt.Println("Using default GitHub URL: https://api.github.com")
	} else {
		fmt.Printf("GitHub URL set to: %s\n", githubURL)
	}
	if serviceAccountName == "" {
		serviceAccountName = "default"
		fmt.Println("Using default service account name: default")
	}
	if githubRunnerTokenDest == "" {
		githubRunnerTokenDest = "/runner-token/runner_token"
		fmt.Println("Using default token destination: /runner-token/runner_token")
	}
}

// Function to get runner registration tokens
func getRunnerToken(url, org, token string) (string, error) {
	var apiURL string
	if url == "https://api.github.com" {
		apiURL = fmt.Sprintf("%s/orgs/%s/actions/runners/registration-token", url, org)
	} else {
		apiURL = fmt.Sprintf("%s/api/v3/orgs/%s/actions/runners/registration-token", url, org)
	}

	fmt.Printf("Requesting runner token from URL: %s\n", apiURL)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Received response with status: %s\n", resp.Status)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response body: %s\n", string(body)) // Print response body in case of failure
		return "", fmt.Errorf("failed to get runner token: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	token, ok := result["token"]
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	fmt.Println("Runner token successfully retrieved")
	return token, nil
}

func main() {
	// Ensure all environment variables are loaded
	fmt.Printf("GITHUB_ORGANIZATION: %s\n", githubOrganization)
	fmt.Printf("GITHUB_URL: %s\n", githubURL)
	fmt.Printf("GITHUB_RUNNER_TOKEN_DEST: %s\n", githubRunnerTokenDest)

	token, err := getRunnerToken(githubURL, githubOrganization, githubToken)
	if err != nil {
		fmt.Printf("Error getting runner token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Writing token to file: %s\n", githubRunnerTokenDest)
	if err := ioutil.WriteFile(githubRunnerTokenDest, []byte(token), 0644); err != nil {
		fmt.Printf("Error writing token to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Runner token successfully written to file")
}
