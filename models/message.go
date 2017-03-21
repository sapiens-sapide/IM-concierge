package models

import (
	"github.com/oklog/ulid"
	"sort"
	"time"
)

type Message struct {
	Body    string    `json:"body"`
	Date    time.Time `json:"date"`
	From    string    `json:"from"`
	Id      ulid.ULID `json:"id"`
	PrevMsg ulid.ULID `json:"prevMsg"`
	Room    string    `json:"room"`
}

/*** sort functions for Messages ***/
// By is the type of a "less" function that defines the ordering of its Message arguments.
type By func(m1, m2 *Message) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(messages []Message) {
	ps := &messageSorter{
		messages: messages,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type messageSorter struct {
	messages []Message
	by       func(m1, m2 *Message) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *messageSorter) Len() int {
	return len(s.messages)
}

// Swap is part of sort.Interface.
func (s *messageSorter) Swap(i, j int) {
	s.messages[i], s.messages[j] = s.messages[j], s.messages[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *messageSorter) Less(i, j int) bool {
	return s.by(&s.messages[i], &s.messages[j])
}

func SortByDate(messages []Message) {
	// Closures that order the Message structure in descending order
	date := func(m1, m2 *Message) bool {
		return m2.Date.Before(m1.Date)
	}
	By(date).Sort(messages) //type conversion
}
