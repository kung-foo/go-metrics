package metrics

import (
	"container/ring"
	"math"
	"testing"
	"time"
)

func TestWindowSink(t *testing.T) {
	key := []string{"foo"}
	ws := NewWindowSink(time.Millisecond*50, 4)

	ws.AddSample(key, 1.0)
	ws.AddSample(key, 2.0)
	ws.AddSample(key, 3.0)
	ws.AddSample(key, 4.0)

	c := ws.Sample(key).Count()
	if c != 4 {
		t.Fatalf("invalid count: %d", c)
	}

	mean := ws.Sample(key).Mean()
	if mean != 2.5 {
		t.Fatalf("unxpected mean: %f", mean)
	}

	min := ws.Sample(key).Min()
	if min != 1.0 {
		t.Fatalf("unxpected min: %f", min)
	}

	max := ws.Sample(key).Max()
	if max != 4.0 {
		t.Fatalf("unxpected max: %f", max)
	}

	stddev := ws.Sample(key).Stddev()
	if stddev != math.Sqrt(float64(5)/3) {
		t.Fatalf("unxpected stddev: %f", stddev)
	}

	ws.AddSample(key, 1.0)

	c = ws.Sample(key).Count()
	if c != 4 {
		t.Fatalf("unxpected new value count: %d", c)
	}

	slice := ws.Sample(key).ToSlice()
	if len(slice) != 4 {
		t.Fatalf("unxpected slice length: %d", len(slice))
	}

	if slice[0].Value != 1.0 {
		t.Fatalf("unxpected first element value: %f", slice[0].Value)
	}

	if slice[3].Value != 4.0 {
		t.Fatalf("unxpected last element value: %f", slice[3].Value)
	}

	time.Sleep(time.Millisecond * 100)

	c = ws.Sample(key).Count()

	if c != 0 {
		t.Fatalf("all elements should have expired: %d", c)
	}

	big := NewWindowSink(time.Millisecond*50, 1E6)
	big.AddSample(key, 1.0)

	c = big.Sample(key).Count()
	if c != 1 {
		t.Fatalf("unxpected new value count: %d", c)
	}

}

func TestWindowSinkFanout(t *testing.T) {
	ws1 := NewWindowSink(time.Millisecond*50, 10)
	ws2 := NewWindowSink(time.Millisecond*200, 40)

	key := []string{"foo"}
	fh := &FanoutSink{ws1, ws2}

	fh.AddSample(key, 1.0)
	fh.AddSample(key, 1.0)

	c1 := ws1.Sample(key).Count()
	c2 := ws2.Sample(key).Count()

	if c1 != 2 {
		t.Fatalf("c1 invalid count: %d", c1)
	}

	if c2 != 2 {
		t.Fatalf("c2 invalid count: %d", c2)
	}

	time.Sleep(time.Millisecond * 100)

	c1 = ws1.Sample(key).Count()
	c2 = ws2.Sample(key).Count()

	if c1 != 0 {
		t.Fatalf("c1 should be 0: %d", c1)
	}

	if c2 != 2 {
		t.Fatalf("c2 invalid count: %d", c2)
	}

	m1 := ws1.Sample(key).Mean()
	m2 := ws2.Sample(key).Mean()

	if m1 != 0.0 {
		t.Fatalf("m1 should be 0.0: %f", m1)
	}

	if m2 != 1.0 {
		t.Fatalf("m2 should be 1.0: %f", m2)
	}
}

func TestWindowSinkShortCircuit(t *testing.T) {
	key := []string{"foo"}
	ws := NewWindowSink(time.Millisecond*50, 1000)

	for i := 0; i < 10; i++ {
		ws.AddSample(key, float32(i))
	}

	var f ringFn

	c := 0
	f = func(r *ring.Ring) bool {
		if r.Value != nil {
			c++
			return true
		}
		return false
	}

	ws.Sample(key).rdo(f)

	if c != 10 {
		t.Fatalf("unexpected short circuit loop count value: %d", c)
	}

	c = 0
	f = func(r *ring.Ring) bool {
		c++
		return true
	}

	ws.Sample(key).rdo(f)

	if c != 1000 {
		t.Fatalf("unexpected short circuit loop count value: %d", c)
	}
}

func BenchmarkSample(b *testing.B) {
	key := []string{"foo"}
	ws := NewWindowSink(time.Minute*1, 1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ws.AddSample(key, 1.0)
	}
}
