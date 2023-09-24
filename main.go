package main

import (
	"bytes"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Task struct {
	ID     int
	Action func()
}

const (
	links_collection = "links"
	links_task_id    = 1
)

var mutex sync.Mutex
var taskQueue = make(map[int]*Task)

func main() {
	docsRepoGHWebhookToken := os.Getenv("DOCS_REPO_GH_WEBHOOK_TOKEN")
	docsRepoGHWebhookUrl := os.Getenv("DOCS_REPO_GH_WEBHOOK_URL")
	app := pocketbase.New()

	app.OnRecordAfterCreateRequest(links_collection).Add(func(e *core.RecordCreateEvent) error {
		scheduleTask(links_task_id, func() { triggerGitHubWorkflow(docsRepoGHWebhookUrl, docsRepoGHWebhookToken) })

		return nil
	})

	app.OnRecordAfterUpdateRequest(links_collection).Add(func(e *core.RecordUpdateEvent) error {
		scheduleTask(links_task_id, func() { triggerGitHubWorkflow(docsRepoGHWebhookUrl, docsRepoGHWebhookToken) })

		return nil
	})

	app.OnRecordAfterDeleteRequest(links_collection).Add(func(e *core.RecordDeleteEvent) error {
		scheduleTask(links_task_id, func() { triggerGitHubWorkflow(docsRepoGHWebhookUrl, docsRepoGHWebhookToken) })

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
		slog.Error("Error creating request:", "error", err)
		return
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending request:", "error", err)
		return
	}
	defer resp.Body.Close()
}

func scheduleTask(taskID int, taskFunc func()) {
	duration := 5 * time.Minute
	mutex.Lock()
	defer mutex.Unlock()

	if existingTask, exists := taskQueue[taskID]; exists {
		slog.Info("Task already exists, resetting the timer:", "taskID", taskID)
		existingTask.Action = taskFunc
	} else {
		slog.Info("Creating new task:", "taskID", taskID)
		task := &Task{ID: taskID, Action: taskFunc}
		taskQueue[taskID] = task

		go func() {
			<-time.After(duration)
			slog.Info("Running task:", "taskID", taskID)
			task.Action()
			mutex.Lock()
			defer mutex.Unlock()
			delete(taskQueue, taskID)
		}()
	}
}
