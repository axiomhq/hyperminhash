package hyperminhash

import (
	"math"
	bits "math/bits"

	metro "github.com/dgryski/go-metro"
)

const (
	p     = 14
	m     = uint32(1 << p) // 16384
	max   = 64 - p
	maxX  = math.MaxUint64 >> max
	alpha = 0.7213 / (1 + 1.079/float64(m))
	q     = 6  // the number of bits for the LogLog hash
	r     = 10 // number of bits for the bbit hash
	_2q   = 1 << q
	_2r   = 1 << r
)

func beta(ez float64) float64 {
	zl := math.Log(ez + 1)
	return -0.370393911*ez +
		0.070471823*zl +
		0.17393686*math.Pow(zl, 2) +
		0.16339839*math.Pow(zl, 3) +
		-0.09237745*math.Pow(zl, 4) +
		0.03738027*math.Pow(zl, 5) +
		-0.005384159*math.Pow(zl, 6) +
		0.00042419*math.Pow(zl, 7)
}

func regSumAndZeros(registers []uint16) (float64, float64) {
	var sum, ez float64
	for _, val := range registers {
		if val == 0 {
			ez++
		}
		sum += 1 / math.Pow(2, float64(val>>r))
	}
	return sum, ez
}

// Sketch is a sketch for cardinality estimation based on LogLog counting
type Sketch struct {
	reg [m]uint16
}

// New returns a Sketch
func New() *Sketch {
	return new(Sketch)
}

// AddHash takes in a "hashed" value (bring your own hashing)
func (sk *Sketch) AddHash(x, y uint64) {
	k := x >> uint(max)
	lz := uint16(bits.LeadingZeros64((x<<p)^maxX)) + 1
	sig := uint16(y) >> p

	val := (lz << r) | sig

	if (sk.reg[k] >> r) < lz {
		sk.reg[k] = val
	} else if (sk.reg[k]>>r) == lz && (sk.reg[k]<<p) > sig {
		sk.reg[k] = val
	}
}

// Add inserts a value into the sketch
func (sk *Sketch) Add(value []byte) {
	h1, h2 := metro.Hash128(value, 1337)
	sk.AddHash(h1, h2)
}

// Cardinality returns the number of unique elements added to the sketch
func (sk *Sketch) Cardinality() uint64 {
	sum, ez := regSumAndZeros(sk.reg[:])
	m := float64(m)
	return uint64(alpha * m * (m - ez) / (beta(ez) + sum))
}

func merge(sk1, sk2 *Sketch) *Sketch {
	m := *sk1
	for i := range m.reg {
		mlz, sklz := m.reg[i]>>r, sk2.reg[i]>>r
		if mlz < sklz {
			m.reg[i] = sk2.reg[i]
		} else if mlz == sklz && m.reg[i]>>p < sk2.reg[i]<<r {
			m.reg[i] = sk2.reg[i]
		}
	}
	return &m
}

// Merge other into sk
func (sk *Sketch) Merge(other *Sketch) {
	*sk = *(merge(sk, other))
}

// Similarity return a Jaccard Index similarity estimation
func (sk *Sketch) Similarity(other *Sketch) float64 {
	var C, N uint64
	for i := range sk.reg {
		if sk.reg[i] == other.reg[i] {
			C++
		}
		if sk.reg[i] != 0 && other.reg[i] != 0 {
			N++
		}
	}
	n := sk.Cardinality()
	m := other.Cardinality()
	ec := sk.expectedCollision(n, m)
	return (float64(C-ec) / float64(N))
}

func (sk *Sketch) expectedCollision(n, m uint64) uint64 {
	var x, b1, b2 float64
	for i := 1.0; i <= _2q; i++ {
		for j := 1.0; j <= _2r; j++ {
			if i != _2q {
				den := math.Pow(2, float64(p+r+i))
				b1 = (_2r + j) / den
				b2 = (_2r + j + 1) / den
			} else {
				den := math.Pow(2, float64(p+r+i-1))
				b1 = j / den
				b2 = (j + 1) / den
			}
			prx := math.Pow(1-b2, float64(n)) - math.Pow(1-b1, float64(n))
			pry := math.Pow(1-b2, float64(m)) - math.Pow(1-b1, float64(m))
			x += (prx * pry)
		}
	}
	return uint64((x * float64(p)) + 0.5)
}

// Intersection returns number of intersections between sk and other
func (sk *Sketch) Intersection(other *Sketch) uint64 {
	sim := sk.Similarity(other)
	m := merge(sk, other)
	return uint64((sim*float64(m.Cardinality()) + 0.5))
}
