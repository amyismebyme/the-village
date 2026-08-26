package external

import "testing"

func TestIdentityKey(t *testing.T) {
	identity := Identity{
		Source:     Source("reddit"),
		ExternalID: "t3_abc123",
	}

	if got := identity.Key(); got != "reddit:t3_abc123" {
		t.Fatalf(
			"unexpected identity key %q",
			got,
		)
	}
}

func TestIdentityValidate(t *testing.T) {
	tests := []struct {
		name     string
		identity Identity
		wantErr  bool
	}{
		{
			name: "valid",
			identity: Identity{
				Source:     Source("reddit"),
				ExternalID: "t3_abc123",
			},
		},
		{
			name: "missing source",
			identity: Identity{
				ExternalID: "t3_abc123",
			},
			wantErr: true,
		},
		{
			name: "missing external id",
			identity: Identity{
				Source: Source("reddit"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.identity.Validate()

			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf(
					"unexpected validation error: %v",
					err,
				)
			}
		})
	}
}

func TestSameIdentity(t *testing.T) {
	left := Identity{
		Source:     Source("reddit"),
		ExternalID: "t3_abc123",
	}

	right := Identity{
		Source:     Source("reddit"),
		ExternalID: "t3_abc123",
	}

	if !SameIdentity(left, right) {
		t.Fatal("expected identities to match")
	}
}

func TestDifferentSourcesAreDifferentIdentities(t *testing.T) {
	reddit := Identity{
		Source:     Source("reddit"),
		ExternalID: "123",
	}

	other := Identity{
		Source:     Source("other"),
		ExternalID: "123",
	}

	if SameIdentity(reddit, other) {
		t.Fatal("expected identities to differ")
	}
}

func TestValidateUniqueIdentitiesRejectsDuplicates(t *testing.T) {
	identities := []Identity{
		{
			Source:     Source("reddit"),
			ExternalID: "123",
		},
		{
			Source:     Source("reddit"),
			ExternalID: "123",
		},
	}

	if err := ValidateUniqueIdentities(identities); err == nil {
		t.Fatal("expected duplicate identity error")
	}
}

func TestValidateUniqueIdentitiesAllowsSameExternalIDFromDifferentSources(
	t *testing.T,
) {
	identities := []Identity{
		{
			Source:     Source("reddit"),
			ExternalID: "123",
		},
		{
			Source:     Source("other"),
			ExternalID: "123",
		},
	}

	if err := ValidateUniqueIdentities(identities); err != nil {
		t.Fatalf(
			"expected identities to be distinct: %v",
			err,
		)
	}
}
