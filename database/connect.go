package database

import (
	"database/sql"
	"errors"
	"fmt"
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
func PostNewUser(db *sql.DB, username string, password string) error {
	var insertUser = "INSERT INTO public.users(username, password, avatar_url, dob, created_at, role) VALUES ($1, $2, $3, $4, current_timestamp,0);"
	avatar := "https://res.cloudinary.com/anhthu1219/image/upload/v1608561347/book_cover/customLogo_ppubap.jpg"
	dob := "1970-01-01"
	_, err := db.Exec(insertUser, username, password, avatar, dob)
	if err != nil {
		return err
	}
	return nil
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
func GetUserWithName(db *sql.DB, username string) (int, string, string, string, error) {
	row, err := db.Query("SELECT id, avatar_url, dob, role FROM public.users WHERE username=$1", username)
	if err != nil {
		return -1, "", "", "", err
	}
	var (
		ID        int
		AvatarURL string
		DOB       string
		Role      string
	)
	for row.Next() {
		err := row.Scan(&ID, &AvatarURL, &DOB, &Role)
		if err != nil {
			return -1, "", "", "", err
		}
	}

	return ID, AvatarURL, DOB, Role, nil
}

func GetUserByID(db *sql.DB, id string) (int, string, string, string, error) {
	row, err := db.Query("SELECT id, username, avatar_url, dob FROM public.users WHERE id=$1", id)
	if err != nil {
		return -1, "", "", "", err
	}
	var (
		ID        int
		Username  string
		AvatarURL string
		DOB       string
	)
	for row.Next() {
		err := row.Scan(&ID, &Username, &AvatarURL, &DOB)
		if err != nil {
			return -1, "", "", "", err
		}
	}

	return ID, Username, AvatarURL, DOB, nil
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

func GetListBookHeaderWithParam(db *sql.DB, queryString string, id string) (ListBookHeader []models.BookHeader, err error) {
	rows, err := db.Query(queryString, id)
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
func GetListBookHeaderWith3Param(db *sql.DB, queryString string, id string) (ListBookHeader []models.BookHeader, err error) {
	rows, err := db.Query(queryString, id, id, id)
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
func GetListReviewofUser(db *sql.DB, queryString string, id string) (ListReview []models.ReviewOfUser, err error) {
	rows, err := db.Query(queryString, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var review models.ReviewOfUser
		err = rows.Scan(&review.ID, &review.BookID, &review.BookTitle, &review.BookCover, &review.Rating, &review.Title, &review.Review, &review.CreatedAt)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		ListReview = append(ListReview, review)
	}

	return
}
func PostANewAuthor(db *sql.DB, queryString string, author string) error {
	//check duplicate
	getDuplicate := "SELECT COUNT (*) FROM public.authors WHERE name=$1"
	row := db.QueryRow(getDuplicate, author)

	var check int
	err := row.Scan(&check)

	if err != nil {
		return err
	}

	if check == 1 {
		return errors.New("Already exist!")
	}
	_, err = db.Exec(queryString, author)
	if err != nil {
		return err
	}
	return nil
}
func GetListAuthor(db *sql.DB, queryString string) (ListAuthor []models.Author, err error) {
	rows, err := db.Query(queryString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var author models.Author
		err = rows.Scan(&author.ID, &author.Name)
		if err != nil {
			return nil, err
		}
		ListAuthor = append(ListAuthor, author)
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

func GetNameCategory(db *sql.DB, queryString string, category_id int) (name string, err error) {
	rows := db.QueryRow(queryString, category_id)
	if err != nil {
		return "", err
	}
	err = rows.Scan(&name)
	if err != nil {
		return "", err
	}
	return
}
func GetListCategoryName(db *sql.DB, queryString string) (ListCategory []models.Category, err error) {
	rows, err := db.Query(queryString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cate models.Category
		err = rows.Scan(&cate.ID, &cate.Name)
		if err != nil {
			return nil, err
		}
		ListCategory = append(ListCategory, cate)
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
func PostANewPublisher(db *sql.DB, addFavourString string, publisherName string) (err error) {

	row := db.QueryRow("SELECT COUNT (*) FROM public.publishers WHERE name=$1", publisherName)
	var check int
	err = row.Scan(&check)

	if err != nil {
		return err
	}

	if check == 1 {
		return errors.New("Already exist!")
	}
	_, err = db.Exec(addFavourString, publisherName)
	if err != nil {
		return err
	}
	return nil
}
func GetListPublisher(db *sql.DB, queryString string) (ListPublisher []models.Publisher, err error) {
	rows, err := db.Query(queryString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var publisher models.Publisher
		err = rows.Scan(&publisher.ID, &publisher.Name)
		if err != nil {
			return nil, err
		}
		ListPublisher = append(ListPublisher, publisher)
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
func PostANewBook(db *sql.DB, addNewBookString string, Title string, Description string, Publisher string, Cover string, Author string) (int, error) {
	row := db.QueryRow("SELECT COUNT (*) FROM public.books WHERE title=$1", Title)
	check := 0
	err := row.Scan(&check)
	fmt.Println(check)
	if err != nil {

		return -1, err
	}
	if check >= 1 {
		return -1, errors.New("Already exist!")
	}
	rowid := db.QueryRow(addNewBookString, Title, Description, Publisher, Cover, Author)
	if err != nil {
		return -1, err
	}
	id := -1
	err = rowid.Scan(&id)
	if err != nil {
		return -1, err
	}
	fmt.Println(" here")
	return id, nil
}
func PostANewBookwithCategory(db *sql.DB, bookID int, categoryID int) error {
	var addNewBookCate = "INSERT INTO public.category_book(book_id, category_id) VALUES ( $1, $2);"
	_, err := db.Exec(addNewBookCate, bookID, categoryID)
	if err != nil {
		return err
	}
	return nil
}
func GetBookbyID(db *sql.DB, getBookString string, id int) (ID int, Title string, Description string, CreatedAt string, PublisherID int, PublisherName string, Cover string, AuthorID int, AuthorName string, Category []models.Category, err error) {
	rows := db.QueryRow(getBookString, id)

	err = rows.Scan(&ID, &Title, &Description, &CreatedAt, &PublisherID, &Cover, &AuthorID)
	if err != nil {
		return -1, "", "", "", -1, "", "", -1, "", nil, err
	}

	var getAuthorString string = "SELECT name FROM public.authors WHERE id=$1"
	rowA := db.QueryRow(getAuthorString, AuthorID)

	err = rowA.Scan(&AuthorName)
	if err != nil {
		return -1, "", "", "", -1, "", "", -1, "", nil, err
	}

	var getPublisherString string = "SELECT name FROM public.publishers WHERE id=$1"
	rowP := db.QueryRow(getPublisherString, PublisherID)

	err = rowP.Scan(&PublisherName)
	if err != nil {
		return -1, "", "", "", -1, "", "", -1, "", nil, err
	}

	var getCategoryString = "SELECT t.id, t.name FROM   public.categories t join public.category_book cb on t.id = cb.category_id join public.books b on b.id = cb.book_id where b.id=$1;"
	rowCs, err := db.Query(getCategoryString, ID)
	if err != nil {
		return -1, "", "", "", -1, "", "", -1, "", nil, err
	}
	for rowCs.Next() {

		var category models.Category
		err = rowCs.Scan(&category.ID, &category.Name)
		if err != nil {
			return -1, "", "", "", -1, "", "", -1, "", nil, err
		}
		Category = append(Category, category)
	}

	return

}
func PostFavourABook(db *sql.DB, addFavourString string, userID string, bookID string) (err error) {
	_, err = db.Exec(addFavourString, userID, bookID)
	if err != nil {
		return err
	}
	return nil
}

func PostReviewABook(db *sql.DB, addReviewString string, userID string, bookID string, rating string, title string, rateReview string, createdAt string) (err error, check bool) {
	var isExist int
	row := db.QueryRow("SELECT COUNT (*) FROM public.list_rating_books where user_id=$1 and book_id=$2", userID, bookID)
	err = row.Scan(&isExist)
	if err != nil {
		fmt.Println("isExist")
		return err, false
	}
	if isExist != 0 {
		fmt.Println("isExist2")
		return nil, false
	}
	fmt.Println(isExist)
	_, err = db.Exec(addReviewString, userID, bookID, rating, title, rateReview, createdAt)
	if err != nil {
		fmt.Println("isExist3")
		return err, false
	}
	return nil, true
}

func GetSumReviewofBook(db *sql.DB, queryString string, bookID string) (count int, err error) {
	row := db.QueryRow(queryString, bookID)
	err = row.Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil

}

func GetListReviewofBook(db *sql.DB, queryString string, bookID string) (ListReview []models.ReviewOfBook, err error) {
	rows, err := db.Query(queryString, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var review models.ReviewOfBook
		err = rows.Scan(&review.ID, &review.UserID, &review.Username, &review.UserAva, &review.Rating, &review.Title, &review.Review, &review.CreatedAt)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		ListReview = append(ListReview, review)
	}

	return
}
