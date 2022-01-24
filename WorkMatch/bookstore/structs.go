package main

type Task struct {
	ID          string
	Language    string
	Description string
	Time        string
	Username    string
	Email       string
	Link        string
	IsLink      bool
}
type Match struct {
	ID1          string
	Language1    string
	Description1 string
	Time1        string
	Username1    string
	Email1       string
	Link1        string
	ID2          string
	Language2    string
	Description2 string
	Time2        string
	Username2    string
	Email2       string
	Link2        string
}

type Task2 struct {
	ID          string
	Language    string
	Description string
	Time        string
	Username    string
	IsLiked     bool
	Email       string
}
