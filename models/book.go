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
