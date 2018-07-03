package main

import (
	"borgdir-media/controller"
	"database/sql"
	"fmt" //main package for interacting w/ HTTP
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var vd *controller.ViewData

func main() {
	db, _ = sql.Open("sqlite3", "data.db")
	defer db.Close()

	mux := mux.NewRouter()
	server := &http.Server{
		Addr:         "localhost:8080",
		Handler:      mux, // context.ClearHandler(mux)
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//serve static ressources w/o having to specify format in HTTP Content-Header
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", controller.Home)
	mux.HandleFunc("/login", controller.Login)
	mux.HandleFunc("/register", controller.Register)

	//TODO Equipment: Sort, Search, Both
	mux.HandleFunc("/equipment", controller.Equipment)
	mux.HandleFunc("/equipment/add/{id:[0-9]+}", controller.AddToCart)
	mux.HandleFunc("/equipment/reserve/{id:[0-9]+}", controller.Reserve)
	//mux.HandleFunc("/equipment/{category:[a-z]+}", controller.EquipmentSort)
	//mux.HandleFunc("/equipment/{category:[a-z]+}/{}", controller.EquipmentSortSearch)
	//mux.HandleFunc("/equipment/{category:[a-z]+}", controller.EquipmentSearch)

	//custom 404 page for whenever the ressource is missing
	mux.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vd := controller.ViewData{}
		w.WriteHeader(http.StatusNotFound)
		if err := controller.NotFoundView.Template.ExecuteTemplate(w,
			controller.NotFoundView.Layout, vd); err != nil {
			panic(err)
		}
	})

	//User-only
	mux.HandleFunc("/logout", controller.Logout)

	mux.HandleFunc("/my-equipment", controller.MyEquipment)
	mux.HandleFunc("/my-equipment/extend", controller.Extend)
	mux.HandleFunc("/my-equipment/delete/{id:[0-9]+}", controller.DeleteFromReserved)

	mux.HandleFunc("/cart", controller.Cart)
	mux.HandleFunc("/cart/order", controller.Order).Methods("POST")
	mux.HandleFunc("/cart/delete/{id:[0-9]+}", controller.DeleteFromOrder)

	mux.HandleFunc("/profile/{id:[0-9]+}", controller.Profile)

	mux.HandleFunc("/profile/{id:[0-9]+}/delete", controller.DeleteClient)
	mux.HandleFunc("/profile/{id:[0-9]+}/block", controller.BlockClient)

	//TODO possibility to parse /profile/{username:[a-z]+}

	//Subrouter for admin
	smux := mux.PathPrefix("/admin").Subrouter()
	smux.HandleFunc("", controller.Admindex)

	smux.HandleFunc("/", controller.Admindex)
	smux.HandleFunc("/clients", controller.Clients)

	//borgdir.media/admin/equipment
	smux.HandleFunc("/equipment", controller.AdmEquip)

	//borgdir.media/equipment/{id:[0-9]+}
	mux.HandleFunc("/equipment/{id:[0-9]+}", controller.Equip)
	mux.HandleFunc("/equipment/{id:[0-9]+}/delete", controller.DeleteEquip)

	mux.HandleFunc("/equipment/new", controller.AddEquip)

	fmt.Println(server.ListenAndServe())

	//for future reference: HTTPS implementation via	server.ListenAndServeTLS("cert.pem", "key.pem")

}
