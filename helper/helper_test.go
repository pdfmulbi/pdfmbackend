package helper

import (
	"testing"
)

func TestGetPresensiThisMonth(t *testing.T) {
	t.Skip("Skipping DB test")
	uri := SRVLookup("mongodb+srv://xx:xxx@cxxx.xxx.mongodb.net/")
	print(uri)
}
