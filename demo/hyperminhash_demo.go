package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dgryski/go-pcgr"

	"github.com/seiflotfy/hyperminhash"
)

func estimateError(got, exp uint64) float64 {
	var delta uint64
	if got > exp {
		delta = got - exp
	} else {
		delta = exp - got
	}
	return float64(delta) / float64(exp)
}

func main() {
	var (
		k     int64 = 10000000
		iters       = 20
		rnd         = pcgr.New(time.Now().UnixNano(), 0)
	)

	fmt.Println("| Set1 | HLL1 | Set2 | HLL2 | S1 ∪ S2 | HLL1 ∪ HLL2 | S1 ∩ S2 | HLL1 ∩ HLL2 |")
	fmt.Println("|---|---|---|---|---|---|---|---|")

	for j := 1; j <= iters; j++ {

		size1 := rnd.Int63() % k
		size2 := rnd.Int63() % k
		sk1 := hyperminhash.New()
		sk2 := hyperminhash.New()

		maxCol := size1
		if maxCol > size2 {
			maxCol = size2
		}

		cols := rnd.Int63() % maxCol
		intersections := 0
		set := make(map[int]uint8)

		for i := 0; i < int(size1); i++ {
			set[i]++
			sk1.Add([]byte(strconv.Itoa(i)))
		}

		for i := int(size1 - cols); i < int(size1-cols+size2); i++ {
			set[i]++
			if set[i] > 1 {
				intersections++
			}
			sk2.Add([]byte(strconv.Itoa(i)))
		}

		card1 := sk1.Cardinality()
		card2 := sk2.Cardinality()
		ints1 := sk1.Intersection(sk2)
		sk1.Merge(sk2)
		mcard := sk1.Cardinality()
		row := fmt.Sprintf("| %d | %d | %d | %d | %d | %d | **%d** (%f%%) | **%d** (%f%%) |", size1, card1, size2, card2, len(set), mcard, cols, float64(int(100*cols)/len(set)), ints1, 100*float64(ints1)/float64(mcard))

		fmt.Println(row)
	}
}
