package tags

import (
	"strings"

	"github.com/google/martian/v3/tags"
	"github.com/pkg/errors"
)

type Key = tags.Key

const (
	// Identifies the source of a trace or spec. See `Source` for values.
	XAkitaSource Key = "x-akita-source"

	// The original filesystem path of an uploaded trace.
	XAkitaTraceLocalPath Key = "x-akita-trace-local-path"
)

// Generic CI tags
const (
	// Identifies the CI framework from which a trace or spec was obtained (e.g.,
	// CircleCI, Travis).
	XAkitaCI Key = "x-akita-ci"
)

// CircleCI tags
const (
	// The contents of the CIRCLE_BUILD_URL environment variable. Attached to
	// traces and specs derived from a CircleCI job.
	XAkitaCircleCIBuildURL Key = "x-akita-circleci-build-url"
)

// Travis tags
const (
	// The contents of the TRAVIS_BUILD_WEB_URL environment variable. Attached to
	// traces and specs derived from a Travis job.
	XAkitaTravisBuildWebURL Key = "x-akita-travis-build-web-url"

	// The contents of the TRAVIS_JOB_WEB_URL environment variable. Attached to
	// traces and specs derived from a Travis job.
	XAkitaTravisJobWebURL Key = "x-akita-travis-job-web-url"
)

// Generic git tags
const (
	// Identifies the git branch from which the trace or spec was derived.
	// Attached to traces or specs obtained from CI.
	XAkitaGitBranch Key = "x-akita-git-branch"

	// Identifies the git commit hash from which the trace or spec was derived.
	// Attached to traces or specs obtained from CI.
	XAkitaGitCommit Key = "x-akita-git-commit"

	// A link to the git repository. Attached to traces or specs obtained from a
	// pull/merge request.
	XAkitaGitRepoURL Key = "x-akita-git-repo-url"
)

// GitHub tags
const (
	// Identifies the GitHub PR number associated with the pull request. Attached
	// to traces or specs obtained from a GitHub pull request.
	XAkitaGitHubPR Key = "x-akita-github-pr"

	// A link to the GitHub pull request. Attached to traces or specs obtained
	// from a GitHub pull request.
	XAkitaGitHubPRURL Key = "x-akita-github-pr-url"

	// Identifies the GitHub repository for which the pull request was made.
	// Attached to traces or specs obtained from a GitHub pull request.
	XAkitaGitHubRepo Key = "x-akita-github-repo"
)

// GitLab tags
const (
	XAkitaGitLabProject Key = "x-akita-gitlab-project"
	XAkitaGitLabMRIID   Key = "x-akita-gitlab-mr-iid"
)

// Packet-capture tags
const (
	// A comma-separated list of interfaces on which packets were captured.
	XAkitaDumpInterfacesFlag Key = "x-akita-dump-interfaces-flag"

	// The packet filter given on the command line to capture packets.
	XAkitaDumpFilterFlag Key = "x-akita-dump-filter-flag"
)

// Determines whether a key is reserved for Akita internal use.
func IsReservedKey(k Key) bool {
	s := strings.ToLower(string(k))
	return strings.HasPrefix(s, "x-akita-")
}

// Returns an error if the key is reserved for Akita internal use.
func CheckReservedKey(k Key) error {
	if !IsReservedKey(k) {
		return nil
	}

	return errors.New(`Tags starting with "x-akita-" are reserved for Akita internal use.`)
}

// Returns a map from parsing a list of "key=value" pairs.
func FromPairs(pairs []string) (map[Key]string, error) {
	results := make(map[Key]string, len(pairs))
	for _, p := range pairs {
		parts := strings.Split(p, "=")
		if len(parts) != 2 {
			return nil, errors.Errorf("%s is not a valid key=value format", p)
		}

		k, v := Key(parts[0]), parts[1]
		if _, ok := results[k]; ok {
			return nil, errors.Errorf("tag with key %s specified more than once", k)
		}

		if err := CheckReservedKey(k); err != nil {
			return nil, err
		}

		results[k] = v
	}
	return results, nil
}
