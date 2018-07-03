package controller

import (
	"borgdir-media/model"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func DeleteClient(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))

	path := strings.Join(strings.Split(fmt.Sprintf("%s", r.URL.Path[9:]), "/delete"), "") //[:1]
	id, _ := strconv.Atoi(path)
	vd.SelectedUser, _ = model.Get(id)

	if session.Values["authenticated"] == false || vd.User.ID != vd.SelectedUser.ID {
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect) //renders forbidden
	} else {
		vd.SelectedUser.DeleteUser()
		session.Values["authenticated"] = false
		session.Values["username"] = ""
		session.AddFlash("Sie haben Ihr Konto erfolgreich gelöscht.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

/*** Controllers w/ actions only accessible to users ***/

func MyEquipment(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))
	vd.User.GetCurrentOrders()
	vd.SeshItems = RetrieveReservedFromSession(session)
	if err := MyEquipView.Template.ExecuteTemplate(w,
		MyEquipView.Layout, vd); err != nil {
		//		panic(err)
		fmt.Println(err)
	}
}

func Profile(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	vd := ViewData{}
	vd.User, _ = model.GetByUsername(fmt.Sprintf("%v", session.Values["username"]))

	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, d := range flashes {
			vd.SeshFlashes = append(vd.SeshFlashes, fmt.Sprintf("%v", d))
		}
		session.Save(r, w)
	}

	var path = fmt.Sprintf("%s", r.URL.Path[9:]) // /profile/
	id, err := strconv.Atoi(path)
	if id == vd.User.ID {
		vd.SelectedUser = vd.User
	} else {
		vd.SelectedUser, err = model.Get(id)
		if err != nil {
			http.Redirect(w, r, "/notfound", http.StatusMovedPermanently)
		}
	}

	switch r.Method {
	case "GET":
		{
			if err := ProfileView.Template.ExecuteTemplate(w,
				ProfileView.Layout, vd); err != nil {
				panic(err)
			}
		}
	case "POST": //update resource by replacing content = edit info
		{
			r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

			if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
				session.AddFlash("Die von Ihnen ausgewählte Datei ist zu groß")
				session.Save(r, w)
			} else {
				fileHeader := r.MultipartForm.File["picture"][0]
				newPath, err := UploadImage("usr/", vd.SelectedUser.ID, fileHeader)
				if err == nil {
					if newPath != "" {
						vd.SelectedUser.PictureURL = newPath
					} else {
					}
				} else {
					session.AddFlash(err.Error())
				}

				if r.FormValue("password") != "" {
					if strings.Compare(r.FormValue("password"), r.FormValue("passwordRep")) == 0 {
						vd.SelectedUser.Password = r.FormValue("password")
					} else {
						session.AddFlash("Die Passwörter stimmen nicht überein. Bitte überprüfen Sie Ihre Angaben")
					}
				} else {
					//see condition in ChangeProfile():if the password hasn't been changed and the fields are empty, the model will update everything but the pwd
					vd.SelectedUser.Password = ""
				}

				usrname := r.FormValue("username")
				email := r.FormValue("email")

				if model.NicknameAlreadyUsed(usrname) {
					if !vd.SelectedUser.HasNickname(usrname) {
						session.AddFlash("Name wird bereits von einem anderen User verwendet. Bitte ändern Sie Ihre Angaben")
					}
				}

				if model.EmailAlreadyUsed(email) {
					if !vd.SelectedUser.HasEmail(email) {
						session.AddFlash("E-Mail wird bereits von einem anderen User verwendet. Bitte ändern Sie Ihre Angaben")
					}
				}

				vd.SelectedUser.Username = usrname
				vd.SelectedUser.Email = email

				session.Save(r, w)
				if flashes := session.Flashes(); len(flashes) > 0 {
					http.Redirect(w, r, "/profile/"+path, http.StatusFound)
				} else {
					vd.SelectedUser.FirstName = r.FormValue("firstname")
					vd.SelectedUser.LastName = r.FormValue("lastname")

					vd.SelectedUser.ChangeProfile()
					session.AddFlash("Erfolg! Sie können jetzt fortfahren.")
					if vd.SelectedUser.ID == vd.User.ID {
						session.Values["username"] = vd.SelectedUser.Username
						session.Save(r, w)
						http.Redirect(w, r, "/", http.StatusFound)
					} else {
						http.Redirect(w, r, "/admin/clients", http.StatusFound)
					}
				}
			}
		}
	}
}

const uploadPath = "./static/images/"
const MaxUploadSize = 2 * 1024 * 1024 //image can be 2 MB max

//TODO loca: either usr/ or equip/
func UploadImage(loca string, id int, fileHeader *multipart.FileHeader) (newPath string, err error) {
	file, err := fileHeader.Open()
	defer file.Close()

	if err == nil {
		data, err := ioutil.ReadAll(file)
		if err == nil {
			//Show uploaded file in browser: fmt.Fprintln(w, string(data))

			filetype := http.DetectContentType(data)
			fmt.Println("filetype", filetype)
			switch filetype {
			case "image/jpeg", "image/jpg", "image/png":
			default:
				{
					return "", nil
				}
			}

			fileName := strconv.Itoa(id) + "-" + time.Now().Format("02-01-2006_15-04-05")
			fileEndings, err := mime.ExtensionsByType(filetype)
			if err != nil {
				return "", errors.New("Fehler beim Einlesen der Datei")
			}
			newPath = filepath.Join(uploadPath+loca, fileName+fileEndings[0])
			//fmt.Printf("FileType: %s, File: %s\n", filetype, newPath)

			newFile, err := os.Create(newPath)
			if err != nil {
				return "", errors.New("Fehler beim Hochladen der Datei")
			}
			defer newFile.Close() // idempotent, okay to call twice
			if _, err := newFile.Write(data); err != nil || newFile.Close() != nil {
				return "", errors.New("Fehler beim Hochladen der Datei")
			}
		}
	}
	return "/" + newPath, err
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["authenticated"] = false
	session.Values["username"] = nil
	session.AddFlash("Sie sind nun ausgeloggt")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
