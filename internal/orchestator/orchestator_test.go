package orchestator

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sklerakuku/calc3.0/internal/config"
)

func TestAddCalculation(t *testing.T) {
	cfg := &config.Config{
		TimeAdditionMS:       100,
		TimeSubtractionMS:    100,
		TimeMultiplicationMS: 100,
		TimeDivisionMS:       100,
	}
	o := New(cfg)

	reqBody := map[string]string{"expression": "2 + 3"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/expressions", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	o.AddCalculation(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response map[string]int
	json.Unmarshal(rr.Body.Bytes(), &response)
	if _, exists := response["id"]; !exists {
		t.Errorf("response does not contain 'id' field")
	}
}

func TestGetExpressions(t *testing.T) {
	cfg := &config.Config{}
	o := New(cfg)

	o.expressions[1] = &Expression{ID: 1, Status: "completed", Result: 5}

	req, _ := http.NewRequest("GET", "/api/v1/expressions", nil)
	rr := httptest.NewRecorder()

	o.GetExpressions(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string][]Expression
	json.Unmarshal(rr.Body.Bytes(), &response)
	if len(response["expressions"]) != 1 {
		t.Errorf("expected 1 expression, got %d", len(response["expressions"]))
	}
}

func TestGetExpression(t *testing.T) {
	cfg := &config.Config{}
	o := New(cfg)

	o.expressions[1] = &Expression{ID: 1, Status: "completed", Result: 5}

	req, _ := http.NewRequest("GET", "/api/v1/expressions/1", nil)
	rr := httptest.NewRecorder()

	o.GetExpression(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]*Expression
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response["expression"].ID != 1 {
		t.Errorf("expected expression with ID 1, got %d", response["expression"].ID)
	}
}

func TestHandleTask(t *testing.T) {
	cfg := &config.Config{}
	o := New(cfg)

	t.Run("GetTask", func(t *testing.T) {
		task := Task{ID: 1, Arg1: "2", Arg2: "3", Operation: "+", OperationTime: 100}
		o.tasks <- task

		req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
		rr := httptest.NewRecorder()

		o.HandleTask(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}

		var response map[string]Task
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response["task"].ID != 1 {
			t.Errorf("expected task with ID 1, got %d", response["task"].ID)
		}
	})

	t.Run("PostTaskResult", func(t *testing.T) {
		o.expressions[1] = &Expression{ID: 1, Status: "pending"}

		reqBody := map[string]interface{}{"id": 1, "result": 5.0}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		o.HandleTask(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}

		if o.expressions[1].Status != "completed" || o.expressions[1].Result != 5.0 {
			t.Errorf("task result not updated correctly")
		}
	})
}

func TestProcessExpression(t *testing.T) {
	cfg := &config.Config{}
	o := New(cfg)

	o.processExpression(1, "2 + 3")

	expr, ok := o.expressions[1]
	if !ok {
		t.Fatalf("expression not found")
	}

	if expr.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", expr.Status)
	}

	if expr.Result != 5 {
		t.Errorf("expected result 5, got %f", expr.Result)
	}
}

func TestEvaluateAST(t *testing.T) {
	cfg := &config.Config{
		TimeAdditionMS:       100,
		TimeSubtractionMS:    100,
		TimeMultiplicationMS: 100,
		TimeDivisionMS:       100,
	}
	o := New(cfg)

	tests := []struct {
		name       string
		expression string
		want       float64
	}{
		{"Addition", "2 + 3", 5},
		{"Subtraction", "5 - 3", 2},
		{"Multiplication", "4 * 3", 12},
		{"Division", "10 / 2", 5},
		{"Complex expression", "2 * 3 + 4 / 2", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := parser.ParseExpr(tt.expression)
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}
			got := o.evaluateAST(tree)
			if got != tt.want {
				t.Errorf("evaluateAST(%s) = %v, want %v", tt.expression, got, tt.want)
			}
		})
	}
}

func TestWaitForResult(t *testing.T) {
	cfg := &config.Config{}
	o := New(cfg)

	taskID := 1
	o.expressions[taskID] = &Expression{ID: taskID, Status: "pending"}

	resultChan := o.waitForResult(taskID)

	go func() {
		time.Sleep(100 * time.Millisecond)
		o.mu.Lock()
		o.expressions[taskID].Status = "completed"
		o.expressions[taskID].Result = 42
		o.mu.Unlock()
	}()

	select {
	case result := <-resultChan:
		if result != 42 {
			t.Errorf("expected result 42, got %f", result)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("timeout waiting for result")
	}
}

func TestServeHTML(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeHTML)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/html; charset=utf-8")
	}
}
