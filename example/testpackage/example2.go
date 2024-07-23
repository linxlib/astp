package testpackage

import "time"

// Fuck fucks
type Fuck struct {
	U string `json:"u"` // u
	I int    //i
}

// Model Hah
type Model struct {
	ID        int64     `json:"id"`         //id
	InDate    string    `json:"in_date"`    //indate
	EditDate  string    `json:"edit_date"`  //editdate
	DeletedAt time.Time `json:"deleted_at"` //deleted_at
}

func TestMethod2() {

}
