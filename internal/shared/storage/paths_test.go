package storage_test

import (
	"testing"

	"github.com/radius/radius-backend/internal/shared/storage"
)

func TestBuildPermanentKeys(t *testing.T) {
	tempKey := "temp/2026-06-06/abc.png"

	avatar := storage.BuildUserAvatarKey(tempKey)
	if avatar == "" || avatar == tempKey {
		t.Fatalf("unexpected avatar key: %q", avatar)
	}
	if got, want := avatar[:len("user-avatar/")], "user-avatar/"; got != want {
		t.Fatalf("avatar prefix = %q, want %q", got, want)
	}

	cover := storage.BuildProjectCoverKey(tempKey)
	if got, want := cover[:len("project-cover/")], "project-cover/"; got != want {
		t.Fatalf("cover prefix = %q, want %q", got, want)
	}

	attachment := storage.BuildAttachmentKey(tempKey)
	if got, want := attachment[:len("attachment/")], "attachment/"; got != want {
		t.Fatalf("attachment prefix = %q, want %q", got, want)
	}
}

func TestValidateTempKey(t *testing.T) {
	if err := storage.ValidateTempKey("temp/2026-06-06/file.png"); err != nil {
		t.Fatalf("expected valid temp key: %v", err)
	}
	if err := storage.ValidateTempKey("user-avatar/file.png"); err == nil {
		t.Fatal("expected invalid temp key error")
	}
}

func TestTrimTempKey(t *testing.T) {
	empty := ""
	value := " temp/key "
	if got := storage.TrimTempKey(nil); got != "" {
		t.Fatalf("nil = %q", got)
	}
	if got := storage.TrimTempKey(&empty); got != "" {
		t.Fatalf("empty = %q", got)
	}
	if got := storage.TrimTempKey(&value); got != "temp/key" {
		t.Fatalf("trimmed = %q", got)
	}
}
