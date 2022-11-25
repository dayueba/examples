package main

type user struct {
	name string
}

func NewUser(name string) *user {
	return &user{name}
}