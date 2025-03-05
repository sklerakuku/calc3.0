package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sklerakuku/calc3.0/internal/config"
	"github.com/sklerakuku/calc3.0/pkg/calculation"
)

type Task struct {
	ID            int    `json:"id"`
	Arg1          string `json:"arg1"`
	Arg2          string `json:"arg2"`
	Operation     string `json:"operation"`
	OperationTime int    `json:"operation_time"`
}

type Agent struct {
	cfg    *config.Config
	client *http.Client
}

func New(cfg *config.Config) *Agent {
	return &Agent{
		cfg:    cfg,
		client: &http.Client{},
	}
}

func (a *Agent) Start() {
	var wg sync.WaitGroup
	for i := 0; i < a.cfg.ComputingPower; i++ {
		wg.Add(1)
		go a.worker(&wg)
	}
	wg.Wait()
}

func (a *Agent) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		task, err := a.getTask()
		if err != nil {
			log.Printf("Error getting task: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if task == nil {
			time.Sleep(time.Second)
			continue
		}

		result := calculation.Calculate(task.Arg1, task.Arg2, task.Operation)
		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

		err = a.sendResult(task.ID, result)
		if err != nil {
			log.Printf("Error sending result: %v", err)
		}
	}
}

func (a *Agent) getTask() (*Task, error) {
	resp, err := a.client.Get("http://127.0.0.1:8080/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var task Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (a *Agent) sendResult(id int, result float64) error {
	data := map[string]interface{}{
		"id":     id,
		"result": result,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := a.client.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
