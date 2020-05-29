package metrics

import "sort"

// Tags returns the internal tag list from a DataDog client instance.
func (c *DataDogClient) Tags() []string {
	sort.Strings(c.tags)
	return c.tags
}
