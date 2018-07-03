package model

import (
	"encoding/base64"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	FirstName      string
	LastName       string
	Email          string
	Username       string
	Password       string
	PasswordHashed string
	PictureURL     string
	IsAdmin        bool
	Status         string
	OrdersList     []Orders
}

//TODO peppering our pwd: move to config upon successful test
var UserPwPepper = "t0p-sEcR3t"

// Get user data by addressing ID -- for displaying profile.gohtml
func Get(id int) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, FirstName, LastName, Email, Username, Password, PictureURL, IsAdmin, Status FROM User WHERE id = $1", id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.PasswordHashed, &user.PictureURL, &user.IsAdmin, &user.Status)
	return
}

func GetByUsername(username string) (user User, err error) {
	user = User{}
	if Db.QueryRow("SELECT Username FROM User WHERE Username = $1", username).Scan(&username) == nil {
		err = Db.QueryRow("SELECT ID, FirstName, LastName, Email, Username, Password, PictureURL, IsAdmin, Status FROM User WHERE Username = $1", username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.PasswordHashed, &user.PictureURL, &user.IsAdmin, &user.Status)
		return
	} else {
		fmt.Println("error:", err)
		return
	}
}

func GetIDByUsername(username string) (id int) {
	if err := Db.QueryRow("SELECT Username FROM User WHERE Username = $1", username).Scan(&username); err == nil {
		err = Db.QueryRow("SELECT ID FROM User WHERE Username = $1", username).Scan(&id)
		return id
	} else {
		fmt.Println("error:", err)
		return 0
	}
}

//temporary login func for checking whether username exists
func GetUsername(username string) bool {
	if Db.QueryRow("SELECT Username FROM User WHERE Username = $1", username).Scan(&username) == nil {
		return true
	} else {
		return false
	}
}

//for registration: check if email or username are already in the database
func EmailAlreadyUsed(email string) bool {
	if err := Db.QueryRow("SELECT Email FROM User WHERE Email = $1", email).Scan(&email); err != nil {
		return false
	}
	return true
}
func NicknameAlreadyUsed(username string) bool {
	if err := Db.QueryRow("SELECT Username FROM User WHERE Username = $1", username).Scan(&username); err != nil {
		return false
	}
	return true
}

//TODO replace
func AlreadyUsed(username, email string) bool {
	if Db.QueryRow("SELECT Username, Email FROM User WHERE Username = $1 OR Email = $2", username, email).Scan(&username, &email) != nil {
		return false
	}
	return true
}

//TODO status "aktiv" und IsAdmin=0 als default-werte in sowohl db als auch structs

// Register
func (user *User) AddUser() (err error) {
	//salt-and-peppering the password
	pwBytes := []byte(user.Password + UserPwPepper) //bcrypt uses byte slices, not strings
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHashed = base64.StdEncoding.EncodeToString(hashedBytes)
	user.Password = "" //having no cleartext pwd is supposedly good practice

	statement := "INSERT INTO User (FirstName, LastName, Username, Password, Email, PictureURL, IsAdmin, Status) values ($1, $2,$3, $4, $5,  'http://mwrsupply.com/wp-content/uploads/2017/12/person-placeholder.jpg', 0, 'Nutzer')"
	stmt, err := Db.Prepare(statement)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer stmt.Close()
	_, err = stmt.Exec(user.FirstName, user.LastName, user.Username, user.PasswordHashed, user.Email)
	return
}

//ADMIN FUNCreturns all Users who aren't admins
//TODO orders einbinden
func GetAllUsers() (users []User, err error) {
	rows, err := Db.Query("SELECT ID, FirstName, LastName, Email, Username, Password, PictureURL, Status FROM User WHERE IsAdmin = 0")
	if err != nil {
		return
	}
	for rows.Next() {
		user := User{}
		err = rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.PasswordHashed, &user.PictureURL, &user.Status)
		if err != nil {
			fmt.Println(err)
			return
		}
		users = append(users, user)
	}
	rows.Close()
	return

}

func (user *User) GetOrders() {
	orders := []Orders{}
	rows, err := Db.Query("SELECT ID, Rented, Return, MadeBy From Orders WHERE Orders.MadeBy = $1 ORDER BY ID DESC;", user.ID)
	if err != nil {
		fmt.Println("error retrieving orders from ", user.ID)
	}
	for rows.Next() {
		order := Orders{}
		rows.Scan(&order.ID, &order.Rented, &order.Return, &order.MadeBy)

		equiprows, err := Db.Query("SELECT OrderEquipment.EquipmentID, OrderEquipment.Quantity From Orders JOIN OrderEquipment ON Orders.ID = OrderEquipment.OrderID WHERE Orders.MadeBy=$1 AND Orders.ID = $2;", user.ID, order.ID)
		if err != nil {
			fmt.Println("error retrieving equipment rows in order")
		}
		all := []OrderedEquipment{}

		for equiprows.Next() {
			one := OrderedEquipment{}
			equiprows.Scan(&one.OrdEq.ID, &one.Quantity)
			one.OrdEq = GetItem(one.OrdEq.ID)
			all = append(all, one)
			order.Contains = all
		}
		orders = append(orders, order)
		user.OrdersList = orders

	}
	return
}

func (user *User) Admin() bool {
	if _, err := Db.Query("SELECT * FROM User WHERE ID = $1 AND IsAdmin = 1", user.ID); err != nil {
		return false
	} else {
		return true
	}
}

func (user *User) ChangeProfile() (err error) {
	//same hashing as in AddUser()
	if user.Password == "" {
		_, err = Db.Exec("UPDATE User SET FirstName = $1, LastName=$2, Username=$3, Email=$4, PictureURL=$6 WHERE ID = $7", user.FirstName, user.LastName, user.Username, user.Email, user.PictureURL, user.ID)
		return err
	} else {

		pwBytes := []byte(user.Password + UserPwPepper)
		hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.PasswordHashed = base64.StdEncoding.EncodeToString(hashedBytes)
		user.Password = ""

		_, err = Db.Exec("UPDATE User SET FirstName = $1, LastName=$2, Username=$3, Email=$4, Password = $5, PictureURL=$6 WHERE ID = $7", user.FirstName, user.LastName, user.Username, user.Email, user.PasswordHashed, user.PictureURL, user.ID)
		return err
	}
}

func (user *User) HasEmail(value string) bool {
	var tempoEmail string
	if err := Db.QueryRow("SELECT Email FROM User WHERE ID = $1;", user.ID).Scan(&tempoEmail); err != nil {
		return false
	} else {

		if tempoEmail == value {
			return true
		} else {

			return false
		}
	}
}
func (user *User) HasNickname(value string) bool {
	var tempoNick string

	if err := Db.QueryRow("SELECT Username FROM User WHERE ID = $1;", user.ID).Scan(&tempoNick); err != nil {
		return false
	} else {

		if tempoNick == value {
			return true
		} else {

			return false
		}
	}
}

func (user *User) BlockUser() (err error) {
	_, err = Db.Exec("UPDATE User SET Status = 'gesperrt' WHERE ID = $1", user.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("successfully blocked")
	return
}
func (user *User) UnblockUser() (err error) {
	_, err = Db.Exec("UPDATE User SET Status = 'Nutzer' WHERE ID = $1", user.ID)
	if err != nil {
		panic(err)
	}
	return
}

func (user *User) DeleteUser() (err error) {
	//cascading delete: first delete objects from other tables that reference our object in question, then the object itself
	rows, err := Db.Query("SELECT ID FROM Orders WHERE MadeBy = $1", user.ID)

	orderids := []int{}
	for rows.Next() {
		var orderid int
		err = rows.Scan(&orderid)
		orderids = append(orderids, orderid)
	}
	for i, _ := range orderids {
		_, err = Db.Exec("DELETE FROM OrderEquipment WHERE OrderID = $1", orderids[i])
	}
	_, err = Db.Exec("DELETE FROM Orders WHERE MadeBy = $1", user.ID)
	_, err = Db.Exec("DELETE FROM User WHERE ID = $1", user.ID)
	return
}
