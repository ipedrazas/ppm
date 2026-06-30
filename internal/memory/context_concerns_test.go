package memory

import "testing"

func TestContextSurfacesConcerns(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	tagProject(t, s, "beta", "Beta") // out of backend scope

	if _, err := s.AddStandard("has-summary", "Has summary", "tag:backend", "warn", "has-summary", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddInitiative("gdpr", "GDPR", "tag:backend", ""); err != nil {
		t.Fatal(err)
	}

	ctx, err := s.Context("alpha", 3)
	if err != nil {
		t.Fatalf("Context: %v", err)
	}
	if len(ctx.Standards) != 1 || ctx.Standards[0].Concern != "has-summary" {
		t.Errorf("standards in scope = %+v, want has-summary", ctx.Standards)
	}
	if ctx.Standards[0].Status != StatusFail {
		t.Errorf("placeholder summary should fail: %v", ctx.Standards[0].Status)
	}
	if len(ctx.Initiatives) != 1 || ctx.Initiatives[0].Concern != "gdpr" {
		t.Errorf("initiatives in scope = %+v, want gdpr", ctx.Initiatives)
	}

	// A project outside the concern scopes sees no obligations.
	ctxBeta, err := s.Context("beta", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(ctxBeta.Standards) != 0 || len(ctxBeta.Initiatives) != 0 {
		t.Errorf("beta should have no obligations: %+v %+v", ctxBeta.Standards, ctxBeta.Initiatives)
	}
}
