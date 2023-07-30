package errors_test

import (
	"errors"
	"fmt"
	xerrors "miopkg/errors"
	"testing"
)

func Test_errors(t *testing.T) {
	base := errors.New("base")
	fmt.Println(base)
	xbase := xerrors.FromError(base)
	fmt.Println(xbase)
	wbase := xerrors.Wrapf(base, "wrap %d", 123)
	fmt.Println(wbase)
	wxbase := xerrors.FromError(wbase)
	fmt.Println(wxbase)
	wwxbase := xerrors.Wrapf(wxbase, "wrap %d", 456)
	fmt.Println(wwxbase)
	fmt.Println(xerrors.Is(wbase, base))
	fmt.Println(xerrors.Is(wwxbase, xbase))
}
