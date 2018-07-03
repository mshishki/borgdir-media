package model

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TodaysDate() string { //for renting equipment / registration purposes
	time := time.Now().Format("02-01-2006")
	return time
}

func ReturnDate() string { //2 weeks added to the date of order placement
	time := time.Now().AddDate(0, 1, 0).Format("02-01-2006")
	return time
}

//prepare dates from the database for further comparison
func ParseDates(files []string) []time.Time {
	dates := []time.Time{}
	for i := 0; i < len(files); i++ {
		a, _ := time.Parse("02-01-2006", files[i])
		dates = append(dates, a)
	}
	return dates
}

//look up nearest return date of reserved item
func NextAvailable(dates []string) string {
	if len(dates) != 0 {
		parsedDates := ParseDates(dates)
		var futureDates []time.Time
		for i := 0; i < len(parsedDates); i++ {
			if parsedDates[i].After(time.Now()) {
				futureDates = append(futureDates, parsedDates[i])
			} else {
				continue
			}
		}
		//TODO error if array is empty
		tmp := futureDates[0]
		for i := 1; i < len(futureDates); i++ {
			if futureDates[i].After(tmp) {
				continue
			} else if futureDates[i].Before(tmp) {
				tmp = futureDates[i]
			}
		}
		return tmp.Format("02-01-2006")
	} else {
		return "nicht verfügbar"
	}
}

//Reserved: Checking the equipment item w/ the nearest available date
//5, 9, 14
//SQLITE3 doesn't have a DATE datatype, but we only have a 1 month limit of renting (see wireframes). So before we start comparing strings,  even using SELECT Return FROM Orders ORDER BY Return DESC LIMIT BY 1 wouldn't be such a big deal in theory
//SELECT Rented FROM Orders WHERE Rented  BETWEEN julianday('2018-04-29') AND julianday('2018-06-29');

//das funktioniert! Nur die innerhalb vom letzten Monat betätigten Bestellungen
// SELECT * FROM Orders WHERE Rented BETWEEN strftime('%d-%m-%Y','now', '-1 month') AND strftime('%d-%m-%Y', 'now');

//8-07-18, 5-07-18, 29-07-18
//SELECT Return FROM Orders JOIN OrderEquipment ON Orders.ID = OrderEquipment.OrderID WHERE OrderEquipment.EquipmentID = '7' ORDER BY strftime('%d-%m-%Y', Rented) LIMIT 1;

//SELECT Rented FROM Orders ORDER BY strftime('%d-%m-%Y', Rented);
