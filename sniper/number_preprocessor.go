package sniper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NumberPreprocessor is the type responsible for cleaning and converting numbers.
type NumberPreprocessor struct {
	compoundNumberRegex *regexp.Regexp
	singleNumberRegex   *regexp.Regexp
	currencyRegex       *regexp.Regexp
	commaRegex          *regexp.Regexp
	hyphenRegex         *regexp.Regexp // Added for phone numbers/ID strings
	ordinalRegex        *regexp.Regexp

	// Maps for text-to-digit conversion
	units map[string]int
	tens  map[string]int
}

// NewNumberPreprocessor initializes the regexes and maps once for efficiency.
func NewNumberPreprocessor() *NumberPreprocessor {
	np := &NumberPreprocessor{
		units: map[string]int{
			"zero": 0, "one": 1, "two": 2, "too": 2, "to": 2, "three": 3, "four": 4,
			"five": 5, "six": 6, "seven": 7, "eight": 8, "nine": 9,
			"ten": 10, "tin": 10, "eleven": 11, "twelve": 12, "thirteen": 13,
			"fourteen": 14, "fifteen": 15, "sixteen": 16,
			"seventeen": 17, "eighteen": 18, "nineteen": 19,
		},
		tens: map[string]int{
			"twenty": 20, "thirty": 30, "forty": 40, "fifty": 50,
			"sixty": 60, "seventy": 70, "eighty": 80, "ninety": 90,
		},
	}

	// Regex to find compound numbers like "twenty-two" or "twenty two"
	// \b ensures word boundaries so we don't match partial words.
	// (?i) makes it case insensitive.
	np.compoundNumberRegex = regexp.MustCompile(`(?i)\b(twenty|thirty|forty|fifty|sixty|seventy|eighty|ninety)[-\s](one|two|three|four|five|six|seven|eight|nine)\b`)

	// Regex to find remaining single words (0-19 and 20, 30, etc.)
	// We build this dynamically from the maps to keep it clean.
	var words []string
	for k := range np.units {
		words = append(words, k)
	}
	for k := range np.tens {
		words = append(words, k)
	}
	words = append(words, "hundred") // specific edge case

	pattern := fmt.Sprintf(`(?i)\b(%s)\b`, strings.Join(words, "|"))
	np.singleNumberRegex = regexp.MustCompile(pattern)

	// Regex to find $ immediately followed by a digit (e.g., $100)
	np.currencyRegex = regexp.MustCompile(`\$(\d)`)

	// Regex to find commas surrounded by digits (e.g., 1,000)
	np.commaRegex = regexp.MustCompile(`(\d),(\d)`)

	// Regex to find hyphens surrounded by digits (e.g., 555-0199)
	np.hyphenRegex = regexp.MustCompile(`(\d)-(\d)`)

	// Regex to find ordinal suffixes attached to digits (e.g., 1st, 2nd, 3rd, 4th, 100th)
	// (?i) = case insensitive
	// (\d+) = capture the number
	// (st|nd|rd|th) = match the suffix
	// \b = word boundary to ensure we don't cut off specific weird codes or hex strings unnecessarily
	np.ordinalRegex = regexp.MustCompile(`(?i)(\d+)(st|nd|rd|th)\b`)

	return np
}

// Process takes a raw string and applies number purification.
func (np *NumberPreprocessor) Process(input string) string {
	processed := input

	// 1. Handle Compound Words (21-99) first
	// We use ReplaceAllStringFunc to calculate the value dynamically
	processed = np.compoundNumberRegex.ReplaceAllStringFunc(processed, func(match string) string {
		// Split on hyphen or space
		parts := strings.FieldsFunc(match, func(r rune) bool {
			return r == '-' || r == ' '
		})

		if len(parts) != 2 {
			return match // Should not happen given the regex
		}

		tensVal := np.tens[strings.ToLower(parts[0])]
		unitVal := np.units[strings.ToLower(parts[1])]
		return strconv.Itoa(tensVal + unitVal)
	})

	// 2. Handle Single Words (0-20, 30, 100, etc.)
	processed = np.singleNumberRegex.ReplaceAllStringFunc(processed, func(match string) string {
		lower := strings.ToLower(match)
		if lower == "hundred" {
			return "100"
		}

		if val, ok := np.units[lower]; ok {
			return strconv.Itoa(val)
		}
		if val, ok := np.tens[lower]; ok {
			return strconv.Itoa(val)
		}
		return match
	})

	// 3. Strip Currency Symbols ($100 -> 100)
	// We replace "$1" with "1"
	processed = np.currencyRegex.ReplaceAllString(processed, "$1")

	// 4. Strip Commas (1,000 -> 1000)
	// We loop this because "1,000,000" has multiple commas.
	// A simple replace all works, but we only want to remove commas *between digits*.
	for {
		if !np.commaRegex.MatchString(processed) {
			break
		}
		processed = np.commaRegex.ReplaceAllString(processed, "$1$2")
	}

	// 5. Strip Hyphens (555-0199 -> 5550199)
	// Just like commas, we loop to ensure we catch multiple hyphens (e.g. 1-800-555-0199)
	for {
		if !np.hyphenRegex.MatchString(processed) {
			break
		}
		processed = np.hyphenRegex.ReplaceAllString(processed, "$1$2")
	}

	// 6. Strip Ordinal Suffixes (1st -> 1, 2nd -> 2, 3rd -> 3, 4th -> 4)
	// We replace the match with just the captured digit ($1)
	processed = np.ordinalRegex.ReplaceAllString(processed, "$1")

	return processed
}
