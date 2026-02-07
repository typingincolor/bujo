package domain

import "time"

type SearchOptions struct {
	Query    string
	Type     *EntryType
	DateFrom *time.Time
	DateTo   *time.Time
	Tags     []string
	Limit    int
}

func NewSearchOptions(query string) SearchOptions {
	return SearchOptions{
		Query: query,
		Limit: 50,
	}
}

func (o SearchOptions) WithType(t EntryType) SearchOptions {
	o.Type = &t
	return o
}

func (o SearchOptions) WithDateRange(from, to time.Time) SearchOptions {
	o.DateFrom = &from
	o.DateTo = &to
	return o
}

func (o SearchOptions) WithTags(tags []string) SearchOptions {
	o.Tags = tags
	return o
}

func (o SearchOptions) WithLimit(limit int) SearchOptions {
	o.Limit = limit
	return o
}
