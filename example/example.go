package main

import (
	"context"

	"github.com/hidori/go-gontext"
)

type contextKey string

const (
	key1 contextKey = "key1"
	key2 contextKey = "key2"
	key3 contextKey = "key3"
)

func main() {
	ctx := context.Background()

	ctx = context.WithValue(ctx, key1, "value")
	value, ok := gontext.Value[string](ctx, key1)
	if ok {
		println("Value:", value)
	} else {
		println("Key not found")
	}

	defaultValue := gontext.ValueOrDefault(ctx, key2, "default")
	println("Default Value:", defaultValue)

	zeroValue := gontext.ValueOrZero[int](ctx, key3)
	println("Zero Value:", zeroValue)
}
