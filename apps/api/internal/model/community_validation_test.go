package model

import "testing"

func TestCommunityValidate(t *testing.T) {

	tests := []struct {
		name      string
		community Community
		wantErr   bool
	}{
		{
			name: "valid",
			community: Community{
				Name:        "Toronto Men",
				Slug:        "toronto-men",
				Description: "Helping men build meaningful friendships.",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			community: Community{
				Name: "",
				Slug: "toronto-men",
			},
			wantErr: true,
		},
		{
			name: "invalid slug",
			community: Community{
				Name: "Toronto",
				Slug: "Toronto-Men",
			},
			wantErr: true,
		},
		{
			name: "description too long",
			community: Community{
				Name:        "Toronto",
				Slug:        "toronto",
				Description: string(make([]byte, 2001)),
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {

		err := tc.community.Validate()

		if tc.wantErr && err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}

		if !tc.wantErr && err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
	}
}
