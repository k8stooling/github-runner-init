# GitHub Runner Token Fetcher

This Go application is a helper function designed to run as an `initContainer` in Kubernetes. It fetches a GitHub Runner registration token, which can then be used to initialize a GitHub runner instance.

## Features

- Retrieves a GitHub Runner registration token for a specified organization.
- Works with both the public GitHub API (`https://api.github.com`) and GitHub Enterprise installations.
- Writes the registration token to a specified file, which can be used later to initialize a GitHub runner.

## Prerequisites

Ensure the following environment variables are set in your Kubernetes manifest or environment where the application will run:

- `GITHUB_TOKEN`: Your GitHub personal access token (with required permissions).
- `GITHUB_ORGANIZATION`: The GitHub organization for which the runner will be registered.
- `GITHUB_URL`: *(Optional)* The GitHub API URL (default: `https://api.github.com`).
- `GITHUB_RUNNER_TOKEN_DEST`: *(Optional)* The file path where the GitHub Runner token will be stored (default: `/runner-token/runner_token`).

## Usage

This application is typically used as an `initContainer` within a Kubernetes pod to prepare a GitHub runner. Below is an example of how you can integrate it into your Kubernetes manifest.

### Example Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: github-runner-diy
  labels:
    app: github-runner
spec:
  replicas: 1
  selector:
    matchLabels:
      app: github-runner
  template:
    metadata:
      labels:
        app: github-runner
    spec:
      initContainers:
      - name: init-runner-token
        image: ghcr.io/k8stooling/github-runner-init:latest
        env:
        - name: GITHUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: github-secrets
              key: GITHUB_TOKEN
        - name: GITHUB_ORGANIZATION
          valueFrom:
            secretKeyRef:
              name: github-secrets
              key: GITHUB_ORGANIZATION
        - name: GITHUB_URL
          valueFrom:
            secretKeyRef:
              name: github-secrets
              key: GITHUB_URL
        volumeMounts:
        - name: runner-token-volume
          mountPath: /runner-token
      containers:
      - name: runner
        image: ghcr.io/k8stooling/github-runner:latest
        env:
        - name: GITHUB_ORGANIZATION
          valueFrom:
            secretKeyRef:
              name: github-secrets
              key: GITHUB_ORGANIZATION
        - name: GITHUB_URL
          valueFrom:
            secretKeyRef:
              name: github-secrets
              key: GITHUB_URL
        command: ["/bin/bash", "-c"]
        args: ["/app/config.sh --unattended --url $GITHUB_URL/$GITHUB_ORGANIZATION --token `cat /runner-token/runner_token` --labels ubuntu-latest && /app/run.sh"]
        volumeMounts:
        - name: runner-token-volume
          mountPath: /runner-token
        lifecycle:
          preStop:
            exec:
              command: ["/bin/bash", "-c", "/app/config.sh remove --token `cat /runner-token/runner_token`"]
      volumes:
      - name: runner-token-volume
        emptyDir: {}

```

### Running Locally

You can also run the application locally for testing:

```bash
export GITHUB_TOKEN=your-github-token
export GITHUB_ORGANIZATION=your-github-organization
export GITHUB_URL=https://api.github.com  # or your GitHub Enterprise URL
export GITHUB_RUNNER_TOKEN_DEST=/path/to/save/token

go run main.go
```

# Error Handling

If the application fails to retrieve the runner token, it will print the error and exit with a non-zero status code. Some common errors include:

- Invalid or missing GitHub token.
- Incorrect GitHub organization.
- Incorrect API URL (especially for GitHub Enterprise users).
