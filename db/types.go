package db

import (
	"errors"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type BigInt struct {
	N *big.Int
}

func (i *BigInt) Equals(j *BigInt) bool {
	isNil := i == nil || i.N == nil
	isNjl := j == nil || j.N == nil
	if isNil && isNjl {
		return true
	}
	if isNil != isNjl {
		return false
	}
	return i.N.Cmp(j.N) == 0
}

func (i *BigInt) Equals2(n *big.Int) bool {
	isNil := i == nil || i.N == nil
	if isNil && n == nil {
		return true
	}
	if (isNil && n != nil) || (!isNil && n == nil) {
		return false
	}
	return i.N.Cmp(n) == 0
}

func (i *BigInt) HasValue() bool {
	return i != nil && i.N != nil
}

func (i *BigInt) Hex() string {
	if i == nil || i.N == nil {
		return ""
	}
	return i.N.Text(16)
}

func (i *BigInt) String() string {
	if i == nil || i.N == nil {
		return ""
	}
	return i.N.Text(10)
}

func (i *BigInt) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if i == nil || i.N == nil {
		return bson.TypeString, nil, nil
	}
	return bson.MarshalValue(i.String())
}

func (i *BigInt) UnmarshalBSONValue(btype bsontype.Type, data []byte) error {
	if btype != bson.TypeString {
		return errors.New("cannot unmarshal non-string bson value to BigInt")
	}
	var s *string
	err := bson.UnmarshalValue(btype, data, &s)
	if err != nil {
		return err
	}
	n, _ := new(big.Int).SetString(*s, 10)
	i.N = n
	return nil
}
