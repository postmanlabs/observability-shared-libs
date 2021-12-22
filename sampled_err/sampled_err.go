package sampled_err

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

// Samples up to SampleCount errors.
type Errors struct {
	Samples     []error
	SampleCount int
	TotalCount  int
}

func (es Errors) Error() string {
	if es.TotalCount == 0 {
		return "no error"
	}

	strs := make([]string, 0, len(es.Samples))
	for _, e := range es.Samples {
		strs = append(strs, e.Error())
	}

	return fmt.Sprintf("encountered %d errors, showing %d samples: %s",
		es.TotalCount, es.SampleCount, strings.Join(strs, ","))
}

func (es *Errors) Add(e error) {
	es.TotalCount += 1
	if len(es.Samples) >= es.SampleCount {
		// Randomly replace one of the existing samples.
		if rand.Intn(2) == 1 {
			es.Samples[rand.Intn(len(es.Samples))] = e
		}
	} else {
		es.Samples = append(es.Samples, e)
	}
}

// A thread-safe version of Errors.
type ConcurrentErrors interface {
	Add(error)
	GetTotalCount() int
	Error() string
}

type concurrentErrors struct {
	err   Errors
	mutex sync.Mutex
}

var _ ConcurrentErrors = (*concurrentErrors)(nil)

func NewConcurrentErrors(sampleCount int) ConcurrentErrors {
	return &concurrentErrors{
		err: Errors{SampleCount: sampleCount},
	}
}

func (errs *concurrentErrors) Add(err error) {
	errs.mutex.Lock()
	defer errs.mutex.Unlock()
	errs.err.Add(err)
}

func (errs *concurrentErrors) GetTotalCount() int {
	errs.mutex.Lock()
	defer errs.mutex.Unlock()
	return errs.err.TotalCount
}

func (errs *concurrentErrors) Error() string {
	errs.mutex.Lock()
	defer errs.mutex.Unlock()
	return errs.err.Error()
}
