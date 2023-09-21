package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	links_collection = "links"
)

func main() {
	docsRepoGHWebhookToken := os.Getenv("DOCS_REPO_GH_WEBHOOK_TOKEN")
	docsRepoGHWebhookUrl := os.Getenv("DOCS_REPO_GH_WEBHOOK_URL")
	app := pocketbase.New()

	app.OnRecordAfterCreateRequest(links_collection).Add(func(e *core.RecordCreateEvent) error {
		triggerGitHubWorkflow(docsRepoGHWebhookUrl, docsRepoGHWebhookToken)
		return nil
	})

	app.OnRecordAfterUpdateRequest(links_collection).Add(func(e *core.RecordUpdateEvent) error {
		triggerGitHubWorkflow(docsRepoGHWebhookUrl, docsRepoGHWebhookToken)
		return nil
	})

	app.OnRecordAfterDeleteRequest(links_collection).Add(func(e *core.RecordDeleteEvent) error {
		triggerGitHubWorkflow(docsRepoGHWebhookUrl, docsRepoGHWebhookToken)
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func triggerGitHubWorkflow(url, token string) {
	if url == "" || token == "" {
		return
	}

	requestBody := []byte(`{"event_type": "webhook"}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
}
