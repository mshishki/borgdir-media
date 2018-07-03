package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Equipment struct {
	ID          int
	Name        string
	Description string
	Category    string
	ItemCount   int
	PictureURL  string
	Storage     string
	AddedBy     int
	Ordered     []Orders
}

type Orders struct {
	ID       int
	Rented   string
	Return   string
	MadeBy   int //fk for User.ID
	Contains []OrderedEquipment
}

type OrderedEquipment struct { //can also be used for viewing whatever admins have added
	OrdEq    Equipment
	Quantity int
}

// Db handle
var Db *sql.DB

func init() {
	var err error
	Db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}

}

/* Actions with Equipment TABLE */
func GetAll() (equipment []Equipment, err error) {
	rows, err := Db.Query("SELECT ID, Name, Description, Category, ItemCount, PictureURL, Storage, AddedBy FROM Equipment")
	if err != nil {
		return
	}
	for rows.Next() {
		equip := Equipment{}
		err = rows.Scan(&equip.ID, &equip.Name, &equip.Description, &equip.Category, &equip.ItemCount, &equip.PictureURL, &equip.Storage, &equip.AddedBy)
		if err != nil {
			fmt.Println(err)
			return
		}

		equipment = append(equipment, equip)
	}
	rows.Close()
	return
}

func GetAllByCategory(category, name string) (equipment []Equipment, err error) {
	var rows *sql.Rows

	if strings.Compare(category, "") == 0 || len(category) == 0 {
		rows, err = Db.Query("SELECT ID, Name, Description, Category, ItemCount, PictureURL, Storage, AddedBy FROM Equipment WHERE Name LIKE $1;", "%"+name+"%")

	} else if strings.Compare(name, "") == 0 || len(name) == 0 {
		rows, err = Db.Query("SELECT ID, Name, Description, Category, ItemCount, PictureURL, Storage, AddedBy FROM Equipment WHERE Category=$1;", category)

	} else {
		rows, err = Db.Query("SELECT ID, Name, Description, Category, ItemCount, PictureURL, Storage, AddedBy FROM Equipment WHERE Category=$1 AND Name LIKE $2;", category, "%"+name+"%")

	}

	if err != nil {
		fmt.Println(err)

		return
	}
	for rows.Next() {
		equip := Equipment{}
		err = rows.Scan(&equip.ID, &equip.Name, &equip.Description, &equip.Category, &equip.ItemCount, &equip.PictureURL, &equip.Storage, &equip.AddedBy)
		if err != nil {
			fmt.Printf("%v , %T", equip.Category, equip.Category)
			fmt.Println(err)
			return
		}
		equipment = append(equipment, equip)
	}
	//fmt.Println(equipment)
	rows.Close()
	return equipment, err
}

func (equip *Equipment) ChangeEquip() (err error) {
	_, err = Db.Exec("UPDATE Equipment SET Name=$1, Description=$2, Category= $3, ItemCount = $4, PictureURL= $5, Storage= $6, AddedBy=$7 WHERE ID = $8", equip.Name, equip.Description, equip.Category, equip.ItemCount, equip.PictureURL, equip.Storage, equip.AddedBy, equip.ID)
	fmt.Println(err)
	return err

}

func GetMyEquipment(id int) (equipment []Equipment, err error) {
	rows, err := Db.Query("SELECT ID, Name, Description, Category, ItemCount, PictureURL, Storage, AddedBy FROM Equipment WHERE AddedBy = $1;", id)
	if err != nil {

		return
	}
	for rows.Next() {
		equip := Equipment{}
		err = rows.Scan(&equip.ID, &equip.Name, &equip.Description, &equip.Category, &equip.ItemCount, &equip.PictureURL, &equip.Storage, &equip.AddedBy)
		if err != nil {
			fmt.Println("getmyeqq err", err)
			return
		}

		equipment = append(equipment, equip)
	}
	rows.Close()
	return
}

func GetItem(id int) (equip Equipment) {
	equip = Equipment{}
	err := Db.QueryRow("SELECT ID, Name, Description, Category, ItemCount, PictureURL, Storage, AddedBy FROM Equipment WHERE id = $1;", id).Scan(&equip.ID, &equip.Name, &equip.Description, &equip.Category, &equip.ItemCount, &equip.PictureURL, &equip.Storage, &equip.AddedBy)
	if err != nil {
		fmt.Println("error retrieving item", id, err)
	}
	return equip
}

//orders that haven't yet been returned
func (user *User) GetCurrentOrders() {
	user.GetOrders()
	var a []Orders
	for _, d := range user.OrdersList {
		tmp, _ := time.Parse("02-01-2006", d.Return)
		if tmp.After(time.Now()) {
			a = append(a, d)
		} else {
			continue
		}
		user.OrdersList = a
	}
	return
}

func (equip *Equipment) WhoOrdered() {
	orders := []Orders{}

	rows, err := Db.Query("SELECT ID, Rented, Return, MadeBy FROM Orders JOIN OrderEquipment ON Orders.ID = OrderEquipment.OrderID WHERE OrderEquipment.EquipmentID = $1;", equip.ID)
	if err != nil {
		fmt.Println("error retrieving orders from ", equip.ID)
	}

	for rows.Next() {
		order := Orders{}
		err = rows.Scan(&order.ID, &order.Rented, &order.Return, &order.MadeBy)

		orders = append(orders, order)
	}
	equip.Ordered = orders

	return
}

func (equip *Equipment) WhoOrderedCurrently() {
	equip.WhoOrdered()
	var a []Orders
	for _, d := range equip.Ordered {
		tmp, _ := time.Parse("02-01-2006", d.Return)
		if tmp.After(time.Now()) {
			a = append(a, d)
		} else {
			continue
		}
	}
	equip.Ordered = a
	return
}

//reserved: see when people who ordered equipment are going to return it
func CheckForReturnDates(id int) string {
	rows, err := Db.Query("SELECT Return FROM Orders JOIN OrderEquipment ON Orders.ID = OrderEquipment.OrderID WHERE OrderEquipment.EquipmentID = $1;", id)
	if err != nil {
		fmt.Println("ERROR CHECKFOR RETURN DATES --------------", err)
	}
	var dates []string
	for rows.Next() {
		var a string
		err := rows.Scan(&a)
		if err != nil {
			fmt.Println("nicht entliehen")
			return "-"
		}
		dates = append(dates, a)
	}
	rows.Close()
	return NextAvailable(dates)
}

func (order *Orders) PlaceOrder() error {
	_, err := Db.Exec("INSERT INTO Orders(Rented, Return, MadeBy) VALUES (strftime('%d-%m-%Y', 'now'),  strftime('%d-%m-%Y', 'now', '+1 month'), $1);", order.MadeBy)
	if err != nil {
		panic(err)
	}
	//get the id of the order that's just been made
	row := Db.QueryRow("SELECT ID FROM Orders WHERE MadeBy=$1 ORDER BY ID DESC LIMIT 1;", order.MadeBy)
	_ = row.Scan(&order.ID)
	for _, d := range order.Contains {
		//add items & quantity to the linking table
		_, err = Db.Exec("INSERT INTO OrderEquipment VALUES ($1, $2, $3);", order.ID, d.OrdEq.ID, d.Quantity)
		//reduce ItemCount
		_, err = Db.Exec("UPDATE Equipment SET ItemCount=ItemCount-$1 WHERE ID=$2;", d.Quantity, d.OrdEq.ID)
	}
	return err
}

//index: display 8 last added items from database in caroussel --
func GetLast() (equipment []Equipment, err error) {
	rows, err := Db.Query("SELECT ID, Name, PictureURL FROM Equipment WHERE ItemCount IS NOT 0 AND PictureURL IS NOT NULL ORDER BY RANDOM() DESC LIMIT 4;")
	//TODO  Equipment table: ORDER BY ID DESC;
	//use JS to append 4 elems at once???
	if err != nil {
		return
	}
	for rows.Next() {
		equip := Equipment{}
		err = rows.Scan(&equip.ID, &equip.Name, &equip.PictureURL)
		if err != nil {
			fmt.Println(err)
			return
		}
		equipment = append(equipment, equip)
	}
	rows.Close()
	return
}

//add equipment
func (equip *Equipment) AddEquip() (err error) {
	statement := "INSERT INTO Equipment (Name, Description, Category, ItemCount, PictureURL, Storage, AddedBy) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	//sessions: $7 is the one who adds equipment -- is it needed? use []Orders in User struct instead?
	stmt, err := Db.Prepare(statement)

	if err != nil {
		return
	}

	defer stmt.Close()
	_, err = stmt.Exec(equip.Name, equip.Description, equip.Category, equip.ItemCount, equip.PictureURL, equip.Storage, equip.AddedBy)
	return
}

func EquipNo() (number int) {
	row := Db.QueryRow("SELECT ID FROM Equipment ORDER BY ID DESC LIMIT 1;")
	_ = row.Scan(&number)
	return number + 1
}

func (equip *Equipment) DeleteEquipment() (err error) {
	_, err = Db.Exec("DELETE FROM OrderEquipment WHERE EquipmentID=$1", equip.ID)
	_, err = Db.Exec("DELETE FROM Equipment WHERE ID = $1", equip.ID)

	return err
}

func (order *Orders) ExtendRent(id, quantity int) (err error) {
	var itemCount int
	row := Db.QueryRow("SELECT ItemCount FROM Equipment WHERE ID = $1", id)
	_ = row.Scan(&itemCount)
	//Extend only if other items w/ the same ID are available (otherwise might be reserved by other users)
	if itemCount == 0 {
		return errors.New("Letzte Instanz von diesem Equipment; Verlängerung unmöglich")
	} else {
		_, err := Db.Exec("DELETE FROM OrderEquipment WHERE OrderID=$1 AND EquipmentID=$2", order.ID, id)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		_, err = Db.Exec("INSERT INTO Orders(Rented, Return, MadeBy) VALUES ($1,  strftime('%d-%m-%Y', 'now', '+1 month'), $2);", order.Rented, order.MadeBy)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		_ = Db.QueryRow("SELECT ID FROM Orders WHERE MadeBy=$1 ORDER BY ID DESC LIMIT 1;", order.MadeBy).Scan(&order.ID)
		_, err = Db.Exec("INSERT INTO OrderEquipment VALUES($1,$2,$3);", order.ID, id, quantity)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)

	}

	return err
}
