package metrics

import (
	"fmt"
	"reflect"
	"strings"
)

// Query provides a mechanism to filter and test metrics for given chainable
// constraints. It allows you to write tests to ensure behavior that is
// independent of metrics fired in upstream or downstream code.
//
//   // Assert that a metric with the given numerical value and tag has
//   // been fired during a test.
//   recorder.Expect("my.metric").Value(1).Tag("foo", "bar")
//
//   // Fail if a metric with the given name and numerical value is found.
//   recorder.If("my.metric").Value(100).Reject()
//
// Custom checks are also possible via `GetCalls`, which returns a slice
// of calls that can be processed however you like. The recorder provides
// a `Fatalf` function you can use that will dump out the current metrics
// stack if your custom check fails.
type Query interface {
	// Reject fails the test at least the minimum number of items are left.
	// Use this with:
	// `recorder.If("...")...Reject()`
	Reject()

	// Accept fails the test if fewer than the minimum number of items are left.
	// It is the opposite of `Reject` and usually you would use the
	// `recorder.Expect()...` shorthand instead. Use it like so:
	// `recorder.If("...")...Accept()`
	Accept()

	// GetCalls returns the currently matching list of calls. The calls may be
	// cast to `MetricCall` or `EventCall` for further processing.
	GetCalls() []Call

	// MinTimes sets the minimum number of calls that should be left before
	// the query is considered a failure. The default value is `1`, so if you
	// expect five calls you can use `recorder.Expect('my.metric').MinTimes(5)`.
	MinTimes(num int) Query

	// Contains filters out any metric or event that does not contain `component`
	// in the serialized representation of the call.
	Contains(component string) Query

	// ID filters out any metric whose name does not match `id` or event whose
	// title does not match `id`. Using `*` will match any ID.
	ID(name string) Query

	// Value filters out any metric whose numeric value does not match `value`.
	// All events are filtered out.
	Value(value interface{}) Query

	// Text filters out any event whose content text does not match `text`. All
	// metrics are filtered out.
	Text(text string) Query

	// Tag filters out any metric or event that does not contain the given tag
	// `name` and `value`.
	Tag(name, value string) Query

	// TagName filters out any metric or event that does not contain a given
	// tag with name `name`. The value does not matter.
	TagName(name string) Query
}

// query is an implementation of the `Query` interface.
type query struct {
	calls []Call
	test  TestFailer

	// The minimum number of calls that should exist after filter operations.
	minCalls int

	// Whether to check the minimum after each filter operation.
	checkMin bool

	// history stores a user-friendly representation of the built query
	history string
}

// Reject fails the test if at least the minimum number of items are left.
func (q *query) Reject() {
	if len(q.calls) >= q.minCalls {
		stack := make([]string, 0, len(q.calls))
		for _, call := range q.calls {
			stack = append(stack, call.String())
		}
		q.fatalf("Expected fewer than %d matching metrics but have '%s'", q.minCalls, strings.Join(stack, "', '"))
	}
}

// Accept fails the test if fewer than the minimum number of items are left.
func (q *query) Accept() {
	if len(q.calls) < q.minCalls {
		if len(q.calls) == 0 {
			q.test.Fatalf("Expected at least %d calls but have none", q.minCalls)
		} else {
			stack := make([]string, 0, len(q.calls))
			for _, call := range q.calls {
				stack = append(stack, call.String())
			}
			q.fatalf("Expected at least %d calls but only have '%s'", q.minCalls, strings.Join(stack, "', '"))
		}
	}
}

// MinTimes sets the minimum required number of calls.
func (q *query) MinTimes(num int) Query {
	q.history = fmt.Sprintf("%s minTimes(%d)", q.history, num)
	q.minCalls = num

	if q.checkMin {
		q.Accept()
	}

	return q
}

// GetCalls returns the currently matching calls.
func (q *query) GetCalls() []Call {
	return q.calls
}

// fatalf passes along a failure message to the test failer with additional
// information about the state of the metrics query.
func (q *query) fatalf(format string, args ...interface{}) {
	q.test.Fatalf(format+". Query was '%s'.", append(args, strings.Trim(q.history, " "))...)
}

// filter will remove calls from the call list given an expected value,
// comparison function, and getter function to get a value given a call
// instance.
func (q *query) filter(pred func(Call) bool) {
	var filtered []Call

	for _, call := range q.calls {
		if pred(call) {
			filtered = append(filtered, call)
		}
	}

	q.calls = filtered
}

// Contains checks whether the serialized metric contains the given
// string value. See `Call.String()` for the serialization format.
func (q *query) Contains(component string) Query {
	q.history = fmt.Sprintf("%s contains(%s)", q.history, component)
	q.filter(func(call Call) bool {
		return strings.Contains(call.String(), component)
	})

	if q.checkMin && len(q.calls) < q.minCalls {
		q.test.Fatalf("Expected metric or event to contain '%s'", component)
	}

	return q
}

// ID expects a metric name or event title.
func (q *query) ID(id string) Query {
	q.history = fmt.Sprintf("%s id(%s)", q.history, id)
	q.filter(func(call Call) bool {
		switch t := call.(type) {
		case *MetricCall:
			if t.Name == id {
				return true
			}
		case *EventCall:
			if t.Event.Title == id {
				return true
			}
		}
		return false
	})

	if q.checkMin && len(q.calls) < q.minCalls {
		q.fatalf("Expected metric or event with ID '%s'", id)
	}

	return q
}

// Value expects a metric value.
func (q *query) Value(value interface{}) Query {
	q.history = fmt.Sprintf("%s value(%v)", q.history, value)
	q.filter(func(call Call) bool {
		if m, ok := call.(*MetricCall); ok {
			return reflect.DeepEqual(m.Value, toFloat64(value))
		}
		return false
	})

	if q.checkMin && len(q.calls) < q.minCalls {
		q.fatalf("Expected metric value '%v'", value)
	}

	return q
}

// Text expects an event with the given text content value.
func (q *query) Text(text string) Query {
	q.history = fmt.Sprintf("%s text(%10s)", q.history, text)
	q.filter(func(call Call) bool {
		if e, ok := call.(*EventCall); ok {
			return e.Event.Text == text
		}
		return false
	})

	if q.checkMin && len(q.calls) < q.minCalls {
		q.fatalf("Expected event text '%v'", text)
	}

	return q
}

// Tag expects a tag name and value to be set with the emitted metric.
func (q *query) Tag(name, value string) Query {
	q.history = fmt.Sprintf("%s tag(%s, %s)", q.history, name, value)
	q.filter(func(call Call) bool {
		switch t := call.(type) {
		case *MetricCall:
			if v, ok := t.TagMap[name]; ok {
				return v == value
			}
		case *EventCall:
			if v, ok := t.TagMap[name]; ok {
				return v == value
			}
		}
		return false
	})

	if q.checkMin && len(q.calls) < q.minCalls {
		q.fatalf("Expected tag '%s' with value '%s'", name, value)
	}

	return q
}

// TagName expects a tag name to exist.
func (q *query) TagName(name string) Query {
	q.history = fmt.Sprintf("%s tag(%s)", q.history, name)
	q.filter(func(call Call) bool {
		switch t := call.(type) {
		case *MetricCall:
			if _, ok := t.TagMap[name]; ok {
				return true
			}
		case *EventCall:
			if _, ok := t.TagMap[name]; ok {
				return true
			}
		}
		return false
	})

	if q.checkMin && len(q.calls) < q.minCalls {
		q.fatalf("Expected tag '%s'", name)
	}

	return q
}
