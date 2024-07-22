package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// messageBodyContent defines the necessary fields for generating the request body.
type messageBodyContent struct {
	title           string
	subtitle        string
	ref             string
	triggeringActor string
	timestamp       string
	clickURL        string
	headerIconURL   string
	eventName       string
	repo            string
}

// generateMessageBodyContent returns messageBodyContent for generating the request body.
// using currentTimestamp as a input is for easier testing on default case.
func generateMessageBodyContent(ghJSON, jobJSON map[string]any, currentTimeStamp time.Time) *messageBodyContent {
	eventName := getMapFieldStringValue(ghJSON, githubContextEventNameKey)
	switch eventName {
	case "issues":
		event, ok := ghJSON[githubContextEventKey].(map[string]any)
		if !ok {
			event = map[string]any{}
		}
		issueContent, ok := event["issue"].(map[string]any)
		if !ok {
			issueContent = map[string]any{}
		}
		return &messageBodyContent{
			title:           fmt.Sprintf("A issue is %s", getMapFieldStringValue(event, githubContextEventObjectActionKey)),
			subtitle:        fmt.Sprintf("Issue title: <b>%s</b>", getMapFieldStringValue(issueContent, "title")),
			ref:             getMapFieldStringValue(ghJSON, githubContextRefKey),
			triggeringActor: getMapFieldStringValue(ghJSON, githubContextTriggeringActorKey),
			timestamp:       getMapFieldStringValue(issueContent, githubEventContenntCreatedAtKey),
			clickURL:        getMapFieldStringValue(issueContent, githubContextEventURLKey),
			eventName:       "issue",
			repo:            getMapFieldStringValue(ghJSON, githubContextRepositoryKey),
			headerIconURL:   successHeaderIconURL,
		}
	case "release":
		event, ok := ghJSON[githubContextEventKey].(map[string]any)
		if !ok {
			event = map[string]any{}
		}
		releaseContent, ok := event["release"].(map[string]any)
		if !ok {
			releaseContent = map[string]any{}
		}
		return &messageBodyContent{
			title:           fmt.Sprintf("A release is %s", getMapFieldStringValue(event, githubContextEventObjectActionKey)),
			subtitle:        fmt.Sprintf("Release name: <b>%s</b>", getMapFieldStringValue(releaseContent, "name")),
			ref:             getMapFieldStringValue(ghJSON, githubContextRefKey),
			triggeringActor: getMapFieldStringValue(ghJSON, githubContextTriggeringActorKey),
			timestamp:       getMapFieldStringValue(releaseContent, githubEventContenntCreatedAtKey),
			clickURL:        getMapFieldStringValue(releaseContent, githubContextEventURLKey),
			eventName:       "release",
			repo:            getMapFieldStringValue(ghJSON, githubContextRepositoryKey),
			headerIconURL:   successHeaderIconURL,
		}
	default:
		res := &messageBodyContent{
			title:           fmt.Sprintf("GitHub workflow %s", getMapFieldStringValue(jobJSON, "status")),
			subtitle:        fmt.Sprintf("Workflow: <b>%s</b>", getMapFieldStringValue(ghJSON, "workflow")),
			ref:             getMapFieldStringValue(ghJSON, githubContextRefKey),
			triggeringActor: getMapFieldStringValue(ghJSON, githubContextTriggeringActorKey),
			timestamp:       currentTimeStamp.UTC().Format(time.RFC3339),
			clickURL:        fmt.Sprintf("https://github.com/%s/actions/runs/%s", getMapFieldStringValue(ghJSON, githubContextRepositoryKey), getMapFieldStringValue(ghJSON, "run_id")),
			eventName:       "workflow",
			repo:            getMapFieldStringValue(ghJSON, githubContextRepositoryKey),
		}

		jobStatus := getMapFieldStringValue(jobJSON, "status")
		if jobStatus == "failure" {
			res.headerIconURL = failureHeaderIconURL
			res.title = "GitHub workflow failed"
			res.subtitle += fmt.Sprintf(" - Job: <b>%s</b>", getMapFieldStringValue(jobJSON, "job"))
			if errorMessage, ok := jobJSON["error_message"].(string); ok && errorMessage != "" {
				res.subtitle += fmt.Sprintf("<br>Error: %s", errorMessage)
			}
		} else {
			res.headerIconURL = successHeaderIconURL
		}
		return res
	}
}

// generateRequestBody returns the body of the request.
func generateRequestBody(m *messageBodyContent) ([]byte, error) {
	jsonData := map[string]any{
		"cardsV2": map[string]any{
			"cardId": "createCardMessage",
			"card": map[string]any{
				"header": map[string]any{
					"title":    m.title,
					"subtitle": m.subtitle,
					"imageUrl": m.headerIconURL,
				},
				"sections": []any{
					map[string]any{
						"collapsible":               true,
						"uncollapsibleWidgetsCount": 1,
						"widgets": []map[string]any{
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"iconUrl": widgetRefIconURL,
									},
									"text": fmt.Sprintf("<b>Repo: </b> %s", m.repo),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"iconUrl": widgetRefIconURL,
									},
									"text": fmt.Sprintf("<b>Ref: </b> %s", m.ref),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"knownIcon": "PERSON",
									},
									"text": fmt.Sprintf("<b>Actor: </b> %s", m.triggeringActor),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"knownIcon": "CLOCK",
									},
									"text": fmt.Sprintf("<b>UTC: </b> %s", m.timestamp),
								},
							},
							{
								"buttonList": map[string]any{
									"buttons": []any{
										map[string]any{
											"text": fmt.Sprintf("Open %s", m.eventName),
											"onClick": map[string]any{
												"openLink": map[string]any{
													"url": m.clickURL,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	res, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error marshal jsonData: %w", err)
	}
	return res, nil
}
