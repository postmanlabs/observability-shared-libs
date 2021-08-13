package tags

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	// Identifies the source of a trace or spec. See `Source` for values.
	XAkitaSource Key = "x-akita-source"

	// Identifies the process by which a trace or spec was created. See
	// `CreatedBy` for values.
	XAkitaCreatedBy Key = "x-akita-created-by"

	// The original filesystem path of an uploaded trace.
	XAkitaTraceLocalPath Key = "x-akita-trace-local-path"
)

// Generic CI tags
const (
	// Identifies the CI framework from which a trace or spec was obtained (e.g.,
	// CircleCI, Travis).
	XAkitaCI Key = "x-akita-ci"

	// Each model derived from a PR (or MR) is automatically diffed against a
	// baseline spec. This tag identifies the AKID for that baseline spec.
	// Attached to specs.
	XAkitaComparedWith Key = "x-akita-compared-with"

	// Each model derived from a PR (or MR) is automatically diffed against a
	// baseline spec. This tag identifies the number of differences that were
	// found in this diff. Attached to specs for which this automatic diffing is
	// done.
	XAkitaNumDifferences Key = "x-akita-num-differences"
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
	// May also be attached to deployment traces if the git commit is known
	// and applicable.
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

// Deployment tags
const (
	// The name of the deployment environment, suggested values "production"
	// or "staging" but may be a user-defined value
	XAkitaDeployment Key = "x-akita-deployment"

	// Used for specifications where the number of witnesses is too large.
	XAkitaTruncated Key = "x-akita-truncated"
)

// AWS deployment tags
const (
	XAkitaAWSRegion Key = "x-akita-aws-region"
)

// Kubernetes deployment tags
const (
	// Kubernetes namespace
	// = metadata.namespace in the Downward API
	XAkitaKubernetesNamespace Key = "x-akita-k8s-namespace"

	// Node (host) on which the collection agent is running
	// = spec.nodeName in the Downward API (v1.4.0+)
	XAkitaKubernetesNode Key = "x-akita-k8s-node"

	// IP address of the Kubernetes node
	// = status.hostIP in the Downward API
	XAkitaKubernetesHostIP Key = "x-akita-k8s-host-ip"

	// Pod in which the collection agent is running; may be
	// a dedicated pod using host networking, or an application
	// pod when running as a sidecar.
	// = metadata.name in the Downward API
	XAkitaKubernetesPod Key = "x-akita-k8s-pod"

	// Pod IP address
	// = status.podIP in the Downward API (v1.7.0+)
	XAkitaKubernetesPodIP Key = "x-akita-k8s-pod-ip"

	// Daemonset used to create the collection agent, if any
	XAkitaKubernetesDaemonset Key = "x-akita-k8s-daemonset"

	// Not included: metadata.uid, metadata.labels, metadata.annotations,
	// resources limits and requests, spec.serviceAccountName
)

// Packet-capture tags
const (
	// A comma-separated list of interfaces on which packets were captured.
	XAkitaDumpInterfacesFlag Key = "x-akita-dump-interfaces-flag"

	// The packet filter given on the command line to capture packets.
	XAkitaDumpFilterFlag Key = "x-akita-dump-filter-flag"
)

// Tags applied to a copy of the spec
const (
	XAkitaOriginalOrganizationID = "x-akita-original-organization-id"

	XAkitaOriginalService = "x-akita-original-service"

	XAkitaOriginalServiceID = "x-akita-original-service-id"

	XAkitaOriginalSpec = "x-akita-original-spec"

	XAkitaOriginalSpecID = "x-akita-original-spec-id"
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
