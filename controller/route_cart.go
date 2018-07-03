package controller

import (
	"borgdir-media/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
)

/*** Controllers responsible for handling the equipment; mostly session-based ***/

type SessionItem struct {
	Item model.Equipment
	Date string //if "cart", ReturnDate(); if "reserved", date of most recent return
}

/*
func init() {
	//a := map[string][]string{"cat": {"orange", "grey"},	"dog": {"black"},
}*/

func AddToCart(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	var path = fmt.Sprintf("%s", r.URL.Path[15:])
	id, _ := strconv.Atoi(path)
	equip := model.GetItem(id)
	if session.Values["cart"] == nil || session.Values["cart"] == "" {
		slice := new([]int)
		session.Values["cart"] = append(*slice, equip.ID)
	} else {
		session.Values["cart"] = append(session.Values["cart"].([]int), equip.ID)
	}
	session.Save(r, w)
	http.Redirect(w, r, "/equipment", http.StatusFound)
}

func Reserve(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	var path = fmt.Sprintf("%s", r.URL.Path[19:])
	id, _ := strconv.Atoi(path)
	equip := model.GetItem(id)
	if session.Values["reserved"] == nil || session.Values["reserved"] == "" {
		slice := new([]int)
		session.Values["reserved"] = append(*slice, equip.ID)
	} else {
		session.Values["reserved"] = append(session.Values["reserved"].([]int), equip.ID)
	}
	fmt.Println(session.Values["reserved"])
	session.Save(r, w)
	http.Redirect(w, r, "/equipment", http.StatusFound)
}

func RetrieveCartFromSession(session *sessions.Session) []SessionItem {
	itemsList := []SessionItem{}
	if session.Values["cart"] != nil {
		for _, id := range session.Values["cart"].([]int) {
			item := SessionItem{model.GetItem(id), model.ReturnDate()}
			itemsList = append(itemsList, item)
		}
		return itemsList
	} else {
		return itemsList
	}
}

func RetrieveReservedFromSession(session *sessions.Session) []SessionItem {
	itemsList := []SessionItem{}
	if session.Values["reserved"] != nil {
		for _, id := range session.Values["reserved"].([]int) {
			item := SessionItem{model.GetItem(id), model.CheckForReturnDates(id)}
			itemsList = append(itemsList, item)
		}
		return itemsList
	} else {
		return itemsList
	}
}

func Cart(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	fmt.Println(session.Values["cart"])
	vd := ViewData{}
	if session.Values["authenticated"] == true && session.Values["username"] != nil && session.Values["username"] != 0 {
		user, _ := model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
		vd.User = user
	}
	vd.SeshItems = RetrieveCartFromSession(session)
	if err := CartView.Template.ExecuteTemplate(w,
		CartView.Layout, vd); err != nil {
		panic(err)
	}
}

func Order(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	if err := r.ParseForm(); err != nil {
		fmt.Print(err)
	} else {
		var a []SessionItem = RetrieveCartFromSession(session)
		var q []string = r.Form["quantity"]
		order := model.Orders{}
		order.MadeBy = model.GetIDByUsername(fmt.Sprintf("%v", session.Values["username"]))
		for i, _ := range a {
			oneOrder := model.OrderedEquipment{}
			oneOrder.OrdEq = a[i].Item
			oneOrder.Quantity, _ = strconv.Atoi(q[i])
			order.Contains = append(order.Contains, oneOrder)
		}
		if err = order.PlaceOrder(); err == nil {
			session.Values["cart"] = nil
		} else {
			fmt.Println(err)
		}
		session.Save(r, w)
		http.Redirect(w, r, "/my-equipment", http.StatusFound)
	}
}

func DeleteFromOrder(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	id, _ := strconv.Atoi(fmt.Sprintf("%s", r.URL.Path[13:]))
	if session.Values["cart"] == nil || session.Values["cart"] == "" {
		http.Redirect(w, r, "/cart", http.StatusGone)
	}
	a := session.Values["cart"].([]int)
	for i, d := range a {
		if id == d {
			if i == 0 {
				session.Values["cart"] = append(a[:0], a[1:]...)
			} else if i == len(a)-1 {
				session.Values["cart"] = append(a[:i], a[len(a):]...)
			} else {
				session.Values["cart"] = append(a[:i], a[i+1:]...)
			}
		}
	}
	session.Save(r, w)
	http.Redirect(w, r, "/cart", http.StatusFound)
}

func DeleteFromReserved(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	id, _ := strconv.Atoi(fmt.Sprintf("%s", r.URL.Path[21:]))
	if session.Values["reserved"] == nil || session.Values["reserved"] == "" {
		http.Redirect(w, r, "/my-equipment", http.StatusGone)
	}
	a := session.Values["reserved"].([]int)
	for i, d := range a {
		if id == d {
			if i == 0 {
				session.Values["reserved"] = append(a[:0], a[1:]...)
			} else if i == len(a)-1 {
				session.Values["reserved"] = append(a[:i], a[len(a):]...)
			} else {
				session.Values["reserved"] = append(a[:i], a[i+1:]...)
			}
		}
	}
	session.Save(r, w)
	http.Redirect(w, r, "/my-equipment", http.StatusFound)
}

//kinda buggy, doesn't delete the old orders if the same return button is pressed quickly multiple times => TODO: optimistic locking
//current solution is to let it Sleep() so that transactions don't get mixed up
func Extend(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	if session.Values["authenticated"] == true {
		r.ParseForm()
		fmt.Println("r.Form", r.Form)
		user, _ := model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
		user.GetCurrentOrders()
		order := model.Orders{}
		for _, d := range user.OrdersList {
			if r.FormValue("return") == d.Return {
				order = d
			}
		}
		quantity, _ := strconv.Atoi(r.FormValue("quantity"))
		equipid, _ := strconv.Atoi(r.FormValue("equip-id"))
		if err := order.ExtendRent(equipid, quantity); err != nil {
			fmt.Println(err)
			return
		}
		session.Save(r, w)
		http.Redirect(w, r, "/my-equipment", http.StatusFound)
	}
}
