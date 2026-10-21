package picker

import (
	"cmp"
	"slices"
	"strings"

	"github.com/brianmargolis/shades/client"

	"github.com/pkg/errors"
)

// Order is how the picker sorts its list. It works as a command line flag.
type Order string

const (
	OrderName       Order = "name"
	OrderDarkFirst  Order = "dark-first"
	OrderLightFirst Order = "light-first"
)

var Orders = []Order{OrderName, OrderDarkFirst, OrderLightFirst}

func (o *Order) String() string {
	if *o == "" {
		return string(OrderName)
	}
	return string(*o)
}

func (o *Order) Set(value string) error {
	if !slices.Contains(Orders, Order(value)) {
		return errors.Errorf("must be one of %s", orderNames())
	}
	*o = Order(value)
	return nil
}

func (o *Order) Type() string {
	return "order"
}

func orderNames() string {
	names := []string{}
	for _, order := range Orders {
		names = append(names, string(order))
	}
	return strings.Join(names, ", ")
}

// next is the order alt-o cycles to. The zero value counts as OrderName.
func (o Order) next() Order {
	index := max(slices.Index(Orders, o), 0)
	return Orders[(index+1)%len(Orders)]
}

// Flags turns the order back into the command line flags that select it, for
// reloading the list.
func (o Order) Flags() []string {
	if o == "" || o == OrderName {
		return []string{}
	}
	return []string{"--order", string(o)}
}

// sortVariants orders variants by how light their background is. Variants come
// in sorted by name, and the sort is stable, so ties stay in name order.
func sortVariants(variants []client.ThemeVariant, order Order) {
	byLuminance := func(a, b client.ThemeVariant) int {
		return cmp.Compare(a.Luminance(), b.Luminance())
	}
	switch order {
	case OrderDarkFirst:
		slices.SortStableFunc(variants, byLuminance)
	case OrderLightFirst:
		slices.SortStableFunc(variants, func(a, b client.ThemeVariant) int {
			return byLuminance(b, a)
		})
	}
}
