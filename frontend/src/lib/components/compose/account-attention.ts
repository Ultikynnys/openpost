import type { components } from '$lib/api/types';

type ResolvedAccountCapability = components['schemas']['ResolvedAccountCapability'];
type ValidationIssue = components['schemas']['ValidationIssue'];

const informationalIssueCodes = new Set(['quota_warning']);

export function isActionableAccountIssue(issue: ValidationIssue): boolean {
	return !informationalIssueCodes.has(issue.code);
}

export function accountCapabilityNeedsAttention(capability: ResolvedAccountCapability): boolean {
	return !capability.compatible || (capability.issues ?? []).some(isActionableAccountIssue);
}
