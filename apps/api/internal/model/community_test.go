package model

import (
	"testing"
	"time"
)

func TestCommunityFields(t *testing.T) {

	now := time.Now()

	c := Community{
		ID: 123,

		Name: "Toronto Men",

		Slug: "toronto-men",

		Category: "C1",

		Description: "Helping men build friendships.",

		ExternalSource: "internal",

		CreatedAt: now,

		UpdatedAt: now,
	}

	if c.ID != 123 {
		t.Fatal("unexpected ID")
	}

	if c.Name != "Toronto Men" {
		t.Fatal("unexpected Name")
	}

	if c.Slug != "toronto-men" {
		t.Fatal("unexpected Slug")
	}

	if c.Category != "C1" {
		t.Fatal("unexpected Category")
	}
	if c.ExternalSource != "internal" {
		t.Fatal("unexpected ExternalSource")
	}

	if c.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set")
	}

	if c.UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt should be set")
	}
}
