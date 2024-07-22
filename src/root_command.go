package main

import (
	"github.com/abcxyz/pkg/cli"
)

var rootCmd = func() cli.Command {
	return &cli.RootCommand{
		Name: "send-google-chat-webhook",
		Commands: map[string]cli.CommandFactory{
			"chat": func() cli.Command {
				return &cli.RootCommand{
					Name:        "workflownotification",
					Description: "notification for workflow",
					Commands: map[string]cli.CommandFactory{
						"workflownotification": func() cli.Command {
							return &WorkflowNotificationCommand{}
						},
					},
				}
			},
		},
	}
}
