package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type resposta struct {
	Operacao  string  `json:"operacao,omitempty"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
	Resultado float64 `json:"resultado"`
}

type erro struct {
	Erro string `json:"erro"`
}

func escreverJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func calcular(op string, a, b float64) (float64, error) {
	switch op {
	case "soma":
		return a + b, nil
	case "sub":
		return a - b, nil
	case "mult":
		return a * b, nil
	case "div":
		if b == 0 {
			return 0, fmt.Errorf("divisão por zero")
		}
		return a / b, nil
	case "pot":
		return math.Pow(a, b), nil
	default:
		return 0, fmt.Errorf("operação inválida: use soma, sub, mult, div ou pot")
	}
}

func handlerCalc(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	op := q.Get("op")

	a, errA := strconv.ParseFloat(q.Get("a"), 64)
	b, errB := strconv.ParseFloat(q.Get("b"), 64)
	if errA != nil || errB != nil {
		escreverJSON(w, http.StatusBadRequest, erro{"parâmetros 'a' e 'b' devem ser números"})
		return
	}

	res, err := calcular(op, a, b)
	if err != nil {
		escreverJSON(w, http.StatusBadRequest, erro{err.Error()})
		return
	}
	log.Printf("calc op=%s a=%v b=%v resultado=%v", op, a, b, res)
	escreverJSON(w, http.StatusOK, resposta{op, a, b, res})
}

func handlerRaiz(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Calculadora Docker")
	fmt.Fprintln(w, "Uso: /calc?op=<soma|sub|mult|div|pot>&a=<numero>&b=<numero>")
	fmt.Fprintln(w, "Exemplo: /calc?op=mult&a=6&b=7")
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	escreverJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// healthcheck faz uma requisição ao próprio servidor e sai com código 0 ou 1.
func healthcheck(porta string) {
	cliente := http.Client{Timeout: 2 * time.Second}
	resp, err := cliente.Get("http://127.0.0.1:" + porta + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	check := flag.Bool("healthcheck", false, "verifica se o servidor está saudável e sai")
	flag.Parse()
	if *check {
		healthcheck(porta)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerRaiz)
	mux.HandleFunc("/calc", handlerCalc)
	mux.HandleFunc("/health", handlerHealth)

	srv := &http.Server{
		Addr:              ":" + porta,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Calculadora ouvindo na porta %s", porta)
	log.Fatal(srv.ListenAndServe())
}
