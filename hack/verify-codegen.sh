#!/usr/bin/env bash
#********************************************************************
# Copyright (c) All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#*******************************************************************/

set -euo pipefail
SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export GOTOOLCHAIN=go1.27.2
TEMPORARY_PARENT="${APIMASTER_CODEGEN_TMPDIR:-/tmp/easehold-upstream-storage}"
mkdir -p "$TEMPORARY_PARENT"
VERIFY_ROOT="$(mktemp -d "$TEMPORARY_PARENT/apimaster-verify.XXXXXX")"
trap 'rm -rf "$VERIFY_ROOT"' EXIT
mkdir -p "$VERIFY_ROOT/am"
(cd "$SCRIPT_ROOT" && tar --exclude=.git -cf - .) | (cd "$VERIFY_ROOT/am" && tar -xf -)
# Preserve the caller's local server substitution, replacing only this module's workspace entry.
if [[ "${GOWORK:-off}" != off && -n "${GOWORK:-}" ]]; then
    cp "$GOWORK" "$VERIFY_ROOT/go.work"
    GOWORK="$VERIFY_ROOT/go.work" go work edit -dropuse="$SCRIPT_ROOT" -use="$VERIFY_ROOT/am"
    export GOWORK="$VERIFY_ROOT/go.work"
fi
"$VERIFY_ROOT/am/hack/update-codegen.sh"
for path in pkg/apis pkg/client pkg/generated/openapi; do
    diff -ru "$SCRIPT_ROOT/$path" "$VERIFY_ROOT/am/$path"
done
printf '%s\n' 'Code generation is reproducible.'
