// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gittuf

import (
	"testing"

	"github.com/gittuf/gittuf/internal/cache"
	"github.com/gittuf/gittuf/internal/dev"
	"github.com/gittuf/gittuf/internal/policy"
	"github.com/gittuf/gittuf/internal/rsl"
	"github.com/stretchr/testify/assert"
)

func TestPopulateCache(t *testing.T) {
	t.Run("successful cache population", func(t *testing.T) {
		t.Setenv(dev.DevModeKey, "1")

		tmpDir := t.TempDir()
		repo := createTestRepositoryWithPolicy(t, tmpDir)

		err := repo.PopulateCache()
		assert.Nil(t, err)

		firstEntry, _, err := rsl.GetFirstEntry(repo.r)
		if err != nil {
			t.Fatal(err)
		}

		latestEntry, err := rsl.GetLatestEntry(repo.r)
		if err != nil {
			t.Fatal(err)
		}

		// This is sorted in order of occurrence for us
		allPolicyEntries, _, err := rsl.GetReferenceEntriesInRangeForRef(repo.r, firstEntry.GetID(), latestEntry.GetID(), policy.PolicyRef)
		if err != nil {
			t.Fatal(err)
		}

		expectedPolicyEntries := []cache.RSLEntryIndex{}
		for _, entry := range allPolicyEntries {
			expectedPolicyEntries = append(expectedPolicyEntries, cache.RSLEntryIndex{EntryNumber: entry.GetNumber(), EntryID: entry.GetID().String()})
		}

		persistentCache, err := cache.LoadPersistentCache(repo.r)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectedPolicyEntries, persistentCache.PolicyEntries)
	})

	t.Run("successful repeated cache population", func(t *testing.T) {
		t.Setenv(dev.DevModeKey, "1")

		tmpDir := t.TempDir()
		repo := createTestRepositoryWithPolicy(t, tmpDir)

		err := repo.PopulateCache()
		assert.Nil(t, err)

		currentCacheID, err := repo.r.GetReference(cache.Ref)
		if err != nil {
			t.Fatal(err)
		}

		err = repo.PopulateCache()
		// No error is reported
		assert.Nil(t, err)

		// However, no changes were committed either, because the cache
		// didn't change.
		newCacheID, err := repo.r.GetReference(cache.Ref)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, currentCacheID, newCacheID)
	})
}