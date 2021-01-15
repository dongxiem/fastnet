package atomic

import (
	"math/rand"
	"sync"
	"testing"
)

// TestBool：测试 Bool
func TestBool(t *testing.T) {
	var isOk Bool
	if isOk.Get() != false {
		t.Fatal("expect false")
	}

	isOk.Set(true)
	if isOk.Get() != true {
		t.Fatal("expect true")
	}

	isOk.Set(false)
	if isOk.Get() != false {
		t.Fatal("expect false")
	}
}

// TestInt32：测试 Int32
func TestInt32(t *testing.T) {
	var count Int32
	if count.Get() != 0 {
		t.Fatal("expect 0 , but get ", count.Get())
	}

	sw := sync.WaitGroup{}

	len := rand.Intn(10000) + 10000
	for i := 0; i < len; i++ {
		sw.Add(1)
		go func() {
			count.Add(1)
			sw.Done()
		}()

		sw.Add(1)
		go func() {
			count.Add(-1)
			sw.Done()
		}()
	}
	sw.Wait()

	if 0 != count.Get() {
		t.Fatal("expect 0 but get ", count.Get())
	}
}

// TestInt64：测试 Int64
func TestInt64(t *testing.T) {
	var count Int64
	if count.Get() != 0 {
		t.Fatal("expect 0 , but get ", count.Get())
	}

	sw := sync.WaitGroup{}

	len := rand.Intn(10000) + 10000
	for i := 0; i < len; i++ {
		sw.Add(1)
		go func() {
			count.Add(1)
			sw.Done()
		}()

		sw.Add(1)
		go func() {
			count.Add(-1)
			sw.Done()
		}()
	}
	sw.Wait()

	if 0 != count.Get() {
		t.Fatal("expect 0 but get ", count.Get())
	}
}
