package controller

import (
	"borgdir-media/model"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/*** Controllers w/ actions only accessible to admins ***/

func Admindex(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session.Values["authenticated"] == false {
		http.Redirect(w, r, "/", http.StatusForbidden)
	} else {
		vd := ViewData{}
		vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
		fmt.Println("user ist admin, eingeloggt als:", vd.User)
		if err := AdminView.Template.ExecuteTemplate(w,
			AdminView.Layout, vd); err != nil {
			panic(err)
		}
	}
	session.Save(r, w)
}

func Clients(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session.Values["authenticated"] == false {
		http.Redirect(w, r, "/", http.StatusForbidden)
	} else {

		vd := ViewData{}
		vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))

		vd.UserList, _ = model.GetAllUsers() //slice mit users vom Type []model.User
		alluserswOrders := []model.User{}
		for _, user := range vd.UserList {
			user.GetOrders()
			alluserswOrders = append(alluserswOrders, user)
		}
		vd.UserList = alluserswOrders
		if err := ClientView.Template.ExecuteTemplate(w,
			ClientView.Layout, vd); err != nil {
			panic(err)
		}
	}
	session.Save(r, w)
}

func DeleteEquip(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
	if session.Values["authenticated"] == false || !vd.User.IsAdmin {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		path := strings.Join(strings.Split(fmt.Sprintf("%s", r.URL.Path[11:]), "/delete"), "") //[:1] /equipment/55
		id, _ := strconv.Atoi(path)
		vd.SelectedItem = model.GetItem(id)
		if err := vd.SelectedItem.DeleteEquipment(); err != nil {
			session.AddFlash("Löschen fehlgeschlagen")
			session.Save(r, w)
			http.Redirect(w, r, "/equipment/"+path, http.StatusFound)

		}
		session.AddFlash("Löschen erfolgreich")
		session.Save(r, w)
		http.Redirect(w, r, "/admin/equipment", http.StatusFound)
	}
}

func BlockClient(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
	if session.Values["authenticated"] == false || !vd.User.IsAdmin {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		path := strings.Join(strings.Split(fmt.Sprintf("%s", r.URL.Path[9:]), "/block"), "") //[:1]
		id, _ := strconv.Atoi(path)
		vd.SelectedUser, _ = model.Get(id)
		if strings.Compare(vd.SelectedUser.Status, "gesperrt") == 0 {
			vd.SelectedUser.UnblockUser()
		} else {
			vd.SelectedUser.BlockUser()
		}
		http.Redirect(w, r, "/admin/clients", http.StatusFound)
	}
	session.Save(r, w)
}

//tmplA.ExecuteTemplate(w, "adminequip", nil)
//GetAll()
//equip.AddEquip(adminid)

func AddEquip(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")

	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))

	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, d := range flashes {
			vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
		}
		session.Save(r, w)
	}

	if vd.User.IsAdmin == false || session.Values["authenticated"] == false {
		//if path is empty: users are redirected to /equipment,
		//admins get an empty form that they are able to fill
		http.Redirect(w, r, "/equipment", http.StatusTemporaryRedirect)
	}

	vd.SelectedItem = model.Equipment{}

	switch r.Method {
	case "GET":
		{
			if err := ItemView.Template.ExecuteTemplate(w,
				ItemView.Layout, vd); err != nil {
				panic(err)
			}
		}
	case "POST":
		{
			equipItem := model.Equipment{}
			r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

			if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
				session.AddFlash("Die von Ihnen ausgewählte Datei ist zu groß")
				session.Save(r, w)
			} else {
				fileHeader := r.MultipartForm.File["picture"][0]
				newPath, err := UploadImage("equip/", model.EquipNo(), fileHeader)
				if err == nil {
					if newPath != "" {
						equipItem.PictureURL = newPath
					}
				} else {
					session.AddFlash(err.Error())
				}

				equipItem.Name = r.FormValue("equipname")
				equipItem.Category = r.FormValue("category")

				equipItem.ID = model.EquipNo()

				equipItem.Storage = r.FormValue("storage")
				equipItem.ItemCount, _ = strconv.Atoi(r.FormValue("itemcount"))
				equipItem.AddedBy = vd.User.ID
				equipItem.Description = r.FormValue("description")

				fmt.Println(equipItem)

				session.Save(r, w)
				if flashes := session.Flashes(); len(flashes) > 0 {
					http.Redirect(w, r, "/equipment/add", http.StatusFound)
				} else {

					if err := equipItem.AddEquip(); err == nil {
						session.Save(r, w)
						http.Redirect(w, r, "/admin/equipment", http.StatusTemporaryRedirect)
					}
				}
			}
		}

	}
}

func Equip(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")

	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))

	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, d := range flashes {
			vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
		}
		session.Save(r, w)
	}

	var path = fmt.Sprintf("%s", r.URL.Path[11:]) // /equipment/
	if strings.Compare(path, "") != 0 {
		id, err := strconv.Atoi(path)
		if err != nil {
			if vd.User.IsAdmin == false || session.Values["authenticated"] == false {
				//if path is empty: users are redirected to /equipment,
				//admins get an empty form that they are able to fill
				http.Redirect(w, r, "/equipment", http.StatusTemporaryRedirect)
			} else {
				fmt.Println("equipment: error retrieving item w id", id)
				http.Redirect(w, r, "/equipment", http.StatusTemporaryRedirect)
			}
		} else {
			vd.SelectedItem = model.GetItem(id)
		}

		switch r.Method {
		case "GET":
			{
				//fmt.Println(model.EquipNo())
				if err := ItemView.Template.ExecuteTemplate(w,
					ItemView.Layout, vd); err != nil {
					panic(err)
				}
			}
		case "POST":
			{
				equipItem := model.Equipment{}
				r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

				if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
					session.AddFlash("Die von Ihnen ausgewählte Datei ist zu groß")
					session.Save(r, w)
				} else {
					fileHeader := r.MultipartForm.File["picture"][0]
					newPath, err := UploadImage("equip/", model.EquipNo(), fileHeader)
					if err == nil {
						if newPath != "" {
							equipItem.PictureURL = newPath
						} else {
							equipItem.PictureURL = vd.SelectedItem.PictureURL
						}
					} else {
						session.AddFlash(err.Error())
					}

					equipItem.Name = r.FormValue("equipname")
					equipItem.Category = r.FormValue("category")

					if strings.Compare(r.FormValue("equip-id"), path) != 0 {
						//we don't want to overwrite old equipment == check if sth w that id already exists
						//if exists, add flash message
						session.AddFlash("Neee, unter dieser Nummer gibt es bestimmt schon etwas")
					} else {
						id, _ := strconv.Atoi(r.FormValue("equip-id"))
						equipItem.ID = id
					}

					equipItem.Storage = r.FormValue("storage")
					equipItem.ItemCount, _ = strconv.Atoi(r.FormValue("itemcount"))
					equipItem.AddedBy = vd.User.ID
					equipItem.Description = r.FormValue("description")
					fmt.Println(equipItem)

					session.Save(r, w)
					if flashes := session.Flashes(); len(flashes) > 0 {
						http.Redirect(w, r, "/equipment/", http.StatusFound)
					} else {

						if err := equipItem.ChangeEquip(); err == nil {
							session.Save(r, w)
							http.Redirect(w, r, "/admin/equipment", http.StatusFound)
						}
					}
				}
			}
		}
	}
}

func AdmEquip(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))

	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, d := range flashes {
			vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
		}
		session.Save(r, w)
	}

	switch r.Method {
	case "GET":
		{
			vd.EquipmentItems, _ = model.GetMyEquipment(vd.User.ID)
			a := []model.Equipment{}
			for _, d := range vd.EquipmentItems {
				d.WhoOrderedCurrently()
				a = append(a, d)
			}
			vd.EquipmentItems = a
			if err := AdmEquipView.Template.ExecuteTemplate(w,
				AdmEquipView.Layout, vd); err != nil {
				panic(err)
			}
		}
	case "POST":
		{
			if err := r.ParseForm(); err != nil {
				fmt.Print("POST err:", err)
			} else {

				var search string = r.FormValue("search")

				vd.EquipmentItems, _ = model.GetAllByCategory("", search)
				a := []model.Equipment{}
				for _, d := range vd.EquipmentItems {
					if d.AddedBy == vd.User.ID {
						d.WhoOrderedCurrently()
						a = append(a, d)
					} else {
						continue
					}
				}
				vd.EquipmentItems = a
				if err := AdmEquipView.Template.ExecuteTemplate(w,
					AdmEquipView.Layout, vd); err != nil {
					panic(err)
				}

			}
		}
	}
}
