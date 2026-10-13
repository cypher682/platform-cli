package logs

import (
	"context"
	"errors"
	"sort"
	"testing"
)

type fake struct{ fail string }

func (f fake) Stream(ctx context.Context, svc string, _ Options, out chan<- Line) error {
	if svc == f.fail {
		return errors.New("boom")
	}
	for _, t := range []string{"a", "b"} {
		out <- Line{Service: svc, Text: t}
	}
	return nil
}

func collect(src Source, svcs []string) ([]string, error) {
	out := make(chan Line)
	done := make(chan error, 1)
	go func() {
		done <- FanIn(context.Background(), src, svcs, Options{}, out)
		close(out)
	}()
	var got []string
	for l := range out {
		got = append(got, l.Service+":"+l.Text)
	}
	return got, <-done
}

func TestFanInMergesAllStreams(t *testing.T) {
	got, err := collect(fake{}, []string{"api", "auth"})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got) // arrival order across goroutines is nondeterministic
	want := []string{"api:a", "api:b", "auth:a", "auth:b"}
	if len(got) != 4 {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestFanInJoinsErrors(t *testing.T) {
	got, err := collect(fake{fail: "auth"}, []string{"api", "auth"})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(got) != 2 {
		t.Fatalf("healthy service should still stream, got %v", got)
	}
}

func TestMatcher(t *testing.T) {
	m, err := NewMatcher("err(or)?")
	if err != nil {
		t.Fatal(err)
	}
	if !m("an error here") || m("all fine") {
		t.Fatal("matcher wrong")
	}
	if _, err := NewMatcher("("); err == nil {
		t.Fatal("expected regex error")
	}
}
