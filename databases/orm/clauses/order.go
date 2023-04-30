// The original package is migrated from beego and modified, you can find orignal from following link:
//    "github.com/beego/beego/"
//
// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clauses

import (
	"strings"
)

type Sort int8

const (
	None       Sort = 0
	Ascending  Sort = 1
	Descending Sort = 2
)

type Option func(order *Order)

type Order struct {
	column string
	sort   Sort
	isRaw  bool
}

func Clause(options ...Option) *Order {
	o := &Order{}
	for _, option := range options {
		option(o)
	}

	return o
}

func (o *Order) GetColumn() string {
	return o.column
}

func (o *Order) GetSort() Sort {
	return o.sort
}

func (o *Order) SortString() string {
	switch o.GetSort() {
	case Ascending:
		return "ASC"
	case Descending:
		return "DESC"
	}

	return ``
}

func (o *Order) IsRaw() bool {
	return o.isRaw
}

func ParseOrder(expressions ...string) []*Order {
	var orders []*Order
	for _, expression := range expressions {
		sort := Ascending
		column := strings.ReplaceAll(expression, ExprSep, ExprDot)
		if column[0] == '-' {
			sort = Descending
			column = column[1:]
		}

		orders = append(orders, &Order{
			column: column,
			sort:   sort,
		})
	}

	return orders
}

func Column(column string) Option {
	return func(order *Order) {
		order.column = strings.ReplaceAll(column, ExprSep, ExprDot)
	}
}

func sort(sort Sort) Option {
	return func(order *Order) {
		order.sort = sort
	}
}

func SortAscending() Option {
	return sort(Ascending)
}

func SortDescending() Option {
	return sort(Descending)
}

func SortNone() Option {
	return sort(None)
}

func Raw() Option {
	return func(order *Order) {
		order.isRaw = true
	}
}
