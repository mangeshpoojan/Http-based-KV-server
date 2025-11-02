package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"strconv"

	_ "github.com/go-sql-driver/mysql"

	// "net"
	"net/http"
	"os"
)

type KV struct {
	Key   int    `json:"key"`
	Value string `json:"value"`
}

type KV_resp struct {
	Status int    `json:"status"`
	Err    string `json:"err,omitempty"`
	Key    int    `json:"key"`
	Data   string `json:"data,omitempty"`
}

var db *sql.DB

func fetch(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		res.Header().Set(
			"Content-Type",
			"application/json",
		)

		var key int
		var err error
		key_to_send := req.URL.Query().Get("key")
		if key, err = strconv.Atoi(key_to_send); err != nil {
			panic(err)
		}

		// default KV_response setup

		resp := KV_resp{
			Status: 200,
			Key:    key,
		}

		// response me data likho

		row := db.QueryRow("SELECT * FROM kv_store WHERE key_col = ?", key)

		var key_from_db int
		var val_from_db string

		if err = row.Scan(&key_from_db, &val_from_db); err == sql.ErrNoRows {

			// key doesn't exist
			resp.Err = "Key doesn't Exists"

		} else if err == nil {

			// write the data into response json
			resp.Data = val_from_db

		} else {
			panic(err)
		}

		// res_data := "dummy data" + key_to_send

		// send the response back to the source
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(resp)

		// fmt.Fprintf(res, res_data)
	} else {
		http.Error(res, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
}
func del(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		res.Header().Set(
			"Content-Type",
			"application/json",
		)

		var key int
		var err error
		key_to_send := req.URL.Query().Get("key")
		if key, err = strconv.Atoi(key_to_send); err != nil {
			panic(err)
		}

		// default KV_response setup

		resp := KV_resp{
			Status: 200,
			Key:    key,
		}

		// delete the key

		if _, err = db.Exec("DELETE FROM kv_store WHERE key_col = ?", key); err != nil {
			panic(err)
		}
		// res_data := "dummy data" + key_to_send

		// send the response back to the source

		// fmt.Fprintf(res, res_data)
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(resp)

	} else {
		http.Error(res, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
}

func create(res http.ResponseWriter, req *http.Request) {

	res.Header().Set(
		"Content-Type",
		"application/json",
	)

	if req.Method == http.MethodPost {

		// read the data from the json

		var data KV
		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			http.Error(res, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// default KV_response setup

		resp := KV_resp{
			Status: 200,
			Key:    data.Key,
		}

		// create the KV

		row := db.QueryRow("SELECT * FROM kv_store WHERE key_col = ?", data.Key)

		var dummy1, dummy2 any

		if err = row.Scan(&dummy1, &dummy2); err == sql.ErrNoRows {

			// insert the data
			if _, err = db.Exec("INSERT INTO kv_store (key_col, value_col) values (?, ?)", data.Key, data.Value); err != nil {
				panic(err)
			}

		} else if err == nil {

			// send the error to the client that the key exists
			resp.Err = "Key already Exists"

		} else {
			panic(err)
		}

		// send the response
		// res_data := "key created"

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(resp)
		// fmt.Fprintf(res, res_data)

	}
}

func update(res http.ResponseWriter, req *http.Request) {

	res.Header().Set(
		"Content-Type",
		"application/json",
	)

	if req.Method == http.MethodPost {

		// read the data from the json

		var data KV
		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			http.Error(res, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// default KV_response setup

		resp := KV_resp{
			Status: 200,
			Key:    data.Key,
		}

		// update the KV

		row := db.QueryRow("SELECT * FROM kv_store WHERE key_col = ?", data.Key)

		var dummy1, dummy2 any

		if err = row.Scan(&dummy1, &dummy2); err == sql.ErrNoRows {

			// send the error to the client that the key doesn't exists
			resp.Err = "Key doesn't exists"

		} else if err == nil {

			// Update the data
			if _, err = db.Exec("UPDATE kv_store SET value_col = ? WHERE key_col = ?", data.Value, data.Key); err != nil {
				panic(err)
			}

		} else {
			panic(err)
		}

		// send the response
		// res_data := "key updated" + strconv.Itoa(data.Key) + data.Value

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(resp)
		// fmt.Fprintf(res, res_data)

	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("<file_name> <port_no>")
		return
	}

	// database connectivity

	data_src_name := "admin:admin123@tcp(127.0.0.1:3306)/KV_store"
	var err error
	db, err = sql.Open("mysql", data_src_name)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to MYSQL KV_store")

	createTable := `
    CREATE TABLE IF NOT EXISTS kv_store (
        key_col VARCHAR(255) PRIMARY KEY,
		value_col VARCHAR(500)
	);`

	if _, err := db.Exec(createTable); err != nil {
		panic(err)
	}

	// http routes

	http.HandleFunc("/create", create)
	http.HandleFunc("/update", update)
	http.HandleFunc("/read", fetch)
	http.HandleFunc("/delete", del)

	http.ListenAndServe(os.Args[1], nil)
}
