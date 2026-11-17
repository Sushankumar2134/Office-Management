package main
import(
	"database/sql"
	"fmt"
	"log"
	_"github.com/jackc/pgx/v5/stdlib"
)
var db *sql.DB
func InitDB() {
	conStr := "host=localhost port=5432 user=postgres password=123 dbname=office_management sslmode=disable"
	var err error

	db, err = sql.Open("pgx", conStr)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
                                                                                                       
	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database: ", err)
	}
	var dbName string

err = db.QueryRow("SELECT current_database()").Scan(&dbName)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Connected Database:", dbName)
	fmt.Println("Successfully connected to the database!")
}

