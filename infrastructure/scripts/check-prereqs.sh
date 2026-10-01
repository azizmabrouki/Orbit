#!/bin/sh
# Check that the tools Orbit needs on the host are installed and recent
# enough. Prints one line per tool and exits non-zero if any check fails.
# Never installs anything: see docs/deployment/local-setup.md.
set -eu

MIN_K3D=5.9.0
MIN_KUBECTL=1.35.0
MIN_GO=1.27.0

failed=""

pass() { printf '  [ok]   %-8s %s\n' "$1" "$2"; }
fail() {
	printf '  [FAIL] %-8s %s\n' "$1" "$2"
	failed="$failed $1"
}

# version_ge A B: true if dotted version A >= B (missing parts count as 0).
version_ge() {
	a=$1
	b=$2
	for _ in 1 2 3; do
		a_part=${a%%.*}
		b_part=${b%%.*}
		a_part=${a_part:-0}
		b_part=${b_part:-0}
		if [ "$a_part" -gt "$b_part" ]; then return 0; fi
		if [ "$a_part" -lt "$b_part" ]; then return 1; fi
		case $a in *.*) a=${a#*.} ;; *) a="" ;; esac
		case $b in *.*) b=${b#*.} ;; *) b="" ;; esac
	done
	return 0
}

# first_version TEXT: first word of TEXT that looks like a version
# ("v1.2.3", "go1.27.1", "1.35.9"), as plain x.y[.z].
first_version() {
	printf '%s\n' "$1" | tr -d '",' | awk '{
		for (i = 1; i <= NF; i++) {
			w = $i
			sub(/^(v|go)/, "", w)
			if (match(w, /^[0-9]+\.[0-9]+(\.[0-9]+)?/)) {
				print substr(w, 1, RLENGTH)
				exit
			}
		}
	}'
}

# check_min NAME FOUND MIN
check_min() {
	if [ -z "$2" ]; then
		fail "$1" "installed, but its version could not be read"
	elif version_ge "$2" "$3"; then
		pass "$1" "$2 (>= $3)"
	else
		fail "$1" "$2 is too old (need >= $3)"
	fi
}

echo "Checking Orbit prerequisites:"

if ! command -v docker >/dev/null 2>&1; then
	fail docker "not found"
elif ! server=$(docker version --format '{{.Server.Version}}' 2>/dev/null); then
	fail docker "installed, but the daemon is not reachable (is it running, and are you in the docker group?)"
else
	pass docker "$server (daemon reachable)"
fi

if command -v k3d >/dev/null 2>&1; then
	check_min k3d "$(first_version "$(k3d version 2>/dev/null | head -n 1)")" "$MIN_K3D"
else
	fail k3d "not found (need >= $MIN_K3D)"
fi

if command -v kubectl >/dev/null 2>&1; then
	check_min kubectl "$(first_version "$(kubectl version --client -o json 2>/dev/null | grep gitVersion)")" "$MIN_KUBECTL"
else
	fail kubectl "not found (need >= $MIN_KUBECTL)"
fi

if command -v go >/dev/null 2>&1; then
	check_min go "$(first_version "$(go version 2>/dev/null)")" "$MIN_GO"
else
	fail go "not found (need >= $MIN_GO)"
fi

if [ -n "$failed" ]; then
	echo "Missing or too old:$failed. See docs/deployment/local-setup.md." >&2
	exit 1
fi
echo "All prerequisites satisfied."
