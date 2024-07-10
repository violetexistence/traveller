package flight

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func GetAllFlights() []FlightPlan {
	db := openDatabase()
	defer db.Close()

	rows, err := db.Query("SELECT id, origin, dest, est_travel_time, created_date FROM plans ORDER BY created_date DESC")
	checkError(err)

	defer rows.Close()

	var plans []FlightPlan

	for rows.Next() {
		plan := FlightPlan{}
		var createdDate string
		err := rows.Scan(&plan.Id, &plan.Origin.Name, &plan.Destination.Name, &plan.EstTravelTime, &createdDate)
		checkError(err)
		plan.CreatedDate, _ = time.Parse(time.RFC3339, createdDate)
		plans = append(plans, plan)
	}

	err = rows.Err()
	checkError(err)

	return plans
}

func CreateFlightPlan(plan FlightPlan) FlightPlan {
	db := openDatabase()
	defer db.Close()

	stmt, _ := db.Prepare("INSERT INTO plans (origin, dest, est_travel_time, created_date) VALUES (?, ?, ?, ?) RETURNING id, origin, dest, est_travel_time, created_date")
	rows, _ := stmt.Query(plan.Origin.Name, plan.Destination.Name, plan.EstTravelTime, plan.CreatedDate.Format(time.RFC3339))
	defer stmt.Close()
	defer rows.Close()

	checkError(rows.Err())

	rows.Next()
	updated := FlightPlan{}
	var createdDate string
	err := rows.Scan(&plan.Id, &plan.Origin.Name, &plan.Destination.Name, &plan.EstTravelTime, &createdDate)
	checkError(err)
	checkError(rows.Err())
	plan.CreatedDate, _ = time.Parse(time.RFC3339, createdDate)

	return updated
}

func DeleteFlightPlan(id int) {
	db := openDatabase()
	defer db.Close()

	stmt, _ := db.Prepare("DELETE FROM plans WHERE id = ?")
	defer stmt.Close()

	result, err := stmt.Exec(id)

	checkError(err)

	_, err = result.RowsAffected()

	checkError(err)
}

func UpdateFlightPlan(plan FlightPlan) (FlightPlan, error) {
	return FlightPlan{}, nil
}

func openDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./flight.db")
	checkError(err)
	return db
}

func CreateTables() {
	createPlans := `
    create table if not exists plans (
      id integer not null primary key,
      origin text not null,
      dest text not null,
      est_travel_time integer not null,
      created_date text not null
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
