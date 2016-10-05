package fourq

import (
	"fmt"
	"math/big"
)

// gfP2 implements a field of size p² as a quadratic extension of the base
// field where i²=-1.
type gfP2 struct {
	x, y *baseFieldElem // value is x+yi.
}

func newGFp2(pool *elemPool) *gfP2 {
	return &gfP2{pool.Get(), pool.Get()}
}

func (e *gfP2) String() string {
	return fmt.Sprintf("[%v, %v]", e.x, e.y)
}

func (e *gfP2) Bytes() []byte {
	ret := make([]byte, 32)
	copy(ret[:16], e.x.Bytes())
	copy(ret[16:], e.y.Bytes())
	return ret
}

func (e *gfP2) Put(pool *elemPool) {
	pool.Put(e.x)
	pool.Put(e.y)
}

func (e *gfP2) Set(a *gfP2) *gfP2 {
	e.x.Set(a.x)
	e.y.Set(a.y)
	return e
}

func (e *gfP2) SetBytes(in []byte) *gfP2 {
	if len(in) != 32 {
		return nil
	}

	e.x.SetBytes(in[:16])
	e.y.SetBytes(in[16:])
	return e
}

func (e *gfP2) SetZero() *gfP2 {
	e.x.SetZero()
	e.y.SetZero()
	return e
}

func (e *gfP2) SetOne() *gfP2 {
	e.x.SetOne()
	e.y.SetZero()
	return e
}

func (e *gfP2) IsZero() bool {
	return e.x.IsZero() && e.y.IsZero()
}

func (e *gfP2) Neg(a *gfP2) *gfP2 {
	e.x.Neg(a.x)
	e.y.Neg(a.y)
	return e
}

func (e *gfP2) Dbl(a *gfP2) *gfP2 {
	e.x.Dbl(a.x)
	e.y.Dbl(a.y)
	return e
}

func (e *gfP2) Add(a, b *gfP2) *gfP2 {
	e.x.Add(a.x, b.x)
	e.y.Add(a.y, b.y)
	return e
}

func (e *gfP2) Sub(a, b *gfP2) *gfP2 {
	e.x.Sub(a.x, b.x)
	e.y.Sub(a.y, b.y)
	return e
}

func (c *gfP2) Exp(a *gfP2, power *big.Int, pool *elemPool) *gfP2 {
	sum := newGFp2(pool)
	sum.SetOne()
	t := newGFp2(pool)

	for i := power.BitLen() - 1; i >= 0; i-- {
		t.Square(sum, pool)
		if power.Bit(i) != 0 {
			sum.Mul(t, a, pool)
		} else {
			sum.Set(t)
		}
	}

	c.Set(sum)

	sum.Put(pool)
	t.Put(pool)

	return c
}

// See "Multiplication and Squaring in Pairing-Friendly Fields",
// http://eprint.iacr.org/2006/471.pdf
func (e *gfP2) Mul(a, b *gfP2, pool *elemPool) *gfP2 {
	tx := pool.Get().Mul(a.x, b.x)
	t := pool.Get().Mul(a.y, b.y)
	tx.Sub(tx, t)

	ty := pool.Get().Mul(a.x, b.y)
	t.Mul(a.y, b.x)
	ty.Add(ty, t)

	e.x.Set(tx)
	e.y.Set(ty)

	pool.Put(tx)
	pool.Put(ty)
	pool.Put(t)

	return e
}

func (e *gfP2) Square(a *gfP2, pool *elemPool) *gfP2 {
	// Complex squaring algorithm:
	// (x+yi)² = (x+y)(x-y) + 2*x*y*i
	t1 := pool.Get().Sub(a.x, a.y)
	t2 := pool.Get().Add(a.x, a.y)
	tx := pool.Get().Mul(t1, t2)

	t1.Mul(a.x, a.y).Dbl(t1)

	e.x.Set(tx)
	e.y.Set(t1)

	pool.Put(t1)
	pool.Put(t2)
	pool.Put(tx)

	return e
}

func (e *gfP2) Invert(a *gfP2, pool *elemPool) *gfP2 {
	// See "Implementing cryptographic pairings", M. Scott, section 3.2.
	// ftp://136.206.11.249/pub/crypto/pairings.pdf
	t := pool.Get().Mul(a.x, a.x)
	t2 := pool.Get().Mul(a.y, a.y)
	t.Add(t, t2)

	inv := pool.Get()
	inv.Invert(t)

	e.y.Neg(a.y)
	e.y.Mul(e.y, inv)

	e.x.Mul(a.x, inv)

	pool.Put(t)
	pool.Put(t2)
	pool.Put(inv)

	return e
}
