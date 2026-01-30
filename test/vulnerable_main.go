package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

// Vulnerability 1: Hardcoded AWS Credentials
const (
	AwsAccessKey = "AKIAIOSFODNN7EXAMPLE"
	AwsSecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
)

func main() {
	http.HandleFunc("/login", loginHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	db, err := sql.Open("postgres", "user=postgres password=secret dbname=mydb sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// Vulnerability 2: SQL Injection
	// User input is directly concatenated into the query string
	query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	defer rows.Close()

	fmt.Fprintf(w, "Logged in as %s", username)
}
