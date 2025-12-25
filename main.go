package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"guide/templates"
)


func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/sse", handleSSE)
	http.HandleFunc("/trigger-toast", handleTriggerToast)
	http.HandleFunc("/spam-toasts", handleSpamToasts)
	http.HandleFunc("/random-quote", handleRandomQuote)
	http.HandleFunc("/delete-item", handleDeleteItem)
	http.HandleFunc("/form-submit", handleFormSubmit)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	templates.Index().Render(r.Context(), w)
}

func handleTriggerToast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	toastType := r.FormValue("type")
	if toastType == "" {
		toastType = "success"
	}

	messages := map[string]string{
		"success": "Operation completed successfully!",
		"error":   "Something went wrong!",
		"info":    "Here's some information for you.",
		"warning": "Please be careful with this action.",
	}

	msg := messages[toastType]
	if msg == "" {
		msg = messages["info"]
	}

	toast := map[string]string{
		"type":    toastType,
		"message": msg,
	}
	data, _ := json.Marshal(toast)
	broadcast.Send(string(data))

	w.WriteHeader(http.StatusOK)
}

func handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock 2 second delay to show loading state
	time.Sleep(2 * time.Second)

	toast := map[string]string{
		"type":    "success",
		"message": "Item deleted successfully!",
	}
	data, _ := json.Marshal(toast)
	broadcast.Send(string(data))

	w.WriteHeader(http.StatusOK)
}

func handleFormSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	// Mock 2 second delay to show loading state
	time.Sleep(2 * time.Second)

	toast := map[string]string{
		"type":    "success",
		"message": "Form submitted! Name: " + name + ", Email: " + email,
	}
	data, _ := json.Marshal(toast)
	broadcast.Send(string(data))

	// Return empty response - toast handles feedback
	w.WriteHeader(http.StatusOK)
}

func handleSpamToasts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send 5 toasts with small delays to demo queue
	types := []string{"success", "error", "info", "warning", "success"}
	messages := []string{
		"First toast incoming!",
		"Oops, an error appeared!",
		"Here's some info for you.",
		"Warning: toast spam detected!",
		"And we're done!",
	}

	go func() {
		for i := 0; i < 5; i++ {
			toast := map[string]string{
				"type":    types[i],
				"message": messages[i],
			}
			data, _ := json.Marshal(toast)
			broadcast.Send(string(data))
			time.Sleep(300 * time.Millisecond)
		}
	}()

	w.WriteHeader(http.StatusOK)
}

var quotes = []string{
	"The best way to predict the future is to invent it. — Alan Kay",
	"Simplicity is the ultimate sophistication. — Leonardo da Vinci",
	"First, solve the problem. Then, write the code. — John Johnson",
	"Code is like humor. When you have to explain it, it's bad. — Cory House",
	"Make it work, make it right, make it fast. — Kent Beck",
	"Any fool can write code that a computer can understand. Good programmers write code that humans can understand. — Martin Fowler",
}

func handleRandomQuote(w http.ResponseWriter, r *http.Request) {
	quote := quotes[rand.Intn(len(quotes))]
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<blockquote class="text-lg italic">"%s"</blockquote>`, quote)
}
