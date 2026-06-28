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

// BuildTaskAttachmentKey returns a permanent object key for a task attachment.
func BuildTaskAttachmentKey(taskID, attachmentID, fileName string) string {
	safeName := SafeFileName(fileName)
	return "attachments/" + taskID + "/" + attachmentID + "/" + safeName
}

// SafeFileName returns a basename safe for object storage keys.
func SafeFileName(fileName string) string {
	base := path.Base(strings.TrimSpace(fileName))
	if base == "" || base == "." || base == ".." {
		return "file"
	}
	var b strings.Builder
	for _, ch := range base {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == '_' {
			b.WriteRune(ch)
		} else {
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "file"
	}
	return out
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
