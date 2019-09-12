package models

// FlagIsSet checks if a given flag is set in flags
func FlagIsSet(flags uint64, flag uint64) bool {
	return flags&flag != 0
}

// FlagIsNotSet checks if a given flag is not in flags
func FlagIsNotSet(flags uint64, flag uint64) bool {
	return !FlagIsSet(flags, flag)
}
