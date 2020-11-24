package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type (
	// TodoModel struct for each todo
	TodoModel struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}
)

const (
	dbDsn = "postgres://mxljoviftxwzhv:bdecd107e2fde2a5fb11b07dcf8f4ea03c15850fb9e8ca804869e62a287ee8aa@ec2-54-247-122-209.eu-west-1.compute.amazonaws.com:5432/dfoigvsgfqvugv"
)

var db *sql.DB
var err error

// Home Function to check if api is working or not
func Home(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"isAlive": true}`)
}

func init() {

	db, err = sql.Open("postgres", dbDsn)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to DB")
	// db.AutoMigrate(&TodoModel{})

}

// CreateTodo func to create a todo
func CreateTodo(w http.ResponseWriter, r *http.Request) {

	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")

	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Need content type application/json but got '%s'", ct)))
		return
	}

	var t TodoModel
	err = json.Unmarshal(bodyBytes, &t)
	fmt.Println(t.Completed)
	t.ID = rand.Intn(100000)
	insertStatement := `INSERT INTO todo (ID, Title, Completed) Values ($1, $2, $3);`
	_, err = db.Exec(insertStatement, t.ID, t.Title, t.Completed)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"CreateTodo": true}`)

}

//GetTodo function to get all the todos present in the database
func GetTodo(w http.ResponseWriter, r *http.Request) {

	var todoSlice []TodoModel
	getStatement := "select * from todo"
	data, err := db.Query(getStatement)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for data.Next() {

		var t TodoModel
		err = data.Scan(&t.ID, &t.Title, &t.Completed)
		if err != nil {
			fmt.Sprintf("Error in data")
			return
		}

		todoSlice = append(todoSlice, t)
	}

	jsonBytes, err := json.Marshal(todoSlice)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}

//UpdateTodo Handler to update todo task need to send the updating id via the body
func UpdateTodo(w http.ResponseWriter, r *http.Request) {

	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")

	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Need content type application/json but got '%s'", ct)))
		return
	}

	var t TodoModel
	err = json.Unmarshal(bodyBytes, &t)
	// params := mux.Vars(r)
	// id, _ := params["ID"]
	id := t.ID
	// t.ID = rand.Intn(100000)
	insertStatement := `UPDATE todo SET COMPLETED = $1 WHERE ID = $2;`
	res, err := db.Exec(insertStatement, t.Completed, id)
	// fmt.Println(t.Completed)
	// fmt.Println(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// io.WriteString(w, `{"RowsUpdated": count}`)

}

//DeleteTodo func to remove the object from the db
func DeleteTodo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(err.Error()))
		return
	}

	deleteQuery := `DELETE FROM TODO WHERE id = $1;`
	res, err := db.Exec(deleteQuery, id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
	w.WriteHeader(http.StatusOK)

}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/home", Home).Methods("GET")
	router.HandleFunc("/todo", CreateTodo).Methods("POST")
	router.HandleFunc("/getAll", GetTodo).Methods("GET")
	router.HandleFunc("/update/{id}", UpdateTodo).Methods("PUT")
	router.HandleFunc("/delete/{id}", DeleteTodo).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))

}
