package testhelper

import (
	"crypto/rand"
	"encoding/binary"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
)

// a polynomial of degree t-1
// f(x) = a_0 + a_1*x + a_2*x^2 + ... + a_t*x^t
// we store the coefficients in the form of a slice
// each coefficient is generated randomly, this is very much like generating nonces
func (s *TestSuite) GeneratePolynomial(degree int64) []*btcec.ModNScalar {
	polynomial := make([]*btcec.ModNScalar, degree+1)
	// the value a_0 is the secret, others should be able to retrieve the secret
	for i := int64(0); i <= degree; i++ {
		var coeff btcec.ModNScalar
		int_secp256k1_rand, err := rand.Int(rand.Reader, btcec.S256().N)
		assert.Nil(s.T, err)
		coeff.SetByteSlice(int_secp256k1_rand.Bytes())
		polynomial[i] = &coeff
	}
	return polynomial
}

// VSS shares are generated by evaluating the polynomial f(i)
//
// 1. Each participant holds a share of the secret, and the secret
// can be reconstructed only if enough shares are combined.
//
// 2. By distributing shares instead of the actual secret, the scheme
// ensures that no single participant can reconstruct the secret on
// their own. Even if some participants collude, as long as their number
// is below the threshold, the secret remains protected.
// polynomial definition: https://byjus.com/maths/polynomial
//
// polynomial is evaluated using Horner's Method
func (s *TestSuite) EvaluatePolynomial(polynomial []*btcec.ModNScalar, x *btcec.ModNScalar) *btcec.ModNScalar {
	result := new(btcec.ModNScalar)
	result.Set(polynomial[len(polynomial)-1])
	for i := len(polynomial) - 2; i >= 0; i-- {
		// term a_n*x + a_n-1
		result.Mul(x)
		result.Add(polynomial[i])
	}
	return result
}

// calculate the Lagrange coefficient at i over a set
// requires exact position, all values start with 1
func (s *TestSuite) CalculateLagrangeCoeff(i int64, set []int64) *btcec.ModNScalar {
	mul_j := new(btcec.ModNScalar).SetInt(1)
	for _, j := range set {
		if j != i {
			x_j := new(btcec.ModNScalar).SetInt(uint32(j))
			x_i := new(btcec.ModNScalar).SetInt(uint32(i))
			numerator := new(btcec.ModNScalar).NegateVal(x_j)
			denominator := new(btcec.ModNScalar).NegateVal(x_j).Add(x_i)
			mul_j.Mul(numerator)
			mul_j.Mul(denominator.InverseNonConst())
		}
	}

	return mul_j
}

func Int64ToBytes(num int64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(num))
	return bytes
}

func BytesToInt64(bytes []byte) int64 {
	return int64(binary.BigEndian.Uint64(bytes))
}
