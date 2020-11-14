package database

import (
	"database/sql"
	"log"
	"server/backend/models"

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
func GetUserWithName(db *sql.DB, username string) (int, string, string, error) {
	row, err := db.Query("SELECT id, avatar_url, dob FROM public.users WHERE username=$1", username)
	if err != nil {
		return -1, "", "", err
	}
	var (
		ID        int
		AvatarURL string
		DOB       string
	)
	for row.Next() {
		err := row.Scan(&ID, &AvatarURL, &DOB)
		if err != nil {
			return -1, "", "", err
		}
	}

	return ID, AvatarURL, DOB, nil
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

func GetListBookHeader(db *sql.DB, queryString string) (ListBookHeader []models.BookHeader, err error) {
	rows, err := db.Query(queryString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookHeader models.BookHeader
		err = rows.Scan(&bookHeader.ID, &bookHeader.Title, &bookHeader.Cover)
		if err != nil {
			return nil, err
		}
		ListBookHeader = append(ListBookHeader, bookHeader)
	}

	return

}

func GetListAuthorBookHeader(db *sql.DB, queryString string, author_id int) (ListBookHeader []models.BookHeader, err error) {
	rows, err := db.Query(queryString, author_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookHeader models.BookHeader
		err = rows.Scan(&bookHeader.ID, &bookHeader.Title, &bookHeader.Cover)
		if err != nil {
			return nil, err
		}
		ListBookHeader = append(ListBookHeader, bookHeader)
	}

	return
}
func GetListCategoryBookHeader(db *sql.DB, queryString string, category_id int) (ListBookHeader []models.BookHeader, err error) {
	rows, err := db.Query(queryString, category_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookHeader models.BookHeader
		err = rows.Scan(&bookHeader.ID, &bookHeader.Title, &bookHeader.Cover)
		if err != nil {
			return nil, err
		}
		ListBookHeader = append(ListBookHeader, bookHeader)
	}

	return
}

func GetListPublisherBookHeader(db *sql.DB, queryString string, publisher_id int) (ListBookHeader []models.BookHeader, err error) {
	rows, err := db.Query(queryString, publisher_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookHeader models.BookHeader
		err = rows.Scan(&bookHeader.ID, &bookHeader.Title, &bookHeader.Cover)
		if err != nil {
			return nil, err
		}
		ListBookHeader = append(ListBookHeader, bookHeader)
	}

	return
}

func GetBookbyID(db *sql.DB, getBookString string, id int) (ID int, Title string, Description string, CreatedAt string, DeletedAt string, PublisherID int, PublisherName string, Cover string, AuthorID int, AuthorName string, Category []models.Category, err error) {
	rows := db.QueryRow(getBookString, id)

	err = rows.Scan(&ID, &Title, &Description, &CreatedAt, &DeletedAt, &PublisherID, &Cover, &AuthorID)
	if err != nil {
		return -1, "", "", "", "", -1, "", "", -1, "", nil, err
	}

	var getAuthorString string = "SELECT name FROM public.authors WHERE id=$1"
	rowA := db.QueryRow(getAuthorString, AuthorID)

	err = rowA.Scan(&AuthorName)
	if err != nil {
		return -1, "", "", "", "", -1, "", "", -1, "", nil, err
	}

	var getPublisherString string = "SELECT name FROM public.publishers WHERE id=$1"
	rowP := db.QueryRow(getPublisherString, PublisherID)

	err = rowP.Scan(&PublisherName)
	if err != nil {
		return -1, "", "", "", "", -1, "", "", -1, "", nil, err
	}

	var getCategoryString = "SELECT t.id, t.name FROM   public.categories t join public.category_book cb on t.id = cb.category_id join public.books b on b.id = cb.book_id where b.id=$1;"
	rowCs, err := db.Query(getCategoryString, ID)
	if err != nil {
		return -1, "", "", "", "", -1, "", "", -1, "", nil, err
	}
	for rowCs.Next() {

		var category models.Category
		err = rowCs.Scan(&category.ID, &category.Name)
		if err != nil {
			return -1, "", "", "", "", -1, "", "", -1, "", nil, err
		}
		Category = append(Category, category)
	}

	return

}

// func GetBookbyPublisher(db *sql.DB) (ListBook []models.Book, err error) {
// 	var getNewestBookString string = "SELECT * FROM public.books WHERE publisher_id=$1 limit 20"
// 	rows, err := db.Query(getNewestBookString)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var book models.Book
// 		err = rows.Scan(&book.ID, &book.Title, &book.Description, &book.CreatedAt, &book.DeletedAt, &book.Publisher.ID, &book.Cover, &book.Author.ID)
// 		if err == sql.ErrNoRows {
// 			return nil, err
// 		}
// 		var getAuthorString string = "SELECT name FROM public.authors WHERE id=$1"
// 		row, err := db.Query(getAuthorString, book.Author.ID)
// 		if err != nil {
// 			book.Author.Name = ""
// 		}
// 		for row.Next() {
// 			err = row.Scan(&book.Author.Name)
// 			if err != nil {
// 				book.Author.Name = ""
// 			}
// 		}
// 		var getPublisherString string = "SELECT name FROM public.publishers WHERE id=$1"
// 		row, err = db.Query(getPublisherString, book.Publisher.ID)
// 		if err != nil {
// 			book.Publisher.Name = ""
// 		}
// 		for row.Next() {
// 			err = row.Scan(&book.Publisher.Name)
// 			if err != nil {
// 				book.Publisher.Name = ""
// 			}
// 		}
// 		fmt.Println("done get")
// 		ListBook = append(ListBook, book)

// 	}

// 	return

// }
