package flight

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func GetAllFlights() []FlightPlan {
	db := openDatabase()
	defer db.Close()

	rows, err := db.Query("SELECT id, origin, dest, est_travel_time FROM plans")
	checkError(err)

	defer rows.Close()

	var plans []FlightPlan

	for rows.Next() {
		plan := FlightPlan{}
		err := rows.Scan(&plan.Id, &plan.Origin.name, &plan.Destination.name, &plan.EstTravelTime)
		checkError(err)
		plans = append(plans, plan)
	}

	err = rows.Err()
	checkError(err)

	return plans
}

func CreateFlightPlan(plan FlightPlan) FlightPlan {
	db := openDatabase()
	defer db.Close()

	stmt, _ := db.Prepare("INSERT INTO plans (origin, dest, est_travel_time) VALUES (?, ?, ?) RETURNING id, origin, dest, est_travel_time")
	rows, _ := stmt.Query(plan.Origin.name, plan.Destination.name, plan.EstTravelTime)
	defer stmt.Close()
	defer rows.Close()

	checkError(rows.Err())

	rows.Next()
	updated := FlightPlan{}
	err := rows.Scan(&plan.Id, &plan.Origin.name, &plan.Destination.name, &plan.EstTravelTime)
	checkError(err)
	checkError(rows.Err())

	return updated
}

func DeleteFlightPlan(id int) {
	log.Printf("Deleting plan with id %d from the db", id)
	db := openDatabase()
	defer db.Close()

	stmt, _ := db.Prepare("DELETE FROM plans WHERE id = ?")
	defer stmt.Close()

	result, err := stmt.Exec(id)

	checkError(err)

	rowsDeleted, err := result.RowsAffected()

	checkError(err)

	log.Printf("Deleted %d plans from the db.", rowsDeleted)
}

func UpdateFlightPlan(plan FlightPlan) (FlightPlan, error) {
	return FlightPlan{}, nil
}

func openDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./flight.db")
	checkError(err)
	return db
}

func createTables() {
	createPlans := `
    create table plans (
      id integer not null primary key,
      origin text not null,
      dest text not null,
      est_travel_time integer
    );
  `
	db := openDatabase()
	defer db.Close()

	_, err := db.Exec(createPlans)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
