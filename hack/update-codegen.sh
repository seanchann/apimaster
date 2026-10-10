#!/usr/bin/env bash
# Generate the retained API marker closure with the repository's pinned tools.
set -euo pipefail
SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export GOTOOLCHAIN=go1.27.2
CODEGEN_BIN="${APIMASTER_CODEGEN_BIN_DIR:-/tmp/easehold-upstream-storage/apimaster-codegen/bin}"
"$SCRIPT_ROOT/hack/tools/install-dep.sh"
cd "$SCRIPT_ROOT"
BOILERPLATE="$SCRIPT_ROOT/hack/boilerplate.go.txt"
MODULE="github.com/seanchann/apimaster"
REPORT_DIR="${APIMASTER_CODEGEN_REPORT_DIR:-/tmp/easehold-upstream-storage/apimaster-codegen/reports}"
mkdir -p "$REPORT_DIR"
versioned=("$MODULE/pkg/apis/coreres/v1" "$MODULE/pkg/apis/rbac/v1")
abac=("$MODULE/pkg/apis/abac/v0" "$MODULE/pkg/apis/abac/v1beta1")
internal=("$MODULE/pkg/apis/coreres" "$MODULE/pkg/apis/rbac" "$MODULE/pkg/apis/abac")
common=(--go-header-file "$BOILERPLATE")
"$CODEGEN_BIN/deepcopy-gen" "${common[@]}" --output-file zz_generated.deepcopy.go "${versioned[@]}" "${abac[@]}" "${internal[@]}"
"$CODEGEN_BIN/defaulter-gen" "${common[@]}" --output-file zz_generated.defaults.go "${versioned[@]}" "${abac[@]}"
"$CODEGEN_BIN/conversion-gen" "${common[@]}" --output-file zz_generated.conversion.go "${versioned[@]}" "${abac[@]}"
"$CODEGEN_BIN/openapi-gen" "${common[@]}" --output-file zz_generated.openapi.go --output-dir "$SCRIPT_ROOT/pkg/generated/openapi" --output-pkg "$MODULE/pkg/generated/openapi" --report-filename "$REPORT_DIR/api_violations.report" "${versioned[@]}" "$MODULE/pkg/apis/abac/v1beta1" k8s.io/apimachinery/pkg/apis/meta/v1 k8s.io/apimachinery/pkg/version k8s.io/apimachinery/pkg/runtime k8s.io/apimachinery/pkg/util/intstr
"$CODEGEN_BIN/client-gen" "${common[@]}" --output-dir "$SCRIPT_ROOT/pkg/client" --output-pkg "$MODULE/pkg/client" --clientset-name clientset --input-base "$MODULE/pkg/apis" --input coreres/v1,rbac/v1
"$CODEGEN_BIN/lister-gen" "${common[@]}" --output-dir "$SCRIPT_ROOT/pkg/client/listers" --output-pkg "$MODULE/pkg/client/listers" "${versioned[@]}"
"$CODEGEN_BIN/informer-gen" "${common[@]}" --output-dir "$SCRIPT_ROOT/pkg/client/informers" --output-pkg "$MODULE/pkg/client/informers" --single-directory --versioned-clientset-package "$MODULE/pkg/client/clientset" --listers-package "$MODULE/pkg/client/listers" "${versioned[@]}"
# Generators prepend build constraints. Keep the Apache template first without changing generated bodies.
python3 - "$SCRIPT_ROOT" <<'PYHEADER'
import re
import sys
from pathlib import Path
root=Path(sys.argv[1])
for directory in ('pkg/apis','pkg/client','pkg/generated/openapi'):
    for file in (root/directory).rglob('*.go'):
        text=file.read_text()
        if 'Code generated' not in text:
            continue
        match=re.match(r'((?://(?:go:build| \+build)[^\n]*\n)+)\n(/\*.*?\*/)(.*)',text,re.S)
        if match:
            constraints='\n'.join(line for line in match[1].splitlines() if line.startswith('//go:build '))+'\n'
            text=match[2]+'\n\n'+constraints+'\n'+match[3].lstrip('\n')
            file.write_text(text)
        if not text.startswith('/*'):
            raise SystemExit(f'generated license must be first: {file}')
PYHEADER
