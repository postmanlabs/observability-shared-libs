package tags

import (
	"strings"

	"github.com/pkg/errors"
)

type Key string

// Identifies the source of a trace or spec. Valid values:
//   - user - trace/spec was manually created by a user
//   - uploaded - trace/spec was manually uploaded by a user
//   - deployment - trace/spec from a staging or production deployment
//   - ci - trace/spec from a CI pipeline
const XAkitaSource Key = "x-akita-source"

// The original filesystem path of an uploaded trace.
const XAkitaTraceLocalPath Key = "x-akita-trace-local-path"

// == Generic CI tags =========================================================

// Identifies the CI framework from which a trace or spec was obtained (e.g.,
// CircleCI, Travis).
const XAkitaCI Key = "x-akita-ci"

// == CircleCI tags ===========================================================

// The contents of the CIRCLE_BUILD_URL environment variable. Attached to
// traces and specs derived from a CircleCI job.
const XAkitaCircleCIBuildURL Key = "x-akita-circleci-build-url"

// == Travis tags =============================================================

// The contents of the TRAVIS_BUILD_WEB_URL environment variable. Attached to
// traces and specs derived from a Travis job.
const XAkitaTravisBuildWebURL Key = "x-akita-travis-build-web-url"

// The contents of the TRAVIS_JOB_WEB_URL environment variable. Attached to
// traces and specs derived from a Travis job.
const XAkitaTravisJobWebURL Key = "x-akita-travis-job-web-url"

// == Generic git tags ========================================================

// Identifies the git branch from which the trace or spec was derived. Attached
// to traces or specs obtained from CI.
const XAkitaGitBranch Key = "x-akita-git-branch"

// Identifies the git commit hash from which the trace or spec was derived.
// Attached to traces or specs obtained from CI.
const XAkitaGitCommit Key = "x-akita-git-commit"

// A link to the git repository. Attached to traces or specs obtained from a
// pull/merge request.
const XAkitaGitRepoURL Key = "x-akita-git-repo-url"

// == GitHub tags =============================================================

// Identifies the GitHub organization in which the pull request was made.
// Attached to traces or specs obtained from a GitHub pull request.
const AkitaGitHubOrganizationID Key = "x-akita-github-organization-id"

// Identifies the GitHub PR number associated with the pull request. Attached
// to traces or specs obtained from a GitHub pull request.
const XAkitaGitHubPR Key = "x-akita-github-pr"

// A link to the GitHub pull request. Attached to traces or specs obtained
// from a GitHub pull request.
const XAkitaGitHubPRURL Key = "x-akita-github-pr-url"

// Identifies the GitHub repository for which the pull request was made.
// Attached to traces or specs obtained from a GitHub pull request.
const XAkitaGitHubRepo Key = "x-akita-github-repo"

// == GitLab tags =============================================================

const XAkitaGitLabProject Key = "x-akita-gitlab-project"
const XAkitaGitLabMRIID Key = "x-akita-gitlab-mr-iid"

// == Methods =================================================================

// Determines whether a key is reserved for Akita internal use.
func (k Key) IsReserved() bool {
	s := strings.ToLower(string(k))
	return strings.HasPrefix(s, "x-akita-")
}

// Returns an error if the key is reserved for Akita internal use.
func (k Key) CheckReserved() error {
	if !k.IsReserved() {
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

		if err := k.CheckReserved(); err != nil {
			return nil, err
		}

		results[k] = v
	}
	return results, nil
}
