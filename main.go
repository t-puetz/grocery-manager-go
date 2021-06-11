package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Starting web server now...\n")
	srvMain(8080)
}

func srvMain(port int) {
	fmt.Println("Rest API v2.0 - Mux Routers")
	router := handleRESTRequests()
	startWebSrv(port, router)
}

func startWebSrv(port int, router *mux.Router) {
	portStr := fmt.Sprint(port)
	fmt.Println(portStr)
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
	fmt.Println("Hit REST end point GET /api/lists")
}
func getListsID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point GET /api/lists/{id}")
}
func postLists(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point POST /api/lists")
}
func patchListsID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point PATCH /api/lists/{id}")
}
func patchListsGroceryItemID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point PATCH /api/lists/{id}/{groceryItemID}")
}
func deleteListsGroceryItemID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point DELETE /api/lists/{id}")
}
func getItems(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point GET /api/items")
}
func postItems(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point POST /api/items")
}
func patchItems(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point PATCH /api/items/{id}")
}
func deleteItems(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit REST end point DELETE /api/items/{id}")
}
