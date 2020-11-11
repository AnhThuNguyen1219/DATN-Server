package database

import (
	"database/sql"
	"log"

	// This package is necessary for connecting to postgresql
	_ "github.com/lib/pq"
)

// Connect initialize a connection to posgresql
func Connect() (db *sql.DB) {
	log.Println("Connecting to database...")
	connectionString := "host=ec2-52-206-15-227.compute-1.amazonaws.com	port=5432 user=lqacqjnhvzsota dbname=de2tl49hiaplua password=a09e77789b5615f760a736e77f0da141eb62c1b6c9ec55b121840bce84056010 sslmode=require"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalln("Can't connect to database")
	} else {
		log.Println("Connected to database successfully.")
	}
	return
}

// IsUserExist check username whether it is stored in db or not
func IsUserExist(db *sql.DB, username string) bool {
	var (
		number int
	)

	rows, err := db.Query("SELECT COUNT (*) FROM public.users WHERE username=$1", username)
	if err != nil {
		log.Fatalf("IsUserExist: Error happened in query: %s", err)
	}
	for rows.Next() {
		err := rows.Scan(&number)
		if err != nil {
			log.Fatalln("IsUserExist: Error happened in traverse rows")
		}
	}
	if number == 1 {
		return true
	}
	return false
}

// GetHashedPassword get hashed password
func GetHashedPassword(db *sql.DB, username string) (hashedPassword string) {
	rows, err := db.Query("SELECT password FROM public.users WHERE username=$1", username)
	if err != nil {
		log.Fatalln("GetHashedPassword: Error happened in query")
	}
	for rows.Next() {
		err := rows.Scan(&hashedPassword)
		if err != nil {
			log.Fatalln("GetHashedPassword: Error happened in traverse rows")
		}
	}
	return
}

func GetUserIDWithName(db *sql.DB, username string) (int, error) {
	row, err := db.Query("SELECT id FROM public.users WHERE username=$1", username)
	if err != nil {
		return 0, err
	}
	var userid int
	for row.Next() {
		err := row.Scan(&userid)
		if err != nil {
			return 0, err
		}
	}
	return userid, nil
}

func CreateContentInterface(db *sql.DB, name string) (int, error) {
	var id int
	var createContentInterfaceQueryString string = "INSERT INTO public.usr_content_interface(name) VALUES ($1);"
	_, err := db.Exec(createContentInterfaceQueryString, name)
	if err != nil {
		return -1, err
	}
	var selectContentInterfaceIDQuery string = "SELECT id FROM public.usr_content_interface WHERE name=$1;"
	row, err := db.Query(selectContentInterfaceIDQuery, name)
	if err != nil {
		return -1, err
	}
	for row.Next() {
		err := row.Scan(&id)
		if err != nil {
			return -1, err
		}
	}
	return id, nil
}

//Code about save content into database go here
func CreateContentText(db *sql.DB, text string, name string) error {
	id, err := CreateContentInterface(db, name)
	if err != nil {
		return err
	}
	var createContentTextQueryString string = "INSERT INTO public.usr_content_text(id, text)VALUES ($1, $2);"
	_, err = db.Exec(createContentTextQueryString, id, text)
	if err != nil {
		return err
	}
	return nil
}

func CreateContentImage(db *sql.DB, title string, originalImgURL string, previewImgURL string) error {
	id, err := CreateContentInterface(db, title)
	if err != nil {
		return err
	}
	var createContentImgQueryString string = "INSERT INTO public.usr_content_image(id, original_img_url, preview_img_url) VALUES ($1, $2, $3);"
	_, err = db.Exec(createContentImgQueryString, id, originalImgURL, previewImgURL)
	if err != nil {
		return err
	}
	return nil
}
