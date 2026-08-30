package version

type ComparisonResult int

const (
	EQUAL     ComparisonResult = 0
	LESS      ComparisonResult = -1
	GREATER   ComparisonResult = 1
	UNDEFINED ComparisonResult = -2
)
