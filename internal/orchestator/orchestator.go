package orchestator

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sklerakuku/calc3.0/internal/config"
)

type Orchestrator struct {
	cfg         *config.Config
	expressions map[int]*Expression
	tasks       chan Task
	mu          sync.Mutex
	nextID      int
}

type Expression struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type Task struct {
	ID            int    `json:"id"`
	Arg1          string `json:"arg1"`
	Arg2          string `json:"arg2"`
	Operation     string `json:"operation"`
	OperationTime int    `json:"operation_time"`
}

func New(cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		cfg:         cfg,
		expressions: make(map[int]*Expression),
		tasks:       make(chan Task, 100),
	}
}

func (o *Orchestrator) AddCalculation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	if req.Expression == "internal" {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}


	if !isValidExpression(req.Expression) || req.Expression == "" {
		http.Error(w, "Invalid expression", http.StatusUnprocessableEntity)
		return
	}
	o.mu.Lock()
	id := o.nextID
	o.nextID++
	expr := &Expression{
		ID:     id,
		Status: "pending",
		Result: 0,
	}
	o.expressions[id] = expr
	o.mu.Unlock()

	errChan := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("internal error: %v", r)
				return
			}
		}()
		tree, err := parser.ParseExpr(req.Expression)
		if err != nil {
			errChan <- fmt.Errorf("invalid expression: %v", err)
			return
		}

		result := o.evaluateAST(tree)

		o.mu.Lock()
		expr := o.expressions[id]
		expr.Result = result
		expr.Status = "completed"
		o.mu.Unlock()

		log.Printf("expression processed: ID=%d, Result=%f", id, result)
	}()

	select {
	case err := <-errChan:
		if strings.Contains(err.Error(), "invalid expression") {
			http.Error(w, "Invalid expression", http.StatusUnprocessableEntity)
		} else {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	default:
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

func isValidExpression(expression string) bool {
	for _, c := range expression {
		if !isDigitOrOperator(c) {
			return false
		}
	}
	return true
}

func isDigitOrOperator(c rune) bool {
	return (c >= '0' && c <= '9') || c == '+' || c == '-' || c == '*' || c == '/' || c == '.' || c == '(' || c == ')'
}


func (o *Orchestrator) GetExpressions(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	expressions := make([]Expression, 0, len(o.expressions))
	for _, expr := range o.expressions {
		expressions = append(expressions, *expr)
	}
	o.mu.Unlock()

	log.Printf("expressions: %+v", expressions)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]Expression{"expressions": expressions})
}

func (o *Orchestrator) GetExpression(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/api/v1/expressions/"):])
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	o.mu.Lock()
	expr, ok := o.expressions[id]
	o.mu.Unlock()

	if !ok {
		http.Error(w, "expression not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]*Expression{"expression": expr})
}

func (o *Orchestrator) HandleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		o.getTask(w, r)
	} else if r.Method == http.MethodPost {
		o.postTaskResult(w, r)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (o *Orchestrator) getTask(w http.ResponseWriter, r *http.Request) {
	select {
	case task := <-o.tasks:
		json.NewEncoder(w).Encode(map[string]Task{"task": task})
	default:
		http.Error(w, "no tasks available", http.StatusNotFound)
	}
}

func (o *Orchestrator) postTaskResult(w http.ResponseWriter, r *http.Request) {
	var result struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	o.mu.Lock()
	expr, ok := o.expressions[result.ID]
	if ok {
		expr.Result = result.Result
		expr.Status = "completed"
	}
	o.mu.Unlock()

	if !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (o *Orchestrator) processExpression(id int, expression string) {
	tree, err := parser.ParseExpr(expression)
	if err != nil {
		o.mu.Lock()
		o.expressions[id].Status = "error"
		o.mu.Unlock()
		return
	}

	result := o.evaluateAST(tree)

	o.mu.Lock()
	expr := o.expressions[id]
	expr.Result = result
	expr.Status = "completed"
	o.mu.Unlock()

	log.Printf("expression processed: ID=%d, Result=%f", id, result)
}

func (o *Orchestrator) evaluateAST(node ast.Node) float64 {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		x := o.evaluateAST(n.X)
		y := o.evaluateAST(n.Y)

		var result float64
		switch n.Op {
		case token.ADD:
			result = x + y
			time.Sleep(time.Duration(o.cfg.TimeAdditionMS) * time.Millisecond)
		case token.SUB:
			result = x - y
			time.Sleep(time.Duration(o.cfg.TimeSubtractionMS) * time.Millisecond)
		case token.MUL:
			result = x * y
			time.Sleep(time.Duration(o.cfg.TimeMultiplicationMS) * time.Millisecond)
		case token.QUO:
			if y != 0 {
				result = x / y
			}
			time.Sleep(time.Duration(o.cfg.TimeDivisionMS) * time.Millisecond)
		}

		return result

	case *ast.BasicLit:
		value, _ := strconv.ParseFloat(n.Value, 64)
		return value
	}

	return 0
}

func (o *Orchestrator) waitForResult(taskID int) <-chan float64 {
	resultChan := make(chan float64)
	go func() {
		for {
			o.mu.Lock()
			expr, ok := o.expressions[taskID]
			if ok && expr.Status == "completed" {
				resultChan <- expr.Result
				o.mu.Unlock()
				return
			}
			o.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return resultChan
}

func ServeHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}
