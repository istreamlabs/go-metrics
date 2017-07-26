package metrics

// TagMap returns the internal tag map from a DataDog client instance.
func (c *DataDogClient) TagMap() map[string]string {
	return c.tagMap
}
