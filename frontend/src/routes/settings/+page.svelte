<script lang="ts">
	import { page } from '$app/state';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import PageContainer from '$lib/components/page-container.svelte';
	import ProfileAvatarUploader from '$lib/components/profile-avatar-uploader.svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';
	import { getApiBase } from '$lib/stores/instance.svelte';
	import { createPasskeyCredential } from '$lib/auth/webauthn';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import SettingsIcon from 'lucide-svelte/icons/settings';
	import SaveIcon from 'lucide-svelte/icons/save';
	import XIcon from 'lucide-svelte/icons/x';
	import ClockIcon from 'lucide-svelte/icons/clock';
	import ImageIcon from 'lucide-svelte/icons/image';
	import CalendarIcon from 'lucide-svelte/icons/calendar';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import TrashIcon from 'lucide-svelte/icons/trash';
	import SparklesIcon from 'lucide-svelte/icons/sparkles';
	import ShieldCheckIcon from 'lucide-svelte/icons/shield-check';
	import SmartphoneIcon from 'lucide-svelte/icons/smartphone';
	import KeyRoundIcon from 'lucide-svelte/icons/key-round';
	import TerminalIcon from 'lucide-svelte/icons/terminal';
	import CreditCardIcon from 'lucide-svelte/icons/credit-card';
	import ExternalLinkIcon from 'lucide-svelte/icons/external-link';
	import ActivityIcon from 'lucide-svelte/icons/activity';
	import UsersIcon from 'lucide-svelte/icons/users';
	import UserPlusIcon from 'lucide-svelte/icons/user-plus';
	import CopyIcon from 'lucide-svelte/icons/copy';
	import MonitorIcon from 'lucide-svelte/icons/monitor';
	import LogOutIcon from 'lucide-svelte/icons/log-out';
	import CameraIcon from 'lucide-svelte/icons/camera';
	import AlertCircleIcon from 'lucide-svelte/icons/alert-circle';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { client } from '$lib/api/client';
	import { getLocaleTag } from '$lib/i18n';
	import { hostedPlanFromSearchParams } from '$lib/billing';
	import { m } from '$lib/paraglide/messages';
	import {
		apiTokenScopeOptions,
		billingMetricLabels,
		billingPlans,
		cleanupDaysOptions,
		getTimezoneLabel,
		inviteRoleOptions,
		timezones,
		type APITokenSummary,
		type AuthSessionSummary,
		type BillingStatus,
		type MCPActivityItem,
		type PostingSchedule,
		type ScheduleRow,
		type SecurityStatus,
		type WorkspaceInvitation,
		type WorkspaceTeam
	} from './settings-data';

	const groupedTimezones = $derived.by(() => {
		const groups: Record<string, typeof timezones> = {};
		for (const tz of timezones) {
			if (!groups[tz.group]) groups[tz.group] = [];
			groups[tz.group].push(tz);
		}
		return groups;
	});

	let saving = $state(false);
	let toastMessage = $state('');
	let profileDisplayName = $state('');
	let profileBusy = $state(false);
	let profileError = $state('');
	let avatarUploaderOpen = $state(false);
	let loadingSecurity = $state(true);
	let securityBusy = $state(false);
	let securityError = $state('');
	let authSessions = $state.raw<AuthSessionSummary[]>([]);
	let authSessionsLoading = $state(true);
	let authSessionsError = $state('');
	let authSessionBusyID = $state('');
	let currentPassword = $state('');
	let totpSetupChallengeId = $state('');
	let totpManualEntryKey = $state('');
	let totpQRCodeDataURL = $state('');
	let totpCode = $state('');
	let newPasskeyName = $state('');

	let securityStatus = $state<SecurityStatus | null>(null);
	let apiTokens = $state<APITokenSummary[]>([]);
	let apiTokensLoading = $state(true);
	let mcpActivity = $state.raw<MCPActivityItem[]>([]);
	let mcpActivityLoading = $state(true);
	let mcpActivityError = $state('');
	let apiTokenBusy = $state(false);
	let apiTokenName = $state('OpenPost MCP');
	let apiTokenScope = $state('mcp:full');
	let apiTokenWorkspaceScope = $state('current');
	let createdAPIToken = $state('');
	let billingBusyPlan = $state('');
	let billingPortalBusy = $state(false);
	let billingError = $state('');
	let billingStatusLoading = $state(false);
	let billingStatus = $state<BillingStatus | null>(null);
	let handledCheckoutPlan = '';
	let teamLoading = $state(false);
	let teamBusy = $state(false);
	let teamError = $state('');
	let workspaceTeam = $state<WorkspaceTeam | null>(null);
	let inviteEmail = $state('');
	let inviteRole = $state<'viewer' | 'editor' | 'admin'>('editor');
	let createdInviteURL = $state('');

	const authState = $derived($auth);
	const currentOrganizationID = $derived(workspaceCtx.currentWorkspace?.organization_id ?? '');
	const weekdayFormatter = $derived(
		new Intl.DateTimeFormat(getLocaleTag(), { weekday: 'short', timeZone: 'UTC' })
	);
	const longWeekdayFormatter = $derived(
		new Intl.DateTimeFormat(getLocaleTag(), { weekday: 'long', timeZone: 'UTC' })
	);
	const passkeyCount = $derived((securityStatus?.passkeys ?? []).length);
	const teamMembers = $derived(workspaceTeam?.members ?? []);
	const pendingInvitations = $derived(workspaceTeam?.invitations ?? []);
	const currentTeamSeats = $derived(workspaceTeam?.current_seats ?? 0);
	const selectedInviteRole = $derived(
		inviteRoleOptions.find((option) => option.value === inviteRole) ?? inviteRoleOptions[0]
	);
	const selectedAPITokenScope = $derived(
		apiTokenScopeOptions.find((option) => option.value === apiTokenScope) ?? apiTokenScopeOptions[0]
	);
	const apiTokenWorkspaceOptions = $derived([
		{
			value: 'current',
			label: m.settings_current_workspace(),
			description: workspaceCtx.currentWorkspace?.name ?? m.settings_current_workspace_body()
		},
		{
			value: 'all',
			label: m.settings_all_workspaces(),
			description: m.settings_all_workspaces_body()
		}
	]);
	const selectedAPITokenWorkspaceScope = $derived(
		apiTokenWorkspaceOptions.find((option) => option.value === apiTokenWorkspaceScope) ??
			apiTokenWorkspaceOptions[0]
	);
	const settingsTabs = $derived([
		{ id: 'profile', label: m.settings_profile() },
		{ id: 'security', label: m.settings_security() },
		{ id: 'developer', label: m.settings_developer() },
		{ id: 'general', label: m.settings_general() },
		{ id: 'schedule', label: m.settings_schedule() },
		{ id: 'media', label: m.settings_media() },
		{ id: 'members', label: m.settings_members() },
		{ id: 'plan', label: m.settings_plan() }
	] as const);
	const activeSettingsTab = $derived.by(() =>
		normalizeSettingsTab(
			page.url.searchParams.get('tab') || page.url.hash.replace(/^#/, '') || null
		)
	);
	const profileEmail = $derived(authState.user?.email ?? '');
	const profileAvatarURL = $derived(authState.user?.avatar_url ?? '');
	const profileInitials = $derived.by(() => {
		const source = profileDisplayName || profileEmail || 'OP';
		const parts = source
			.replace(/@.*/, '')
			.split(/[\s._-]+/)
			.filter(Boolean);
		return (parts[0]?.[0] ?? 'O').toUpperCase() + (parts[1]?.[0] ?? '').toUpperCase();
	});
	const currentBillingPlan = $derived(
		billingPlans.find((plan) => plan.id === billingStatus?.plan_id) ?? null
	);
	const hasActiveBillingPlan = $derived(
		Boolean(
			billingStatus?.plan_id &&
			['active', 'trialing'].includes((billingStatus.status ?? '').toLowerCase())
		)
	);
	const activeSettingsTitle = $derived.by(() => {
		if (activeSettingsTab === 'profile') return m.settings_profile();
		if (activeSettingsTab === 'security') return m.settings_security();
		if (activeSettingsTab === 'developer') return m.settings_developer();
		if (activeSettingsTab === 'members') return m.settings_team_members();
		if (activeSettingsTab === 'plan') return m.settings_plan();
		if (activeSettingsTab === 'schedule') return m.settings_schedule();
		if (activeSettingsTab === 'media') return m.settings_media();
		return m.settings_general();
	});
	const activeSettingsDescription = $derived.by(() => {
		if (activeSettingsTab === 'profile') return m.settings_profile_description();
		if (activeSettingsTab === 'security') return m.settings_security_description();
		if (activeSettingsTab === 'developer') return m.settings_developer_description();
		if (activeSettingsTab === 'members') return m.settings_members_description();
		if (activeSettingsTab === 'plan') return m.settings_plan_description();
		if (activeSettingsTab === 'schedule') return m.settings_schedule_description();
		if (activeSettingsTab === 'media') return m.settings_media_description();
		return m.settings_general_description({
			workspace: workspaceCtx.currentWorkspace?.name || m.settings_workspace()
		});
	});
	const requestedBillingPlan = $derived.by(() => {
		const planID = hostedPlanFromSearchParams(page.url.searchParams);
		return billingPlans.some((plan) => plan.id === planID) ? planID : '';
	});
	const monthlyBillingUsageRows = $derived.by(() => {
		if (!billingStatus) return [];
		return Object.entries(billingStatus.limits)
			.filter(([metric]) => metric.endsWith('_monthly'))
			.map(([metric, limit]) => ({
				metric,
				label: billingMetricLabels[metric] ?? metric.replaceAll('_', ' '),
				current: billingStatus?.usage[metric] ?? 0,
				limit
			}));
	});
	function isSettingsTab(value: string): value is (typeof settingsTabs)[number]['id'] {
		return settingsTabs.some((tab) => tab.id === value);
	}

	function localizedWeekday(dayIndex: number, format: 'short' | 'long' = 'short') {
		const date = new Date(Date.UTC(2026, 6, 5 + dayIndex));
		return (format === 'long' ? longWeekdayFormatter : weekdayFormatter).format(date);
	}

	function normalizeSettingsTab(value: string | null) {
		if (value === 'billing' || value === 'organization') return 'plan';
		if (value === 'team') return 'members';
		if (value === 'tokens' || value === 'account')
			return value === 'tokens' ? 'developer' : 'profile';
		if (value === 'workspace' || value === 'social-accounts') return 'general';
		return value && isSettingsTab(value) ? value : 'general';
	}

	function openSettingsTab(tab: (typeof settingsTabs)[number]['id']) {
		goto(resolve(`/settings?tab=${tab}` as '/settings'));
	}

	async function saveProfile(event: SubmitEvent) {
		event.preventDefault();
		profileBusy = true;
		profileError = '';
		try {
			const { data, error: err } = await client.PATCH('/auth/profile', {
				body: { display_name: profileDisplayName }
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			auth.setUser(data);
			profileDisplayName = data.display_name ?? '';
			toastMessage = m.settings_profile_updated();
		} catch (e) {
			profileError = (e as Error).message;
		} finally {
			profileBusy = false;
		}
	}

	function handleAvatarUploaded(avatarURL: string) {
		if (authState.user) {
			auth.setUser({ ...authState.user, avatar_url: avatarURL });
		}
		toastMessage = m.settings_picture_updated();
	}

	async function removeAvatar() {
		if (!profileAvatarURL) return;
		profileBusy = true;
		profileError = '';
		try {
			const { error: err } = await client.DELETE('/auth/profile/avatar', {});
			if (err) throw new Error(err.detail || m.settings_action_failed());
			if (authState.user) {
				auth.setUser({ ...authState.user, avatar_url: '' });
			}
			toastMessage = m.settings_picture_removed();
		} catch (e) {
			profileError = (e as Error).message;
		} finally {
			profileBusy = false;
		}
	}

	async function loadSecurityStatus() {
		loadingSecurity = true;
		securityError = '';
		try {
			const { data, error: err } = await client.GET('/auth/security');
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			securityStatus = data;
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			loadingSecurity = false;
		}
	}

	async function loadAuthSessions() {
		authSessionsLoading = true;
		authSessionsError = '';
		try {
			const { data, error: err } = await client.GET('/auth/sessions');
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			authSessions = data as AuthSessionSummary[];
		} catch (e) {
			authSessions = [];
			authSessionsError = (e as Error).message;
		} finally {
			authSessionsLoading = false;
		}
	}

	async function loadAPITokens() {
		apiTokensLoading = true;
		securityError = '';
		try {
			const { data, error: err } = await client.GET('/api-tokens');
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			apiTokens = data as APITokenSummary[];
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			apiTokensLoading = false;
		}
	}

	async function loadMCPActivity() {
		mcpActivityLoading = true;
		mcpActivityError = '';
		try {
			const { data, error: err } = await client.GET('/mcp/activity', {
				params: { query: { limit: 8 } }
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			mcpActivity = data as MCPActivityItem[];
		} catch (e) {
			mcpActivityError = (e as Error).message;
			mcpActivity = [];
		} finally {
			mcpActivityLoading = false;
		}
	}

	async function loadWorkspaceTeam() {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		teamLoading = true;
		teamError = '';
		try {
			const { data, error: err } = await client.GET('/workspaces/{id}/team', {
				params: { path: { id: workspaceID } }
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			workspaceTeam = data as WorkspaceTeam;
		} catch (e) {
			workspaceTeam = null;
			teamError = (e as Error).message;
		} finally {
			teamLoading = false;
		}
	}

	async function createWorkspaceInvitation(event: SubmitEvent) {
		event.preventDefault();
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		teamBusy = true;
		teamError = '';
		createdInviteURL = '';
		try {
			const { data, error: err } = await client.POST('/workspaces/{id}/invitations', {
				params: { path: { id: workspaceID } },
				body: {
					email: inviteEmail.trim(),
					role: inviteRole
				}
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			const invitation = data as WorkspaceInvitation;
			createdInviteURL =
				invitation.accept_url ||
				(invitation.token ? `${window.location.origin}/invite?token=${invitation.token}` : '');
			inviteEmail = '';
			inviteRole = 'editor';
			await loadWorkspaceTeam();
			toastMessage = m.settings_invite_created();
		} catch (e) {
			teamError = (e as Error).message;
		} finally {
			teamBusy = false;
		}
	}

	async function revokeWorkspaceInvitation(invitationID: string) {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		if (!confirm('Revoke this workspace invitation? The invite link will stop working.')) return;
		teamBusy = true;
		teamError = '';
		try {
			const { error: err } = await client.DELETE('/workspaces/{id}/invitations/{invitation_id}', {
				params: { path: { id: workspaceID, invitation_id: invitationID } }
			});
			if (err) throw new Error(err.detail || m.settings_action_failed());
			await loadWorkspaceTeam();
			toastMessage = m.settings_invitation_revoked();
		} catch (e) {
			teamError = (e as Error).message;
		} finally {
			teamBusy = false;
		}
	}

	async function copyCreatedInviteURL() {
		if (!createdInviteURL) return;
		await navigator.clipboard.writeText(createdInviteURL);
		toastMessage = m.settings_invite_copied();
	}

	async function createAPIToken() {
		apiTokenBusy = true;
		securityError = '';
		createdAPIToken = '';
		const fallbackName = apiTokenScope === 'mcp:full' ? 'OpenPost MCP' : 'OpenPost CLI';
		const workspaceID =
			apiTokenWorkspaceScope === 'current' ? (workspaceCtx.currentWorkspace?.id ?? '') : '';
		try {
			const { data, error: err } = await client.POST('/api-tokens', {
				body: {
					name: apiTokenName.trim() || fallbackName,
					scope: apiTokenScope,
					...(workspaceID ? { workspace_id: workspaceID } : {})
				}
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			createdAPIToken = data.token;
			apiTokenName = fallbackName;
			await loadAPITokens();
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			apiTokenBusy = false;
		}
	}

	async function revokeAuthSession(session: AuthSessionSummary) {
		const prompt = session.current
			? 'Sign out of this browser? You will need to log in again.'
			: 'Revoke this browser session?';
		if (!confirm(prompt)) return;
		authSessionBusyID = session.id;
		authSessionsError = '';
		try {
			const { data, error: err } = await client.DELETE('/auth/sessions/{session_id}', {
				params: { path: { session_id: session.id } }
			});
			if (err) throw new Error(err.detail || m.settings_action_failed());
			if (data?.revoked_current || session.current) {
				await auth.logout();
				await goto(resolve('/login'));
				return;
			}
			await loadAuthSessions();
			toastMessage = m.settings_session_revoked();
		} catch (e) {
			authSessionsError = (e as Error).message;
		} finally {
			authSessionBusyID = '';
		}
	}

	async function revokeAPIToken(tokenID: string) {
		if (!confirm('Revoke this API token? Any CLI or automation using it will stop working.'))
			return;
		apiTokenBusy = true;
		securityError = '';
		try {
			const { error: err } = await client.DELETE('/api-tokens/{id}', {
				params: { path: { id: tokenID } }
			});
			if (err) throw new Error(err.detail || m.settings_action_failed());
			await loadAPITokens();
			toastMessage = m.settings_token_revoked();
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			apiTokenBusy = false;
		}
	}

	async function loadBillingStatus() {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		billingStatusLoading = true;
		billingError = '';
		try {
			const { data, error: err } = currentOrganizationID
				? await client.GET('/organizations/{id}/billing/status', {
						params: { path: { id: currentOrganizationID } }
					})
				: await client.GET('/billing/status', {
						params: { query: { workspace_id: workspaceID } }
					});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			billingStatus = data as BillingStatus;
		} catch (e) {
			billingStatus = null;
			billingError = (e as Error).message;
		} finally {
			billingStatusLoading = false;
		}
	}

	async function startCheckout(planID: string) {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		billingBusyPlan = planID;
		billingError = '';
		try {
			const { data, error: err } = currentOrganizationID
				? await client.POST('/organizations/{id}/billing/checkout', {
						params: { path: { id: currentOrganizationID } },
						body: { plan_id: planID }
					})
				: await client.POST('/billing/checkout', {
						body: { workspace_id: workspaceID, plan_id: planID }
					});
			if (err || !data?.url) throw new Error(err?.detail || m.settings_action_failed());
			window.location.assign(data.url);
		} catch (e) {
			billingError = (e as Error).message;
		} finally {
			billingBusyPlan = '';
		}
	}

	async function openBillingPortal() {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		billingPortalBusy = true;
		billingError = '';
		try {
			const { data, error: err } = currentOrganizationID
				? await client.POST('/organizations/{id}/billing/portal', {
						params: { path: { id: currentOrganizationID } }
					})
				: await client.POST('/billing/portal', {
						body: { workspace_id: workspaceID }
					});
			if (err || !data?.url) throw new Error(err?.detail || m.settings_action_failed());
			window.location.assign(data.url);
		} catch (e) {
			billingError = (e as Error).message;
		} finally {
			billingPortalBusy = false;
		}
	}

	async function startTOTPSetup() {
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await client.POST('/auth/security/totp/setup', {
				body: { current_password: currentPassword }
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			totpSetupChallengeId = data.challenge_id;
			totpManualEntryKey = data.manual_entry_key;
			totpQRCodeDataURL = data.qr_code_data_url;
			totpCode = '';
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function confirmTOTPSetup() {
		if (!totpSetupChallengeId) return;
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await client.POST('/auth/security/totp/confirm', {
				body: {
					challenge_id: totpSetupChallengeId,
					code: totpCode
				}
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			securityStatus = data;
			totpSetupChallengeId = '';
			totpManualEntryKey = '';
			totpQRCodeDataURL = '';
			totpCode = '';
			currentPassword = '';
			toastMessage = m.settings_authenticator_enabled_notice();
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function disableTOTP() {
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await client.POST('/auth/security/totp/disable', {
				body: { current_password: currentPassword }
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			securityStatus = data;
			currentPassword = '';
			toastMessage = m.settings_authenticator_disabled_notice();
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function addPasskey() {
		securityBusy = true;
		securityError = '';
		try {
			const { data: beginData, error: beginError } = await client.POST(
				'/auth/security/passkeys/begin',
				{
					body: {
						current_password: currentPassword,
						name: newPasskeyName
					}
				}
			);
			if (beginError || !beginData) {
				throw new Error(beginError?.detail || m.settings_action_failed());
			}

			const credential = await createPasskeyCredential(beginData.options);
			const { data, error: err } = await client.POST('/auth/security/passkeys/finish', {
				body: {
					challenge_id: beginData.challenge_id,
					name: newPasskeyName,
					credential
				}
			});
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			securityStatus = data;
			currentPassword = '';
			newPasskeyName = '';
			toastMessage = m.settings_passkey_added();
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function removePasskey(passkeyId: string) {
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await client.POST(
				'/auth/security/passkeys/{passkey_id}/remove',
				{
					params: { path: { passkey_id: passkeyId } },
					body: { current_password: currentPassword }
				}
			);
			if (err || !data) throw new Error(err?.detail || m.settings_action_failed());
			securityStatus = data;
			currentPassword = '';
			toastMessage = m.settings_passkey_removed();
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function saveSettings() {
		saving = true;
		try {
			await workspaceCtx.saveSettings({
				avatar_url: workspaceCtx.settings.avatar_url,
				timezone: workspaceCtx.settings.timezone,
				week_start: workspaceCtx.settings.week_start,
				media_cleanup_days: workspaceCtx.settings.media_cleanup_days,
				random_delay_minutes: workspaceCtx.settings.random_delay_minutes,
				draft_gap_minutes: workspaceCtx.settings.draft_gap_minutes,
				slot_start_hour: workspaceCtx.settings.slot_start_hour,
				slot_end_hour: workspaceCtx.settings.slot_end_hour,
				slot_interval_minutes: workspaceCtx.settings.slot_interval_minutes
			});
			toastMessage = m.settings_saved();
		} catch (e) {
			toastMessage = (e as Error).message;
		} finally {
			saving = false;
		}
	}

	function parseDurationInput(input: string, allowZero: boolean = false): number | null {
		input = input.trim().toLowerCase();
		const direct = parseInt(input, 10);
		if (!isNaN(direct) && String(direct) === input && (direct > 0 || (allowZero && direct === 0))) {
			return direct;
		}
		const hourMatch = input.match(/(\d+)\s*h/);
		const minMatch = input.match(/(\d+)\s*m/);
		let total = 0;
		if (hourMatch) total += parseInt(hourMatch[1], 10) * 60;
		if (minMatch) total += parseInt(minMatch[1], 10);
		if (total > 0) return total;
		return null;
	}

	let intervalInput = $state(String(workspaceCtx.settings.slot_interval_minutes));
	let intervalError = $state('');
	let draftGapInput = $state(String(workspaceCtx.settings.draft_gap_minutes));
	let draftGapError = $state('');

	function handleIntervalChange(value: string) {
		intervalInput = value;
		const parsed = parseDurationInput(value);
		if (parsed !== null && parsed >= 1 && parsed <= 180) {
			intervalError = '';
			workspaceCtx.settings.slot_interval_minutes = parsed;
		} else if (value.trim() !== '') {
			intervalError = m.settings_interval_invalid();
		}
	}

	function handleDraftGapChange(value: string) {
		draftGapInput = value;
		const parsed = parseDurationInput(value, true);
		if (parsed !== null && parsed >= 0 && parsed <= 24 * 60) {
			draftGapError = '';
			workspaceCtx.settings.draft_gap_minutes = parsed;
		} else if (value.trim() !== '') {
			draftGapError = m.settings_draft_gap_invalid();
		}
	}

	let schedules = $state<PostingSchedule[]>([]);
	let loadingSchedules = $state(false);
	let showSuggestSchedule = $state(false);
	let suggestedPostsPerDay = $state(3);
	let generatingSchedule = $state(false);
	let newTimeInput = $state('09:00');
	let newTimeError = $state('');
	let newTimeDays = $state<number[]>([1, 2, 3, 4, 5]);

	const dayOrder = $derived.by(() => {
		const start = workspaceCtx.settings.week_start === 0 ? 0 : 1;
		return Array.from({ length: 7 }, (_, index) => (start + index) % 7);
	});

	const scheduleRows = $derived.by(() => {
		const rows: Record<string, ScheduleRow> = {};
		for (const schedule of schedules) {
			const key = `${schedule.local_hour}:${schedule.local_minute}`;
			if (!rows[key]) {
				rows[key] = {
					key,
					local_hour: schedule.local_hour,
					local_minute: schedule.local_minute,
					label: schedule.label ?? '',
					days: {}
				};
			}
			const row = rows[key];
			row.days[schedule.local_day_of_week] = schedule;
			if (!row.label && schedule.label) {
				row.label = schedule.label;
			}
		}
		return Object.values(rows).sort(
			(a, b) => a.local_hour * 60 + a.local_minute - (b.local_hour * 60 + b.local_minute)
		);
	});

	async function loadSchedules() {
		if (!workspaceCtx.currentWorkspace) return;
		loadingSchedules = true;
		try {
			const { data, error: err } = await client.GET('/posting-schedules', {
				params: { query: { workspace_id: workspaceCtx.currentWorkspace.id } }
			});
			if (!err && data) {
				schedules = data;
			}
		} catch (e) {
			console.error('Failed to load schedules:', e);
		} finally {
			loadingSchedules = false;
		}
	}

	function parseClockInput(value: string): { hour: number; minute: number } | null {
		const match = value.trim().match(/^(\d{1,2}):(\d{2})$/);
		if (!match) return null;
		const hour = Number(match[1]);
		const minute = Number(match[2]);
		if (hour < 0 || hour > 23 || minute < 0 || minute > 59) return null;
		return { hour, minute };
	}

	async function createSchedule(dayOfWeek: number, localHour: number, localMinute: number) {
		if (!workspaceCtx.currentWorkspace) return;
		const { error: err } = await client.POST('/posting-schedules', {
			body: {
				workspace_id: workspaceCtx.currentWorkspace.id,
				local_day_of_week: dayOfWeek,
				local_hour: localHour,
				local_minute: localMinute,
				day_of_week: 0,
				utc_hour: 0,
				utc_minute: 0,
				label: ''
			}
		});
		if (err) throw err;
	}

	async function addTimeRow() {
		const parsed = parseClockInput(newTimeInput);
		if (!parsed) {
			newTimeError = 'Use HH:MM in 24-hour format.';
			return;
		}
		if (newTimeDays.length === 0) {
			newTimeError = 'Select at least one day.';
			return;
		}
		newTimeError = '';
		try {
			for (const day of newTimeDays) {
				const exists = schedules.some(
					(schedule) =>
						schedule.local_day_of_week === day &&
						schedule.local_hour === parsed.hour &&
						schedule.local_minute === parsed.minute
				);
				if (!exists) {
					await createSchedule(day, parsed.hour, parsed.minute);
				}
			}
			await loadSchedules();
			toastMessage = m.settings_time_added();
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to add schedule row';
		}
	}

	async function deleteSchedule(id: string) {
		try {
			const { error: err } = await client.DELETE('/posting-schedules/{id}', {
				params: { path: { id } }
			});
			if (err) throw err;
			await loadSchedules();
			toastMessage = m.settings_schedule_deleted();
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to delete schedule';
		}
	}

	async function toggleScheduleCell(row: ScheduleRow, dayOfWeek: number) {
		try {
			const existing = row.days[dayOfWeek];
			if (existing) {
				await deleteSchedule(existing.id);
				return;
			}
			await createSchedule(dayOfWeek, row.local_hour, row.local_minute);
			await loadSchedules();
			toastMessage = m.settings_schedule_updated();
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to update schedule';
		}
	}

	async function removeTimeRow(row: ScheduleRow) {
		try {
			for (const schedule of Object.values(row.days)) {
				if (schedule) {
					const { error: err } = await client.DELETE('/posting-schedules/{id}', {
						params: { path: { id: schedule.id } }
					});
					if (err) throw err;
				}
			}
			await loadSchedules();
			toastMessage = m.settings_time_removed();
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to remove schedule row';
		}
	}

	function toggleNewDay(dayOfWeek: number) {
		if (newTimeDays.includes(dayOfWeek)) {
			newTimeDays = newTimeDays.filter((value) => value !== dayOfWeek);
			return;
		}
		newTimeDays = [...newTimeDays, dayOfWeek].sort((a, b) => a - b);
	}

	async function generateSuggestedSchedule() {
		if (!workspaceCtx.currentWorkspace) return;
		generatingSchedule = true;
		try {
			const { error: err } = await client.POST('/posting-schedules/suggest', {
				body: {
					workspace_id: workspaceCtx.currentWorkspace.id,
					posts_per_day: suggestedPostsPerDay
				}
			});
			if (err) throw err;
			showSuggestSchedule = false;
			await loadSchedules();
			toastMessage = `Generated suggested schedule with ${suggestedPostsPerDay} posts per day`;
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to generate schedule';
		} finally {
			generatingSchedule = false;
		}
	}

	function formatTime(hour: number, minute: number): string {
		return new Date(Date.UTC(2024, 0, 1, hour, minute)).toLocaleTimeString(getLocaleTag(), {
			hour: 'numeric',
			minute: '2-digit',
			timeZone: 'UTC'
		});
	}

	function formatBillingValue(metric: string, value: number): string {
		if (metric.includes('bytes')) {
			return formatBytes(value);
		}
		return new Intl.NumberFormat(getLocaleTag()).format(value);
	}

	function formatPlanPrice(amount: number): string {
		return new Intl.NumberFormat(getLocaleTag(), {
			style: 'currency',
			currency: 'EUR',
			maximumFractionDigits: 0
		}).format(amount);
	}

	function formatBytes(value: number): string {
		if (value >= 1_000_000_000) {
			return `${(value / 1_000_000_000).toFixed(value % 1_000_000_000 === 0 ? 0 : 1)} GB`;
		}
		if (value >= 1_000_000) {
			return `${(value / 1_000_000).toFixed(value % 1_000_000 === 0 ? 0 : 1)} MB`;
		}
		return `${new Intl.NumberFormat(getLocaleTag()).format(value)} B`;
	}

	function formatSessionUserAgent(value: string): string {
		const trimmed = value.trim();
		if (!trimmed) return 'Unknown browser';

		const browser = sessionBrowserName(trimmed);
		const device = sessionDeviceName(trimmed);
		return `${browser} on ${device}`;
	}

	function sessionBrowserName(userAgent: string): string {
		if (/Edg\//.test(userAgent)) return 'Edge';
		if (/OPR\//.test(userAgent) || /Opera\//.test(userAgent)) return 'Opera';
		if (/Firefox\//.test(userAgent)) return 'Firefox';
		if (/CriOS\//.test(userAgent)) return 'Chrome';
		if (/Chrome\//.test(userAgent) || /Chromium\//.test(userAgent)) return 'Chrome';
		if (/Safari\//.test(userAgent)) return 'Safari';
		return 'Browser';
	}

	function sessionDeviceName(userAgent: string): string {
		if (/iPad/.test(userAgent)) return 'iPad';
		if (/iPhone/.test(userAgent)) return 'iPhone';
		if (/Android/.test(userAgent))
			return /Mobile/.test(userAgent) ? 'Android phone' : 'Android tablet';
		if (/Macintosh|Mac OS X/.test(userAgent)) return 'MacBook';
		if (/Windows NT/.test(userAgent)) return 'Windows';
		if (/Linux/.test(userAgent)) return 'Linux';
		return 'device';
	}

	function formatSessionTime(value: string): string {
		if (!value || value.startsWith('0001-01-01')) return 'Never';
		return new Date(value).toLocaleString();
	}

	let lastProfileUserID = $state('');
	$effect(() => {
		const user = authState.user;
		if (user?.id && user.id !== lastProfileUserID) {
			lastProfileUserID = user.id;
			profileDisplayName = user.display_name || '';
		}
	});

	$effect(() => {
		if (workspaceCtx.currentWorkspace) {
			loadBillingStatus();
			loadWorkspaceTeam();
			loadSchedules();
		}
	});

	$effect(() => {
		if (
			requestedBillingPlan &&
			workspaceCtx.currentWorkspace &&
			handledCheckoutPlan !== requestedBillingPlan &&
			!billingBusyPlan
		) {
			handledCheckoutPlan = requestedBillingPlan;
			startCheckout(requestedBillingPlan);
		}
	});

	$effect(() => {
		if (authState.isAuthenticated) {
			loadSecurityStatus();
			loadAuthSessions();
			loadAPITokens();
			loadMCPActivity();
		}
	});

	$effect(() => {
		intervalInput = String(workspaceCtx.settings.slot_interval_minutes);
		draftGapInput = String(workspaceCtx.settings.draft_gap_minutes);
	});

	function keepActiveSettingsTabVisible(settingsNavigation: HTMLElement) {
		const activeTab = activeSettingsTab;
		const activeButton = settingsNavigation.querySelector<HTMLElement>(
			`[data-settings-tab="${activeTab}"]`
		);
		if (!activeButton) return;
		const navigationRect = settingsNavigation.getBoundingClientRect();
		const activeButtonRect = activeButton.getBoundingClientRect();
		const centeredOffset =
			activeButtonRect.left -
			navigationRect.left -
			(navigationRect.width - activeButtonRect.width) / 2;
		settingsNavigation.scrollLeft += centeredOffset;
	}

	function handleTimezoneChange(value: string) {
		workspaceCtx.settings.timezone = value;
	}

	function handleWeekStartChange(value: number) {
		workspaceCtx.settings.week_start = value;
	}

	function handleCleanupDaysChange(value: number) {
		workspaceCtx.settings.media_cleanup_days = value;
	}
</script>

<svelte:head>
	<title>{m.settings_page_title()}</title>
</svelte:head>

{#if toastMessage}
	<div
		class="pointer-events-auto fixed right-4 bottom-4 z-50 mb-4 flex items-center gap-2 rounded-lg border bg-background px-4 py-3 shadow-lg"
	>
		<span class="text-sm" role="status" aria-live="polite">{toastMessage}</span>
		<Button
			variant="ghost"
			size="icon-sm"
			class="-my-1 -mr-2"
			onclick={() => (toastMessage = '')}
			aria-label={m.common_dismiss_notification()}
		>
			<XIcon class="size-4" />
		</Button>
	</div>
{/if}

<PageContainer
	title={activeSettingsTitle}
	description={activeSettingsDescription}
	icon={SettingsIcon}
	loading={!workspaceCtx.currentWorkspace}
	loadingMessage={m.settings_loading_workspace()}
>
	<div class="grid min-w-0 items-start gap-8 lg:grid-cols-[13rem_minmax(0,1fr)]">
		<aside class="min-w-0 lg:sticky lg:top-6">
			<nav
				{@attach keepActiveSettingsTabVisible}
				class="flex max-w-full gap-2 overflow-x-auto pb-2 lg:flex-col lg:overflow-visible lg:pb-0"
				aria-label={m.settings_nav_label()}
			>
				<p
					class="hidden px-2 text-[11px] font-medium tracking-[0.12em] text-muted-foreground uppercase lg:block"
				>
					{m.settings_personal()}
				</p>
				{#each settingsTabs.slice(0, 3) as tab (tab.id)}
					<button
						type="button"
						data-settings-tab={tab.id}
						class={[
							'min-h-10 shrink-0 rounded-md px-3 py-2 text-left text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none lg:w-full',
							activeSettingsTab === tab.id
								? 'bg-accent text-foreground'
								: 'text-muted-foreground hover:bg-accent/60 hover:text-foreground'
						]}
						onclick={() => openSettingsTab(tab.id)}
						aria-current={activeSettingsTab === tab.id ? 'page' : undefined}
					>
						{tab.label}
					</button>
				{/each}
				<p
					class="hidden px-2 pt-4 text-[11px] font-medium tracking-[0.12em] text-muted-foreground uppercase lg:block"
				>
					{m.settings_workspace()}
				</p>
				{#each settingsTabs.slice(3, 6) as tab (tab.id)}
					<button
						type="button"
						data-settings-tab={tab.id}
						class={[
							'min-h-10 shrink-0 rounded-md px-3 py-2 text-left text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none lg:w-full',
							activeSettingsTab === tab.id
								? 'bg-accent text-foreground'
								: 'text-muted-foreground hover:bg-accent/60 hover:text-foreground'
						]}
						onclick={() => openSettingsTab(tab.id)}
						aria-current={activeSettingsTab === tab.id ? 'page' : undefined}
					>
						{tab.label}
					</button>
				{/each}
				<button
					type="button"
					class="min-h-10 shrink-0 rounded-md px-3 py-2 text-left text-sm font-medium text-muted-foreground transition-colors hover:bg-accent/60 hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none lg:w-full"
					onclick={() => goto(resolve('/accounts'))}
				>
					{m.accounts_heading()}
				</button>
				<p
					class="hidden px-2 pt-4 text-[11px] font-medium tracking-[0.12em] text-muted-foreground uppercase lg:block"
				>
					{m.settings_team_billing()}
				</p>
				{#each settingsTabs.slice(6) as tab (tab.id)}
					<button
						type="button"
						data-settings-tab={tab.id}
						class={[
							'min-h-10 shrink-0 rounded-md px-3 py-2 text-left text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none lg:w-full',
							activeSettingsTab === tab.id
								? 'bg-accent text-foreground'
								: 'text-muted-foreground hover:bg-accent/60 hover:text-foreground'
						]}
						onclick={() => openSettingsTab(tab.id)}
						aria-current={activeSettingsTab === tab.id ? 'page' : undefined}
					>
						{tab.label}
					</button>
				{/each}
			</nav>
		</aside>

		<div class="min-w-0 space-y-10">
			<section id="profile" class:hidden={activeSettingsTab !== 'profile'} class="scroll-mt-24">
				{#if avatarUploaderOpen}
					<ProfileAvatarUploader
						bind:open={avatarUploaderOpen}
						onComplete={handleAvatarUploaded}
						onError={(message) => (profileError = message)}
					/>
				{/if}

				<form onsubmit={saveProfile} class="space-y-6">
					<div class="flex flex-col gap-6 sm:flex-row sm:items-center">
						<div class="group relative h-24 w-24 shrink-0">
							{#if profileAvatarURL}
								<img
									src={profileAvatarURL}
									alt="Profile avatar"
									class="h-24 w-24 rounded-full border bg-muted object-cover"
								/>
							{:else}
								<div
									class="flex h-24 w-24 items-center justify-center rounded-full border border-dashed bg-muted text-xl font-semibold text-muted-foreground"
								>
									{profileInitials}
								</div>
							{/if}
							<button
								type="button"
								onclick={() => (avatarUploaderOpen = true)}
								class="absolute inset-0 flex items-center justify-center rounded-full bg-black/45 text-white opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
								aria-label={m.settings_change_profile_picture()}
							>
								<CameraIcon class="h-6 w-6" />
							</button>
						</div>

						<div class="min-w-0 flex-1 space-y-3">
							<div class="space-y-2">
								<Label for="profile-display-name">{m.settings_display_name()}</Label>
								<Input
									id="profile-display-name"
									bind:value={profileDisplayName}
									placeholder={m.settings_your_name()}
									maxlength={120}
								/>
							</div>
							<p class="text-sm text-muted-foreground">{profileEmail}</p>
							<div class="flex flex-wrap gap-2">
								<Button type="button" variant="outline" onclick={() => (avatarUploaderOpen = true)}>
									<CameraIcon class="mr-2 h-4 w-4" />
									{m.settings_change_picture()}
								</Button>
								{#if profileAvatarURL}
									<Button
										type="button"
										variant="ghost"
										class="text-destructive hover:text-destructive"
										onclick={removeAvatar}
										disabled={profileBusy}
									>
										<TrashIcon class="mr-2 h-4 w-4" />
										{m.settings_remove()}
									</Button>
								{/if}
							</div>
						</div>
					</div>

					{#if profileError}
						<div
							class="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
						>
							{profileError}
						</div>
					{/if}

					<div class="flex justify-end">
						<Button type="submit" disabled={profileBusy}>
							{#if profileBusy}
								<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
							{:else}
								<SaveIcon class="mr-2 h-4 w-4" />
							{/if}
							{m.settings_save_profile()}
						</Button>
					</div>
				</form>
			</section>

			<section
				id="workspace"
				class:hidden={activeSettingsTab !== 'general'}
				class="scroll-mt-24 space-y-4"
			>
				<div class="rounded-lg border bg-muted/20 p-4">
					<div class="flex flex-col gap-4 sm:flex-row sm:items-center">
						<div
							class="flex size-16 shrink-0 items-center justify-center overflow-hidden rounded-lg bg-muted text-lg font-semibold text-muted-foreground"
						>
							{#if workspaceCtx.settings.avatar_url}
								<img
									src={workspaceCtx.settings.avatar_url}
									alt={workspaceCtx.currentWorkspace?.name || 'Workspace'}
									class="h-full w-full object-cover"
								/>
							{:else}
								{(workspaceCtx.currentWorkspace?.name?.[0] ?? 'W').toUpperCase()}
							{/if}
						</div>
						<div class="min-w-0 flex-1 space-y-3">
							<div class="flex flex-col gap-1">
								<span class="text-sm font-medium">{workspaceCtx.currentWorkspace?.name}</span>
								<span class="text-sm text-muted-foreground">
									{workspaceCtx.currentWorkspace?.organization_name ||
										m.settings_personal_workspace()}
								</span>
							</div>
							<div class="space-y-2">
								<Label for="workspace-avatar-url">{m.settings_workspace_image_url()}</Label>
								<Input
									id="workspace-avatar-url"
									type="url"
									bind:value={workspaceCtx.settings.avatar_url}
									placeholder="https://example.com/app-icon.png"
									maxlength={1000}
								/>
							</div>
						</div>
					</div>
				</div>
				<div
					class="flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between"
				>
					<div>
						<p class="text-sm font-medium">{m.settings_connected_channels()}</p>
						<p class="text-sm text-muted-foreground">{m.settings_connected_channels_body()}</p>
					</div>
					<Button variant="outline" onclick={() => goto(resolve('/accounts'))}
						>{m.settings_manage_accounts()}</Button
					>
				</div>
			</section>

			<section id="team" class:hidden={activeSettingsTab !== 'members'} class="scroll-mt-24">
				<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
					<div>
						<h2 class="flex items-center gap-2 text-lg font-semibold">
							<UsersIcon class="h-5 w-5 text-muted-foreground" />
							{m.settings_team()}
						</h2>
						<p class="mt-2 text-sm text-muted-foreground">
							{m.settings_team_body()}
						</p>
					</div>
					<div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
						<span class="font-medium">{currentTeamSeats}</span>
						<span class="text-muted-foreground"> seats reserved</span>
					</div>
				</div>

				{#if teamError}
					<div
						data-testid="team-error"
						class="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
					>
						{teamError}
					</div>
				{/if}

				<form
					onsubmit={createWorkspaceInvitation}
					class="mb-4 grid gap-3 lg:grid-cols-[minmax(0,1fr)_220px_auto]"
				>
					<div class="space-y-2">
						<Label for="team-invite-email">{m.settings_invite_email()}</Label>
						<Input
							id="team-invite-email"
							data-testid="team-invite-email"
							type="email"
							bind:value={inviteEmail}
							placeholder="teammate@example.com"
							autocomplete="email"
							required
						/>
					</div>
					<div class="space-y-2">
						<Label for="team-invite-role">{m.settings_role()}</Label>
						<Select.Root
							type="single"
							value={inviteRole}
							onValueChange={(value) => {
								if (value === 'viewer' || value === 'editor' || value === 'admin') {
									inviteRole = value;
								}
							}}
						>
							<Select.Trigger id="team-invite-role" data-testid="team-invite-role" class="w-full">
								{selectedInviteRole.label}
							</Select.Trigger>
							<Select.Content>
								{#each inviteRoleOptions as option (option.value)}
									<Select.Item value={option.value}>
										<div class="flex flex-col gap-0.5 text-left">
											<span>{option.label}</span>
											<span class="text-xs text-muted-foreground">{option.description}</span>
										</div>
									</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="flex items-end">
						<Button type="submit" disabled={teamBusy || !inviteEmail.trim()}>
							{#if teamBusy}
								<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
							{:else}
								<UserPlusIcon class="mr-2 h-4 w-4" />
							{/if}
							{m.settings_send_invite()}
						</Button>
					</div>
				</form>

				{#if createdInviteURL}
					<div
						data-testid="team-invite-link"
						class="mb-4 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4"
					>
						<p class="text-sm font-medium text-emerald-900">{m.settings_invite_created()}</p>
						<div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
							<p
								class="min-w-0 flex-1 rounded-md bg-background px-3 py-2 font-mono text-xs break-all"
							>
								{createdInviteURL}
							</p>
							<Button type="button" variant="outline" size="sm" onclick={copyCreatedInviteURL}>
								<CopyIcon class="mr-2 h-4 w-4" />
								{m.common_copy()}
							</Button>
						</div>
					</div>
				{/if}

				{#if teamLoading}
					<div class="grid gap-3 lg:grid-cols-2">
						<Skeleton class="h-28 rounded-lg" />
						<Skeleton class="h-28 rounded-lg" />
					</div>
				{:else}
					<div class="grid gap-4 lg:grid-cols-2">
						<div>
							<h3 class="mb-2 text-sm font-semibold">{m.settings_members_heading()}</h3>
							<div data-testid="team-members-list" class="space-y-2">
								{#each teamMembers as member (member.user_id)}
									<div
										class="flex flex-col gap-2 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
									>
										<div class="min-w-0">
											<p class="truncate text-sm font-medium">{member.email}</p>
										</div>
										<span
											class="inline-flex w-fit items-center rounded-full border px-2 py-0.5 text-xs font-medium capitalize"
										>
											{member.role}
										</span>
									</div>
								{:else}
									<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
										{m.settings_no_members()}
									</p>
								{/each}
							</div>
						</div>

						<div>
							<h3 class="mb-2 text-sm font-semibold">{m.settings_pending_invitations()}</h3>
							<div data-testid="team-invitations-list" class="space-y-2">
								{#each pendingInvitations as invitation (invitation.id)}
									<div
										class="flex flex-col gap-2 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
									>
										<div class="min-w-0">
											<p class="truncate text-sm font-medium">{invitation.email}</p>
											<p class="text-xs text-muted-foreground">
												{invitation.role} · expires
												{new Date(invitation.expires_at).toLocaleDateString()}
											</p>
										</div>
										<Button
											type="button"
											variant="ghost"
											size="sm"
											class="text-destructive hover:text-destructive"
											onclick={() => revokeWorkspaceInvitation(invitation.id)}
											disabled={teamBusy}
										>
											{m.settings_revoke()}
										</Button>
									</div>
								{:else}
									<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
										{m.settings_no_invitations()}
									</p>
								{/each}
							</div>
						</div>
					</div>
				{/if}
			</section>

			<section id="billing" class:hidden={activeSettingsTab !== 'plan'} class="scroll-mt-24">
				<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
					<div>
						<h2 class="flex items-center gap-2 text-lg font-semibold">
							<CreditCardIcon class="h-5 w-5 text-muted-foreground" />
							{m.settings_billing()}
						</h2>
						<p class="mt-2 text-sm text-muted-foreground">
							{m.settings_billing_body()}
						</p>
					</div>
					<Button variant="outline" onclick={openBillingPortal} disabled={billingPortalBusy}>
						{#if billingPortalBusy}
							<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
						{:else}
							<ExternalLinkIcon class="mr-2 h-4 w-4" />
						{/if}
						{m.settings_customer_portal()}
					</Button>
				</div>

				{#if billingStatusLoading}
					<div class="mb-4 grid gap-3 lg:grid-cols-2">
						<Skeleton class="h-24 rounded-lg" />
						<Skeleton class="h-24 rounded-lg" />
					</div>
				{:else if billingStatus}
					<div class="mb-4 grid gap-3 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]">
						<div class="rounded-lg border bg-muted/20 p-4">
							<p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
								{m.settings_current_plan()}
							</p>
							<div class="mt-2 flex flex-wrap items-end gap-x-3 gap-y-1">
								<p class="text-2xl font-semibold">
									{currentBillingPlan?.name ??
										(billingStatus.plan_id || m.settings_no_active_plan())}
								</p>
								<p class="pb-1 text-sm text-muted-foreground capitalize">{billingStatus.status}</p>
							</div>
							{#if billingStatus.current_period_end}
								<p class="mt-2 text-sm text-muted-foreground">
									Period ends {new Date(billingStatus.current_period_end).toLocaleDateString()}
									{#if billingStatus.cancel_at_period_end}
										· cancels after this period
									{/if}
								</p>
							{:else if hasActiveBillingPlan}
								<p class="mt-2 text-sm text-muted-foreground">
									{m.settings_active_plan()}
								</p>
							{:else}
								<p class="mt-2 text-sm text-muted-foreground">
									{m.settings_start_checkout()}
								</p>
							{/if}
						</div>

						<div class="rounded-lg border bg-muted/20 p-4">
							<p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
								{m.settings_usage_month()}
							</p>
							{#if monthlyBillingUsageRows.length}
								<div class="mt-3 grid gap-3 sm:grid-cols-2">
									{#each monthlyBillingUsageRows as row (row.metric)}
										<div>
											<div class="mb-1 flex items-center justify-between gap-2 text-sm">
												<span>{row.label}</span>
												<span class="text-muted-foreground">
													{formatBillingValue(row.metric, row.current)} / {formatBillingValue(
														row.metric,
														row.limit
													)}
												</span>
											</div>
											<div class="h-2 overflow-hidden rounded-full bg-muted">
												<div
													class="h-full rounded-full bg-primary"
													style:width={`${Math.min(100, Math.round((row.current / Math.max(row.limit, 1)) * 100))}%`}
												></div>
											</div>
										</div>
									{/each}
								</div>
							{:else}
								<p class="mt-2 text-sm text-muted-foreground">
									{m.settings_usage_empty()}
								</p>
							{/if}
						</div>
					</div>
				{/if}

				<details class="border-t pt-4" open={!hasActiveBillingPlan}>
					<summary
						class="cursor-pointer text-sm font-medium focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
					>
						{hasActiveBillingPlan ? m.settings_compare_plan() : m.settings_choose_plan()}
					</summary>
					<div class="mt-4 grid gap-3 lg:grid-cols-3">
						{#each billingPlans as plan (plan.id)}
							<article
								class={`rounded-lg border p-4 ${plan.featured ? 'border-primary bg-primary/5 shadow-sm' : 'bg-background'}`}
							>
								<div class="mb-3 flex items-start justify-between gap-3">
									<div>
										<h3 class="font-semibold">{plan.name}</h3>
										<p class="text-sm text-muted-foreground">{plan.description}</p>
									</div>
									<div class="text-right">
										<div class="text-xl font-semibold">{formatPlanPrice(plan.monthlyPriceEur)}</div>
										<div class="text-xs text-muted-foreground">/mo</div>
									</div>
								</div>
								<ul class="mb-4 space-y-1 text-sm text-muted-foreground">
									{#each plan.limits as limit (limit)}
										<li>{limit}</li>
									{/each}
								</ul>
								<Button
									class="w-full"
									variant={plan.featured ? 'default' : 'outline'}
									onclick={() => startCheckout(plan.id)}
									disabled={Boolean(billingBusyPlan) || hasActiveBillingPlan}
								>
									{#if billingBusyPlan === plan.id}
										<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
									{/if}
									{#if hasActiveBillingPlan && billingStatus?.plan_id === plan.id}
										{m.settings_current_plan()}
									{:else if hasActiveBillingPlan}
										{m.settings_use_portal()}
									{:else}
										Choose {plan.name}
									{/if}
								</Button>
							</article>
						{/each}
					</div>
				</details>

				{#if billingError}
					<div
						class="mt-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
					>
						{billingError}
					</div>
				{/if}
			</section>

			<section id="security" class:hidden={activeSettingsTab !== 'security'} class="scroll-mt-24">
				<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
					<ShieldCheckIcon class="h-5 w-5 text-muted-foreground" />
					{m.settings_account_security()}
				</h2>
				<p class="mb-4 text-sm text-muted-foreground">
					{m.settings_account_security_body()}
				</p>

				{#if loadingSecurity}
					<div class="space-y-3">
						<Skeleton class="h-24 rounded-lg" />
						<Skeleton class="h-40 rounded-lg" />
					</div>
				{:else}
					<div class="space-y-4">
						<div class="rounded-lg border bg-muted/20 p-4">
							<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
								<div>
									<p class="text-sm font-medium">{securityStatus?.user.email}</p>
									<p class="text-sm text-muted-foreground">
										{m.settings_active_methods()}
										{securityStatus?.methods?.length
											? (securityStatus.methods ?? []).join(', ')
											: m.settings_none_configured()}
									</p>
								</div>
								<p class="text-sm text-muted-foreground">
									Passkeys: {passkeyCount}
								</p>
							</div>
						</div>

						<div class="rounded-lg border p-4">
							<div class="mb-4 flex items-center justify-between gap-3">
								<div>
									<h3 class="flex items-center gap-2 font-medium">
										<MonitorIcon class="h-4 w-4 text-muted-foreground" />
										{m.settings_active_sessions()}
									</h3>
									<p class="mt-1 text-sm text-muted-foreground">
										{m.settings_active_sessions_body()}
									</p>
								</div>
								<Button
									variant="outline"
									size="sm"
									onclick={loadAuthSessions}
									disabled={authSessionsLoading}
								>
									{#if authSessionsLoading}
										<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
									{/if}
									{m.common_refresh()}
								</Button>
							</div>

							{#if authSessionsError}
								<div
									class="mb-3 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
								>
									{authSessionsError}
								</div>
							{/if}

							{#if authSessionsLoading}
								<div class="space-y-2">
									<Skeleton class="h-16 rounded-md" />
									<Skeleton class="h-16 rounded-md" />
								</div>
							{:else if authSessions.length === 0}
								<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
									{m.settings_no_sessions()}
								</p>
							{:else}
								<div class="space-y-2" data-testid="auth-session-list">
									{#each authSessions as session (session.id)}
										<div
											class="flex flex-col gap-3 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
											data-testid="auth-session-row"
										>
											<div class="min-w-0">
												<div class="flex flex-wrap items-center gap-2">
													<p class="truncate text-sm font-medium" title={session.user_agent}>
														{session.device_name || formatSessionUserAgent(session.user_agent)}
													</p>
													{#if session.current}
														<span
															class="rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary"
														>
															{m.settings_current()}
														</span>
													{/if}
												</div>
												<p class="mt-1 text-xs text-muted-foreground">
													{session.ip_address || m.settings_unknown_ip()} · {m.settings_last_used()}
													{formatSessionTime(session.last_used_at)} · {m.settings_expires()}
													{formatSessionTime(session.expires_at)}
												</p>
											</div>
											<Button
												variant="ghost"
												size="sm"
												class="self-start text-destructive hover:text-destructive sm:self-center"
												onclick={() => revokeAuthSession(session)}
												disabled={Boolean(authSessionBusyID)}
											>
												{#if authSessionBusyID === session.id}
													<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
												{:else}
													<LogOutIcon class="mr-2 h-4 w-4" />
												{/if}
												{session.current ? m.settings_sign_out() : m.settings_revoke()}
											</Button>
										</div>
									{/each}
								</div>
							{/if}
						</div>

						<div class="grid gap-4 lg:grid-cols-2">
							<div class="rounded-lg border p-4">
								<div class="mb-3 flex items-center gap-2">
									<SmartphoneIcon class="h-4 w-4 text-muted-foreground" />
									<h3 class="font-medium">{m.settings_authenticator()}</h3>
								</div>
								<p class="mb-4 text-sm text-muted-foreground">
									{m.settings_authenticator_body()}
								</p>

								{#if securityStatus?.totp_enabled}
									<div class="space-y-3">
										<div class="rounded-md bg-emerald-500/10 px-3 py-2 text-sm text-emerald-700">
											{m.settings_authenticator_enabled()}
										</div>
										<div class="space-y-2">
											<Label for="disable-password">{m.settings_current_password()}</Label>
											<Input
												id="disable-password"
												type="password"
												bind:value={currentPassword}
												placeholder="Required to disable"
											/>
										</div>
										<Button
											variant="outline"
											onclick={disableTOTP}
											disabled={securityBusy || !currentPassword.trim()}
										>
											{m.settings_disable_authenticator()}
										</Button>
									</div>
								{:else}
									<div class="space-y-3">
										<div class="space-y-2">
											<Label for="totp-password">{m.settings_current_password()}</Label>
											<Input
												id="totp-password"
												type="password"
												bind:value={currentPassword}
												placeholder="Required to start setup"
											/>
										</div>
										<Button
											onclick={startTOTPSetup}
											disabled={securityBusy || !currentPassword.trim()}
										>
											{m.settings_start_authenticator()}
										</Button>

										{#if totpSetupChallengeId}
											<div class="space-y-3 rounded-lg border bg-muted/20 p-4">
												<img
													src={totpQRCodeDataURL}
													alt="TOTP QR code"
													class="mx-auto h-48 w-48 rounded-lg border bg-white p-2"
												/>
												<div class="space-y-1">
													<p class="text-sm font-medium">{m.settings_manual_key()}</p>
													<p class="font-mono text-xs break-all text-muted-foreground">
														{totpManualEntryKey}
													</p>
												</div>
												<div class="space-y-2">
													<Label for="totp-code">{m.settings_totp_code()}</Label>
													<Input
														id="totp-code"
														bind:value={totpCode}
														inputmode="numeric"
														autocomplete="one-time-code"
														maxlength={6}
														placeholder="123456"
													/>
												</div>
												<Button
													onclick={confirmTOTPSetup}
													disabled={securityBusy || totpCode.trim().length !== 6}
												>
													{m.settings_confirm_authenticator()}
												</Button>
											</div>
										{/if}
									</div>
								{/if}
							</div>

							<div class="rounded-lg border p-4">
								<div class="mb-3 flex items-center gap-2">
									<KeyRoundIcon class="h-4 w-4 text-muted-foreground" />
									<h3 class="font-medium">{m.settings_passkeys()}</h3>
								</div>
								<p class="mb-4 text-sm text-muted-foreground">
									{m.settings_passkeys_body()}
								</p>

								<div class="space-y-3">
									<div class="space-y-2">
										<Label for="passkey-password">{m.settings_current_password()}</Label>
										<Input
											id="passkey-password"
											type="password"
											bind:value={currentPassword}
											placeholder="Required to add or remove passkeys"
										/>
									</div>
									<div class="space-y-2">
										<Label for="passkey-name">{m.settings_passkey_name()}</Label>
										<Input
											id="passkey-name"
											bind:value={newPasskeyName}
											placeholder="MacBook, iPhone, YubiKey"
										/>
									</div>
									<Button onclick={addPasskey} disabled={securityBusy || !currentPassword.trim()}>
										{m.settings_add_passkey()}
									</Button>
								</div>

								<div class="mt-4 space-y-2">
									{#if (securityStatus?.passkeys ?? []).length}
										{#each securityStatus?.passkeys ?? [] as passkey (passkey.id)}
											<div class="flex items-center justify-between rounded-md border px-3 py-2">
												<div>
													<p class="text-sm font-medium">{passkey.name}</p>
													<p class="text-xs text-muted-foreground">
														{#if passkey.last_used_at && passkey.last_used_at !== '0001-01-01T00:00:00Z'}
															Last used {new Date(passkey.last_used_at).toLocaleString()}
														{:else}
															Added {new Date(passkey.created_at).toLocaleString()}
														{/if}
													</p>
												</div>
												<Button
													variant="ghost"
													size="sm"
													class="text-destructive hover:text-destructive"
													onclick={() => removePasskey(passkey.id)}
													disabled={securityBusy || !currentPassword.trim()}
												>
													{m.settings_remove()}
												</Button>
											</div>
										{/each}
									{:else}
										<p class="text-sm text-muted-foreground">{m.settings_no_passkeys()}</p>
									{/if}
								</div>
							</div>
						</div>

						{#if securityError}
							<div
								class="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
							>
								{securityError}
							</div>
						{/if}
					</div>
				{/if}
			</section>

			<section id="tokens" class:hidden={activeSettingsTab !== 'developer'} class="scroll-mt-24">
				<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
					<TerminalIcon class="h-5 w-5 text-muted-foreground" />
					{m.settings_tokens_heading()}
				</h2>
				<p class="mb-4 text-sm text-muted-foreground">
					{m.settings_tokens_body()}
				</p>

				<div class="mb-4 grid gap-3 lg:grid-cols-[1fr_240px_240px_auto]">
					<div class="space-y-2">
						<Label for="api-token-name">{m.settings_token_name()}</Label>
						<Input
							id="api-token-name"
							bind:value={apiTokenName}
							placeholder="ChatGPT App, MacBook CLI, GitHub CI"
						/>
					</div>
					<div class="space-y-2">
						<Label for="api-token-scope">{m.settings_token_scope()}</Label>
						<Select.Root
							type="single"
							value={apiTokenScope}
							onValueChange={(value) => value && (apiTokenScope = value)}
						>
							<Select.Trigger id="api-token-scope" data-testid="api-token-scope" class="w-full">
								{selectedAPITokenScope.label}
							</Select.Trigger>
							<Select.Content>
								{#each apiTokenScopeOptions as option (option.value)}
									<Select.Item value={option.value}>
										<div class="flex flex-col gap-0.5 text-left">
											<span>{option.label}</span>
											<span class="text-xs text-muted-foreground">{option.description}</span>
										</div>
									</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="space-y-2">
						<Label for="api-token-workspace">{m.settings_access_boundary()}</Label>
						<Select.Root
							type="single"
							value={apiTokenWorkspaceScope}
							onValueChange={(value) => value && (apiTokenWorkspaceScope = value)}
						>
							<Select.Trigger id="api-token-workspace" class="w-full">
								{selectedAPITokenWorkspaceScope.label}
							</Select.Trigger>
							<Select.Content>
								{#each apiTokenWorkspaceOptions as option (option.value)}
									<Select.Item value={option.value}>
										<div class="flex flex-col gap-0.5 text-left">
											<span>{option.label}</span>
											<span class="text-xs text-muted-foreground">{option.description}</span>
										</div>
									</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="flex items-end">
						<Button
							onclick={createAPIToken}
							disabled={apiTokenBusy ||
								(apiTokenWorkspaceScope === 'current' && !workspaceCtx.currentWorkspace)}
						>
							{#if apiTokenBusy}
								<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
							{/if}
							{m.settings_create_token()}
						</Button>
					</div>
				</div>

				{#if createdAPIToken}
					<div
						class="mb-4 rounded-lg border border-amber-300/50 bg-amber-50 p-4 text-sm text-amber-950"
					>
						<p class="font-medium">{m.settings_copy_token_now()}</p>
						<p class="mt-2 font-mono text-xs break-all">{createdAPIToken}</p>
					</div>
				{/if}

				{#if apiTokensLoading}
					<div class="space-y-2">
						<Skeleton class="h-14 rounded-md" />
						<Skeleton class="h-14 rounded-md" />
					</div>
				{:else if apiTokens.length === 0}
					<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
						{m.settings_no_tokens()}
					</p>
				{:else}
					<div class="space-y-2">
						{#each apiTokens as token (token.id)}
							<div
								class="flex flex-col gap-3 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
							>
								<div>
									<p class="text-sm font-medium">{token.name}</p>
									<p class="text-xs text-muted-foreground">
										Prefix <span class="font-mono">{token.token_prefix}</span> · {token.scope} · Created
										{new Date(token.created_at).toLocaleString()}
										{#if token.workspace_id}
											· Workspace <span class="font-mono">{token.workspace_id}</span>
										{:else}
											· All workspaces
										{/if}
										{#if token.last_used_at}
											· Last used {new Date(token.last_used_at).toLocaleString()}
										{/if}
									</p>
								</div>
								<Button
									variant="ghost"
									size="sm"
									class="text-destructive hover:text-destructive"
									onclick={() => revokeAPIToken(token.id)}
									disabled={apiTokenBusy}
								>
									{m.settings_revoke()}
								</Button>
							</div>
						{/each}
					</div>
				{/if}

				<div class="mt-6 border-t pt-6">
					<div class="mb-4 flex items-center justify-between gap-3">
						<div>
							<h3 class="flex items-center gap-2 text-sm font-semibold">
								<ActivityIcon class="h-4 w-4 text-muted-foreground" />
								{m.settings_mcp_activity()}
							</h3>
							<p class="mt-1 text-sm text-muted-foreground">
								{m.settings_mcp_activity_body()}
							</p>
						</div>
						<Button
							variant="outline"
							size="sm"
							onclick={loadMCPActivity}
							disabled={mcpActivityLoading}
						>
							{#if mcpActivityLoading}
								<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
							{/if}
							{m.common_refresh()}
						</Button>
					</div>

					{#if mcpActivityError}
						<div
							data-testid="mcp-activity-error"
							class="mb-3 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
						>
							{mcpActivityError}
						</div>
					{/if}

					{#if mcpActivityLoading}
						<div class="space-y-2">
							<Skeleton class="h-16 rounded-md" />
							<Skeleton class="h-16 rounded-md" />
						</div>
					{:else if mcpActivity.length === 0}
						<p
							data-testid="mcp-activity-empty"
							class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground"
						>
							{m.settings_no_mcp_activity()}
						</p>
					{:else}
						<div data-testid="mcp-activity-list" class="space-y-2">
							{#each mcpActivity as call (call.id)}
								<div class="rounded-md border px-3 py-3">
									<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
										<div class="min-w-0">
											<p class="truncate text-sm font-medium">{call.tool_name}</p>
											<p class="mt-1 text-xs text-muted-foreground">
												{new Date(call.created_at).toLocaleString()} · {call.duration_ms} ms
												{#if call.workspace_id}
													· Workspace <span class="font-mono">{call.workspace_id}</span>
												{/if}
											</p>
											{#if call.client_name || call.client_scope}
												<p class="mt-1 truncate text-xs text-muted-foreground">
													Client {call.client_name || call.client_scope}
													{#if call.client_token_prefix}
														· <span class="font-mono">{call.client_token_prefix}</span>
													{/if}
												</p>
											{/if}
										</div>
										<span
											class={[
												'inline-flex w-fit items-center rounded-full border px-2 py-0.5 text-xs font-medium',
												call.status === 'success'
													? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700'
													: 'border-destructive/30 bg-destructive/10 text-destructive'
											]}
										>
											{call.status}
										</span>
									</div>
									{#if call.error_message}
										<p class="mt-2 text-xs text-destructive">{call.error_message}</p>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</section>

			<section
				id="date-time"
				class:hidden={activeSettingsTab !== 'general'}
				class="scroll-mt-24 space-y-4"
			>
				<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
					<ClockIcon class="h-5 w-5 text-muted-foreground" />
					{m.settings_date_time()}
				</h2>
				<div class="grid gap-4 sm:grid-cols-2">
					<div class="space-y-2">
						<label class="text-sm font-medium" for="timezone-select">{m.settings_timezone()}</label>
						<Select.Root
							type="single"
							value={workspaceCtx.settings.timezone}
							onValueChange={handleTimezoneChange}
						>
							<Select.Trigger id="timezone-select" class="w-full">
								{getTimezoneLabel(workspaceCtx.settings.timezone)}
							</Select.Trigger>
							<Select.Content class="max-h-80 overflow-y-auto">
								{#each Object.entries(groupedTimezones) as [group, tzs] (group)}
									<Select.Group>
										<Select.GroupHeading class="text-xs">{group}</Select.GroupHeading>
										{#each tzs as tz (tz.value)}
											<Select.Item value={tz.value}>{tz.label}</Select.Item>
										{/each}
									</Select.Group>
								{/each}
							</Select.Content>
						</Select.Root>
						<p class="text-sm text-muted-foreground">
							{m.settings_timezone_body()}
						</p>
					</div>

					<div class="space-y-2">
						<label class="text-sm font-medium" for="week-start-select"
							>{m.settings_week_starts()}</label
						>
						<Select.Root
							type="single"
							value={String(workspaceCtx.settings.week_start)}
							onValueChange={(v) => handleWeekStartChange(Number(v))}
						>
							<Select.Trigger id="week-start-select" class="w-full">
								{workspaceCtx.settings.week_start === 0 ? m.settings_sunday() : m.settings_monday()}
							</Select.Trigger>
							<Select.Content>
								<Select.Item value="0">{m.settings_sunday()}</Select.Item>
								<Select.Item value="1">{m.settings_monday()}</Select.Item>
							</Select.Content>
						</Select.Root>
						<p class="text-sm text-muted-foreground">
							{m.settings_week_start_body()}
						</p>
					</div>
				</div>
			</section>

			<section
				id="media-cleanup"
				class:hidden={activeSettingsTab !== 'media'}
				class="scroll-mt-24 space-y-4"
			>
				<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
					<ImageIcon class="h-5 w-5 text-muted-foreground" />
					{m.settings_media_cleanup()}
				</h2>
				<div class="space-y-2">
					<label class="text-sm font-medium" for="cleanup-select"
						>{m.settings_auto_delete_media()}</label
					>
					<Select.Root
						type="single"
						value={String(workspaceCtx.settings.media_cleanup_days)}
						onValueChange={(v) => handleCleanupDaysChange(Number(v))}
					>
						<Select.Trigger id="cleanup-select" class="w-full">
							{cleanupDaysOptions.find((o) => o.value === workspaceCtx.settings.media_cleanup_days)
								?.label || m.settings_disabled()}
						</Select.Trigger>
						<Select.Content>
							{#each cleanupDaysOptions as option (option.value)}
								<Select.Item value={String(option.value)}>{option.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
					<p class="text-sm text-muted-foreground">
						{m.settings_auto_delete_media_body()}
					</p>
				</div>
			</section>

			<section
				id="posting-schedule"
				class:hidden={activeSettingsTab !== 'schedule'}
				class="scroll-mt-24"
			>
				<div class="mb-4 flex items-center justify-between">
					<h2 class="flex items-center gap-2 text-lg font-semibold">
						<CalendarIcon class="h-5 w-5 text-muted-foreground" />
						{m.settings_posting_schedule()}
					</h2>
					<Button
						onclick={() => (showSuggestSchedule = !showSuggestSchedule)}
						variant="outline"
						size="sm"
					>
						<SparklesIcon class="mr-2 h-4 w-4" />
						{m.settings_suggest_pattern()}
					</Button>
				</div>
				<p class="mb-4 text-sm text-muted-foreground">
					{m.settings_schedule_body()}
				</p>

				<div class="mb-4 rounded-xl border bg-muted/20 p-4">
					<div class="grid gap-4 lg:grid-cols-[180px_1fr_auto]">
						<div class="space-y-2">
							<label class="text-sm font-medium" for="new-time">{m.settings_add_time_row()}</label>
							<Input id="new-time" bind:value={newTimeInput} type="time" step="900" />
						</div>
						<div class="space-y-2">
							<span class="text-sm font-medium">{m.settings_active_days()}</span>
							<div class="flex flex-wrap gap-3">
								{#each dayOrder as dayIndex (dayIndex)}
									<label
										class="flex items-center gap-2 rounded-md border bg-background px-3 py-2 text-sm"
									>
										<Checkbox
											checked={newTimeDays.includes(dayIndex)}
											onCheckedChange={() => toggleNewDay(dayIndex)}
										/>
										<span>{localizedWeekday(dayIndex)}</span>
									</label>
								{/each}
							</div>
						</div>
						<div class="flex items-end">
							<Button onclick={addTimeRow} class="w-full lg:w-auto">
								<PlusIcon class="mr-2 h-4 w-4" />
								{m.settings_add_time()}
							</Button>
						</div>
					</div>
					{#if newTimeError}
						<p class="mt-3 text-xs text-destructive">{newTimeError}</p>
					{:else}
						<p class="mt-3 text-xs text-muted-foreground">
							{m.settings_new_rows_timezone({
								timezone: getTimezoneLabel(workspaceCtx.settings.timezone)
							})}
						</p>
					{/if}
				</div>

				{#if showSuggestSchedule}
					<div class="mb-4 rounded-xl border bg-background p-4">
						<div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
							<div class="space-y-2">
								<label class="text-sm font-medium" for="posts-per-day"
									>{m.settings_suggested_posts_day()}</label
								>
								<Select.Root
									type="single"
									value={String(suggestedPostsPerDay)}
									onValueChange={(v) => (suggestedPostsPerDay = Number(v))}
								>
									<Select.Trigger id="posts-per-day" class="w-28">
										{suggestedPostsPerDay}
									</Select.Trigger>
									<Select.Content class="max-h-60 overflow-y-auto">
										{#each Array.from({ length: 10 }, (_, i) => i + 1) as n (n)}
											<Select.Item value={String(n)}>{n}</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>
							<div class="flex gap-2">
								<Button onclick={() => (showSuggestSchedule = false)} variant="outline" size="sm"
									>{m.common_cancel()}</Button
								>
								<Button onclick={generateSuggestedSchedule} size="sm" disabled={generatingSchedule}>
									{#if generatingSchedule}
										<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
									{/if}
									{m.settings_generate()}
								</Button>
							</div>
						</div>
					</div>
				{/if}

				{#if loadingSchedules}
					<div class="space-y-2">
						<Skeleton class="h-14 rounded-md" />
						<Skeleton class="h-14 rounded-md" />
						<Skeleton class="h-14 rounded-md" />
					</div>
				{:else}
					<div class="overflow-hidden rounded-xl border">
						<div
							class="grid grid-cols-[120px_repeat(7,minmax(56px,1fr))_52px] border-b bg-muted/30"
						>
							<div
								class="px-4 py-3 text-xs font-semibold tracking-wide text-muted-foreground uppercase"
							>
								{m.settings_time()}
							</div>
							{#each dayOrder as dayIndex (dayIndex)}
								<div
									class="px-2 py-3 text-center text-xs font-semibold tracking-wide text-muted-foreground uppercase"
								>
									{localizedWeekday(dayIndex)}
								</div>
							{/each}
							<div class="px-2 py-3"></div>
						</div>

						{#if scheduleRows.length === 0}
							<div class="px-4 py-10 text-center text-sm text-muted-foreground">
								{m.settings_no_posting_times()}
							</div>
						{:else}
							{#each scheduleRows as row (row.key)}
								<div
									class="grid grid-cols-[120px_repeat(7,minmax(56px,1fr))_52px] border-b last:border-b-0"
								>
									<div class="px-4 py-3">
										<div class="font-medium">{formatTime(row.local_hour, row.local_minute)}</div>
										{#if row.label}
											<div class="text-xs text-muted-foreground">{row.label}</div>
										{/if}
									</div>
									{#each dayOrder as dayIndex (`${row.key}-${dayIndex}`)}
										<div class="flex items-center justify-center px-2 py-3">
											<Checkbox
												checked={Boolean(row.days[dayIndex])}
												onCheckedChange={() => toggleScheduleCell(row, dayIndex)}
												aria-label={m.settings_toggle_schedule_cell({
													day: localizedWeekday(dayIndex, 'long'),
													time: formatTime(row.local_hour, row.local_minute)
												})}
											/>
										</div>
									{/each}
									<div class="flex items-center justify-center px-2 py-3">
										<Button
											variant="ghost"
											size="icon"
											class="h-8 w-8"
											onclick={() => removeTimeRow(row)}
											aria-label={m.settings_remove_time_row({
												time: formatTime(row.local_hour, row.local_minute)
											})}
										>
											<TrashIcon class="h-4 w-4" />
										</Button>
									</div>
								</div>
							{/each}
						{/if}
					</div>
				{/if}
			</section>

			<section
				id="natural-posting"
				class:hidden={activeSettingsTab !== 'schedule'}
				class="scroll-mt-24 space-y-4"
			>
				<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
					<ClockIcon class="h-5 w-5 text-muted-foreground" />
					{m.settings_advanced_scheduling()}
				</h2>
				<div class="space-y-4">
					<p class="text-sm text-muted-foreground">
						{m.settings_advanced_scheduling_body()}
					</p>
					<div class="space-y-2">
						<label class="text-sm font-medium" for="random-delay"
							>{m.settings_time_variation()}</label
						>
						<Select.Root
							type="single"
							value={String(workspaceCtx.settings.random_delay_minutes)}
							onValueChange={(v) => (workspaceCtx.settings.random_delay_minutes = Number(v))}
						>
							<Select.Trigger id="random-delay" class="w-full sm:w-64">
								{#if workspaceCtx.settings.random_delay_minutes === 0}
									{m.settings_no_delay()}
								{:else}
									±{m.settings_minutes({
										minutes: workspaceCtx.settings.random_delay_minutes
									})}
								{/if}
							</Select.Trigger>
							<Select.Content>
								<Select.Item value="0">{m.settings_no_delay()}</Select.Item>
								{#each [5, 10, 15, 30, 45] as delay (delay)}
									<Select.Item value={String(delay)}
										>±{m.settings_minutes({ minutes: delay })}</Select.Item
									>
								{/each}
								<Select.Item value="60">±{m.settings_one_hour()}</Select.Item>
							</Select.Content>
						</Select.Root>
					</div>
					<div class="space-y-2">
						<label class="text-sm font-medium" for="draft-gap">{m.settings_queue_full()}</label>
						<Input
							id="draft-gap"
							type="text"
							value={draftGapInput}
							oninput={(e) => handleDraftGapChange((e.target as HTMLInputElement).value)}
							placeholder={m.settings_draft_gap_placeholder()}
							class={draftGapError ? 'border-destructive' : ''}
						/>
						{#if draftGapError}
							<p class="text-xs text-destructive">{draftGapError}</p>
						{:else}
							<p class="text-xs text-muted-foreground">
								{m.settings_queue_spillover_body({
									minutes: workspaceCtx.settings.draft_gap_minutes
								})}
							</p>
						{/if}
					</div>
				</div>
			</section>

			<section
				id="slot-defaults"
				class:hidden={activeSettingsTab !== 'schedule'}
				class="scroll-mt-24 space-y-4"
			>
				<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
					<ClockIcon class="h-5 w-5 text-muted-foreground" />
					{m.settings_time_picker_range()}
				</h2>
				<div class="space-y-4">
					<p class="text-sm text-muted-foreground">
						{m.settings_time_picker_range_body()}
					</p>
					<div class="grid gap-4 sm:grid-cols-3">
						<div class="space-y-2">
							<label class="text-sm font-medium" for="start-time">{m.settings_start_time()}</label>
							<Select.Root
								type="single"
								value={String(workspaceCtx.settings.slot_start_hour)}
								onValueChange={(v) => (workspaceCtx.settings.slot_start_hour = Number(v))}
							>
								<Select.Trigger id="start-time" class="w-full">
									{workspaceCtx.settings.slot_start_hour.toString().padStart(2, '0')}:00
								</Select.Trigger>
								<Select.Content class="max-h-60 overflow-y-auto">
									{#each Array.from({ length: 24 }, (_, i) => i) as hour (hour)}
										<Select.Item value={String(hour)}
											>{hour.toString().padStart(2, '0')}:00</Select.Item
										>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<div class="space-y-2">
							<label class="text-sm font-medium" for="end-time">{m.settings_end_time()}</label>
							<Select.Root
								type="single"
								value={String(workspaceCtx.settings.slot_end_hour)}
								onValueChange={(v) => (workspaceCtx.settings.slot_end_hour = Number(v))}
							>
								<Select.Trigger id="end-time" class="w-full">
									{workspaceCtx.settings.slot_end_hour.toString().padStart(2, '0')}:00
								</Select.Trigger>
								<Select.Content class="max-h-60 overflow-y-auto">
									{#each Array.from({ length: 24 }, (_, i) => i) as hour (hour)}
										<Select.Item value={String(hour)}
											>{hour.toString().padStart(2, '0')}:00</Select.Item
										>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<div class="space-y-2">
							<label class="text-sm font-medium" for="interval">{m.settings_interval()}</label>
							<input
								id="interval"
								type="text"
								value={intervalInput}
								oninput={(e) => handleIntervalChange((e.target as HTMLInputElement).value)}
								placeholder={m.settings_interval_placeholder()}
								class="h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm {intervalError
									? 'border-destructive'
									: ''}"
							/>
							{#if intervalError}
								<p class="text-xs text-destructive">{intervalError}</p>
							{:else}
								<p class="text-xs text-muted-foreground">
									{m.settings_current_interval({
										minutes: workspaceCtx.settings.slot_interval_minutes
									})}
								</p>
							{/if}
						</div>
					</div>
				</div>
			</section>

			<div
				class:hidden={!['general', 'media', 'schedule'].includes(activeSettingsTab)}
				class="sticky bottom-3 z-10 flex justify-end rounded-lg border bg-background/95 p-3 shadow-sm backdrop-blur"
			>
				<Button onclick={saveSettings} disabled={saving}>
					{#if saving}
						<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
					{:else}
						<SaveIcon class="mr-2 h-4 w-4" />
					{/if}
					{m.settings_save_changes()}
				</Button>
			</div>
		</div>
	</div>
</PageContainer>
