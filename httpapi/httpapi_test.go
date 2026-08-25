package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"membership13/domain"
	"membership13/service"
	"membership13/store"
)

func TestHTTPAPI(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "http.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s).WithClock(func() time.Time { return time.Unix(10, 0) })
	if err := svc.RegisterUser(domain.NewUser("u13", 13, "Member", "gold", time.Unix(1, 0))); err != nil {
		t.Fatal(err)
	}
	h := New(svc).Routes()
	req := httptest.NewRequest(http.MethodPost, "/records", strings.NewReader(`{"id":"r1","member_id":13,"benefit_code":"basic-gift"}`))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("create status %d", res.Code)
	}
	check := httptest.NewRecorder()
	h.ServeHTTP(check, httptest.NewRequest(http.MethodGet, "/records?member=13", nil))
	if check.Code != http.StatusOK {
		t.Fatalf("query status %d", check.Code)
	}
}
