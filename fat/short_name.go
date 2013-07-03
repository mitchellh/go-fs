package fat

import (
	"bytes"
	"fmt"
	"strings"
)

// generateShortName takes a list of existing short names and a long
// name and generates the next valid short name. This process is done
// according to the MS specification.
func generateShortName(longName string, used []string) (string, error) {
	// Remove leading periods and uppercase, as required
	longName = strings.TrimLeft(longName, ".")
	longName = strings.ToUpper(longName)

	// Split the string at the final "."
	dotIdx := strings.LastIndex(longName, ".")
	if dotIdx == -1 {
		dotIdx = len(longName)-1
	}

	ext := cleanShortString(longName[dotIdx+1:len(longName)])
	ext = ext[0:3]
	rawName := longName[0:dotIdx]
	name := cleanShortString(rawName)
	simpleName := fmt.Sprintf("%s.%s", name, ext)

	doSuffix := name != rawName || len(name) > 8
	if !doSuffix {
		for _, usedSingle := range used {
			if strings.ToUpper(usedSingle) == simpleName {
				doSuffix = true
				break
			}
		}
	}

	if doSuffix {
		found := false
		for i := 1; i < 99999; i++ {
			serial := fmt.Sprintf("~%d", i)

			nameOffset := 8 - len(serial)
			if len(name) < nameOffset {
				nameOffset = len(name)
			}

			serialName := fmt.Sprintf("%s%s", name[0:nameOffset], serial)
			simpleName = fmt.Sprintf("%s.%s", serialName, ext)

			exists := false
			for _, usedSingle := range used {
				if strings.ToUpper(usedSingle) == simpleName {
					exists = true
					break
				}
			}

			if !exists {
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("could not generate short name for %s", longName)
		}
	}

	return simpleName, nil
}

func cleanShortString(v string) string {
	var result bytes.Buffer
	for _, char := range v {
		// We skip these chars
		if char == '.' || char == ' ' {
			continue
		}

		if !validShortChar(char) {
			char = '_'
		}

		result.WriteRune(char)
	}

	return result.String()
}

func validShortChar(char rune) bool {
	if char >= 'A' && char <= 'Z' {
		return true
	}

	if char >= '0' && char <= '9' {
		return true
	}

	validShortSymbols := []rune{
		'_', '^', '$', '~', '!', '#', '%', '&', '-', '{', '}', '(',
		')', '@', '\'', '`',
	}

	for _, valid := range validShortSymbols {
		if char == valid {
			return true
		}
	}

	return false
}
