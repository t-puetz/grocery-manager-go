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
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// In the case of go-sqlite3, the underscore import is used for the side-effect
// of registering the sqlite3 driver as a database driver in the init() function,
// without importing any other functions
// init functions run even before main does!

type listTableJSON struct {
	ID    int     `json: "id"`
	Title *string `json: "title,omitempty"`
}

type listItemTableJSON struct {
	GroceryItemID int  `json: "groceryItemID"`
	Quantity      *int `json: "quantity,omitempty"`
	Checked       *int `json: "checked,omitempty"`
	Position      *int `json: "position,omitempty"`
	OnList        *int `json: "onList,omitempty"`
}

type groceryItemTableJSON struct {
	ID      int     `json: "ID"`
	Name    *string `json: "name,omitempty"`
	Current *int    `json: "current,omitempty"`
	Minimum *int    `json: "minimum,omitempty`
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

func ifErrorLogPanicError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func createSinkSliceForRecords(numColumns int, pContent *[]*string) *[]interface{} {
	//SQL Scan method parameter values must be of type interface{}. So we need an intermediate slice
	ims := make([]interface{}, numColumns)
	content := *pContent

	for i := range ims {
		ims[i] = &content[i]
	}

	return &ims
}

/* func concatInsertIntoSQLStatement(table string, columns []string, formatters []string) string {
	fixedFmtStrOne := "INSERT INTO "
	fixedFmtStrTwo := " ("
	fixedFmtStrThree := ") VALUES ("
	fixedFmtStrFour := ");"

	sqlStatementBase := fixedFmtStrOne + table + fixedFmtStrTwo + strings.Join(columns, ",") + fixedFmtStrThree + strings.Join(formatters, ",") + fixedFmtStrFour
	log.Printf("concatInsertIntoSQLStatement(): Returned SQL-Statement:\n%s\n", sqlStatementBase)
	return sqlStatementBase
}*/

func concatUpdateSQLStatement(table string, foreignKeys []string, columns []string, formatters []string) string {
	//sqlStatement := fmt.Sprintf("UPDATE list_item SET quantity = %d, checked = %d, position = %d WHERE on_list = %s AND grocery_item_id = %s;", quantity, checked, position, listID, groceryItemID)
	fixedFmtStrOne := "UPDATE " + table + " SET "

	for i := range columns {
		if i != len(columns)-1 {
			fixedFmtStrOne += columns[i] + " = " + formatters[i] + ", "
		} else {
			fixedFmtStrOne += columns[i] + " = " + formatters[i]
		}
	}

	formatters = formatters[len(foreignKeys):]

	for i := range foreignKeys {
		if i == 0 {
			fixedFmtStrOne += " WHERE " + foreignKeys[i] + " = " + formatters[i]
		} else if i > 0 {
			fixedFmtStrOne += " AND " + foreignKeys[i] + " = " + formatters[i]
		} else {
			log.Panic("concatUpdateSQLStatement() no foreign key(s) slice provided as second arg. Returning empty string.")
			return ""
		}
	}

	fixedFmtStrOne += ";"
	sqlStatementBase := fixedFmtStrOne
	log.Printf("concatUpdateSQLStatement(): Returned SQL-Statement:\n%s\n", sqlStatementBase)
	return sqlStatementBase
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
	router.HandleFunc(apiBasePath+"/lists/{id}", deleteLists).Methods("DELETE")
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
	ifErrorLogPanicError(err)

	log.Printf("Endpoint GET /api/lists raw DB return as map:\n%v", rows)

	var rowSinks []listTableJSON

	for rows.Next() {
		var rowSink listTableJSON

		//SQL Scan method parameter values must be of type interface{}. So we need an intermediate slice
		content := make([]*string, 2)
		pIms := createSinkSliceForRecords(2, &content)
		ims := *pIms

		err = rows.Scan(ims[0], ims[1])
		ifErrorLogPanicError(err)

		rowSink.ID, _ = strconv.Atoi(*(content[0]))
		rowSink.Title = content[1]
		rowSinks = append(rowSinks, rowSink)
	}
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
	ifErrorLogPanicError(err)

	var rowSink listTableJSON

	// Although we can be SURE there will only be one Result
	// sql still expects us to call .Next() or else it will fail
	for rows.Next() {
		content := make([]*string, 2)
		pIms := createSinkSliceForRecords(2, &content)
		ims := *pIms

		err = rows.Scan(ims[0], ims[1])
		ifErrorLogPanicError(err)

		rowSink.ID, _ = strconv.Atoi(*(content[0]))
		rowSink.Title = content[1]

		json.NewEncoder(w).Encode(rowSink)
	}
}

func postLists(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point POST /api/lists")

	db := openDB()

	var rowSink listTableJSON

	reqBody, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(reqBody, &rowSink)
	ifErrorLogPanicError(err)

	sqlStatement := ""
	listTitle := "-1"
	listIDStr := strconv.Itoa(rowSink.ID)

	if rowSink.Title != nil {
		listTitle = *(rowSink.Title)
	}

	if listTitle != "-1" {
		sqlStatement = fmt.Sprintf("INSERT INTO list (id,title) VALUES (%s,\"%s\");", listIDStr, listTitle)
	} else {
		log.Println("Your JSON payload did not provide the title attribute: Naming list an empty string.")
		sqlStatement = fmt.Sprintf("INSERT INTO list (id,title) VALUES (%s,\"%s\");", listIDStr, "")
	}

	_, err = db.Exec(sqlStatement)
	ifErrorLogPanicError(err)

	// Respond back to client
	err = json.NewEncoder(w).Encode(rowSink)

	ifErrorLogPanicError(err)
}

func patchListsID(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point PATCH /api/lists/{id}")
	db := openDB()

	var rowSink listTableJSON
	reqBody, _ := ioutil.ReadAll(r.Body)

	vars := mux.Vars(r)
	listID := vars["id"]

	err := json.Unmarshal(reqBody, &rowSink)
	ifErrorLogPanicError(err)

	sqlStatement := ""
	listTitle := "-1"

	if rowSink.Title != nil {
		listTitle = *(rowSink.Title)
	}

	if listTitle != "-1" {
		sqlStatement = fmt.Sprintf("UPDATE list SET title = \"%s\" WHERE id = %s;", listTitle, listID)
	} else {
		log.Println("Your JSON payload did not provide the title attribute: Not triggering any UPDATE!")
	}

	_, err = db.Exec(sqlStatement)
	ifErrorLogPanicError(err)

	// Respond back to client
	err = json.NewEncoder(w).Encode(rowSink)
	ifErrorLogPanicError(err)
}

func patchListsGroceryItemID(w http.ResponseWriter, r *http.Request) {
	// Our goal is it to update a list_item record
	log.Println("Hit REST end point PATCH /api/lists/{id}/{groceryItemID}")
	db := openDB()

	vars := mux.Vars(r)

	for key := range vars {
		log.Printf("%v", key)
	}

	listID := vars["id"]
	groceryItemID := vars["groceryItemID"]

	var rowSink listItemTableJSON
	reqBody, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(reqBody, &rowSink)
	ifErrorLogPanicError(err)

	quantity := -1
	checked := -1
	position := -1
	sqlStatement := ""
	sqlStatementBase := ""

	if rowSink.Quantity != nil {
		quantity = *(rowSink.Quantity)
	}

	if rowSink.Checked != nil {
		checked = *(rowSink.Checked)
	}

	if rowSink.Position != nil {
		position = *(rowSink.Position)
	}

	if quantity == -1 && checked == -1 {
		sqlStatementBase = concatUpdateSQLStatement("list_item", []string{"on_list", "grocery_item_id"}, []string{"position"}, []string{"%d", "%d", "%d"})
		sqlStatement = fmt.Sprintf(sqlStatementBase, position, listID, groceryItemID)
	} else if quantity == -1 && position == -1 {
		sqlStatementBase = concatUpdateSQLStatement("list_item", []string{"on_list", "grocery_item_id"}, []string{"checked"}, []string{"%d", "%d", "%d"})
		sqlStatement = fmt.Sprintf(sqlStatementBase, checked, listID, groceryItemID)
	} else if checked == -1 && position == -1 {
		sqlStatementBase = concatUpdateSQLStatement("list_item", []string{"on_list", "grocery_item_id"}, []string{"quantity"}, []string{"%d", "%d", "%d"})
		sqlStatement = fmt.Sprintf(sqlStatementBase, quantity, listID, groceryItemID)
	} else if quantity == -1 {
		sqlStatementBase = concatUpdateSQLStatement("list_item", []string{"on_list", "grocery_item_id"}, []string{"position", "checked"}, []string{"%d", "%d", "%d", "%d"})
		sqlStatement = fmt.Sprintf(sqlStatementBase, checked, position, listID, groceryItemID)
	} else if position == -1 {
		log.Println("Case: Only position attribute was not provied!")
		sqlStatementBase = concatUpdateSQLStatement("list_item", []string{"on_list", "grocery_item_id"}, []string{"quantity", "checked"}, []string{"%d", "%d", "\"%s\"", "\"%s\""})
		sqlStatement = fmt.Sprintf(sqlStatementBase, quantity, checked, listID, groceryItemID)
	} else if checked == -1 {
		sqlStatementBase = concatUpdateSQLStatement("list_item", []string{"on_list", "grocery_item_id"}, []string{"quantity", "position"}, []string{"%d", "%d", "%d", "%d"})
		sqlStatement = fmt.Sprintf(sqlStatementBase, quantity, position, listID, groceryItemID)
	} else {
		sqlStatementBase = concatUpdateSQLStatement("list_item", []string{"on_list", "grocery_item_id"}, []string{"quantity", "checked", "position"}, []string{"%d", "%d", "%d", "%d", "%d"})
		sqlStatement = fmt.Sprintf(sqlStatementBase, quantity, checked, position, listID, groceryItemID)
	}

	log.Printf("patchListsGroceryItemID() SQL-Statement:\n%s\n", sqlStatement)

	_, err = db.Exec(sqlStatement)
	ifErrorLogPanicError(err)
	err = json.NewEncoder(w).Encode(rowSink)
	ifErrorLogPanicError(err)
}

func deleteLists(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point DELETE /api/lists/{id}")
	db := openDB()

	vars := mux.Vars(r)
	id := vars["id"]

	// DB returns a string even though on DB level ID is an INTEGER!
	sqlStatement := fmt.Sprintf("DELETE FROM list WHERE id = %s;", id)
	_, err := db.Exec(sqlStatement)
	ifErrorLogPanicError(err)
}

func getItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point GET /api/items")

	db := openDB()
	sqlStatement := fmt.Sprintf("SELECT * from list_item;")
	rows, err := db.Query(sqlStatement)
	ifErrorLogPanicError(err)

	var rowSinks []listItemTableJSON

	for rows.Next() {
		var rowSink listItemTableJSON

		content := make([]*string, 5)
		pIms := createSinkSliceForRecords(5, &content)
		ims := *pIms

		err = rows.Scan(ims[0], ims[1], ims[2], ims[3], ims[4])
		ifErrorLogPanicError(err)

		id, _ := strconv.Atoi(*(content[1]))
		checked, _ := strconv.Atoi(*(content[0]))
		quanity, _ := strconv.Atoi(*(content[2]))
		position, _ := strconv.Atoi(*(content[3]))
		onList, _ := strconv.Atoi(*(content[4]))

		rowSink.GroceryItemID = id
		rowSink.Checked = &checked
		rowSink.Quantity = &quanity
		rowSink.Position = &position
		rowSink.OnList = &onList

		rowSinks = append(rowSinks, rowSink)
	}
	json.NewEncoder(w).Encode(rowSinks)
}

func postItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point POST /api/items")
	db := openDB()

	var rowSink groceryItemTableJSON

	reqBody, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(reqBody, &rowSink)
	ifErrorLogPanicError(err)

	// DB returns a string even though on DB level ID is an INTEGER!
	sqlStatement := ""
	name := "-1"
	minimum := -1
	current := -1

	if rowSink.Name != nil {
		name = *(rowSink.Name)
	}

	if rowSink.Minimum != nil {
		minimum = *(rowSink.Minimum)
	}

	if rowSink.Current != nil {
		current = *(rowSink.Current)
	}

	if name != "-1" && minimum != -1 && current != -1 {
		sqlStatement = fmt.Sprintf("INSERT INTO grocery_item (name,current,minimum) VALUES (\"%s\", %d, %d);", name, minimum, current)
	} else {
		log.Println("Invalid JSON payload for POST /api/items: {\"name\": \"some_name\", \"minimum\": some_positive_integer, \"current\": some_positive_integer} expected")
	}

	_, err = db.Exec(sqlStatement)
	ifErrorLogPanicError(err)

	// Respond back to client
	err = json.NewEncoder(w).Encode(rowSink)
	ifErrorLogPanicError(err)
}

func patchItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point PATCH /api/items/{id}")
	db := openDB()

	vars := mux.Vars(r)
	id := vars["id"]

	var rowSink groceryItemTableJSON
	reqBody, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(reqBody, &rowSink)
	ifErrorLogPanicError(err)

	name := "-1"
	current := -1
	minimum := -1
	sqlStatementBase := ""
	sqlStatement := ""

	if rowSink.Name != nil {
		name = *(rowSink.Name)
	}

	if rowSink.Current != nil {
		current = *(rowSink.Current)
	}

	if rowSink.Minimum != nil {
		minimum = *(rowSink.Minimum)
	}

	if name == "-1" && current == -1 {
		sqlStatementBase = concatUpdateSQLStatement("grocery_item", []string{"id"}, []string{"minimum"}, []string{"%d", "\"%s\""})
		sqlStatement = fmt.Sprintf(sqlStatementBase, minimum, id)
	}

	log.Printf("patchItems() SQL-Statement:\n%s\n", sqlStatement)

	//sqlStatement := fmt.Sprintf("UPDATE grocery_item SET name = \"%s\", current = %d, minimum = %d WHERE id = %s;", name, current, minimum, id)
	_, err = db.Exec(sqlStatement)
	ifErrorLogPanicError(err)

	// Respond back to client
	err = json.NewEncoder(w).Encode(rowSink)
	ifErrorLogPanicError(err)
}

func deleteItems(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit REST end point DELETE /api/items/{id}")
	db := openDB()

	vars := mux.Vars(r)
	id := vars["id"]

	// DB returns a string even though on DB level ID is an INTEGER!
	sqlStatement := fmt.Sprintf("DELETE FROM grocery_item WHERE id = %s;", id)
	_, err := db.Exec(sqlStatement)
	ifErrorLogPanicError(err)
}

// End webserver section

// Start sqlite DB section

func openDB() *sql.DB {
	db, err := sql.Open("sqlite3", "grocery-manager-go.sqlite")

	ifErrorLogPanicError(err)
	_, err = db.Exec("PRAGMA foreign_keys = on;")
	ifErrorLogPanicError(err)
	log.Println("sqlite3 DB opened successfully (grocery-manager-go.sqlite)")

	return db
}

func readInDBSchemeDefinition() string {
	content, err := ioutil.ReadFile("db.scheme")
	ifErrorLogPanicError(err)
	contentStr := string(content)
	return contentStr
}

func createScheme(db *sql.DB, dbscheme string) {
	defer db.Close()

	_, err := db.Exec(dbscheme)
	ifErrorLogPanicError(err)
	log.Println("sqlite DB grocery-manager-go.sqlite is setup and ready...")
}

// End sqlite DB section
