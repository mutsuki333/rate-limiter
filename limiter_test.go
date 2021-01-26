package limiter

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	ip = "192.168.0."
)

var testCount = 0
var l = Default()

func examine(limit int, testIP string) error {
	for i := 1; i < limit+20; i++ {
		rate, err := l.HitOrError(testIP)
		if i <= limit {
			if rate != i {
				return fmt.Errorf("rate should be %v not %v", i, rate)
			}
			if err != nil {
				return fmt.Errorf("hit %v should not be err, but get %s", i, err.Error())
			}
		} else {
			if err == nil {
				return errors.New("there should be err, but none get")
			}
		}
	}
	return nil
}

func test(limit int, interval time.Duration, testIP string) error {
	var err error
	err = examine(limit, testIP)
	if err != nil {
		return err
	}
	time.Sleep(interval + time.Second)
	err = examine(limit, testIP)
	if err != nil {
		return err
	}
	return err
}

func Test60in60sec(t *testing.T) {
	testCount++
	testIP := ip + strconv.Itoa(testCount)
	err := test(60, time.Minute, testIP)
	if err != nil {
		t.Fatal(err)
	}
}

func Test5in10sec(t *testing.T) {
	testCount++
	testIP := ip + strconv.Itoa(testCount)
	l.Limit = 5
	l.Interval = 10 * time.Second
	err := test(5, time.Second*10, testIP)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParallel(t *testing.T) {
	l.Limit = 5
	l.Interval = 10 * time.Second
	var wg sync.WaitGroup
	testP := func(testIP string, wg *sync.WaitGroup) {
		defer wg.Done()
		err := test(5, time.Second*10, testIP)
		if err != nil {
			t.Fatal(err)
		}
	}
	wg.Add(3)
	testCount++
	testIP := ip + strconv.Itoa(testCount)
	go testP(testIP, &wg)
	testCount++
	testIP = ip + strconv.Itoa(testCount)
	go testP(testIP, &wg)
	testCount++
	testIP = ip + strconv.Itoa(testCount)
	go testP(testIP, &wg)
	wg.Wait()
}
