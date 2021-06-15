package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// In the case of go-sqlite3, the underscore import is used for the side-effect
// of registering the sqlite3 driver as a database driver in the init() function,
// without importing any other functions
// init functions run even before main does!

type listTableJSON struct {
	ID    int    `json: "id"`
	Title string `json: "title"`
}

type listItemTableJSON struct {
	GroceryItemID int `json: "groceryItemID"`
	Quantity      int `json: "quantity"`
	Checked       int `json: "checked"`
	Position      int `json: "position"`
	OnList        int `json: "onList"`
}

type groceryItemTableJSON struct {
	ID      int    `json: "ID"`
	Name    string `json: "name"`
	Current int    `json: "current"`
	Maximum int    `json: "maximum"`
}

func main() {
	flags := parseFlags()
	log.Printf("Flags provided: %v", flags)

	db := openDB()
	dbscheme := readInDBSchemeDefinition()
	createScheme(db, dbscheme)

	portStr := flags["--port"]
	portInt, _ := strconv.Atoi(portStr)
	srvMain(portInt)
}

func parseFlags() map[string]string {
	flagMap := make(map[string]string)

	portArg := flag.String("port", "", `Port you want the webserver to run at.`)
	flag.Parse()

	if *portArg == "" {
		defaultPort := "8080"
		pDefaultPort := &defaultPort
		portArg = pDefaultPort
		log.Printf("You did not provide --port flag. Will try to bind webserver at port %s.", *portArg)
	}

	port := *portArg
	flagMap["--port"] = port

	return flagMap
}

// Start webserver section

func srvMain(port int) {
	router := handleRESTRequests()
	startWebSrv(port, router)
}

func startWebSrv(port int, router *mux.Router) {
	portStr := fmt.Sprint(port)
	log.Printf("Starting web server now at port %v...\n", portStr)
	log.Fatal(http.ListenAndServe(":"+portStr, router))
}

func handleRESTRequests() *mux.Router {
	apiBasePath := "/api"
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc(apiBasePath+"/lists", getLists).Methods("GET")
	router.HandleFunc(apiBasePath+"/lists/{id}", getListsID).Methods("GET")
	router.HandleFunc(apiBasePath+"/lists", postLists).Methods("POST")
	router.HandleFunc(apiBasePath+"/lists/{id}", patchListsID).Methods("PATCH")
	router.HandleFunc(apiBasePath+"/lists/{id}/{groceryItemID}", patchListsGroceryItemID).Methods("PATCH")
	router.HandleFunc(apiBasePath+"/lists/{id}", deleteListsGroceryItemID).Methods("DELETE")
	router.HandleFunc(apiBasePath+"/items", getItems).Methods("GET")
	router.HandleFunc(apiBasePath+"/items", postItems).Methods("POST")
	router.HandleFunc(apiBasePath+"/items/{id}", patchItems).Methods("PATCH")
	router.HandleFunc(apiBasePath+"/items/{id}", deleteItems).Methods("DELETE")

	return router
}

func getLists(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point GET /api/lists")

	db := openDB()
	rows, err := db.Query("SELECT * FROM list;")

	if err != nil {
		log.Panic(err)
	}

	log.Printf("Endpoint GET /api/lists raw DB return as map:\n%v", rows)

	var rowSinks []listTableJSON

	for rows.Next() {
		var rowSink listTableJSON
		content := make([]*string, 3)

		//SQL Scan method parameter values must be of type interface{}. So we need an intermediate slice

		ims := make([]interface{}, 3)

		for i := range ims {
			ims[i] = &content[i]
		}

		err = rows.Scan(ims[0], ims[1], ims[2])

		if err != nil {
			log.Panic(err)
		}

		rowSink.ID, _ = strconv.Atoi(*(content[0]))
		rowSink.Title = *(content[1])

		rowSinks = append(rowSinks, rowSink)
	}
	log.Printf("Control print of list struct before encoding to JSON:\n%v", rowSinks)
	json.NewEncoder(w).Encode(rowSinks)
}

func getListsID(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point GET /api/lists/{id}")

	vars := mux.Vars(r)
	key := vars["id"]

	db := openDB()
	// DB returns a string even though on DB level ID is an INTEGER!
	query := fmt.Sprintf("SELECT * FROM list where id = %s;", key)
	rows, err := db.Query(query)

	if err != nil {
		log.Panic(err)
	}

	var rowSink listTableJSON
	content := make([]*string, 3)

	//SQL Scan method parameter values must be of type interface{}. So we need an intermediate slice

	ims := make([]interface{}, 3)

	// Although we can be SURE there will only be one Result
	// sql still expects us to call .Next() or else it will fail
	for rows.Next() {

		for i := range ims {
			ims[i] = &content[i]
		}

		err = rows.Scan(ims[0], ims[1], ims[2])

		if err != nil {
			log.Panic(err)
		}

		rowSink.ID, _ = strconv.Atoi(*(content[0]))
		rowSink.Title = *(content[1])

		log.Printf("Endpoint GET /api/list/{id} raw DB return as map:\n%v", rows)
		log.Printf("Endpoint GET /api/list/{id} raw Go struct representing a row:\n%v", rowSink)

		json.NewEncoder(w).Encode(rowSink)
	}
}

func postLists(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point POST /api/lists")
	db := openDB()

	var rowSink listTableJSON
	reqBody, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(reqBody, &rowSink)

	if err != nil {
		log.Panic(err)
	}

	err = json.NewEncoder(w).Encode(rowSink)
	log.Printf("Potential err encoded JSON: %v\n", err)

	if err != nil {
		log.Panic(err)
	}
	log.Printf("Go struct generated from requestBody:\n%+v", rowSink)

	json.NewEncoder(w).Encode(rowSink)
	// DB returns a string even though on DB level ID is an INTEGER!
	listIDStr := strconv.Itoa(rowSink.ID)
	listTitle := rowSink.Title
	sqlStatement := fmt.Sprintf("INSERT INTO list (id,title) VALUES (%s,%s);", listIDStr, listTitle)
	_, err = db.Exec(sqlStatement)

	if err != nil {
		log.Panic(err)
	}

}

func patchListsID(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point PATCH /api/lists/{id}")
}

func patchListsGroceryItemID(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point PATCH /api/lists/{id}/{groceryItemID}")
}

func deleteListsGroceryItemID(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point DELETE /api/lists/{id}")
}

func getItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point GET /api/items")
}

func postItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point POST /api/items")
}

func patchItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point PATCH /api/items/{id}")
}

func deleteItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point DELETE /api/items/{id}")
}

// End webserver section

// Start sqlite DB section

func openDB() *sql.DB {
	db, err := sql.Open("sqlite3", "grocery-manager-go.sqlite")

	if err != nil {
		log.Panic(err)
	}

	log.Println("sqlite3 DB opened successfully (grocery-manager-go.sqlite)")

	return db
}

func readInDBSchemeDefinition() string {
	content, err := ioutil.ReadFile("db.scheme")

	if err != nil {
		log.Panic(err)
	}

	contentStr := string(content)
	return contentStr
}

func createScheme(db *sql.DB, dbscheme string) {
	defer db.Close()

	_, err := db.Exec(dbscheme)

	if err != nil {
		log.Panic(err)
	}

	log.Println("sqlite DB grocery-manager-go.sqlite is setup and ready...")
}

// End sqlite DB section
