// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Every struct {
	Quantity uint
	Unit     EveryUnit
}

type EveryUnit string

const (
	EveryMinute EveryUnit = "minute"
	EveryHour             = "hour"
	EveryDay              = "day"
	EveryWeek             = "week"
	EveryMonth            = "month"
	EveryYear             = "year"
)

var EveryUnits = map[EveryUnit]EveryUnit{
	EveryMinute: EveryMinute,
	EveryHour:   EveryHour,
	EveryDay:    EveryDay,
	EveryWeek:   EveryWeek,
	EveryMonth:  EveryMonth,
	EveryYear:   EveryYear,
}

func (every *Every) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprintf("Invalid string value: %s", value))
	}

	result, err := ParseEvery(str)
	*every = result
	return err
}

func (every Every) Value() (driver.Value, error) {
	return every.String(), nil
}

func (every Every) String() string {
	return fmt.Sprintf("%d.%s", every.Quantity, every.Unit)
}

func ParseEvery(str string) (Every, error) {
	re := regexp.MustCompile("^([0-9]+)\\.(minute|hour|day|week|month|year)$")
	matches := re.FindStringSubmatch(str)

	if len(matches) != 3 {
		return Every{}, errors.New("Invalid every string format")
	}
	// it should not fail, regex should handle that already
	c, _ := strconv.Atoi(matches[1])

	if u, ok := EveryUnits[EveryUnit(matches[2])]; ok {
		return Every{uint(c), u}, nil
	}

	return Every{}, errors.New("Unsupported every unit")
}
