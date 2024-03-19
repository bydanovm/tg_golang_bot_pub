package models

type SearchResults struct {
	Ready   bool
	Query   string
	Results []Result
}

type Result struct {
	Name, Description, Url string
}
