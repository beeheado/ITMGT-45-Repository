package main

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func generateSessionToken() string {
	raw := make([]byte, 16)
	_, _ = rand.Read(raw)
	return base64.StdEncoding.EncodeToString(raw)
}

// handlers

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	var token string
	for _, c := range r.Cookies() {
		if c.Name == "cafego_session" {
			token = c.Value
			break
		}
	}
	user := getUserFromSessionToken(token)
	tmpl, _ := template.ParseFiles("./templates/index.html")
	tmpl.Execute(w, map[string]interface{}{
		"Username": user.Username,
		"Products": getProducts(),
	})
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var product Product
	for _, p := range getProducts() {
		if p.Id == id {
			product = p
			break
		}
	}
	if product == (Product{}) {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("./templates/product.html")
		tmpl.Execute(w, product)
		return
	}

	if r.Method == http.MethodPost {
		var token string
		for _, c := range r.Cookies() {
			if c.Name == "cafego_session" {
				token = c.Value
				break
			}
		}
		user := getUserFromSessionToken(token)
		if user == (User{}) {
			http.Error(w, "Please log in first.", http.StatusUnauthorized)
			return
		}

		qty, _ := strconv.Atoi(r.FormValue("quantity"))
		if qty > 0 {
			createCartItem(user.Id, product.Id, qty)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tmpl, _ := template.ParseFiles("./templates/login.html")
		tmpl.Execute(w, nil)
	case http.MethodPost:
		username := r.FormValue("username")
		password := r.FormValue("password")

		var user User
		for _, u := range getUsers() {
			if u.Username == username && u.Password == password {
				user = u
				break
			}
		}
		if user == (User{}) {
			http.Error(w, "Invalid login", http.StatusUnauthorized)
			return
		}
		token := generateSessionToken()
		setSession(token, user)
		http.SetCookie(w, &http.Cookie{Name: "cafego_session", Value: token, Path: "/"})
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func cartHandler(w http.ResponseWriter, r *http.Request) {
	var token string
	for _, c := range r.Cookies() {
		if c.Name == "cafego_session" {
			token = c.Value
			break
		}
	}
	user := getUserFromSessionToken(token)
	if user == (User{}) {
		http.Error(w, "Please log in first.", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("./templates/cart.html")
		tmpl.Execute(w, map[string]interface{}{
			"User":      user,
			"CartItems": getCartItemsByUser(user),
		})
	} else if r.Method == http.MethodPost {
		checkoutItemsForUser(user)
		http.Redirect(w, r, "/cart/confirmation", http.StatusFound)
	}
}

func confirmationHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("./templates/confirmation.html")
	tmpl.Execute(w, nil)
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	var token string
	for _, c := range r.Cookies() {
		if c.Name == "cafego_session" {
			token = c.Value
			break
		}
	}
	user := getUserFromSessionToken(token)
	if user == (User{}) {
		http.Error(w, "Please log in first.", http.StatusUnauthorized)
		return
	}

	history := getTransactionsByUser(user)
	tmpl, _ := template.ParseFiles("./templates/transactions.html")
	tmpl.Execute(w, map[string]interface{}{
		"User":         user,
		"Transactions": history,
	})
}

// main
func main() {
	initDB()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/product/", productHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/cart/", cartHandler)
	http.HandleFunc("/cart/confirmation", confirmationHandler)
	http.HandleFunc("/transactions/", transactionsHandler)

	log.Println("Server running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
