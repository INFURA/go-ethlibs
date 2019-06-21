package eth

// OptionalString can be used to generate a string pointer from a static string easily
func OptionalString(s string) *string { return &s }
