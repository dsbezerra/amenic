package timeutil

import (
	"testing"
	"time"
)

func TestIsToday(t *testing.T) {
	testdate := time.Now()
	expectedResult := true
	actualResult := IsToday(&testdate)
	if actualResult != expectedResult {
		t.Fatalf("Expected %t but got %t", expectedResult, actualResult)
	}

	testdate = testdate.AddDate(0, 0, 1)
	expectedResult = false
	actualResult = IsToday(&testdate)
	if actualResult != expectedResult {
		t.Fatalf("Expected %t but got %t", expectedResult, actualResult)
	}

	testdate = testdate.AddDate(0, 0, -2)
	expectedResult = false
	actualResult = IsToday(&testdate)
	if actualResult != expectedResult {
		t.Fatalf("Expected %t but got %t", expectedResult, actualResult)
	}
}
