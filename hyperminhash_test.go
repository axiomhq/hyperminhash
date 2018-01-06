package hyperminhash

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func estimateError(got, exp uint64) float64 {
	var delta uint64
	if got > exp {
		delta = got - exp
	} else {
		delta = exp - got
	}
	return float64(delta) / float64(exp)
}

func TestZeros(t *testing.T) {
	registers := [m]uint16{}
	exp := 0.0
	for i := range registers {
		val := uint16(rand.Intn(32))
		if val == 0 {
			exp++
		}
		registers[i] = val << r
	}
	_, got := regSumAndZeros(registers[:])
	if got != exp {
		t.Errorf("expected %.2f, got %.2f", exp, got)
	}
}

func RandStringBytesMaskImprSrc(n uint32) string {
	b := make([]byte, n)
	for i := uint32(0); i < n; i++ {
		b[i] = letterBytes[rand.Int()%len(letterBytes)]
	}
	return string(b)
}

func TestCardinality(t *testing.T) {
	sk := New()
	step := 10000
	unique := map[string]bool{}

	for i := 1; len(unique) <= 1000000; i++ {
		str := RandStringBytesMaskImprSrc(rand.Uint32() % 32)
		sk.Add([]byte(str))
		unique[str] = true

		if len(unique)%step == 0 {
			exact := uint64(len(unique))
			res := uint64(sk.Cardinality())
			step *= 10

			ratio := 100 * estimateError(res, exact)
			if ratio > 2 {
				t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
			}

		}
	}
}

func TestMerge(t *testing.T) {
	sk1 := New()
	sk2 := New()

	unique := map[string]bool{}

	for i := 1; i <= 3500000; i++ {
		str := RandStringBytesMaskImprSrc(rand.Uint32() % 32)
		sk1.Add([]byte(str))
		unique[str] = true

		str = RandStringBytesMaskImprSrc(rand.Uint32() % 32)
		sk2.Add([]byte(str))
		unique[str] = true
	}

	sk1.Merge(sk2)
	exact := len(unique)
	res := int(sk1.Cardinality())

	ratio := 100 * math.Abs(float64(res-exact)) / float64(exact)
	expectedError := 1.04 / math.Sqrt(float64(m))

	if float64(res) < float64(exact)-(float64(exact)*expectedError) || float64(res) > float64(exact)+(float64(exact)*expectedError) {
		t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
	}

	sk1.Merge(sk2)
	exact = res
	res = int(sk1.Cardinality())

	if float64(res) < float64(exact)-(float64(exact)*expectedError) || float64(res) > float64(exact)+(float64(exact)*expectedError) {
		t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
	}
}

func TestCollision(t *testing.T) {
	sk1 := &Sketch{}
	sk2 := &Sketch{}
	unique := map[string]uint{}

	for i := 0; i < 10000000; i++ {
		str := strconv.Itoa(i)
		sk1.Add([]byte(str))
		unique[str]++
	}

	for i := 10000000; i < 20000000; i++ {
		str := strconv.Itoa(i)
		sk2.Add([]byte(str))
		unique[str]++
	}

	col := 0
	for _, count := range unique {
		if count > 1 {
			col++
		}
	}
	fmt.Println(sk1.Cardinality())
	fmt.Println(sk1.Similarity(sk2))
	fmt.Println(sk1.Intersection(sk2))
}
