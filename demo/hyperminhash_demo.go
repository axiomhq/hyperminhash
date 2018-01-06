package main

import (
	"fmt"
	"strconv"

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

	iters := 20
	k := 1000000

	for j := 1; j <= iters; j++ {

		sk1 := &hyperminhash.Sketch{}
		sk2 := &hyperminhash.Sketch{}
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

		fmt.Printf(
			"| [0 - %d]  | [%d - %d]| %d | %d |\n",
			k,
			int(float64(k)*frac),
			2*k-int(float64(k)*frac),
			exact,
			res,
		)
	}
}
