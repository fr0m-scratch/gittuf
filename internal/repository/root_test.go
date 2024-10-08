// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"testing"

	"github.com/gittuf/gittuf/internal/policy"
	"github.com/gittuf/gittuf/internal/signerverifier"
	"github.com/gittuf/gittuf/internal/signerverifier/dsse"
	sslibdsse "github.com/gittuf/gittuf/internal/third_party/go-securesystemslib/dsse"
	sslibsv "github.com/gittuf/gittuf/internal/third_party/go-securesystemslib/signerverifier"
	"github.com/gittuf/gittuf/internal/tuf"
	"github.com/stretchr/testify/assert"
)

func TestInitializeRoot(t *testing.T) {
	// The helper also runs InitializeRoot for this test
	r, rootKeyBytes := createTestRepositoryWithRoot(t, "")

	key, err := tuf.LoadKeyFromBytes(rootKeyBytes)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := sslibsv.NewVerifierFromSSLibKey(key)
	if err != nil {
		t.Fatal(err)
	}

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	assert.Nil(t, err)
	assert.Equal(t, key.KeyID, rootMetadata.Roles[policy.RootRoleName].KeyIDs[0])
	assert.Equal(t, key.KeyID, state.RootEnvelope.Signatures[0].KeyID)

	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{verifier}, 1)
	assert.Nil(t, err)
}

func TestAddRootKey(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}
	originalKeyID, err := sv.KeyID()
	if err != nil {
		t.Fatal(err)
	}

	var newRootKey *sslibsv.SSLibKey

	newRootKey, err = tuf.LoadKeyFromBytes(targetsKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddRootKey(testCtx, sv, newRootKey, false)
	assert.Nil(t, err)

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	assert.Nil(t, err)
	assert.Equal(t, []string{originalKeyID, newRootKey.KeyID}, rootMetadata.Roles[policy.RootRoleName].KeyIDs)
	assert.Equal(t, originalKeyID, state.RootEnvelope.Signatures[0].KeyID)
	assert.Equal(t, 2, len(state.RootPublicKeys))

	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)
}

func TestRemoveRootKey(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	rootKey, err := tuf.LoadKeyFromBytes(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	originalSigner, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddRootKey(testCtx, originalSigner, rootKey, false)
	if err != nil {
		t.Fatal(err)
	}

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	// We should have no additions as we tried to add the same key
	assert.Equal(t, 1, len(state.RootPublicKeys))
	assert.Equal(t, 1, len(rootMetadata.Roles[policy.RootRoleName].KeyIDs))

	newRootKey, err := tuf.LoadKeyFromBytes(targetsPubKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddRootKey(testCtx, originalSigner, newRootKey, false)
	if err != nil {
		t.Fatal(err)
	}

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}
	rootMetadata, err = state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, rootMetadata.Roles[policy.RootRoleName].KeyIDs, rootKey.KeyID)
	assert.Contains(t, rootMetadata.Roles[policy.RootRoleName].KeyIDs, newRootKey.KeyID)
	assert.Equal(t, 2, len(state.RootPublicKeys))

	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{originalSigner}, 1)
	assert.Nil(t, err)

	newSigner, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(targetsKeyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	// We can use the newly added root key to revoke the old one
	err = r.RemoveRootKey(testCtx, newSigner, rootKey.KeyID, false)
	assert.Nil(t, err)

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err = state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, rootMetadata.Roles[policy.RootRoleName].KeyIDs, newRootKey.KeyID)
	assert.Equal(t, 1, len(rootMetadata.Roles[policy.RootRoleName].KeyIDs))
	assert.Equal(t, 1, len(state.RootPublicKeys))

	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{newSigner}, 1)
	assert.Nil(t, err)
}

func TestAddTopLevelTargetsKey(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	key, err := tuf.LoadKeyFromBytes(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddTopLevelTargetsKey(testCtx, sv, key, false)
	assert.Nil(t, err)

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	assert.Nil(t, err)
	assert.Equal(t, key.KeyID, rootMetadata.Roles[policy.RootRoleName].KeyIDs[0])
	assert.Equal(t, key.KeyID, rootMetadata.Roles[policy.TargetsRoleName].KeyIDs[0])
	assert.Equal(t, key.KeyID, state.RootEnvelope.Signatures[0].KeyID)

	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)
}

func TestRemoveTopLevelTargetsKey(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	rootKey, err := tuf.LoadKeyFromBytes(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddTopLevelTargetsKey(testCtx, sv, rootKey, false)
	if err != nil {
		t.Fatal(err)
	}

	targetsKey, err := tuf.LoadKeyFromBytes(targetsKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddTopLevelTargetsKey(testCtx, sv, targetsKey, false)
	if err != nil {
		t.Fatal(err)
	}

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rootKey.KeyID, rootMetadata.Roles[policy.TargetsRoleName].KeyIDs[0])
	assert.Contains(t, rootMetadata.Roles[policy.TargetsRoleName].KeyIDs, rootKey.KeyID)
	assert.Contains(t, rootMetadata.Roles[policy.TargetsRoleName].KeyIDs, targetsKey.KeyID)
	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)

	err = r.RemoveTopLevelTargetsKey(testCtx, sv, rootKey.KeyID, false)
	assert.Nil(t, err)

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err = state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, rootMetadata.Roles[policy.TargetsRoleName].KeyIDs, targetsKey.KeyID)
	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)
}

func TestAddGitHubAppKey(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	key, err := tuf.LoadKeyFromBytes(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddGitHubAppKey(testCtx, sv, key, false)
	assert.Nil(t, err)

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	assert.Nil(t, err)

	assert.Equal(t, key.KeyID, rootMetadata.Roles[policy.GitHubAppRoleName].KeyIDs[0])
	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)
}

func TestRemoveGitHubAppKey(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	key, err := tuf.LoadKeyFromBytes(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	err = r.AddGitHubAppKey(testCtx, sv, key, false)
	if err != nil {
		t.Fatal(err)
	}

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, key.KeyID, rootMetadata.Roles[policy.GitHubAppRoleName].KeyIDs[0])
	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)

	err = r.RemoveGitHubAppKey(testCtx, sv, false)
	assert.Nil(t, err)

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err = state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Empty(t, rootMetadata.Roles[policy.GitHubAppRoleName].KeyIDs)
	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)
}

func TestTrustGitHubApp(t *testing.T) {
	t.Run("GitHub app role not defined", func(t *testing.T) {
		r, keyBytes := createTestRepositoryWithRoot(t, "")

		_, err := tuf.LoadKeyFromBytes(keyBytes)
		if err != nil {
			t.Fatal(err)
		}
		sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
		if err != nil {
			t.Fatal(err)
		}

		err = r.TrustGitHubApp(testCtx, sv, false)
		assert.Nil(t, err)

		_, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
		assert.ErrorIs(t, err, policy.ErrNoGitHubAppRoleDeclared)
	})

	t.Run("GitHub app role defined", func(t *testing.T) {
		r, keyBytes := createTestRepositoryWithRoot(t, "")

		key, err := tuf.LoadKeyFromBytes(keyBytes)
		if err != nil {
			t.Fatal(err)
		}
		sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
		if err != nil {
			t.Fatal(err)
		}

		state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
		if err != nil {
			t.Fatal(err)
		}

		rootMetadata, err := state.GetRootMetadata()
		assert.Nil(t, err)

		assert.False(t, rootMetadata.GitHubApprovalsTrusted)

		err = r.AddGitHubAppKey(testCtx, sv, key, false)
		assert.Nil(t, err)

		err = r.TrustGitHubApp(testCtx, sv, false)
		assert.Nil(t, err)

		state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
		if err != nil {
			t.Fatal(err)
		}

		rootMetadata, err = state.GetRootMetadata()
		assert.Nil(t, err)

		assert.True(t, rootMetadata.GitHubApprovalsTrusted)
		_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
		assert.Nil(t, err)

		// Test if we can trust again if already trusted
		err = r.TrustGitHubApp(testCtx, sv, false)
		assert.Nil(t, err)
	})
}

func TestUntrustGitHubApp(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	key, err := tuf.LoadKeyFromBytes(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	assert.Nil(t, err)

	assert.False(t, rootMetadata.GitHubApprovalsTrusted)

	err = r.AddGitHubAppKey(testCtx, sv, key, false)
	assert.Nil(t, err)

	err = r.TrustGitHubApp(testCtx, sv, false)
	assert.Nil(t, err)

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err = state.GetRootMetadata()
	assert.Nil(t, err)

	assert.True(t, rootMetadata.GitHubApprovalsTrusted)
	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)

	err = r.UntrustGitHubApp(testCtx, sv, false)
	assert.Nil(t, err)

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err = state.GetRootMetadata()
	assert.Nil(t, err)

	assert.False(t, rootMetadata.GitHubApprovalsTrusted)
	_, err = dsse.VerifyEnvelope(testCtx, state.RootEnvelope, []sslibdsse.Verifier{sv}, 1)
	assert.Nil(t, err)
}

func TestUpdateRootThreshold(t *testing.T) {
	r, _ := createTestRepositoryWithRoot(t, "")

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(rootMetadata.Roles[policy.RootRoleName].KeyIDs))
	assert.Equal(t, 1, rootMetadata.Roles[policy.RootRoleName].Threshold)

	signer, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(rootKeyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	secondKey, err := tuf.LoadKeyFromBytes(targetsKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	if err := r.AddRootKey(testCtx, signer, secondKey, false); err != nil {
		t.Fatal(err)
	}

	err = r.UpdateRootThreshold(testCtx, signer, 2, false)
	assert.Nil(t, err)

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err = state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(rootMetadata.Roles[policy.RootRoleName].KeyIDs))
	assert.Equal(t, 2, rootMetadata.Roles[policy.RootRoleName].Threshold)
}

func TestUpdateTopLevelTargetsThreshold(t *testing.T) {
	r, keyBytes := createTestRepositoryWithRoot(t, "")

	key, err := tuf.LoadKeyFromBytes(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(keyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	if err := r.AddTopLevelTargetsKey(testCtx, sv, key, false); err != nil {
		t.Fatal(err)
	}

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(rootMetadata.Roles[policy.TargetsRoleName].KeyIDs))
	assert.Equal(t, 1, rootMetadata.Roles[policy.TargetsRoleName].Threshold)

	targetsKey, err := tuf.LoadKeyFromBytes(targetsKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	if err := r.AddTopLevelTargetsKey(testCtx, sv, targetsKey, false); err != nil {
		t.Fatal(err)
	}

	err = r.UpdateTopLevelTargetsThreshold(testCtx, sv, 2, false)
	assert.Nil(t, err)

	state, err = policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	rootMetadata, err = state.GetRootMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(rootMetadata.Roles[policy.TargetsRoleName].KeyIDs))
	assert.Equal(t, 2, rootMetadata.Roles[policy.TargetsRoleName].Threshold)
}

func TestSignRoot(t *testing.T) {
	r, _ := createTestRepositoryWithRoot(t, "")

	rootSigner, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(rootKeyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	// Add targets key as a root key
	secondKey, err := tuf.LoadKeyFromBytes(targetsPubKeyBytes)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.AddRootKey(testCtx, rootSigner, secondKey, false); err != nil {
		t.Fatal(err)
	}

	secondSigner, err := signerverifier.NewSignerVerifierFromSecureSystemsLibFormat(targetsKeyBytes) //nolint:staticcheck
	if err != nil {
		t.Fatal(err)
	}

	// Add signature to root
	err = r.SignRoot(testCtx, secondSigner, false)
	assert.Nil(t, err)

	state, err := policy.LoadCurrentState(testCtx, r.r, policy.PolicyStagingRef)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(state.RootEnvelope.Signatures))
}
