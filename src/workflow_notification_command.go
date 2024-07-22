package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abcxyz/pkg/cli"
)

type WorkflowNotificationCommand struct {
	cli.BaseCommand
	flagWebhookURL string
}

func (c *WorkflowNotificationCommand) Desc() string {
	return "Send a message to a Google Chat space"
}

func (c *WorkflowNotificationCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

  The chat command sends messages to Google Chat spaces.
`
}

func (c *WorkflowNotificationCommand) Flags() *cli.FlagSet {
	set := c.NewFlagSet()

	f := set.NewSection("COMMAND OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "webhook-url",
		Example: "https://chat.googleapis.com/v1/spaces/<SPACE_ID>/messages?key=<KEY>&token=<TOKEN>",
		Target:  &c.flagWebhookURL,
		Usage:   `Webhook URL from google chat`,
	})

	return set
}

func (c *WorkflowNotificationCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	args = f.Args()
	if len(args) != 0 {
		return fmt.Errorf("expected 0 arguments, got %q", args)
	}

	ghJSONStr := c.GetEnv(githubContextEnvKey)
	if ghJSONStr == "" {
		return fmt.Errorf("environment var %s not set", githubContextEnvKey)
	}
	jobJSONStr := c.GetEnv(jobContextEnvKey)
	if jobJSONStr == "" {
		return fmt.Errorf("environment var %s not set", jobContextEnvKey)
	}

	ghJSON := map[string]any{}
	jobJSON := map[string]any{}
	if err := json.Unmarshal([]byte(ghJSONStr), &ghJSON); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", githubContextEnvKey, err)
	}
	if err := json.Unmarshal([]byte(jobJSONStr), &jobJSON); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", jobContextEnvKey, err)
	}

	// Check if the job status is a failure
	if jobStatus, ok := jobJSON["status"].(string); !ok || jobStatus != "failure" {
		return nil // Exit without sending a notification if it's not a failure
	}

	b, err := generateRequestBody(generateMessageBodyContent(ghJSON, jobJSON, time.Now()))
	if err != nil {
		return fmt.Errorf("failed to generate message body: %w", err)
	}

	url := c.flagWebhookURL

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("creating http request failed: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("sending http request failed: %w", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read")
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("unexpected HTTP status code %d (%s)\n got body: %s", got, http.StatusText(got), bodyString)
	}

	return nil
}
