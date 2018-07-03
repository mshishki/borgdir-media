package controller

import (
	"borgdir-media/model"
	"html/template"
)

//Normal Views
var HomeView, EquipView, LoginView, ProfileView, RegisterView, MyEquipView, CartView, ItemView, NotFoundView *View

//Admin views
var AdminView, ClientView, EditClientView, AdmEquipView *View

func init() {
	HomeView = NewView("layout", "views/index.gohtml")
	EquipView = NewView("layout", "views/equipment.gohtml")
	LoginView = NewView("layout", "views/login.gohtml")
	ProfileView = NewView("layout", "views/profile.gohtml")
	RegisterView = NewView("layout", "views/register.gohtml")
	MyEquipView = NewView("layout", "views/my-equipment.gohtml")
	CartView = NewView("layout", "views/cart.gohtml")
	NotFoundView = NewView("layout", "views/templates/notfound.gohtml")
	ItemView = NewView("layout", "views/item.gohtml")

	AdminView = NewView("layout", "views/admin/index.gohtml")
	ClientView = NewView("layout", "views/admin/clients.gohtml")
	AdmEquipView = NewView("layout", "views/admin/equipment.gohtml")
}

func NewView(layout string, files ...string) *View {
	files = append(files,
		"views/templates/header.gohtml",
		"views/templates/layout.gohtml",
		"views/templates/item-view.gohtml",
		"views/templates/forbidden.gohtml",
		"views/templates/session-messages.gohtml")
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

type ViewData struct {
	User           model.User        //header info
	EquipmentItems []model.Equipment //two should be enough for several purposes where equipment slices displayed at once multiple times: 1. my equipment - display both ordered & reserved; index: should somehow parse two slices in a carroussel
	MoreItems      []model.Equipment
	SelectedItem   model.Equipment
	SelectedUser   model.User   // viewing & editing(admin only) user profile; no need if you want to view your own profile
	UserList       []model.User //lst for admins
	SeshItems      []SessionItem
	SeshFlashes    []string
}

type View struct {
	Template *template.Template
	Layout   string
}
