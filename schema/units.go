package schema

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Unit is a description of a single scale of measurement, such as a "second". If there are multiple scales, such as
// "minute", "second", etc. then multiple of these unit classes can be composed into units.
type Unit struct {
	NameShortSingular string `json:"name_short_singular" name:"Short name (singular)" description:"Short name that can be printed in a few characters, singular form." examples:"[\"B\", \"char\"]"`
	NameShortPlural   string `json:"name_short_plural" name:"Short name (plural)" description:"Short name that can be printed in a few characters, plural form." examples:"[\"B\", \"chars\"]"`
	NameLongSingular  string `json:"name_long_singular" name:"Long name (singular)" description:"Longer name for this unit in singular form." examples:"[\"byte\", \"character\"]"`
	NameLongPlural    string `json:"name_long_plural" name:"Long name (plural)" description:"Longer name for this unit in plural form." examples:"[\"bytes\", \"characters\"]"`
}

// FormatShortInt formats an amount according to this unit.
func (u Unit) FormatShortInt(amount int64, displayZero bool) string {
	return FormatNumberUnitShort(amount, u, displayZero)
}

// FormatShortFloat formats an amount according to this unit.
func (u Unit) FormatShortFloat(amount float64, displayZero bool) string {
	return FormatNumberUnitShort(amount, u, displayZero)
}

// FormatLongInt formats an amount according to this unit.
func (u Unit) FormatLongInt(amount int64, displayZero bool) string {
	return FormatNumberUnitLong(amount, u, displayZero)
}

// FormatLongFloat formats an amount according to this unit.
func (u Unit) FormatLongFloat(amount float64, displayZero bool) string {
	return FormatNumberUnitLong(amount, u, displayZero)
}

// FormatNumberUnitShort formats a number with a single unit.
func FormatNumberUnitShort[T NumberType](amount T, unit Unit, displayZero bool) string {
	var formatString string
	switch any(amount).(type) {
	case int64:
		formatString = "%d"
	case float64:
		formatString = "%f"
	}
	switch {
	case amount == 1 || amount == -1:
		return strings.TrimRight(fmt.Sprintf(formatString, amount), "0") + unit.NameShortSingular
	case amount != 0:
		return strings.TrimRight(fmt.Sprintf(formatString, amount), "0") + unit.NameShortPlural
	case displayZero:
		return strings.TrimRight(fmt.Sprintf(formatString, amount), "0") + unit.NameShortPlural
	default:
		return ""
	}
}

// FormatNumberUnitLong formats a number with a single unit.
func FormatNumberUnitLong[T NumberType](amount T, unit Unit, displayZero bool) string {
	var formatString string
	switch any(amount).(type) {
	case int64:
		formatString = "%d"
	case float64:
		formatString = "%f"
	}
	switch {
	case amount == 1 || amount == -1:
		return fmt.Sprintf(formatString, amount) + unit.NameLongSingular
	case amount != 0:
		return fmt.Sprintf(formatString, amount) + unit.NameLongPlural
	case displayZero:
		return fmt.Sprintf(formatString, amount) + unit.NameLongPlural
	default:
		return ""
	}
}

// Units holds several scales of magnitude of the same unit, for example 5m30s.
type Units struct {
	BaseUnit               Unit
	Multipliers            map[int64]Unit
	sortedMultipliersCache []int64
	reCache                *regexp.Regexp
	reSubExpNames          map[string]int
}

// FormatShortInt formats the passed int according to the unit multipliers.
func (u *Units) FormatShortInt(data int64) string {
	return FormatNumberUnitsShort(data, *u)
}

// FormatShortFloat formats the passed float according to the unit multipliers.
func (u *Units) FormatShortFloat(data float64) string {
	return FormatNumberUnitsShort(data, *u)
}

// FormatLongInt formats the passed int according to the unit multipliers.
func (u *Units) FormatLongInt(data int64) string {
	return FormatNumberUnitsLong(data, *u)
}

// FormatLongFloat formats the passed float according to the unit multipliers.
func (u *Units) FormatLongFloat(data float64) string {
	return FormatNumberUnitsLong(data, *u)
}

// FormatNumberUnitsShort is a generic way to format a number with a unit.
func FormatNumberUnitsShort[T NumberType](data T, units Units) string {
	if data == 0 {
		return FormatNumberUnitShort(data, units.BaseUnit, true)
	}
	remainder := data
	output := ""
	for _, multiplier := range units.getSortedMultipliersCache() {
		base := int64(math.Floor(float64(remainder) / float64(multiplier)))
		remainder -= T(base * multiplier)
		output += FormatNumberUnitShort(base, units.Multipliers[multiplier], false)
	}
	output += FormatNumberUnitShort(remainder, units.BaseUnit, false)
	return output
}

// FormatNumberUnitsLong is a generic way to format a number with a unit.
func FormatNumberUnitsLong[T NumberType](data T, units Units) string {
	if data == 0 {
		return FormatNumberUnitLong(data, units.BaseUnit, true)
	}
	remainder := data
	output := ""
	for _, multiplier := range units.getSortedMultipliersCache() {
		base := int64(math.Floor(float64(remainder) / float64(multiplier)))
		remainder -= T(base * multiplier)
		output += FormatNumberUnitLong(remainder, units.Multipliers[multiplier], false)
	}
	output += FormatNumberUnitLong(remainder, units.BaseUnit, false)
	return output
}

func (u *Units) getSortedMultipliersCache() []int64 {
	if u.sortedMultipliersCache == nil {
		var multipliers []int64
		for multiplier := range u.Multipliers {
			multipliers = append(multipliers, multiplier)
		}
		sort.SliceStable(multipliers, func(i, j int) bool {
			return multipliers[i] > multipliers[j]
		})
		u.sortedMultipliersCache = multipliers
	}
	return u.sortedMultipliersCache
}

func (u *Units) parse(data string) (any, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return 0, &UnitParseError{
			Message: "Empty string cannot be parsed as " + u.BaseUnit.NameLongPlural,
		}
	}
	if u.reCache == nil {
		u.updateReCache()
	}
	match := u.reCache.FindStringSubmatch(data)
	if match == nil {
		return u.buildUnitParseError(data)
	}

	var isFloat bool
	var floatNumber float64
	var intNumber int64
	var err error
	for _, multiplier := range u.getSortedMultipliersCache() {
		matchGroupID := u.reSubExpNames[fmt.Sprintf("g%d", multiplier)]
		result := match[matchGroupID]

		intNumber, floatNumber, isFloat, err = u.handleParseMultiplier(
			result,
			multiplier,
			intNumber,
			floatNumber,
			isFloat,
		)
		if err != nil {
			return 0, err
		}
	}
	baseMatchGroup := match[u.reSubExpNames["g1"]]
	intNumber, floatNumber, isFloat, err = u.handleParseMultiplier(
		baseMatchGroup,
		1,
		intNumber,
		floatNumber,
		isFloat,
	)
	if err != nil {
		return 0, err
	}
	if isFloat {
		return floatNumber, nil
	}
	return intNumber, nil
}

func (u *Units) handleParseMultiplier(
	result string,
	multiplier int64,
	intNumber int64,
	floatNumber float64,
	isFloat bool,
) (int64, float64, bool, error) {
	if result == "" {
		return intNumber, floatNumber, isFloat, nil
	}
	if strings.Contains(result, ".") {
		i, err := strconv.ParseFloat(result, 64)
		if err != nil {
			return intNumber, floatNumber, isFloat, BadArgumentError{
				Message: fmt.Sprintf("Failed to parse number as float: %s", result),
			}
		}
		floatNumber += i * float64(multiplier)
		isFloat = true
	} else {
		i, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			return intNumber, floatNumber, isFloat, BadArgumentError{
				Message: fmt.Sprintf("Failed to parse number as int: %s", result),
			}
		}
		floatNumber += float64(i * multiplier)
		if !isFloat {
			intNumber += i * multiplier
		}
	}
	return intNumber, floatNumber, isFloat, nil
}

func (u *Units) updateReCache() {
	var parts []string
	if u.Multipliers != nil {
		for _, multiplier := range u.getSortedMultipliersCache() {
			unit := u.Multipliers[multiplier]
			parts = append(parts, fmt.Sprintf(
				"(?:|(?P<g%s>[0-9]+)\\s*(%s|%s|%s|%s))",
				regexp.QuoteMeta(fmt.Sprintf("%d", multiplier)),
				regexp.QuoteMeta(unit.NameShortSingular),
				regexp.QuoteMeta(unit.NameShortPlural),
				regexp.QuoteMeta(unit.NameLongSingular),
				regexp.QuoteMeta(unit.NameLongPlural),
			))
		}
	}
	parts = append(parts, fmt.Sprintf(
		"(?:|(?P<g1>[0-9]+(|.[0-9]+))\\s*(|%s|%s|%s|%s))",
		regexp.QuoteMeta(u.BaseUnit.NameShortSingular),
		regexp.QuoteMeta(u.BaseUnit.NameShortPlural),
		regexp.QuoteMeta(u.BaseUnit.NameLongSingular),
		regexp.QuoteMeta(u.BaseUnit.NameLongPlural),
	))
	regex := "^\\s*" + strings.Join(parts, "\\s*") + "\\s*$"
	u.reCache = regexp.MustCompile(regex)
	u.reSubExpNames = map[string]int{}
	for i, subExpName := range u.reCache.SubexpNames() {
		u.reSubExpNames[subExpName] = i
	}
}

func (u *Units) buildUnitParseError(data string) (any, error) {
	validUnits := []string{
		u.BaseUnit.NameShortSingular,
		u.BaseUnit.NameShortPlural,
		u.BaseUnit.NameLongSingular,
		u.BaseUnit.NameLongPlural,
	}
	for _, multiplier := range u.getSortedMultipliersCache() {
		validUnits = append(
			validUnits,
			u.Multipliers[multiplier].NameShortSingular,
			u.Multipliers[multiplier].NameShortPlural,
			u.Multipliers[multiplier].NameLongSingular,
			u.Multipliers[multiplier].NameLongPlural,
		)
	}
	return 0, UnitParseError{
		Message: fmt.Sprintf(
			"Cannot parse '%s' as '%s': invalid format, valid unit types are: '%s",
			data,
			u.BaseUnit.NameLongPlural,
			strings.Join(validUnits, "', '"),
		),
	}
}

// ParseInt parses a string into an integer.
func (u *Units) ParseInt(data string) (int64, error) {
	result, err := u.parse(data)
	if err != nil {
		return 0, err
	}
	if i, ok := result.(int64); ok {
		return i, nil
	}
	return 0, BadArgumentError{
		Message: fmt.Sprintf("Failed to parse %s as an integer, float found.", data),
	}
}

// ParseFloat parses a string into a floating point number.
func (u *Units) ParseFloat(data string) (float64, error) {
	result, err := u.parse(data)
	if err != nil {
		return 0, err
	}
	if i, ok := result.(int64); ok {
		return float64(i), nil
	}
	return result.(float64), nil
}

// UnitBytes is scaling, byte-based unit.
var UnitBytes = Units{
	BaseUnit: Unit{
		"B",
		"B",
		"byte",
		"bytes",
	},
	Multipliers: map[int64]Unit{
		1024: {
			"kB",
			"kB",
			"kilobyte",
			"kilobytes",
		},
		1048576: {
			"MB",
			"MB",
			"megabyte",
			"megabytes",
		},
		1073741824: {
			"GB",
			"GB",
			"gigabyte",
			"gigabytes",
		},
		1099511627776: {
			"TB",
			"TB",
			"terabyte",
			"terabytes",
		},
		1125899906842624: {
			"PB",
			"PB",
			"petabyte",
			"petabytes",
		},
	},
}

// UnitDurationNanoseconds is a nanosecond-based unit for time durations.
var UnitDurationNanoseconds = Units{
	BaseUnit: Unit{
		"ns",
		"ns",
		"nanosecond",
		"nanoseconds",
	},
	Multipliers: map[int64]Unit{
		1000: {
			"ms",
			"ms",
			"microsecond",
			"microseconds",
		},
		1000000: {
			"s",
			"s",
			"second",
			"seconds",
		},
		60000000: {
			"m",
			"m",
			"minute",
			"minutes",
		},
		3600000000: {
			"H",
			"H",
			"hour",
			"hours",
		},
		86400000000: {
			"d",
			"d",
			"day",
			"days",
		},
	},
}

// UnitDurationSeconds is a second-based unit for time durations.
var UnitDurationSeconds = Units{
	BaseUnit: Unit{
		"s",
		"s",
		"second",
		"seconds",
	},
	Multipliers: map[int64]Unit{
		60: {
			"m",
			"m",
			"minute",
			"minutes",
		},
		3600: {
			"H",
			"H",
			"hour",
			"hours",
		},
		86400: {
			"d",
			"d",
			"day",
			"days",
		},
	},
}

// UnitCharacters is a single unit for characters.
var UnitCharacters = Units{
	BaseUnit: Unit{
		"char",
		"chars",
		"character",
		"characters",
	},
}

// UnitPercentage is a single unit for percentages.
var UnitPercentage = Units{
	BaseUnit: Unit{
		"%",
		"%",
		"percent",
		"percent",
	},
}
