package main

import (
	"fmt"
	"illusionman1212/kotapi-go/db"
	"illusionman1212/kotapi-go/routes"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	db.InitializeDB()

	router := mux.NewRouter()
	router.HandleFunc("/", routes.RandomHandler).Methods("GET")
	router.HandleFunc("/addkot", routes.AddKotHandler).Methods("POST")
	router.HandleFunc("/{id}", routes.IdHandler).Methods("GET")

	router.HandleFunc("/kots/compressed/{filename}", routes.KotsCompressedHandler).Methods("GET")
	router.HandleFunc("/kots/{filename}", routes.KotsHandler).Methods("GET")

	fmt.Println("Listening on 8080")
	http.ListenAndServe(":8080", router)
}
