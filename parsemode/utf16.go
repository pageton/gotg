package parsemode

import (
	"unicode/utf16"
	"unicode/utf8"
)

// UTF16RuneCountInString returns the number of UTF-16 code units in the string.
// Telegram uses UTF-16 for calculating entity offsets and lengths.
func UTF16RuneCountInString(s string) int32 {
	// Fast path for ASCII strings
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return int32(len(s))
	}

	// Convert to UTF-16 and count
	runes := utf16.Encode([]rune(s))
	return int32(len(runes))
}

// UTF16OffsetToByteOffset converts a UTF-16 code unit offset to a byte offset.
// Returns -1 if the offset is invalid.
func UTF16OffsetToByteOffset(s string, utf16Offset int32) int {
	if utf16Offset < 0 {
		return -1
	}

	// Fast path for ASCII
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		if int(utf16Offset) > len(s) {
			return -1
		}
		return int(utf16Offset)
	}

	// Convert to UTF-16
	utf16Units := utf16.Encode([]rune(s))
	if int(utf16Offset) > len(utf16Units) {
		return -1
	}

	// Count runes up to the offset
	runeCount := 0
	byteOffset := 0
	for _, r := range s {
		runeCount++
		if runeCount > int(utf16Offset) {
			break
		}
		byteOffset += utf8.RuneLen(r)
	}

	return byteOffset
}

// ByteOffsetToUTF16Offset converts a byte offset to a UTF-16 code unit offset.
func ByteOffsetToUTF16Offset(s string, byteOffset int) int32 {
	if byteOffset < 0 || byteOffset > len(s) {
		return -1
	}

	// Fast path for ASCII
	ascii := true
	for i := 0; i < byteOffset; i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return int32(byteOffset)
	}

	// Count UTF-16 units up to byte offset
	runesBefore := []rune(s[:byteOffset])
	utf16Units := utf16.Encode(runesBefore)
	return int32(len(utf16Units))
}

// SubstringByUTF16Offset extracts a substring using UTF-16 offsets.
// Returns empty string if offsets are invalid.
func SubstringByUTF16Offset(s string, start, end int32) string {
	if start < 0 || end < start {
		return ""
	}

	// Convert to UTF-16 for easier indexing
	utf16Units := utf16.Encode([]rune(s))
	if int(end) > len(utf16Units) {
		end = int32(len(utf16Units))
	}

	// Extract the slice and convert back to string
	slicedUnits := utf16Units[int(start):int(end)]
	return string(utf16.Decode(slicedUnits))
}

// utf16RuneCountInString is an alias for UTF16RuneCountInString for internal use.
func utf16RuneCountInString(s string) int32 {
	return UTF16RuneCountInString(s)
}
