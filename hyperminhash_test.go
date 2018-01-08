package hyperminhash

import (
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
	registers := [m]register{}
	exp := 0.0
	for i := range registers {
		val := register(rand.Intn(math.MaxUint16))
		if val.lz() == 0 {
			exp++
		}
		registers[i] = val
	}
	_, got := regSumAndZeros(registers[:])
	if got != exp {
		t.Errorf("expected %.2f, got %.2f", exp, got)
	}
}

func TestAllZeros(t *testing.T) {
	registers := [m]register{}
	exp := 16384.00
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

	msk := sk1.Merge(sk2)
	exact := uint64(len(unique))
	res := msk.Cardinality()

	ratio := 100 * estimateError(res, exact)

	if ratio > 2 {
		t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
	}

	sk1.Merge(sk2)
	exact = res
	res = sk1.Cardinality()

	if ratio > 2 {
		t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
	}
}

func TestIntersection(t *testing.T) {

	iters := 20
	k := 1000000

	for j := 1; j <= iters; j++ {

		sk1 := &Sketch{}
		sk2 := &Sketch{}
		unique := map[string]uint{}

		frac := float64(j) / float64(iters)

		for i := 0; i < k; i++ {
			str := strconv.Itoa(i)
			sk1.Add([]byte(str))
			unique[str]++
		}

		for i := int(float64(k) * frac); i < 2*k; i++ {
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

		exact := uint64(k - int(float64(k)*frac))
		res := sk1.Intersection(sk2)

		ratio := 100 * estimateError(res, exact)
		if ratio > 10 {
			t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
		}
	}
}

func TestNoIntersection(t *testing.T) {

	sk1 := New()
	sk2 := New()

	for i := 0; i < 1000000; i++ {
		sk1.Add([]byte(strconv.Itoa(i)))
	}

	for i := 1000000; i < 2000000; i++ {
		sk2.Add([]byte(strconv.Itoa(i)))
	}

	if got := sk1.Intersection(sk2); got != 0 {
		t.Errorf("Expected no intersection, got %v", got)
	}
}
