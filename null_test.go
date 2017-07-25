package metrics_test

import "github.com/istreamlabs/go-metrics"

func ExampleNullClient() {
	client := metrics.NewNullClient()
	client.WithTags(map[string]string{
		"tag1": "value1",
		"tag2": "value2",
	}).Incr("requests.count")
	client.Incr("other")

	// Output:
}
