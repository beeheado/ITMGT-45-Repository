package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// types

type Product struct {
	Id          int
	Name        string
	Price       int
	Description string
}

type User struct {
	Id       int
	Username string
	Password string
}

type Session struct {
	Token  string
	UserId int
}

type CartItem struct {
	Id          int
	UserId      int
	ProductId   int
	Quantity    int
	ProductName string
}

type Transaction struct {
	Id        int
	UserId    int
	CreatedAt time.Time
}

type TransactionLine struct {
	ProductName string
	Quantity    int
}

var database *sql.DB

func initDB() {
	db, err := sql.Open("sqlite3", "./db")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	database = db

	queries := []string{
		`CREATE TABLE IF NOT EXISTS cgo_user (username TEXT, password TEXT)`,
		`CREATE TABLE IF NOT EXISTS cgo_product (name TEXT, price INTEGER, description TEXT)`,
		`CREATE TABLE IF NOT EXISTS cgo_session (token TEXT, user_id INTEGER)`,
		`CREATE TABLE IF NOT EXISTS cgo_cart_item (product_id INTEGER, quantity INTEGER, user_id INTEGER)`,
		`CREATE TABLE IF NOT EXISTS cgo_transaction (user_id INTEGER, created_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS cgo_line_item (transaction_id INTEGER, product_id INTEGER, quantity INTEGER)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatal(err)
		}
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM cgo_user`).Scan(&count); err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		users := []User{
			{Username: "zagreus", Password: "cerberus"},
			{Username: "melinoe", Password: "b4d3ec1"},
		}
		for _, u := range users {
			if _, err := db.Exec(`INSERT INTO cgo_user (username, password) VALUES (?, ?)`, u.Username, u.Password); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM cgo_product`).Scan(&count); err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		products := []Product{
			{Name: "Americano", Price: 100, Description: "Espresso, diluted for a lighter experience"},
			{Name: "Cappuccino", Price: 110, Description: "Espresso with steamed milk"},
			{Name: "Espresso", Price: 90, Description: "A strong shot of coffee"},
			{Name: "Macchiato", Price: 120, Description: "Espresso with a small amount of milk"},
		}
		for _, p := range products {
			if _, err := db.Exec(`INSERT INTO cgo_product (name, price, description) VALUES (?, ?, ?)`, p.Name, p.Price, p.Description); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func getUsers() []User {
	rows, err := database.Query(`SELECT rowid, username, password FROM cgo_user`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Username, &u.Password); err != nil {
			log.Fatal(err)
		}
		result = append(result, u)
	}
	return result
}

func getProducts() []Product {
	rows, err := database.Query(`SELECT rowid, name, price, description FROM cgo_product`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var result []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.Id, &p.Name, &p.Price, &p.Description); err != nil {
			log.Fatal(err)
		}
		result = append(result, p)
	}
	return result
}

func setSession(token string, user User) {
	if _, err := database.Exec(`INSERT INTO cgo_session (token, user_id) VALUES (?, ?)`, token, user.Id); err != nil {
		log.Fatal(err)
	}
}

func getUserFromSessionToken(token string) User {
	const q = `
	SELECT
		cgo_session.user_id,
		cgo_user.username,
		cgo_user.password
	FROM cgo_session
	INNER JOIN cgo_user ON cgo_session.user_id = cgo_user.rowid
	WHERE cgo_session.token = ?
	LIMIT 1;`
	var u User
	err := database.QueryRow(q, token).Scan(&u.Id, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return User{}
	} else if err != nil {
		log.Fatal(err)
	}
	return u
}

func createCartItem(userId int, productId int, quantity int) {
	if _, err := database.Exec(`INSERT INTO cgo_cart_item (user_id, product_id, quantity) VALUES (?, ?, ?)`, userId, productId, quantity); err != nil {
		log.Fatal(err)
	}
}

func getCartItemsByUser(user User) []CartItem {
	q := `
	SELECT
		cgo_cart_item.rowid,
		cgo_cart_item.user_id,
		cgo_cart_item.product_id,
		cgo_cart_item.quantity,
		cgo_product.name
	FROM cgo_cart_item
	LEFT JOIN cgo_product ON cgo_cart_item.product_id = cgo_product.rowid
	WHERE cgo_cart_item.user_id = ?`
	rows, err := database.Query(q, user.Id)
	if err == sql.ErrNoRows {
		return []CartItem{}
	} else if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var result []CartItem
	for rows.Next() {
		var ci CartItem
		if err := rows.Scan(&ci.Id, &ci.UserId, &ci.ProductId, &ci.Quantity, &ci.ProductName); err != nil {
			log.Fatal(err)
		}
		result = append(result, ci)
	}
	return result
}

// checkout stuff
func checkoutItemsForUser(user User) {
	tx, err := database.Begin()
	if err != nil {
		log.Fatal(err)
	}

	res, err := tx.Exec(`INSERT INTO cgo_transaction (user_id, created_at) VALUES (?, ?)`,
		user.Id, time.Now().Format(time.RFC3339))
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	tid, _ := res.LastInsertId()

	rows, err := tx.Query(`SELECT product_id, quantity FROM cgo_cart_item WHERE user_id = ?`, user.Id)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var pid, qty int
		if err := rows.Scan(&pid, &qty); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
		if _, err := tx.Exec(`INSERT INTO cgo_line_item (transaction_id, product_id, quantity) VALUES (?, ?, ?)`,
			tid, pid, qty); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}

	if _, err := tx.Exec(`DELETE FROM cgo_cart_item WHERE user_id = ?`, user.Id); err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

// transaction history
func getTransactionsByUser(user User) []struct {
	Id        int
	CreatedAt string
	Lines     []TransactionLine
} {
	q := `SELECT rowid, created_at FROM cgo_transaction WHERE user_id = ? ORDER BY rowid DESC`
	rows, err := database.Query(q, user.Id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var result []struct {
		Id        int
		CreatedAt string
		Lines     []TransactionLine
	}

	for rows.Next() {
		var id int
		var created string
		if err := rows.Scan(&id, &created); err != nil {
			log.Fatal(err)
		}

		lineRows, err := database.Query(`
			SELECT cgo_product.name, cgo_line_item.quantity
			FROM cgo_line_item
			JOIN cgo_product ON cgo_line_item.product_id = cgo_product.rowid
			WHERE cgo_line_item.transaction_id = ?`, id)
		if err != nil {
			log.Fatal(err)
		}

		var lines []TransactionLine
		for lineRows.Next() {
			var tl TransactionLine
			if err := lineRows.Scan(&tl.ProductName, &tl.Quantity); err != nil {
				log.Fatal(err)
			}
			lines = append(lines, tl)
		}
		lineRows.Close()

		result = append(result, struct {
			Id        int
			CreatedAt string
			Lines     []TransactionLine
		}{id, created, lines})
	}
	return result
}
