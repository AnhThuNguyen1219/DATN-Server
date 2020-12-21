package route

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"server/backend/auth"
	"server/backend/database"
	"server/backend/models"
	"server/backend/utils"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v7"
	"github.com/julienschmidt/httprouter"
)

var (
	db = database.Connect()
)

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params, client *redis.Client) {
	// Convert json data from request to type User
	var user models.User
	err := utils.DecodeJSONBody(w, r, &user)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if govalidator.IsNull(user.Username) || govalidator.IsNull(user.Password) {
		utils.JSON(w, http.StatusBadRequest, "Input cannot be empty")
		return
	}

	// Escape string before query to sql server
	user.Username = models.Santize(user.Username)
	user.Password = models.Santize(user.Password)

	// Check with regex
	match, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]{4,255}", user.Username)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !match {
		utils.JSON(w, http.StatusBadRequest, "Username must longer than 3 characters can only include: a-z, A-Z, 0-9")
		return
	}

	match, err = regexp.MatchString(".{5,255}$", user.Password)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !match {
		utils.JSON(w, http.StatusBadRequest, "Password length must between 5 and 255 characters")
		return
	}

	isUsernameExist := database.IsUserExist(db, user.Username)

	if isUsernameExist == false {
		utils.JSON(w, http.StatusBadRequest, "Username or Password incorrect") // Username is not exist, but show incorrect
		return
	}

	hashedPassword := database.GetHashedPassword(db, user.Username)

	check := models.CheckPasswordHash(hashedPassword, user.Password)

	if check != true {
		utils.JSON(w, http.StatusUnauthorized, "Username or Password incorrect")
		return
	}

	token, errCreate := auth.CreateToken(user.Username)

	if errCreate != nil {
		log.Println(errCreate.Error())
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	saveErr := auth.SaveAuthRedis(client, user.Username, token)
	if saveErr != nil {
		utils.JSON(w, http.StatusUnprocessableEntity, saveErr.Error())
		return
	}
	ID, AvatarURL, DOB, Role, err := database.GetUserWithName(db, user.Username)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, "Infomation is incorrect") // Username is not exist, but show incorrect
		return
	}
	utils.JSON(w, http.StatusOK, struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ID           int    `json:"id"`
		Username     string `json:"username"`
		AvatarURL    string `json:"avatar_url"`
		DOB          string `json:"dob"`
		Role         string `json:"role"`
	}{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ID:           ID,
		Username:     user.Username,
		AvatarURL:    AvatarURL,
		DOB:          DOB,
		Role:         Role,
	})
}
func SignUp(w http.ResponseWriter, r *http.Request, _ httprouter.Params, client *redis.Client) {
	// Convert json data from request to type User
	var user models.User
	err := utils.DecodeJSONBody(w, r, &user)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if govalidator.IsNull(user.Username) || govalidator.IsNull(user.Password) {
		utils.JSON(w, http.StatusBadRequest, "Input cannot be empty")
		return
	}

	// Escape string before query to sql server
	user.Username = models.Santize(user.Username)
	user.Password = models.Santize(user.Password)

	// Check with regex
	match, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]{4,255}", user.Username)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !match {
		utils.JSON(w, http.StatusBadRequest, "Username must longer than 3 characters can only include: a-z, A-Z, 0-9")
		return
	}

	match, err = regexp.MatchString(".{5,255}$", user.Password)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !match {
		utils.JSON(w, http.StatusBadRequest, "Password length must between 5 and 255 characters")
		return
	}

	isUsernameExist := database.IsUserExist(db, user.Username)

	if isUsernameExist != false {
		utils.JSON(w, http.StatusBadRequest, "Username or Password incorrect") // Username is exist, but show incorrect
		return
	}
	//New a User

	err = database.PostNewUser(db, user.Username, user.Password)
	if err != nil {
		utils.JSON(w, http.StatusUnauthorized, "Username or Password incorrect")
		return
	}

	hashedPassword := database.GetHashedPassword(db, user.Username)

	check := models.CheckPasswordHash(hashedPassword, user.Password)

	if check != true {
		utils.JSON(w, http.StatusUnauthorized, "Username or Password incorrect")
		return
	}

	token, errCreate := auth.CreateToken(user.Username)

	if errCreate != nil {
		log.Println(errCreate.Error())
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	saveErr := auth.SaveAuthRedis(client, user.Username, token)
	if saveErr != nil {
		utils.JSON(w, http.StatusUnprocessableEntity, saveErr.Error())
		return
	}
	ID, AvatarURL, DOB, Role, err := database.GetUserWithName(db, user.Username)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, "Infomation is incorrect") // Username is not exist, but show incorrect
		return
	}
	utils.JSON(w, http.StatusOK, struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ID           int    `json:"id"`
		Username     string `json:"username"`
		AvatarURL    string `json:"avatar_url"`
		DOB          string `json:"dob"`
		Role         string `json:"role"`
	}{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ID:           ID,
		Username:     user.Username,
		AvatarURL:    AvatarURL,
		DOB:          DOB,
		Role:         Role,
	})
}
func RefreshTokenAPI(w http.ResponseWriter, r *http.Request, _ httprouter.Params, client *redis.Client) {
	token, status, statusMsg := auth.RefreshToken(client, r)
	if token != nil {
		// Create new token successfully
		utils.JSON(w, status, struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		}{
			AccessToken:  token["access_token"],
			RefreshToken: token["refresh_token"],
		})
	} else {
		utils.JSON(w, status, struct {
			ErrorMsg string `json:"error"`
		}{
			ErrorMsg: statusMsg,
		})
	}
}

//For USER API
// type NUser struct {
// 	Username string `json:"username"`
// }

// func NewUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

// }
func GetUserByID(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	ID, Username, AvatarURL, DOB, err := database.GetUserByID(db, id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	if ID == -1 {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	utils.JSON(w, http.StatusOK, struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		AvatarURL string `json:"avatar_url"`
		DOB       string `json:"dob"`
	}{
		ID:        ID,
		Username:  Username,
		AvatarURL: AvatarURL,
		DOB:       DOB,
	})
}

//For BOOK API
func GetSearchBook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	q := r.URL.Query()
	key := q.Get("key") + "%"
	var bookSearchString = "SELECT b.id, b.title, b.cover FROM public.books b join public.authors a on a.id=b.author_id	join public.publishers p on p.id = b.publisher_id	where LOWER (b.title) like Lower($1) or LOWER(a.name) like Lower($2) or LOWER (p.name) like Lower($3);"
	rows, err := database.GetListBookHeaderWith3Param(db, bookSearchString, key)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	type Book struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Cover string `json:"cover"`
	}
	type ListBook []Book
	var listBook ListBook
	for _, boo := range rows {
		var bo Book
		bo.ID = boo.ID
		bo.Title = boo.Title
		bo.Cover = boo.Cover
		listBook = append(listBook, bo)
	}

	utils.JSON(w, http.StatusOK, struct {
		Books []Book `json:"books"`
	}{
		Books: listBook,
	})
	return
}

type BookJson struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PublisherID string `json:"publisherId"`
	Cover       string `json:"cover"`
	AuthorID    string `json:"authorId"`
	Category    []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}

func PostANewBook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var data BookJson
	err := utils.DecodeJSONBody(w, r, &data)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)

		} else {
			log.Println(err.Error())

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		}
		return
	}

	var addNewBookString = "INSERT INTO public.books(title, description, created_at, publisher_id, cover, author_id) VALUES ($1, $2, current_timestamp, $3, $4, $5) returning id"
	id, err := database.PostANewBook(db, addNewBookString, data.Title, data.Description, data.PublisherID, data.Cover, data.AuthorID)
	for _, cate := range data.Category {
		err = database.PostANewBookwithCategory(db, id, cate.ID)
		if err != nil {
			utils.JSON(w, http.StatusInternalServerError, "Can't add into list, please try again later")
			return
		}
	}
	if err != nil {
		if err.Error() == "Already exist!" {
			utils.JSON(w, http.StatusUnprocessableEntity, "Already exist!")
			return
		}
		utils.JSON(w, http.StatusInternalServerError, "Can't add into list, please try again later")
		fmt.Println("There")
		return
	}
	utils.JSON(w, http.StatusOK, "Add favour successfully")
}
func GetBookbyID(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		fmt.Println(err)
	}
	var getBookString string = "SELECT id, title, description, created_at, publisher_id, cover, author_id FROM public.books WHERE id=$1"
	ID, Title, Description, CreatedAt, PublisherID, PublisherName, Cover, AuthorID, AuthorName, Categories, err := database.GetBookbyID(db, getBookString, id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	if ID == -1 {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	fmt.Println("Yes")
	type Publisher struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type Author struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type Category struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var cates []Category
	for _, categoryy := range Categories {
		var cate Category
		cate.ID = categoryy.ID
		cate.Name = categoryy.Name
		cates = append(cates, cate)
	}
	utils.JSON(w, http.StatusOK, struct {
		ID          int        `json:"id"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		CreatedAt   string     `json:"created_at"`
		Publisher   Publisher  `json:"publisher"`
		Cover       string     `json:"cover"`
		Author      Author     `json:"author"`
		Category    []Category `json:"category"`
	}{
		ID:          ID,
		Title:       Title,
		Description: Description,
		CreatedAt:   CreatedAt,
		Publisher: Publisher{
			ID:   PublisherID,
			Name: PublisherName,
		},
		Cover: Cover,
		Author: Author{
			ID:   AuthorID,
			Name: AuthorName,
		},
		Category: cates,
	})
	return
}
func GetListFavourBookofUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user_id := p.ByName("id")
	var getFavourBookString string = "SELECT b.id, b.title, b.cover FROM public.books b join public.list_favourite_books fb on b.id = fb.book_id join public.users u on u.id = fb.user_id where u.id=$1;"
	rows, err := database.GetListBookHeaderWithParam(db, getFavourBookString, user_id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	fmt.Println("Yes")
	type Book struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Cover string `json:"cover"`
	}
	type ListBook []Book
	var listBook ListBook
	for _, boo := range rows {
		var bo Book
		bo.ID = boo.ID
		bo.Title = boo.Title
		bo.Cover = boo.Cover
		listBook = append(listBook, bo)
	}

	utils.JSON(w, http.StatusOK, struct {
		Books []Book `json:"books"`
	}{
		Books: listBook,
	})
	return
}

func GetListReviewofUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user_id := p.ByName("id")
	var getReviewBookString string = "SELECT rb.id, b.id, b.title, b.cover, rb.rating , rb.rate_title, rb.rate_review, rb.created_at FROM public.list_rating_books rb join public.books b on b.id = rb.book_id join public.users u on u.id = rb.user_id where u.id=$1 order by created_at DESC;"
	rows, err := database.GetListReviewofUser(db, getReviewBookString, user_id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database 1")
		return
	}
	fmt.Println("Yes")
	type Review struct {
		ID        int    `json:"id"`
		BookID    int    `json:"book_id"`
		BookTitle string `json:"book_title"`
		BookCover string `json:"book_cover"`
		Rating    int    `json:"rating"`
		Title     string `json:"title"`
		Review    string `json:"review"`
		CreatedAt string `json:"created_at"`
	}
	type ListReview []Review
	var listReview ListReview
	for _, boo := range rows {
		var bo Review
		bo.ID = boo.ID
		bo.BookID = boo.BookID
		bo.BookTitle = boo.BookTitle
		bo.BookCover = boo.BookCover
		bo.Rating = boo.Rating
		bo.Title = boo.Title
		bo.Review = boo.Review
		bo.CreatedAt = boo.CreatedAt

		listReview = append(listReview, bo)
	}

	utils.JSON(w, http.StatusOK, struct {
		Reviews []Review `json:"review"`
	}{
		Reviews: listReview,
	})
	return
}

func GetListNewestBookHeader(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var getNewestBookString string = "SELECT id, title, cover FROM public.books ORDER BY created_at DESC limit 10"
	rows, err := database.GetListBookHeader(db, getNewestBookString)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	fmt.Println("Yes")
	type Book struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Cover string `json:"cover"`
	}
	type ListBook []Book
	var listBook ListBook
	for _, boo := range rows {
		var bo Book
		bo.ID = boo.ID
		bo.Title = boo.Title
		bo.Cover = boo.Cover
		listBook = append(listBook, bo)
	}

	utils.JSON(w, http.StatusOK, struct {
		Books []Book `json:"books"`
	}{
		Books: listBook,
	})
	return
}

type NewAuthor struct {
	AuthorName string `json:"author-name"`
}

func PostANewAuthor(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var NewAuthor NewAuthor
	err := utils.DecodeJSONBody(w, r, &NewAuthor)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)

		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		}
		return
	}
	//Check same

	var addAuthorString = "INSERT INTO public.authors (name) VALUES ($1)"
	err = database.PostANewAuthor(db, addAuthorString, NewAuthor.AuthorName)
	if err != nil {
		if err.Error() == "Already exist!" {
			utils.JSON(w, http.StatusUnprocessableEntity, "Publisher already exist!")
		}
		utils.JSON(w, http.StatusInternalServerError, "Can't add into list, please try again later")
		return
	}
	utils.JSON(w, http.StatusOK, "Add favour successfully")
}

func GetListAuthor(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	getListAuthor := "SELECT id, name FROM public.authors order by name ASC"
	rows, err := database.GetListAuthor(db, getListAuthor)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, "Cannot get data")
		return
	}
	type Author struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type ListAu []Author
	var listAu ListAu
	for _, au := range rows {
		var a Author
		a.ID = au.ID
		a.Name = au.Name

		listAu = append(listAu, a)
	}

	utils.JSON(w, http.StatusOK, struct {
		ListAuthor []Author `json:"author"`
	}{
		ListAuthor: listAu,
	})
	return
}

func GetListAuthorBook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	author_id, _ := strconv.Atoi(p.ByName("author-id"))
	var getAuthorBookString string = "SELECT id, title, cover FROM public.books where author_id=$1 order by created_at DESC "
	rows, err := database.GetListAuthorBookHeader(db, getAuthorBookString, author_id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
	}
	fmt.Println("Yes")
	type Book struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Cover string `json:"cover"`
	}
	type ListBook []Book
	var listBook ListBook
	for _, boo := range rows {
		var bo Book
		bo.ID = boo.ID
		bo.Title = boo.Title
		bo.Cover = boo.Cover

		listBook = append(listBook, bo)
	}

	utils.JSON(w, http.StatusOK, listBook)
	return
}
func GetListCategoryName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var getAllCategory = "SELECT id, name from public.categories"
	rows, err := database.GetListCategoryName(db, getAllCategory)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, "Cannot get data")
		return
	}
	type Category struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type ListCate []Category
	var listCate ListCate
	for _, cate := range rows {
		var c Category
		c.ID = cate.ID
		c.Name = cate.Name

		listCate = append(listCate, c)
	}

	utils.JSON(w, http.StatusOK, struct {
		ListCategory []Category `json:"categories"`
	}{
		ListCategory: listCate,
	})
	return
}
func GetListCategoryBook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	category_id, _ := strconv.Atoi(p.ByName("category-id"))
	var getCategoryBookString string = "SELECT b.id, b.title, b.cover FROM public.books b join public.category_book cb on b.id = cb.book_id	join public.categories t on t.id = cb.category_id where t.id=$1 ORDER BY b.created_at DESC;"
	rows, err := database.GetListCategoryBookHeader(db, getCategoryBookString, category_id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
	}
	var getCategoryName string = "SELECT name from public.categories where id=$1"
	CategoryName, err := database.GetNameCategory(db, getCategoryName, category_id)
	fmt.Println("Yes")
	type Book struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Cover string `json:"cover"`
	}
	type ListBook []Book
	var listBook ListBook
	for _, boo := range rows {
		var bo Book
		bo.ID = boo.ID
		bo.Title = boo.Title
		bo.Cover = boo.Cover

		listBook = append(listBook, bo)
	}

	utils.JSON(w, http.StatusOK, struct {
		CategoryName string `json:"category_name"`
		Books        []Book `json:"books"`
	}{
		CategoryName: CategoryName,
		Books:        listBook,
	})
	return
}

func GetListPublisher(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var getAllPublisher = "SELECT id, name from public.publishers"
	rows, err := database.GetListPublisher(db, getAllPublisher)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, "Cannot get data")
		return
	}
	type Publisher struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type ListPub []Publisher
	var listPub ListPub
	for _, pub := range rows {
		var p Publisher
		p.ID = pub.ID
		p.Name = pub.Name

		listPub = append(listPub, p)
	}

	utils.JSON(w, http.StatusOK, struct {
		ListPublisher []Publisher `json:"publisher"`
	}{
		ListPublisher: listPub,
	})
	return
}

type NewPublisher struct {
	PublisherName string `json:"publisher-name"`
}

func PostANewPublisher(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var NewPublisher NewPublisher
	err := utils.DecodeJSONBody(w, r, &NewPublisher)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)

		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		}
		return
	}
	//Check same

	var addPublisherString = "INSERT INTO public.publishers (name) VALUES ($1)"
	err = database.PostANewPublisher(db, addPublisherString, NewPublisher.PublisherName)
	if err != nil {
		if err.Error() == "Already exist!" {
			utils.JSON(w, http.StatusUnprocessableEntity, "Publisher already exist!")
		}
		utils.JSON(w, http.StatusInternalServerError, "Can't add into list, please try again later")
		return
	}
	utils.JSON(w, http.StatusOK, "Add favour successfully")
}
func GetListPublisherBook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	publisher_id, _ := strconv.Atoi(p.ByName("publisher-id"))
	var getPublisherBookString string = "SELECT id, title, cover FROM public.books where publisher_id=$1 order by created_at DESC "
	rows, err := database.GetListPublisherBookHeader(db, getPublisherBookString, publisher_id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
	}
	fmt.Println("Yes")
	type Book struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Cover string `json:"cover"`
	}
	type ListBook []Book
	var listBook ListBook
	for _, boo := range rows {
		var bo Book
		bo.ID = boo.ID
		bo.Title = boo.Title
		bo.Cover = boo.Cover

		listBook = append(listBook, bo)
	}

	utils.JSON(w, http.StatusOK, listBook)
	return
}

//Post list favour book
type AddListFavourInfor struct {
	UserID string `json:"user_id"`
	BookID string `json:"book_id"`
}

func PostFavourABook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("here")
	var data AddListFavourInfor
	err := utils.DecodeJSONBody(w, r, &data)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)

		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		}
		return
	}
	var addFavourString = "INSERT INTO public.list_favourite_books(user_id, book_id) VALUES ($1, $2);"
	err = database.PostFavourABook(db, addFavourString, data.UserID, data.BookID)
	if err != nil {
		log.Println(err.Error())
		utils.JSON(w, http.StatusInternalServerError, "Can't add into list, please try again later")
		fmt.Println("There")
		return
	}
	utils.JSON(w, http.StatusOK, "Add favour successfully")

}

type AddListReviewInfor struct {
	UserID     string `json:"user_id"`
	BookID     string `json:"book_id"`
	Rating     string `json:"rating"`
	Title      string `json:"title"`
	RateReview string `json:"rate_review"`
	CreatedAt  string `json:"time"`
}

func PostReviewABook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("here review")
	var data AddListReviewInfor
	err := utils.DecodeJSONBody(w, r, &data)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)

		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		}
		return
	}
	var addFavourString = "insert into public.list_rating_books (user_id, book_id, rating, rate_title, rate_review, created_at) values ($1, $2, $3, $4, $5, $6)"
	err, check := database.PostReviewABook(db, addFavourString, data.UserID, data.BookID, data.Rating, data.Title, data.RateReview, data.CreatedAt)
	fmt.Println("here review2")
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Can't add into list, please try again later")
		fmt.Println("There")
		return
	}
	if check == false {
		utils.JSON(w, http.StatusBadRequest, "Can't add into list, please try again later")
		fmt.Println("There")
		return
	}

	utils.JSON(w, http.StatusOK, "Add review successfully")

}

func GetSumReviewofBook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	book_id := p.ByName("id")
	var getSumReviewofBook = "SELECT count (*) FROM public.list_rating_books where book_id=$1;"
	count, err := database.GetSumReviewofBook(db, getSumReviewofBook, book_id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database")
		return
	}
	utils.JSON(w, http.StatusOK, struct {
		Reviews int `json:"count_review"`
	}{
		Reviews: count,
	})
	return
}

func GetListReviewofBook(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	book_id := p.ByName("id")
	var getReviewofBOOK = "SELECT rb.id, u.id, u.username, u.avatar_url, rb.rating , rb.rate_title, rb.rate_review, rb.created_at FROM public.list_rating_books rb join public.users u on u.id = rb.user_id where rb.book_id=$1 order by created_at DESC;"
	rows, err := database.GetListReviewofBook(db, getReviewofBOOK, book_id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Cannot get database 1")
		return
	}
	fmt.Println("Yes")
	type Review struct {
		ID        int    `json:"id"`
		UserID    int    `json:"user_id"`
		Username  string `json:"username"`
		UserAva   string `json:"userava"`
		Rating    int    `json:"rating"`
		Title     string `json:"title"`
		Review    string `json:"review"`
		CreatedAt string `json:"created_at"`
	}
	type ListReview []Review
	var listReview ListReview
	for _, boo := range rows {
		var bo Review
		bo.ID = boo.ID
		bo.UserID = boo.UserID
		bo.Username = boo.Username
		bo.UserAva = boo.UserAva
		bo.Rating = boo.Rating
		bo.Title = boo.Title
		bo.Review = boo.Review
		bo.CreatedAt = boo.CreatedAt

		listReview = append(listReview, bo)
	}

	utils.JSON(w, http.StatusOK, struct {
		Reviews []Review `json:"review"`
	}{
		Reviews: listReview,
	})
	return
}
