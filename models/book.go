package models

type Book struct {
	ID          int
	Title       string
	Description string
	CreatedAt   string
	DeletedAt   string
	Publisher   Publisher
	Cover       string
	Author      Author
}

type Publisher struct {
	ID   int
	Name string
}

type Author struct {
	ID   int
	Name string
}

type Category struct {
	ID   int
	Name string
}
type BookHeader struct {
	ID    int
	Title string
	Cover string
}
type ReviewOfUser struct {
	ID        int
	BookID    int
	BookTitle string
	BookCover string
	Rating    int
	Title     string
	Review    string
	CreatedAt string
}
type ReviewOfBook struct {
	ID        int
	UserID    int
	Username  string
	UserAva   string
	Rating    int
	Title     string
	Review    string
	CreatedAt string
}

//for json
