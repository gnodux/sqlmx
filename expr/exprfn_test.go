/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import (
	"fmt"
	"testing"
)

func TestFilters(t *testing.T) {
	var filters Filters
	filters.Select(UseLimits(10, 20))
	filters.Append(AllToOr)
	fmt.Println(filters)
}
