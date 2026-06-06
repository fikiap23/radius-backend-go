package storage

import (
	"path"
	"strings"

	"github.com/google/uuid"
)

func extensionFromKey(key string) string {
	return SafeExtension(path.Base(key))
}

// BuildUserAvatarKey returns a permanent object key for a user avatar.
func BuildUserAvatarKey(tempKey string) string {
	return buildPermanentKey("user-avatar", tempKey)
}

// BuildProjectCoverKey returns a permanent object key for a project cover image.
func BuildProjectCoverKey(tempKey string) string {
	return buildPermanentKey("project-cover", tempKey)
}

// BuildAttachmentKey returns a permanent object key for a generic attachment.
func BuildAttachmentKey(tempKey string) string {
	return buildPermanentKey("attachment", tempKey)
}

func buildPermanentKey(prefix, tempKey string) string {
	ext := extensionFromKey(tempKey)
	return prefix + "/" + uuid.NewString() + ext
}

// TrimTempKey returns a non-empty trimmed temp key, or empty when absent/blank.
func TrimTempKey(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
