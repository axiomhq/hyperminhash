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
	_2p   = 1 << p
	_2q   = 1 << q
	_2r   = 1 << r
	c     = 0.169919487159739093975315012348
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

func regSumAndZeros(registers []register) (float64, float64) {
	var sum, ez float64
	for _, val := range registers {
		lz := val.lz()
		if lz == 0 {
			ez++
		}
		sum += 1 / math.Pow(2, float64(lz))
	}
	return sum, ez
}

type register uint16

func (reg register) lz() uint8 {
	return uint8(uint16(reg) >> (16 - q))
}

func (reg register) sig() uint16 {
	return uint16(reg) << (16 - r) >> (16 - r)
}

func newReg(lz uint8, sig uint16) register {
	return register((uint16(lz) << r) | sig)
}

// Sketch is a sketch for cardinality estimation based on LogLog counting
type Sketch struct {
	reg [m]register
	raw map[uint64]uint16
}

// New returns a Sketch
func New() *Sketch {
	return new(Sketch)
}

// AddHash takes in a "hashed" value (bring your own hashing)
func (sk *Sketch) AddHash(x, y uint64) {
	k := x >> uint(max)
	lz := uint8(bits.LeadingZeros64((x<<p)^maxX)) + 1
	sig := uint16(y << (64 - r) >> (64 - r))
	if sk.reg[k].lz() < lz {
		sk.reg[k] = newReg(lz, sig)
	} else if sk.reg[k].lz() == lz && sk.reg[k].sig() > sig {
		sk.reg[k] = newReg(lz, sig)
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

// Merge returns a new union sketch of both sk and other
func (sk *Sketch) Merge(other *Sketch) *Sketch {
	m := *sk
	for i := range m.reg {
		if m.reg[i] < other.reg[i] || (m.reg[i] == other.reg[i] && m.reg[i] < other.reg[i]) {
			m.reg[i] = other.reg[i]
		}
	}
	return &m
}

// Similarity return a Jaccard Index similarity estimation
func (sk *Sketch) Similarity(other *Sketch) float64 {
	var C, N float64
	for i := range sk.reg {
		if sk.reg[i] != 0 && sk.reg[i] == other.reg[i] {
			C++
		}
		if sk.reg[i] != 0 || other.reg[i] != 0 {
			N++
		}
	}
	if C == 0 {
		return 0
	}

	n := float64(sk.Cardinality())
	m := float64(other.Cardinality())
	ec := sk.approximateExpectedCollisions(n, m)
	return float64(C-ec) / float64(N)
}

func (sk *Sketch) approximateExpectedCollisions(n, m float64) float64 {
	if n < m {
		n, m = m, n
	}
	if float64(n) > math.Pow(2, math.Pow(2, q)+r) {
		return math.MaxUint64
	} else if float64(n) > math.Pow(2, p+5) {
		d := float64(4*n/m) / math.Pow(float64((1+n)/m), 2)
		return float64((c * math.Pow(2, p-r) * d) + 0.5)
	} else {
		return sk.expectedCollision(n, m) / float64(p)
	}
}

func (sk *Sketch) expectedCollision(n, m float64) float64 {
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
	return (x * float64(p)) + 0.5
}

// Intersection returns number of intersections between sk and other
func (sk *Sketch) Intersection(other *Sketch) uint64 {
	sim := sk.Similarity(other)
	return uint64((sim*float64(sk.Merge(other).Cardinality()) + 0.5))
}
