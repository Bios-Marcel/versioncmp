package versioncmp_test

import (
	"testing"

	"github.com/Bios-Marcel/versioncmp"
	"github.com/stretchr/testify/assert"
)

func Test_CompareVersion(t *testing.T) {
	defaultRules := versioncmp.VersionCompareRules{}

	t.Run("meta values ignored", func(t *testing.T) {
		assert.Empty(t, versioncmp.Compare("1", "1-bmsrtbq23ui", defaultRules))

		t.Parallel()
	})

	t.Run("nightly equals", func(t *testing.T) {
		assert.Empty(t, versioncmp.Compare("1.0.0-nightly-12412", "1.0.0-nightly-187623", defaultRules))
		assert.Empty(t, versioncmp.Compare("1.0.0-nightly-12412", "2.0.0-nightly-187623", defaultRules))

		t.Parallel()
	})

	t.Run("pre-releases", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "1.0.0-pre-3", versioncmp.Compare("1.0.0-pre-2", "1.0.0-pre-3", defaultRules))
	})

	t.Run("semver-like", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, versioncmp.Compare("0", "0", defaultRules))
		assert.Empty(t, versioncmp.Compare("1", "1", defaultRules))
		assert.Empty(t, versioncmp.Compare("0.0.1", "0.0.1", defaultRules))
		assert.Empty(t, versioncmp.Compare("1.0.0", "1.0.0", defaultRules))
		assert.Empty(t, versioncmp.Compare("1.0", "1.0.0", defaultRules))
		assert.Empty(t, versioncmp.Compare("1", "1.0.0", defaultRules))
		assert.Empty(t, versioncmp.Compare("2.0.0", "2.0", defaultRules))

		assert.Equal(t, "2", versioncmp.Compare("2", "1", defaultRules))
		assert.Equal(t, "2.0", versioncmp.Compare("2.0", "1", defaultRules))
		assert.Equal(t, "2.0.0", versioncmp.Compare("2.0.0", "1", defaultRules))
		assert.Equal(t, "2", versioncmp.Compare("2", "1.9", defaultRules))
	})

	t.Run("stability precedence", func(t *testing.T) {
		assert.Equal(t, "1", versioncmp.Compare("1", "1-dev", defaultRules))
		assert.Equal(t, "1", versioncmp.Compare("1", "1-alpha", defaultRules))
		assert.Equal(t, "1", versioncmp.Compare("1", "1-beta", defaultRules))
		assert.Equal(t, "1", versioncmp.Compare("1", "1-rc", defaultRules))
		assert.Equal(t, "1", versioncmp.Compare("1", "1-rc2", defaultRules))
		assert.Equal(t, "1", versioncmp.Compare("1", "1-pre", defaultRules))
		assert.Equal(t, "1", versioncmp.Compare("1", "1-prerelease", defaultRules))
		assert.Equal(t, "1", versioncmp.Compare("1", "1-pre-release", defaultRules))

		assert.Equal(t, "1-pre", versioncmp.Compare("1-pre", "1-dev", defaultRules))
		assert.Equal(t, "1-pre", versioncmp.Compare("1-pre", "1-alpha", defaultRules))
		assert.Equal(t, "1-pre", versioncmp.Compare("1-pre", "1-beta", defaultRules))
		assert.Equal(t, "1-pre", versioncmp.Compare("1-pre", "1-rc", defaultRules))

		assert.Equal(t, "1-rc", versioncmp.Compare("1-rc", "1-dev", defaultRules))
		assert.Equal(t, "1-rc", versioncmp.Compare("1-rc", "1-alpha", defaultRules))
		assert.Equal(t, "1-rc", versioncmp.Compare("1-rc", "1-beta", defaultRules))

		assert.Equal(t, "1-beta", versioncmp.Compare("1-beta", "1-dev", defaultRules))
		assert.Equal(t, "1-beta", versioncmp.Compare("1-beta", "1-alpha", defaultRules))

		assert.Equal(t, "1-alpha", versioncmp.Compare("1-alpha", "1-dev", defaultRules))
	})

	t.Run("release candidates", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, versioncmp.Compare("1.0.0-rc", "1.0.0-rc", defaultRules))
		assert.Empty(t, versioncmp.Compare("1.0.0-rc1", "1.0.0-rc1", defaultRules))
		assert.Empty(t, versioncmp.Compare("1.0.0-rc2", "1.0.0-rc2", defaultRules))

		assert.Equal(t, "1.0.0-rc2", versioncmp.Compare("1.0.0-rc1", "1.0.0-rc2", defaultRules))
		assert.Equal(t, "2.0.0-rc1", versioncmp.Compare("2.0.0-rc1", "1.0.0-rc2", defaultRules))
	})

	t.Run("reverse release candidates notation", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, versioncmp.Compare("1.0.0-1rc", "1.0.0-1rc", defaultRules))
		assert.Empty(t, versioncmp.Compare("1.0.0-2rc", "1.0.0-2rc", defaultRules))

		assert.Equal(t, "1.0.0-2rc", versioncmp.Compare("1.0.0-1rc", "1.0.0-2rc", defaultRules))
		assert.Equal(t, "2.0.0-1rc", versioncmp.Compare("2.0.0-1rc", "1.0.0-2rc", defaultRules))
	})

	t.Run("reverse dates", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, versioncmp.Compare("01-02-2024", "01-02-2024", defaultRules))
		assert.Empty(t, versioncmp.Compare("01.02.2024", "01.02.2024", defaultRules))

		assert.Equal(t, "01-02-2024", versioncmp.Compare("01-02-2024", "02-01-2024", defaultRules))
		assert.Equal(t, "02-01-2025", versioncmp.Compare("01-02-2024", "02-01-2025", defaultRules))
		assert.Equal(t, "01-02-2028", versioncmp.Compare("01-02-2028", "02-01-2027", defaultRules))

		// Ensure no false positive, all of these need to be treated as
		// major.minor.patch, as we are sure these can't be years / we'll be
		// dead so we don't care.
		assert.Equal(t, "02.02.1700", versioncmp.Compare("02.02.1700", "02.01.1700", defaultRules))
		assert.Equal(t, "02.02.1700", versioncmp.Compare("02.02.1700", "02.01.1701", defaultRules))

		// Interpret as major minor, so major 10 is bigger major 8.
		assert.Equal(t, "10.02.8354", versioncmp.Compare("10.02.8354", "08.05.8354", defaultRules))
		// The patch component is bigger, but irrelevant due to the difference
		// in mjaor. This proves we aren't treating it as a date.
		assert.Equal(t, "05.01.8354", versioncmp.Compare("05.01.8354", "02.01.8355", defaultRules))
	})

	t.Run("dates", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, versioncmp.Compare("2024-02-01", "2024-02-01", defaultRules))
		assert.Empty(t, versioncmp.Compare("2024.02.01", "2024.02.01", defaultRules))

		assert.Equal(t, "2024-02-01", versioncmp.Compare("2024-02-01", "2024-01-02", defaultRules))
		assert.Equal(t, "2025-01-02", versioncmp.Compare("2024-02-01", "2025-01-02", defaultRules))
	})

	t.Run("inkscapes wild mix", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "1.3.1_2023-11-16_91b66b0783",
			versioncmp.Compare("1.3_2023-07-21_0e150ed6c4", "1.3.1_2023-11-16_91b66b0783", defaultRules))
	})

	t.Run("chromium", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "123.0.6312.87-r1262506",
			versioncmp.Compare("123.0.6312.59-r1262506", "123.0.6312.87-r1262506", defaultRules))
	})
}
