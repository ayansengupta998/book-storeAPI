package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/subosito/gotenv"

	"github.com/gorilla/mux"
)

//Book - struct to keep records of books
type Book struct {
	ID     int    `json:id`
	Title  string `json:title`
	Author string `json:author`
	Year   string `json:year`
}

//Books - slice of book struct
var Books []Book

var db *sql.DB

func init() {
	gotenv.Load()
	fmt.Println("entered init")
}
func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	var err error

	db, err = sql.Open("postgres", os.Getenv("ELEPHANTSQL_URL"))
	logFatal(err)
	err = db.Ping()
	logFatal(err)

	var router = mux.NewRouter()

	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books", addBook).Methods("POST")
	router.HandleFunc("/books", updateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", removeBook).Methods("DELETE")

	http.ListenAndServe(":8080", router)
}

func getBooks(w http.ResponseWriter, r *http.Request) {

	var book Book
	books := []Book{}

	rows, err := db.Query("select * from books")
	logFatal(err)

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
		logFatal(err)
		books = append(books, book)
	}
	json.NewEncoder(w).Encode(books)

}
func getBook(w http.ResponseWriter, r *http.Request) {

	var parms = mux.Vars(r)
	var book Book

	rows := db.QueryRow("select * from books where id=$1", parms["id"])

	err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
	logFatal(err)
	json.NewEncoder(w).Encode(book)

}
func addBook(w http.ResponseWriter, r *http.Request) {

	var book Book
	var bookID int

	json.NewDecoder(r.Body).Decode(&book)

	err := db.QueryRow("insert into books(title,author,year) values($1, $2, $3) RETURNING id;", book.Title, book.Author, book.Year).Scan(&bookID)
	logFatal(err)
	json.NewEncoder(w).Encode(bookID)

}
func updateBook(w http.ResponseWriter, r *http.Request) {

	var book Book

	json.NewDecoder(r.Body).Decode(&book)
	sqlStatement := `UPDATE books
	SET title=$2,author=$3,year=$4
	WHERE id =$1 
	RETURNING id`
	res, err := db.Exec(sqlStatement, book.ID, book.Title, book.Author, book.Year)
	logFatal(err)

	rowsUp, err := res.RowsAffected()

	logFatal(err)

	json.NewEncoder(w).Encode(rowsUp)

}
func removeBook(w http.ResponseWriter, r *http.Request) {

	var params = mux.Vars(r)

	rows, err := db.Exec("delete from books where id = $1", params["id"])

	rowsUp, err := rows.RowsAffected()

	logFatal(err)

	json.NewEncoder(w).Encode(rowsUp)

}
