package client

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/go-github/v36/github"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

// hmacSecret contains a hmac token and the time when it's created.
type hmacSecret struct {
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

// hmacsForRepo contains all hmac tokens configured for a repo, org or globally.
type hmacsForRepo []hmacSecret

type genericEvent struct {
	Sender github.User       `json:"sender"`
	Repo   github.Repository `json:"repository"`
}

// ValidatePayload ensures that the request payload signature matches the key.
func ValidatePayload(payload []byte, sig string, tokenGenerator func() []byte) bool {
	var event genericEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		logrus.WithError(err).Info("validatePayload couldn't unmarshal the github event payload")

		return false
	}

	if !strings.HasPrefix(sig, "sha1=") {
		return false
	}

	sig = sig[5:]
	sb, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	hmacs, err := extractHmacs(event.Repo.GetFullName(), tokenGenerator)
	if err != nil {
		logrus.WithError(err).Error("couldn't unmarshal the hmac secret")

		return false
	}

	// If we have a match with any valid hmac, we can validate successfully.
	for _, key := range hmacs {
		mac := hmac.New(sha1.New, key)
		mac.Write(payload)
		expected := mac.Sum(nil)

		if hmac.Equal(sb, expected) {
			return true
		}
	}

	return false
}

// PayloadSignature returns the signature that matches the payload.
func PayloadSignature(payload []byte, key []byte) string {
	mac := hmac.New(sha1.New, key)
	mac.Write(payload)
	sum := mac.Sum(nil)

	return "sha1=" + hex.EncodeToString(sum)
}

// extractHmacs returns all *valid* HMAC tokens for given repository/organization.
// It considers only the tokens at the most specific level configured for the given repo.
// For example : if a token for repo is present and it doesn't match the repo, we will
// not try to find a match with org level token. However if no token is present for repo,
// we will try to match with org level.
func extractHmacs(repo string, tokenGenerator func() []byte) ([][]byte, error) {
	t := tokenGenerator()
	repoToTokenMap := map[string]hmacsForRepo{}

	if err := yaml.Unmarshal(t, &repoToTokenMap); err != nil {
		// To keep backward compatibility, we are going to assume that in case of error,
		// whole file is a single line hmac token.
		logrus.WithError(err).Trace("Couldn't unmarshal the hmac secret as hierarchical file. Parsing as single token format")

		return [][]byte{t}, nil
	}

	orgName := strings.Split(repo, "/")[0]

	if val, ok := repoToTokenMap[repo]; ok {
		return extractTokens(val), nil
	}

	if val, ok := repoToTokenMap[orgName]; ok {
		return extractTokens(val), nil
	}

	if val, ok := repoToTokenMap["*"]; ok {
		return extractTokens(val), nil
	}

	return nil, errors.New("invalid content in secret file, global token doesn't exist")
}

// extractTokens return tokens for any given level of tree.
func extractTokens(allTokens hmacsForRepo) [][]byte {
	validTokens := make([][]byte, len(allTokens))
	for i := range allTokens {
		validTokens[i] = []byte(allTokens[i].Value)
	}

	return validTokens
}
