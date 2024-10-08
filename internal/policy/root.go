// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"time"

	"github.com/gittuf/gittuf/internal/tuf"
)

var (
	ErrCannotMeetThreshold = errors.New("insufficient keys to meet threshold")
	ErrRootMetadataNil     = errors.New("rootMetadata is nil")
	ErrRootKeyNil          = errors.New("root key not found")
	ErrTargetsMetadataNil  = errors.New("targetsMetadata not found")
	ErrTargetsKeyNil       = errors.New("targetsKey is nil")
	ErrGitHubAppKeyNil     = errors.New("app key is nil")
	ErrKeyIDEmpty          = errors.New("keyID is empty")
)

const GitHubAppRoleName = "github-app"

// InitializeRootMetadata initializes a new instance of tuf.RootMetadata with
// default values and a given key. The default values are version set to 1,
// expiry date set to one year from now, and the provided key is added.
func InitializeRootMetadata(key *tuf.Key) *tuf.RootMetadata {
	rootMetadata := tuf.NewRootMetadata()
	rootMetadata.SetExpires(time.Now().AddDate(1, 0, 0).Format(time.RFC3339))
	rootMetadata.AddKey(key)

	rootMetadata.AddRole(RootRoleName, tuf.Role{
		KeyIDs:    []string{key.KeyID},
		Threshold: 1,
	})

	return rootMetadata
}

// AddRootKey adds rootKey as a trusted public key in rootMetadata for the
// Root role.
func AddRootKey(rootMetadata *tuf.RootMetadata, rootKey *tuf.Key) *tuf.RootMetadata {
	if _, ok := rootMetadata.Roles[RootRoleName]; !ok {
		return rootMetadata
	}

	rootMetadata.AddKey(rootKey)

	rootRole := rootMetadata.Roles[RootRoleName]

	for _, keyID := range rootRole.KeyIDs {
		if keyID == rootKey.KeyID {
			return rootMetadata
		}
	}

	rootRole.KeyIDs = append(rootRole.KeyIDs, rootKey.KeyID)
	rootMetadata.Roles[RootRoleName] = rootRole

	return rootMetadata
}

// DeleteRootKey removes keyID from the list of trusted Root
// public keys in rootMetadata. It does not remove the key entry itself as it
// does not check if other roles can be verified using the same key.
func DeleteRootKey(rootMetadata *tuf.RootMetadata, keyID string) (*tuf.RootMetadata, error) {
	if _, ok := rootMetadata.Roles[RootRoleName]; !ok {
		return rootMetadata, nil
	}

	rootRole := rootMetadata.Roles[RootRoleName]
	if len(rootRole.KeyIDs) <= rootRole.Threshold {
		return nil, ErrCannotMeetThreshold
	}
	for i, k := range rootRole.KeyIDs {
		if k == keyID {
			rootRole.KeyIDs = append(rootRole.KeyIDs[:i], rootRole.KeyIDs[i+1:]...)
			break
		}
	}
	rootMetadata.Roles[RootRoleName] = rootRole

	return rootMetadata, nil
}

// AddTargetsKey adds the 'targetsKey' as a trusted public key in 'rootMetadata'
// for the top level Targets role.
func AddTargetsKey(rootMetadata *tuf.RootMetadata, targetsKey *tuf.Key) (*tuf.RootMetadata, error) {
	if rootMetadata == nil {
		return nil, ErrRootMetadataNil
	}
	if targetsKey == nil {
		return nil, ErrTargetsKeyNil
	}

	rootMetadata.Keys[targetsKey.KeyID] = targetsKey

	if _, ok := rootMetadata.Roles[TargetsRoleName]; !ok {
		rootMetadata.AddRole(TargetsRoleName, tuf.Role{
			KeyIDs:    []string{targetsKey.KeyID},
			Threshold: 1,
		})
		return rootMetadata, nil
	}

	targetsRole := rootMetadata.Roles[TargetsRoleName]
	for _, keyID := range targetsRole.KeyIDs {
		if keyID == targetsKey.KeyID {
			return rootMetadata, nil
		}
	}

	targetsRole.KeyIDs = append(targetsRole.KeyIDs, targetsKey.KeyID)
	rootMetadata.Roles[TargetsRoleName] = targetsRole

	return rootMetadata, nil
}

// DeleteTargetsKey removes the key matching 'keyID' from trusted public keys
// for top level Targets role in 'rootMetadata'. Note: It doesn't remove the key
// entry itself as it doesn't check if other roles can use the same key.
func DeleteTargetsKey(rootMetadata *tuf.RootMetadata, keyID string) (*tuf.RootMetadata, error) {
	if rootMetadata == nil {
		return nil, ErrRootMetadataNil
	}
	if keyID == "" {
		return nil, ErrKeyIDEmpty
	}
	if _, ok := rootMetadata.Roles[TargetsRoleName]; !ok {
		return rootMetadata, nil
	}

	targetsRole := rootMetadata.Roles[TargetsRoleName]

	if len(targetsRole.KeyIDs) <= targetsRole.Threshold {
		return nil, ErrCannotMeetThreshold
	}

	newKeyIDs := []string{}
	for _, k := range targetsRole.KeyIDs {
		if k != keyID {
			newKeyIDs = append(newKeyIDs, k)
		}
	}
	targetsRole.KeyIDs = newKeyIDs

	rootMetadata.Roles[TargetsRoleName] = targetsRole

	return rootMetadata, nil
}

// AddGitHubAppKey adds the 'appKey' as a trusted public key in 'rootMetadata'
// for the special GitHub app role. This key is used to verify GitHub pull
// request approval attestation signatures.
func AddGitHubAppKey(rootMetadata *tuf.RootMetadata, appKey *tuf.Key) (*tuf.RootMetadata, error) {
	if rootMetadata == nil {
		return nil, ErrRootMetadataNil
	}
	if appKey == nil {
		return nil, ErrGitHubAppKeyNil
	}

	// TODO: support multiple keys / threshold for app
	rootMetadata.Keys[appKey.KeyID] = appKey
	role := tuf.Role{
		KeyIDs:    []string{appKey.KeyID},
		Threshold: 1,
	}
	rootMetadata.AddRole(GitHubAppRoleName, role) // AddRole replaces the specified role if it already exists
	return rootMetadata, nil
}

// DeleteGitHubAppKey removes the special GitHub app role from the root
// metadata.
func DeleteGitHubAppKey(rootMetadata *tuf.RootMetadata) (*tuf.RootMetadata, error) {
	if rootMetadata == nil {
		return nil, ErrRootMetadataNil
	}

	// TODO: support multiple keys / threshold for app
	delete(rootMetadata.Roles, GitHubAppRoleName)
	return rootMetadata, nil
}

// EnableGitHubAppApprovals sets GitHubApprovalsTrusted to true in the
// root metadata.
func EnableGitHubAppApprovals(rootMetadata *tuf.RootMetadata) (*tuf.RootMetadata, error) {
	if rootMetadata == nil {
		return nil, ErrRootMetadataNil
	}

	rootMetadata.GitHubApprovalsTrusted = true
	return rootMetadata, nil
}

// DisableGitHubAppApprovals sets GitHubApprovalsTrusted to false in the root
// metadata.
func DisableGitHubAppApprovals(rootMetadata *tuf.RootMetadata) (*tuf.RootMetadata, error) {
	if rootMetadata == nil {
		return nil, ErrRootMetadataNil
	}

	rootMetadata.GitHubApprovalsTrusted = false
	return rootMetadata, nil
}

// UpdateRootThreshold sets the threshold for the Root role.
func UpdateRootThreshold(rootMetadata *tuf.RootMetadata, threshold int) (*tuf.RootMetadata, error) {
	rootRole, ok := rootMetadata.Roles[RootRoleName]
	if !ok {
		return nil, ErrTargetsMetadataNil
	}

	if len(rootRole.KeyIDs) < threshold {
		return nil, ErrCannotMeetThreshold
	}

	rootRole.Threshold = threshold
	rootMetadata.Roles[RootRoleName] = rootRole

	return rootMetadata, nil
}

// UpdateTargetsThreshold sets the threshold for the top level Targets role.
func UpdateTargetsThreshold(rootMetadata *tuf.RootMetadata, threshold int) (*tuf.RootMetadata, error) {
	targetsRole, ok := rootMetadata.Roles[TargetsRoleName]
	if !ok {
		return nil, ErrTargetsMetadataNil
	}

	if len(targetsRole.KeyIDs) < threshold {
		return nil, ErrCannotMeetThreshold
	}

	targetsRole.Threshold = threshold
	rootMetadata.Roles[TargetsRoleName] = targetsRole

	return rootMetadata, nil
}
