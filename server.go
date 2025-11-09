package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"strconv"

	"sync"

	_ "github.com/go-sql-driver/mysql"

	// "net"
	"net/http"
	"os"
)

// req data format
type KV struct {
	Key   int    `json:"key"`
	Value string `json:"value"`
}

// reply
type KV_resp struct {
	Status int    `json:"status"`
	Err    string `json:"err,omitempty"`
	Key    int    `json:"key"`
	Data   string `json:"data,omitempty"`
}

// DLL node
type Node struct {
	key        int
	prev, next *Node
}

// map entry
type entry struct {
	value string
	node  *Node
}

var lock sync.Mutex // lock variable

var db *sql.DB // sql connection

var log = 1 // log flag

var cache *LRUCache // cache

// Cache

type LRUCache struct {
	capacity int
	cache    map[int]*entry
	head     *Node // MRU
	tail     *Node // LRU
}

// Cache initializer
func Initializer(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*entry),
	}
}

// Cache Get
func (some *LRUCache) Get(key int) (string, int) {

	// lock now and unlock in case of error or when function exits

	lock.Lock()
	defer lock.Unlock()

	if e, exists := some.cache[key]; exists {
		some.moveToFront(e.node)
		return e.value, 1 // Found
	}
	return "", 0 // Not found
}

// Cache Set
func (some *LRUCache) Set(key int, value string) {

	// lock now and unlock in case of error or when function exits

	lock.Lock()
	defer lock.Unlock()

	// If key exists

	if e, exists := some.cache[key]; exists { // update value and move node to front
		e.value = value
		some.moveToFront(e.node)
		return
	}

	// If capacity hits

	if len(some.cache) == some.capacity { // remove least recently used so we can insert the new one
		lruKey := some.tail.key
		some.removeNode(some.tail)
		delete(some.cache, lruKey)

		// WB the data to DB
		if _, err := db.Exec("UPDATE kv_store SET value_col = ? WHERE key_col = ?", some.cache[some.tail.key], some.tail.key); err != nil {
			panic(err)
		}
		if log == 1 {
			fmt.Println("evicted :", some.tail.key)
		}
	}

	// If key Doesn't exists

	newNode := &Node{key: key} // Create new node for list
	some.addToFront(newNode)

	// Store value and pointer in map

	some.cache[key] = &entry{value: value, node: newNode}
}

// Cache Del
func (l *LRUCache) Del(key int) int {
	lock.Lock()
	defer lock.Unlock()

	if e, exists := l.cache[key]; exists {
		l.removeNode(e.node)
		delete(l.cache, key)
		return 1 // deleted
	}
	return 0 // not found
}

// Sub-Functions of cache

func (some *LRUCache) moveToFront(node *Node) {
	some.removeNode(node)
	some.addToFront(node)
}

func (some *LRUCache) addToFront(node *Node) {
	node.prev = nil
	node.next = some.head
	if some.head != nil {
		some.head.prev = node
	}
	some.head = node
	if some.tail == nil {
		some.tail = node
	}
}

func (some *LRUCache) removeNode(node *Node) {

	// remove the node
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		some.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		some.tail = node.prev
	}
}

// Print the Cache
func (some *LRUCache) PrintCache() {
	lock.Lock()
	defer lock.Unlock()

	fmt.Println("Cache State (MRU → LRU):")
	for n := some.head; n != nil; n = n.next {
		fmt.Printf("Key=%d, Value=%q → ", n.key, some.cache[n.key].value)
	}
	fmt.Println()
}

// CRUD with cache

func C_fetch(res http.ResponseWriter, req *http.Request) {

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

		val_from_cache, found := cache.Get(key)

		if found == 0 { // if not found

			// check the DB

			row := db.QueryRow("SELECT * FROM kv_store WHERE key_col = ?", key)

			var key_from_db int
			var val_from_db string

			if err = row.Scan(&key_from_db, &val_from_db); err == sql.ErrNoRows { // if not found in DB

				// key doesn't exist
				resp.Err = "Key doesn't Exists"

			} else if err == nil { // if found in DB

				// write the data into response json
				resp.Data = val_from_db

				// write the data in cache as well
				cache.Set(key, val_from_db)

			} else {
				panic(err)
			}

		} else { // if found

			// write the data into response json
			resp.Data = val_from_cache
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

	if log == 1 {
		cache.PrintCache()
	}
}
func C_del(res http.ResponseWriter, req *http.Request) {

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

		// delete the key from cache

		cache.Del(key)

		// delete the key from DB

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

	if log == 1 {
		cache.PrintCache()
	}
}

func C_create(res http.ResponseWriter, req *http.Request) {

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

		// find key in the cache

		_, found := cache.Get(data.Key)

		if found == 0 { // if not found in cache

			// we check in DB

			row := db.QueryRow("SELECT * FROM kv_store WHERE key_col = ?", data.Key)

			var dummy1, dummy2 any

			if err = row.Scan(&dummy1, &dummy2); err == sql.ErrNoRows { // if not in DB

				// insert the data into DB

				if _, err = db.Exec("INSERT INTO kv_store (key_col, value_col) values (?, ?)", data.Key, data.Value); err != nil {
					panic(err)
				}

				// insert the data into cache

				cache.Set(data.Key, data.Value)

			} else if err == nil { // if rows are found then

				// send the error to the client that the key exists
				resp.Err = "Key already Exists"

			} else {
				panic(err)
			}

		} else { // if found in cache

			// send the error to the client that the key exists
			resp.Err = "Key already Exists"
		}

		// send the response
		// res_data := "key created"

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(resp)
		// fmt.Fprintf(res, res_data)

	} else { // if it ain't POST req
		http.Error(res, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if log == 1 {
		cache.PrintCache()
	}
}

func C_update(res http.ResponseWriter, req *http.Request) {

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

		_, found := cache.Get(data.Key)

		if found == 0 { // if not found in cache

			// check in db

			row := db.QueryRow("SELECT * FROM kv_store WHERE key_col = ?", data.Key)

			var dummy1, dummy2 any

			if err = row.Scan(&dummy1, &dummy2); err == sql.ErrNoRows { // if not found in db

				// send the error to the client that the key doesn't exists
				resp.Err = "Key doesn't exists"

			} else if err == nil { // if found in db

				// Update the data
				if _, err = db.Exec("UPDATE kv_store SET value_col = ? WHERE key_col = ?", data.Value, data.Key); err != nil {
					panic(err)
				}

				// insert data in the cache
				cache.Set(data.Key, data.Value)

			} else {
				panic(err)
			}

		} else { // if found in cache

			// update the data of the cache
			cache.Set(data.Key, data.Value)

		}

		// send the response
		// res_data := "key updated" + strconv.Itoa(data.Key) + data.Value

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(resp)
		// fmt.Fprintf(res, res_data)

	} else { // if it ain't POST method
		http.Error(res, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if log == 1 {
		cache.PrintCache()
	}
}

// CRUD without cache

func D_fetch(res http.ResponseWriter, req *http.Request) {

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
func D_del(res http.ResponseWriter, req *http.Request) {

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

func D_create(res http.ResponseWriter, req *http.Request) {

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

func D_update(res http.ResponseWriter, req *http.Request) {

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
	if len(os.Args) < 3 {
		fmt.Println("<file_name> <port_no> <C or D>")
		return
	}

	// Initializing the Cache

	cache = Initializer(32)

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

	if os.Args[2] == "C" {

		// http routes (with cache)

		http.HandleFunc("/C/create", C_create)
		http.HandleFunc("/C/update", C_update)
		http.HandleFunc("/C/read", C_fetch)
		http.HandleFunc("/C/delete", C_del)

	} else if os.Args[2] == "D" {

		// http routes (without cache)

		http.HandleFunc("/D/create", D_create)
		http.HandleFunc("/D/update", D_update)
		http.HandleFunc("/D/read", D_fetch)
		http.HandleFunc("/D/delete", D_del)

	} else {
		fmt.Println("<file_name> <port_no> <C or D>")
		return
	}

	// Port number

	http.ListenAndServe(os.Args[1], nil)
}
