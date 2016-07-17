package main

// Context Context will store information about the connection
type Context struct {
	Bind string `json:"bind"`
	File string `json:"keys>file"`
}
