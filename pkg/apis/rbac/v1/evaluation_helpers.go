/********************************************************************
* Copyright (c) 2008 - 2024. Authors: seanchann <seandev@foxmail.com>
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*         http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************/

package v1

import (
	"fmt"
	"strings"
)

func VerbMatches(rule *PolicyRule, requestedVerb string) bool {
	for _, ruleVerb := range rule.Verbs {
		if ruleVerb == VerbAll {
			return true
		}
		if ruleVerb == requestedVerb {
			return true
		}
	}

	return false
}

func APIGroupMatches(rule *PolicyRule, requestedGroup string) bool {
	for _, ruleGroup := range rule.APIGroups {
		if ruleGroup == APIGroupAll {
			return true
		}
		if ruleGroup == requestedGroup {
			return true
		}
	}

	return false
}

func ResourceMatches(rule *PolicyRule, combinedRequestedResource, requestedSubresource string) bool {
	for _, ruleResource := range rule.Resources {
		// if everything is allowed, we match
		if ruleResource == ResourceAll {
			return true
		}
		// if we have an exact match, we match
		if ruleResource == combinedRequestedResource {
			return true
		}

		// We can also match a */subresource.
		// if there isn't a subresource, then continue
		if len(requestedSubresource) == 0 {
			continue
		}
		// if the rule isn't in the format */subresource, then we don't match, continue
		if len(ruleResource) == len(requestedSubresource)+2 &&
			strings.HasPrefix(ruleResource, "*/") &&
			strings.HasSuffix(ruleResource, requestedSubresource) {
			return true

		}
	}

	return false
}

func ResourceNameMatches(rule *PolicyRule, requestedName string) bool {
	if len(rule.ResourceNames) == 0 {
		return true
	}

	for _, ruleName := range rule.ResourceNames {
		if ruleName == requestedName {
			return true
		}
	}

	return false
}

func NonResourceURLMatches(rule *PolicyRule, requestedURL string) bool {
	for _, ruleURL := range rule.NonResourceURLs {
		if ruleURL == NonResourceAll {
			return true
		}
		if ruleURL == requestedURL {
			return true
		}
		if strings.HasSuffix(ruleURL, "*") && strings.HasPrefix(requestedURL, strings.TrimRight(ruleURL, "*")) {
			return true
		}
	}

	return false
}

// CompactString exposes a compact string representation for use in escalation error messages
func CompactString(r PolicyRule) string {
	formatStringParts := []string{}
	formatArgs := []interface{}{}
	if len(r.APIGroups) > 0 {
		formatStringParts = append(formatStringParts, "APIGroups:%q")
		formatArgs = append(formatArgs, r.APIGroups)
	}
	if len(r.Resources) > 0 {
		formatStringParts = append(formatStringParts, "Resources:%q")
		formatArgs = append(formatArgs, r.Resources)
	}
	if len(r.NonResourceURLs) > 0 {
		formatStringParts = append(formatStringParts, "NonResourceURLs:%q")
		formatArgs = append(formatArgs, r.NonResourceURLs)
	}
	if len(r.ResourceNames) > 0 {
		formatStringParts = append(formatStringParts, "ResourceNames:%q")
		formatArgs = append(formatArgs, r.ResourceNames)
	}
	if len(r.Verbs) > 0 {
		formatStringParts = append(formatStringParts, "Verbs:%q")
		formatArgs = append(formatArgs, r.Verbs)
	}
	formatString := "{" + strings.Join(formatStringParts, ", ") + "}"
	return fmt.Sprintf(formatString, formatArgs...)
}

func (r *PolicyRule) String() string {
	if r == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PolicyRule{`,
		`Verbs:` + fmt.Sprintf("%v", r.Verbs) + `,`,
		`APIGroups:` + fmt.Sprintf("%v", r.APIGroups) + `,`,
		`Resources:` + fmt.Sprintf("%v", r.Resources) + `,`,
		`ResourceNames:` + fmt.Sprintf("%v", r.ResourceNames) + `,`,
		`NonResourceURLs:` + fmt.Sprintf("%v", r.NonResourceURLs) + `,`,
		`}`,
	}, "")
	return s
}

type SortableRuleSlice []PolicyRule

func (s SortableRuleSlice) Len() int      { return len(s) }
func (s SortableRuleSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortableRuleSlice) Less(i, j int) bool {
	return strings.Compare(s[i].String(), s[j].String()) < 0
}
