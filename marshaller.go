package listen

import (
	"context"
	"io"

	"github.com/pkg/errors"
)

var (
	ErrUnImplement = errors.New("unimplement method")
)

type Marshaller interface {
	Marshal(ctx context.Context, i interface{}) ([]byte, error)
	MarshalTo(ctx context.Context, i interface{}, w io.Writer) error

	UnMarshal(ctx context.Context, b []byte) (interface{}, error)
	UnMarshalTo(ctx context.Context, b []byte, i interface{}) error
}

type UnImplementMarshaller struct{}

func (UnImplementMarshaller) Marshal(ctx context.Context, i interface{}) ([]byte, error) {
	return nil, errors.Wrap(ErrUnImplement, "Marshal")
}

func (UnImplementMarshaller) MarshalTo(ctx context.Context, i interface{}, w io.Writer) error {
	return errors.Wrap(ErrUnImplement, "MarshalTo")
}

func (UnImplementMarshaller) UnMarshal(ctx context.Context, b []byte) (interface{}, error) {
	return nil, errors.Wrap(ErrUnImplement, "UnMarshal")
}

func (UnImplementMarshaller) UnMarshalTo(ctx context.Context, b []byte, i interface{}) error {
	return errors.Wrap(ErrUnImplement, "UnMarshalTo")
}

type Encoder interface {
	Encode() ([]byte, error)
}
