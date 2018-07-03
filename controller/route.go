package controller

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"borgdir-media/model"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var UserPwPepper = "t0p-sEcR3t"

/* Controllers w access granted to anyone*/

var store *sessions.CookieStore

func init() {
	key := make([]byte, 32)
	rand.Read(key)
	store = sessions.NewCookieStore(key)
	fmt.Println("Key: ", key)
}

func Register(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	vd := ViewData{}
	vd.User.Username = ""
	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, d := range flashes {
			vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
		}
		session.Save(r, w)
	}

	if session.Values["authenticated"] == false || session.Values["authenticated"] == nil {
		switch r.Method {
		case "GET":
			{
				if err := RegisterView.Template.ExecuteTemplate(w,
					RegisterView.Layout, vd); err != nil {
					panic(err)
				}
			}
		case "POST":
			{
				if err := r.ParseForm(); err != nil {
					fmt.Print(" register - POST err:", err)
				} else {
					firstName := r.FormValue("firstNameReg")
					lastName := r.FormValue("lastNameReg")
					username := r.FormValue("usernameReg")
					password := r.FormValue("passwordReg")
					email := r.FormValue("emailReg")
					fmt.Println("Username", username, ", Das eingegebene Passwort ist", password)
					if strings.Compare(password, r.FormValue("passwordRep")) != 0 {
						session.AddFlash("Passwörter stimmen nicht überein. Bitte überprüfen Sie Ihre Angaben")
					}
					if model.AlreadyUsed(username, email) {
						session.AddFlash("Name oder E-Mail werden bereits verwendet. Bitte ändern Sie Ihre Angaben")
					}
					session.Save(r, w)
					if flashes := session.Flashes(); len(flashes) > 0 {
						http.Redirect(w, r, "/register", http.StatusFound)
					} else {
						userAcc := model.User{FirstName: firstName, LastName: lastName, Username: username, Password: password, Email: email}
						userAcc.AddUser()
						//fmt.Println("erfolgreich!", userAcc)
						session.AddFlash("Erfolg! Sie können jetzt fortfahren.")
						session.Save(r, w)
						//TODO redirect only works with 3xx -- solution???
						http.Redirect(w, r, "/", http.StatusFound)
					}
				}
			}
		}
	} else { //if already authenticated
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func Equipment(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")

	vd := ViewData{}
	if session.Values["authenticated"] == true && session.Values["username"] != nil && session.Values["username"] != 0 {
		vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
	}

	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, d := range flashes {
			vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
		}
		session.Save(r, w)
	}
	switch r.Method {
	case "GET":
		{
			vd.EquipmentItems, _ = model.GetAll()
			err := EquipView.Template.ExecuteTemplate(w,
				EquipView.Layout, vd)
			if err != nil {
				panic(err)
			}
		}
	case "POST":
		{
			if err := r.ParseForm(); err != nil {
				fmt.Print("POST err:", err)
			} else {

				category := r.FormValue("kategorie")
				var search string = r.FormValue("search")
				fmt.Printf("Kategorie: %v von Typ %T, Suche: %v von Typ %T", category, category, search, search)

				a, _ := model.GetAllByCategory(category, search)
				vd.EquipmentItems = a
				err := EquipView.Template.ExecuteTemplate(w,
					EquipView.Layout, vd)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	vd := ViewData{}
	vd.EquipmentItems, _ = model.GetLast()
	vd.MoreItems, _ = model.GetLast()

	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, d := range flashes {
			vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
		}
		session.Save(r, w)
	}
	if session.Values["authenticated"] == true && session.Values["username"] != nil && session.Values["username"] != 0 {

		user, _ := model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
		vd.User = user
		err := HomeView.Template.ExecuteTemplate(w,
			HomeView.Layout, vd)
		if err != nil {
			panic(err)
		}
	} else {
		err := HomeView.Template.ExecuteTemplate(w,
			HomeView.Layout, vd)
		if err != nil {
			panic(err)
		}
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	vd := ViewData{}
	if session.Values["authenticated"] == false || session.Values["authenticated"] == nil {
		if flashes := session.Flashes(); len(flashes) > 0 {
			for _, d := range flashes {
				vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
			}
			session.Save(r, w)
		}
		switch r.Method {
		case "GET":
			if err := LoginView.Template.ExecuteTemplate(w, LoginView.Layout, vd); err != nil {
				panic(err)
			}
		case "POST":
			{
				if err := r.ParseForm(); err != nil {
					fmt.Print("err", err)
				} else {
					username := r.FormValue("usernameLog")
					password := r.FormValue("passwordLog")
					// Authentication
					if model.NicknameAlreadyUsed(username) {
						user, _ := model.GetByUsername(username)
						// decode base64 String to []byte & compare w peppered pw
						passwordDB, _ := base64.StdEncoding.DecodeString(user.PasswordHashed)

						err := bcrypt.CompareHashAndPassword(passwordDB, []byte(password+UserPwPepper))
						if err == nil {
							session.Values["authenticated"] = true
							session.Values["username"] = username
							session.Save(r, w)
							http.Redirect(w, r, "/", http.StatusFound)
						} else {
							session.AddFlash("Benutzername oder Passwort falsch")
						}
					} else {
						session.AddFlash("Benutzer existiert nicht")
					}
				}
				session.Save(r, w)
				if flashes := session.Flashes(); len(flashes) > 0 {
					http.Redirect(w, r, "/login", http.StatusFound)
				}
			}
		}
	} else {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
