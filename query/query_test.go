package query

import (
	"testing"
	"time"

	"membership13/domain"
)

func TestQueryRecords(t *testing.T) {
	items := []domain.Record{domain.NewRecord("a", 13, "basic", time.Unix(1, 0)), domain.NewRecord("b", 14, "basic", time.Unix(2, 0))}
	if len(FilterRecords(items, RecordFilter{MemberID: 13})) != 1 {
		t.Fatal("member filter failed")
	}
	if Paginate(items, 1, 1).Total != 2 {
		t.Fatal("pagination total failed")
	}
}
