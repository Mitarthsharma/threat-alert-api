package main

import "sync"

type Store struct {
	mu        sync.Mutex
	Companies map[string]Company
	Users     map[string]User
	Tokens    map[string]string
	Apps      map[string]App
	AppKeys   map[string]string
	Alerts    map[string]Alert
}

func NewStore() *Store {
	return &Store{
		Companies: make(map[string]Company),
		Users:     make(map[string]User),
		Tokens:    make(map[string]string),
		Apps:      make(map[string]App),
		AppKeys:   make(map[string]string),
		Alerts:    make(map[string]Alert),
	}
}
