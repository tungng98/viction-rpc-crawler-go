package db

import (
	"errors"
	"math/big"
	"time"

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

type Timestamp struct {
	t time.Time
}

func (t *Timestamp) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if t == nil {
		return bson.TypeDateTime, nil, nil
	}
	return bson.MarshalValue(t.t.Unix())
}

func (t *Timestamp) UnmarshalBSONValue(btype bsontype.Type, data []byte) error {
	if btype != bson.TypeDateTime {
		return errors.New("cannot unmarshal non-numeric bson value to Timestamp")
	}
	var i *int64
	err := bson.UnmarshalValue(btype, data, &i)
	if err != nil {
		return err
	}
	t.t = time.Unix(*i, 0)
	return nil
}

type TimestampNano struct {
	t time.Time
}

func (t *TimestampNano) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if t == nil {
		return bson.TypeInt64, nil, nil
	}
	return bson.MarshalValue(t.t.UnixNano())
}

func (t *TimestampNano) UnmarshalBSONValue(btype bsontype.Type, data []byte) error {
	if btype != bson.TypeInt64 {
		return errors.New("cannot unmarshal non-numeric bson value to Timestamp")
	}
	var i *int64
	err := bson.UnmarshalValue(btype, data, &i)
	if err != nil {
		return err
	}
	t.t = time.Unix(*i/int64(1_000_000_000), *i%int64(1_000_000_000))
	return nil
}
