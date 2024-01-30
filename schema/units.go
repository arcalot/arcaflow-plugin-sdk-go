package schema

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Unit is a description of a single scale of measurement, such as a "second". If there are multiple scales, such as
// "minute", "second", etc. then multiple of these UnitDefinition classes can be composed into UnitsDefinition.
type Unit interface {
	NameShortSingular() string
	NameShortPlural() string
	NameLongSingular() string
	NameLongPlural() string

	FormatShortInt(amount int64, displayZero bool) string
	FormatShortFloat(amount float64, displayZero bool) string
	FormatLongInt(amount int64, displayZero bool) string
	FormatLongFloat(amount float64, displayZero bool) string
}

// NewUnit defines a new UnitDefinition with the given parameters.
func NewUnit(nameSortSingular string, nameShortPlural string, nameLongSingular string, nameLongPlural string) *UnitDefinition {
	return &UnitDefinition{
		nameSortSingular,
		nameShortPlural,
		nameLongSingular,
		nameLongPlural,
	}
}

type UnitDefinition struct {
	NameShortSingularValue string `json:"name_short_singular"`
	NameShortPluralValue   string `json:"name_short_plural"`
	NameLongSingularValue  string `json:"name_long_singular"`
	NameLongPluralValue    string `json:"name_long_plural"`
}

func (u *UnitDefinition) NameShortSingular() string {
	return u.NameShortSingularValue
}

func (u *UnitDefinition) NameShortPlural() string {
	return u.NameShortPluralValue
}

func (u *UnitDefinition) NameLongSingular() string {
	return u.NameLongSingularValue
}

func (u *UnitDefinition) NameLongPlural() string {
	return u.NameLongPluralValue
}

func (u *UnitDefinition) FormatShortInt(amount int64, displayZero bool) string {
	return formatNumberUnitShort(amount, u, displayZero)
}

func (u *UnitDefinition) FormatShortFloat(amount float64, displayZero bool) string {
	return formatNumberUnitShort(amount, u, displayZero)
}

func (u *UnitDefinition) FormatLongInt(amount int64, displayZero bool) string {
	return formatNumberUnitLong(amount, u, displayZero)
}

func (u *UnitDefinition) FormatLongFloat(amount float64, displayZero bool) string {
	return formatNumberUnitLong(amount, u, displayZero)
}

func formatNumberUnitShort[T NumberType](amount T, unit *UnitDefinition, displayZero bool) string {
	var formatString string
	switch any(amount).(type) {
	case int64:
		formatString = "%d"
	case float64:
		formatString = "%f"
	}
	switch {
	case amount == 1 || amount == -1:
		return strings.TrimRight(fmt.Sprintf(formatString, amount), "0.") + unit.NameShortSingular()
	case amount != 0:
		return strings.TrimRight(fmt.Sprintf(formatString, amount), "0.") + unit.NameShortPlural()
	case displayZero:
		return strings.TrimRight(fmt.Sprintf(formatString, amount), "0.") + unit.NameShortPlural()
	default:
		return ""
	}
}

func formatNumberUnitLong[T NumberType](amount T, unit Unit, displayZero bool) string {
	var formatString string
	switch any(amount).(type) {
	case int64:
		formatString = "%d"
	case float64:
		formatString = "%f"
	}
	switch {
	case amount == 1 || amount == -1:
		return fmt.Sprintf(formatString, amount) + unit.NameLongSingular()
	case amount != 0:
		return fmt.Sprintf(formatString, amount) + unit.NameLongPlural()
	case displayZero:
		return fmt.Sprintf(formatString, amount) + unit.NameLongPlural()
	default:
		return ""
	}
}

// Units holds several scales of magnitude of the same UnitDefinition, for example 5m30s.
type Units interface {
	BaseUnit() *UnitDefinition
	Multipliers() map[int64]*UnitDefinition

	ParseInt(data string) (int64, error)
	ParseFloat(data string) (float64, error)

	FormatShortInt(data int64) string
	FormatShortFloat(data float64) string
	FormatLongInt(data int64) string
	FormatLongFloat(data float64) string
}

// NewUnits defines a new set of UnitsDefinition with the given parameters.
func NewUnits(baseUnit *UnitDefinition, multipliers map[int64]*UnitDefinition) *UnitsDefinition {
	ud := &UnitsDefinition{
		BaseUnitValue:    baseUnit,
		MultipliersValue: multipliers,
	}
	//ud.updateReCache()
	return ud
}

type UnitsDefinition struct {
	BaseUnitValue          *UnitDefinition           `json:"base_unit"`
	MultipliersValue       map[int64]*UnitDefinition `json:"multipliers"`
	sortedMultipliersCache []int64
	reCache                *regexp.Regexp
	reSubExpNames          map[string]int
}

func (u *UnitsDefinition) BaseUnit() *UnitDefinition {
	return u.BaseUnitValue
}

func (u *UnitsDefinition) Multipliers() map[int64]*UnitDefinition {
	return u.MultipliersValue
}

// FormatShortInt formats the passed int according to the UnitDefinition multipliers.
func (u *UnitsDefinition) FormatShortInt(data int64) string {
	if data == 0 {
		return u.BaseUnit().FormatShortInt(data, true)
	}
	remainder := data
	output := ""
	for _, multiplier := range u.getSortedMultipliersCache() {
		base := int64(math.Floor(float64(remainder) / float64(multiplier)))
		remainder -= base * multiplier
		output += formatNumberUnitShort(base, u.Multipliers()[multiplier], false)
	}
	output += formatNumberUnitShort(remainder, u.BaseUnit(), false)
	return output
}

// FormatShortFloat formats the passed float according to the UnitDefinition multipliers.
func (u *UnitsDefinition) FormatShortFloat(data float64) string {
	if data == 0 {
		return u.BaseUnit().FormatShortFloat(data, true)
	}
	remainder := data
	output := ""
	for _, multiplier := range u.getSortedMultipliersCache() {
		base := int64(math.Floor(remainder / float64(multiplier)))
		remainder -= float64(base * multiplier)
		output += u.Multipliers()[multiplier].FormatShortFloat(float64(base), false)
	}
	output += u.BaseUnit().FormatShortFloat(remainder, false)
	return output
}

// FormatLongInt formats the passed int according to the UnitDefinition multipliers.
func (u *UnitsDefinition) FormatLongInt(data int64) string {
	if data == 0 {
		return u.BaseUnitValue.FormatLongInt(data, true)
	}
	remainder := data
	output := ""
	for _, multiplier := range u.getSortedMultipliersCache() {
		base := int64(math.Floor(float64(remainder) / float64(multiplier)))
		remainder -= base * multiplier
		output += u.Multipliers()[multiplier].FormatLongInt(remainder, false)
	}
	output += u.BaseUnit().FormatLongInt(remainder, false)
	return output
}

// FormatLongFloat formats the passed float according to the UnitDefinition multipliers.
func (u *UnitsDefinition) FormatLongFloat(data float64) string {
	if data == 0 {
		return u.BaseUnitValue.FormatLongFloat(data, true)
	}
	remainder := data
	output := ""
	for _, multiplier := range u.getSortedMultipliersCache() {
		base := int64(math.Floor(remainder / float64(multiplier)))
		remainder -= float64(base * multiplier)
		output += u.Multipliers()[multiplier].FormatLongFloat(float64(base), false)
	}
	output += u.BaseUnit().FormatLongFloat(remainder, false)
	return output
}

func (u *UnitsDefinition) getSortedMultipliersCache() []int64 {
	if u.sortedMultipliersCache == nil {
		var multipliers []int64
		for multiplier := range u.MultipliersValue {
			multipliers = append(multipliers, multiplier)
		}
		sort.SliceStable(multipliers, func(i, j int) bool {
			return multipliers[i] > multipliers[j]
		})
		u.sortedMultipliersCache = multipliers
	}
	return u.sortedMultipliersCache
}

func (u *UnitsDefinition) parse(data string) (any, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return 0, &UnitParseError{
			Message: "Empty string cannot be parsed as " + u.BaseUnitValue.NameLongPlural(),
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

func (u *UnitsDefinition) handleParseMultiplier(
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

func (u *UnitsDefinition) updateReCache() {
	var parts []string
	if u.MultipliersValue != nil {
		for _, multiplier := range u.getSortedMultipliersCache() {
			unit := u.MultipliersValue[multiplier]
			parts = append(parts, fmt.Sprintf(
				"(?:|(?P<g%s>[0-9]+)\\s*(%s|%s|%s|%s))",
				regexp.QuoteMeta(fmt.Sprintf("%d", multiplier)),
				regexp.QuoteMeta(unit.NameShortSingular()),
				regexp.QuoteMeta(unit.NameShortPlural()),
				regexp.QuoteMeta(unit.NameLongSingular()),
				regexp.QuoteMeta(unit.NameLongPlural()),
			))
		}
	}
	parts = append(parts, fmt.Sprintf(
		"(?:|(?P<g1>[0-9]+(|.[0-9]+))\\s*(|%s|%s|%s|%s))",
		regexp.QuoteMeta(u.BaseUnitValue.NameShortSingular()),
		regexp.QuoteMeta(u.BaseUnitValue.NameShortPlural()),
		regexp.QuoteMeta(u.BaseUnitValue.NameLongSingular()),
		regexp.QuoteMeta(u.BaseUnitValue.NameLongPlural()),
	))
	regex := "^\\s*" + strings.Join(parts, "\\s*") + "\\s*$"
	u.reCache = regexp.MustCompile(regex)
	u.reSubExpNames = map[string]int{}
	for i, subExpName := range u.reCache.SubexpNames() {
		u.reSubExpNames[subExpName] = i
	}
}

func (u *UnitsDefinition) buildUnitParseError(data string) (any, error) {
	validUnits := []string{
		u.BaseUnitValue.NameShortSingular(),
		u.BaseUnitValue.NameShortPlural(),
		u.BaseUnitValue.NameLongSingular(),
		u.BaseUnitValue.NameLongPlural(),
	}
	for _, multiplier := range u.getSortedMultipliersCache() {
		validUnits = append(
			validUnits,
			u.MultipliersValue[multiplier].NameShortSingular(),
			u.MultipliersValue[multiplier].NameShortPlural(),
			u.MultipliersValue[multiplier].NameLongSingular(),
			u.MultipliersValue[multiplier].NameLongPlural(),
		)
	}
	return 0, UnitParseError{
		Message: fmt.Sprintf(
			"Cannot parse '%s' as '%s': invalid format, valid UnitDefinition types are: '%s",
			data,
			u.BaseUnitValue.NameLongPlural(),
			strings.Join(validUnits, "', '"),
		),
	}
}

// ParseInt parses a string into an integer.
func (u *UnitsDefinition) ParseInt(data string) (int64, error) {
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
func (u *UnitsDefinition) ParseFloat(data string) (float64, error) {
	result, err := u.parse(data)
	if err != nil {
		return 0, err
	}
	if i, ok := result.(int64); ok {
		return float64(i), nil
	}
	return result.(float64), nil
}

// UnitBytes is scaling, byte-based UnitDefinition.
var UnitBytes = NewUnits(
	NewUnit(
		"B",
		"B",
		"byte",
		"bytes",
	),
	map[int64]*UnitDefinition{
		1024: NewUnit(
			"kB",
			"kB",
			"kilobyte",
			"kilobytes",
		),
		1048576: NewUnit(
			"MB",
			"MB",
			"megabyte",
			"megabytes",
		),
		1073741824: NewUnit(
			"GB",
			"GB",
			"gigabyte",
			"gigabytes",
		),
		1099511627776: NewUnit(
			"TB",
			"TB",
			"terabyte",
			"terabytes",
		),
		1125899906842624: NewUnit(
			"PB",
			"PB",
			"petabyte",
			"petabytes",
		),
	},
)

// UnitDurationNanoseconds is a nanosecond-based UnitDefinition for time durations.
var UnitDurationNanoseconds = NewUnits(
	NewUnit(
		"ns",
		"ns",
		"nanosecond",
		"nanoseconds",
	),
	map[int64]*UnitDefinition{
		int64(time.Microsecond): NewUnit(
			"μs",
			"μs",
			"microsecond",
			"microseconds",
		),
		int64(time.Millisecond): NewUnit(
			"ms",
			"ms",
			"milliseconds",
			"milliseconds",
		),
		int64(time.Second): NewUnit(
			"s",
			"s",
			"second",
			"seconds",
		),
		int64(time.Minute): NewUnit(
			"m",
			"m",
			"minute",
			"minutes",
		),
		int64(time.Hour): NewUnit(
			"H",
			"H",
			"hour",
			"hours",
		),
		int64(24 * time.Hour): NewUnit(
			"d",
			"d",
			"day",
			"days",
		),
	},
)

// UnitDurationSeconds is a second-based UnitDefinition for time durations.
var UnitDurationSeconds = NewUnits(
	NewUnit(
		"s",
		"s",
		"second",
		"seconds",
	),
	map[int64]*UnitDefinition{
		60: NewUnit(
			"m",
			"m",
			"minute",
			"minutes",
		),
		3600: NewUnit(
			"H",
			"H",
			"hour",
			"hours",
		),
		86400: NewUnit(
			"d",
			"d",
			"day",
			"days",
		),
	},
)

// UnitCharacters is a single UnitDefinition for characters.
var UnitCharacters = NewUnits(
	NewUnit(
		"char",
		"chars",
		"character",
		"characters",
	),
	nil,
)

// UnitPercentage is a single UnitDefinition for percentages.
var UnitPercentage = NewUnits(
	NewUnit(
		"%",
		"%",
		"percent",
		"percent",
	),
	nil,
)
