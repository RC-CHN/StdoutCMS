package task

import (
	"slices"
	"testing"
)

func TestFindOrphans_EmptyInputs(t *testing.T) {
	if o := findOrphans(nil, nil); len(o) != 0 {
		t.Errorf("expected no orphans from nil inputs, got %v", o)
	}
	if o := findOrphans([]string{}, []string{}); len(o) != 0 {
		t.Errorf("expected no orphans from empty inputs, got %v", o)
	}
}

func TestFindOrphans_NoKeys(t *testing.T) {
	orphans := findOrphans([]string{}, []string{"![](/img/images/abc.png)"})
	if len(orphans) != 0 {
		t.Errorf("no keys should mean no orphans, got %v", orphans)
	}
}

func TestFindOrphans_NoContents(t *testing.T) {
	keys := []string{"images/a.png", "images/b.png"}
	orphans := findOrphans(keys, nil)
	if len(orphans) != 2 {
		t.Errorf("all keys should be orphans when there's no content, got %d", len(orphans))
	}
}

func TestFindOrphans_AllReferenced(t *testing.T) {
	keys := []string{"images/cat.png", "images/dog.jpg"}
	contents := []string{
		"Check out my cat: ![](/img/images/cat.png)",
		"Also here is a dog: ![](/img/images/dog.jpg)",
	}

	orphans := findOrphans(keys, contents)
	if len(orphans) != 0 {
		t.Errorf("expected no orphans, got %v", orphans)
	}
}

func TestFindOrphans_SomeReferenced(t *testing.T) {
	keys := []string{"images/used.png", "images/orphan.png", "images/also-used.jpg"}

	contents := []string{
		"Here is used.png: ![](/img/images/used.png)",
		"images/also-used.jpg appears here as well",
		// orphan.png is never mentioned
	}

	orphans := findOrphans(keys, contents)
	if len(orphans) != 1 || orphans[0] != "images/orphan.png" {
		t.Errorf("expected [images/orphan.png], got %v", orphans)
	}
}

func TestFindOrphans_ContentSpansMultiplePosts(t *testing.T) {
	keys := []string{"images/x.png", "images/y.png"}

	// x.png is in post 1, y.png is in post 2 — both should be referenced
	contents := []string{
		"post one: ![](images/x.png)",
		"post two: ![](images/y.png)",
	}

	orphans := findOrphans(keys, contents)
	if len(orphans) != 0 {
		t.Errorf("keys spread across posts should all be referenced, got %v", orphans)
	}
}

func TestFindOrphans_RealWorldMarkdown(t *testing.T) {
	keys := []string{
		"images/550e8400-e29b-41d4-a716-446655440000.png",
		"images/6ba7b810-9dad-11d1-80b4-00c04fd430c8.png",
		"images/unused-uuid-here.png",
	}

	contents := []string{
		`## My Blog Post

Here is a screenshot:

![screenshot](/img/images/550e8400-e29b-41d4-a716-446655440000.png)

And another one:

<img src="/img/images/6ba7b810-9dad-11d1-80b4-00c04fd430c8.png" alt="diagram" />

The end.`,
	}

	orphans := findOrphans(keys, contents)
	if len(orphans) != 1 || orphans[0] != "images/unused-uuid-here.png" {
		t.Errorf("expected only unused-uuid-here to be orphan, got %v", orphans)
	}
}

func TestFindOrphans_Deduplication(t *testing.T) {
	// Same key referenced multiple times — should be found once
	keys := []string{"images/dup.png", "images/single.png"}
	contents := []string{
		"ref one: images/dup.png",
		"ref two: images/dup.png",
		// single.png not referenced
	}

	orphans := findOrphans(keys, contents)
	if len(orphans) != 1 || orphans[0] != "images/single.png" {
		t.Errorf("expected only single.png to be orphan, got %v", orphans)
	}
}

func TestFindOrphans_StableOrder(t *testing.T) {
	keys := []string{"images/a.png", "images/b.png", "images/c.png"}
	contents := []string{}

	orphans := findOrphans(keys, contents)
	// All should be orphans, in original order
	if !slices.Equal(orphans, keys) {
		t.Errorf("orphans should preserve input order, got %v", orphans)
	}
}
