<script lang="ts">
	import { onMount, type Snippet } from 'svelte';
	import { client, type SocialAccount } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { getAuthenticatedMediaByID } from '$lib/media-url';
	import { isSupportedMediaFile, uploadMediaFile } from '$lib/media-upload-client';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { ui } from '$lib/stores/ui.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Calendar } from '$lib/components/ui/calendar';
	import { Input } from '$lib/components/ui/input';
	import * as Popover from '$lib/components/ui/popover';
	import { Textarea } from '$lib/components/ui/textarea';
	import ComposerAccountMenu from './composer-account-menu.svelte';
	import InlineNotice from './inline-notice.svelte';
	import PageLoading from './page-loading.svelte';
	import PlatformIcon from './platform-icon.svelte';
	import { getLocaleTag } from '$lib/i18n';
	import { getPlatformKey, getPlatformName } from '$lib/utils';
	import { CalendarDate, isEqualDay } from '@internationalized/date';
	import {
		workspaceClock,
		workspaceScheduleFromISO,
		workspaceScheduleToISO
	} from './compose/schedule-timezone';
	import {
		buildFocusedPublicationPayload,
		composerMode,
		isAccountCompatibleWithMode,
		roleFieldsForMode,
		type ComposerModeKey,
		type FocusedComposerFields,
		type FocusedFieldKey
	} from './compose/modes';
	import {
		defaultFocusedSchedulingSettings,
		isFocusedProviderReadinessReady,
		isFutureSchedule,
		snapshotFocusedSchedulingSettings,
		type FocusedSchedulingSettings
	} from './compose/focused-workspace';
	import AlertCircleIcon from 'lucide-svelte/icons/alert-circle';
	import CalendarClockIcon from 'lucide-svelte/icons/calendar-clock';
	import ImagePlusIcon from 'lucide-svelte/icons/image-plus';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import SaveIcon from 'lucide-svelte/icons/save';
	import SendIcon from 'lucide-svelte/icons/send';
	import Settings2Icon from 'lucide-svelte/icons/settings-2';
	import UploadIcon from 'lucide-svelte/icons/upload';
	import XIcon from 'lucide-svelte/icons/x';
	import { m } from '$lib/paraglide/messages';

	type Capability = components['schemas']['Capability'];
	type Publication = components['schemas']['PublicationResponse'];
	type Rendition = components['schemas']['RenditionResponse'];
	type MediaSummary = components['schemas']['MediaSummary'];
	type ValidationIssue = components['schemas']['ValidationIssue'];
	type ProviderReadinessItem = components['schemas']['ProviderReadinessItem'];
	type SettingField = components['schemas']['SettingField'];

	interface FocusedMedia {
		id: string;
		mime_type: string;
		url: string;
		size?: number;
		filename?: string;
		role?: string;
	}

	interface Props {
		mode: ComposerModeKey;
		initialPublication?: Publication | null;
		onSuccess?: () => void;
		onCancel?: () => void;
		modeControl?: Snippet;
	}

	let { mode, initialPublication = null, onSuccess, onCancel, modeControl }: Props = $props();

	let publicationId = $state('');
	let hydratedPublicationId = $state('');
	let selectedWorkspaceId = $state('');
	let accounts = $state<SocialAccount[]>([]);
	let selectedAccountIds = $state<string[]>([]);
	let capabilities = $state<Capability[]>([]);
	let providerReadiness = $state<ProviderReadinessItem[]>([]);
	let providerReadinessWorkspaceId = $state('');
	let providerReadinessLoading = $state(false);
	let providerReadinessError = $state('');
	let fields = $state<FocusedComposerFields>({});
	let media = $state<FocusedMedia[]>([]);
	let thumbnailMedia = $state<FocusedMedia | null>(null);
	let thumbnailMediaId = $state('');
	let settingsByAccount = $state<Record<string, Record<string, unknown>>>({});
	let customizeAccountId = $state('');
	let selectedDate = $state<CalendarDate | undefined>(undefined);
	let selectedTime = $state<string | null>(null);
	let showSchedulePopover = $state(false);
	let validationIssues = $state<ValidationIssue[]>([]);
	let loading = $state(true);
	let accountsLoading = $state(false);
	let uploading = $state(false);
	let uploadingThumbnail = $state(false);
	let saving = $state(false);
	let error = $state('');
	let success = $state('');
	let schedulingSettings = $state<FocusedSchedulingSettings>(defaultFocusedSchedulingSettings());
	let schedulingSettingsWorkspaceId = $state('');
	let workspaceChangeSequence = 0;
	let accountsRequestSequence = 0;
	let readinessRequestSequence = 0;
	let mediaUploadRequestSequence = 0;
	let thumbnailUploadRequestSequence = 0;
	let publicationContextRequestId = '';

	const modeMeta = $derived(composerMode(mode));
	const compatibleAccounts = $derived(accounts.filter(isAccountCompatible));
	const selectedAccounts = $derived(
		accounts.filter((account) => selectedAccountIds.includes(account.id))
	);
	const roleFields = $derived(roleFieldsForMode(mode, selectedAccounts));
	const selectedCapabilities = $derived(
		selectedAccounts
			.map((account) => capabilityForAccount(account))
			.filter((capability): capability is Capability => capability !== null)
	);
	const hasYouTubeTarget = $derived(
		selectedAccounts.some((account) => getPlatformKey(account.platform) === 'youtube')
	);
	const blockingIssues = $derived(validationIssues.filter((issue) => issue.severity === 'error'));
	const warningIssues = $derived(validationIssues.filter((issue) => issue.severity === 'warning'));
	const localBlockers = $derived(formBlockers());
	const canSaveDraft = $derived(Boolean(selectedWorkspaceId) && !saving);
	const selectedReadinessProviders = $derived(
		selectedAccounts.map((account) => getPlatformKey(account.platform))
	);
	const loadedReadinessProviders = $derived(providerReadiness.map((item) => item.provider));
	const selectedProviderReadinessReady = $derived(
		isFocusedProviderReadinessReady(
			selectedWorkspaceId,
			providerReadinessWorkspaceId,
			providerReadinessLoading,
			providerReadinessError,
			selectedReadinessProviders,
			loadedReadinessProviders
		)
	);
	const canQueue = $derived(
		canSaveDraft && selectedProviderReadinessReady && localBlockers.length === 0
	);
	const isEditMode = $derived(Boolean(initialPublication?.id));
	const selectedWorkspaceSettingsReady = $derived(
		Boolean(selectedWorkspaceId) && schedulingSettingsWorkspaceId === selectedWorkspaceId
	);
	const canSchedule = $derived(canQueue && selectedWorkspaceSettingsReady);
	const scheduleTimezoneLabel = $derived(schedulingSettings.timezone);
	const isToday = $derived(
		selectedDate ? isEqualDay(selectedDate, workspaceClock(scheduleTimezoneLabel).date) : false
	);
	const allTimeSlots = $derived.by(() => {
		const start = schedulingSettings.slotStartHour;
		const end = schedulingSettings.slotEndHour;
		const interval = Math.max(1, schedulingSettings.slotIntervalMinutes);
		const slots: string[] = [];
		for (let hour = start; hour <= end; hour++) {
			for (let minute = 0; minute < 60; minute += interval) {
				if (hour === end && minute > 0) break;
				slots.push(`${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`);
			}
		}
		return slots;
	});
	const timeSlots = $derived.by(() => {
		const date = selectedDate;
		if (!date) return allTimeSlots;
		const validSlots = allTimeSlots.filter((slot) =>
			Boolean(workspaceScheduleToISO(date, slot, scheduleTimezoneLabel))
		);
		if (!isToday) return validSlots;
		const currentMinutes = workspaceClock(scheduleTimezoneLabel).minutes;
		return validSlots.filter((slot) => {
			const [hour, minute] = slot.split(':').map(Number);
			return hour * 60 + minute > currentMinutes;
		});
	});

	onMount(async () => {
		await loadInitialData();
	});

	$effect(() => {
		if (
			!loading &&
			initialPublication &&
			initialPublication.id !== hydratedPublicationId &&
			initialPublication.id !== publicationContextRequestId
		) {
			publicationContextRequestId = initialPublication.id;
			void loadInitialPublicationContext(initialPublication);
		}
	});

	$effect(() => {
		const workspaceId = workspaceCtx.currentWorkspace?.id ?? '';
		if (!loading && !isEditMode && workspaceId && workspaceId !== selectedWorkspaceId) {
			void changeWorkspace(workspaceId);
		}
	});

	async function loadInitialData() {
		loading = true;
		error = '';
		try {
			if (initialPublication) {
				publicationContextRequestId = initialPublication.id;
				selectedWorkspaceId = initialPublication.workspace_id;
				await loadSchedulingSettings(selectedWorkspaceId);
				hydrateInitialPublication(initialPublication);
			}
			const { data: capabilityData, error: capError } = await client.GET('/capabilities', {});
			if (capError) throw new Error(capError.detail || m.compose_load_capabilities_failed());
			capabilities = capabilityData?.capabilities ?? [];
			selectedWorkspaceId = selectedWorkspaceId || workspaceCtx.currentWorkspace?.id || '';
			if (selectedWorkspaceId) {
				if (schedulingSettingsWorkspaceId !== selectedWorkspaceId) {
					await loadSchedulingSettings(selectedWorkspaceId);
				}
				await Promise.all([
					loadAccounts(selectedWorkspaceId),
					loadProviderReadiness(selectedWorkspaceId)
				]);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : m.compose_load_composer_failed();
		} finally {
			loading = false;
		}
	}

	async function loadInitialPublicationContext(publication: Publication) {
		loading = true;
		selectedWorkspaceId = publication.workspace_id;
		hydratedPublicationId = '';
		resetWorkspaceScopedState();
		fields = {};
		try {
			const settingsReady = await loadSchedulingSettings(publication.workspace_id);
			if (!settingsReady || initialPublication?.id !== publication.id) return;
			hydrateInitialPublication(publication);
			await Promise.all([
				loadAccounts(publication.workspace_id),
				loadProviderReadiness(publication.workspace_id)
			]);
		} catch (err) {
			if (initialPublication?.id === publication.id) {
				error = err instanceof Error ? err.message : m.compose_load_composer_failed();
			}
		} finally {
			loading = false;
		}
	}

	async function loadSchedulingSettings(workspaceId: string): Promise<boolean> {
		await ensureSchedulingWorkspace(workspaceId);
		if (selectedWorkspaceId !== workspaceId) return false;

		schedulingSettings = snapshotFocusedSchedulingSettings(workspaceCtx.settings);
		schedulingSettingsWorkspaceId = workspaceId;
		return true;
	}

	async function ensureSchedulingWorkspace(workspaceId: string) {
		const workspace = workspaceCtx.workspaces.find((candidate) => candidate.id === workspaceId);
		if (!workspace) throw new Error(m.compose_load_workspaces_failed());

		if (workspaceCtx.currentWorkspace?.id !== workspaceId) {
			await workspaceCtx.setWorkspace(workspace);
		} else if (!workspaceCtx.settingsReady) {
			await workspaceCtx.loadSettings(workspaceId);
		}

		if (!workspaceCtx.settingsReady || workspaceCtx.currentWorkspace?.id !== workspaceId) {
			throw new Error(m.compose_load_workspace_settings_failed());
		}
	}

	async function loadAccounts(workspaceId = selectedWorkspaceId) {
		if (!workspaceId) return;
		const requestSequence = ++accountsRequestSequence;
		accountsLoading = true;
		try {
			const { data, error: err } = await client.GET('/accounts', {
				params: { query: { workspace_id: workspaceId } }
			});
			if (err) throw new Error(err.detail || m.compose_load_accounts_failed());
			if (requestSequence !== accountsRequestSequence || selectedWorkspaceId !== workspaceId)
				return;
			accounts = (data ?? []).filter((account) => account.is_active);
			normalizeSelectedAccounts();
			settingsByAccount = normalizeAllAccountSettings(settingsByAccount);
		} catch (err) {
			if (requestSequence === accountsRequestSequence && selectedWorkspaceId === workspaceId) {
				error = err instanceof Error ? err.message : m.compose_load_accounts_failed();
			}
		} finally {
			if (requestSequence === accountsRequestSequence && selectedWorkspaceId === workspaceId) {
				accountsLoading = false;
			}
		}
	}

	async function loadProviderReadiness(workspaceId = selectedWorkspaceId) {
		if (!workspaceId) return;
		const requestSequence = ++readinessRequestSequence;
		providerReadiness = [];
		providerReadinessWorkspaceId = '';
		providerReadinessLoading = true;
		providerReadinessError = '';
		try {
			const { data, error: err } = await client.GET('/provider-readiness', {
				params: { query: { workspace_id: workspaceId } }
			});
			if (err) throw new Error(err.detail || m.compose_load_readiness_failed());
			if (requestSequence !== readinessRequestSequence || selectedWorkspaceId !== workspaceId)
				return;
			providerReadiness = data?.providers ?? [];
			providerReadinessWorkspaceId = workspaceId;
		} catch (err) {
			if (requestSequence === readinessRequestSequence && selectedWorkspaceId === workspaceId) {
				providerReadinessError =
					err instanceof Error ? err.message : m.compose_load_readiness_failed();
				error = providerReadinessError;
			}
		} finally {
			if (requestSequence === readinessRequestSequence && selectedWorkspaceId === workspaceId) {
				providerReadinessLoading = false;
			}
		}
	}

	function hydrateInitialPublication(publication: Publication) {
		hydratedPublicationId = publication.id;
		publicationId = publication.id;
		selectedWorkspaceId = publication.workspace_id;
		fields = fieldsFromPublication(publication);
		media = (publication.media ?? []).map(mediaSummaryToFocusedMedia);
		selectedAccountIds = (publication.renditions ?? []).map(
			(rendition) => rendition.social_account_id
		);
		settingsByAccount = Object.fromEntries(
			(publication.renditions ?? []).map((rendition) => [
				rendition.social_account_id,
				{ ...(rendition.settings ?? {}) }
			])
		);
		thumbnailMediaId = youtubeThumbnailId(publication.renditions ?? []);
		hydrateSchedule(publication.scheduled_at);
	}

	function fieldsFromPublication(publication: Publication): FocusedComposerFields {
		const renditions = publication.renditions ?? [];
		const youtube = renditions.find(
			(rendition) => getPlatformKey(rendition.platform) === 'youtube'
		);
		const social = renditions.find((rendition) => getPlatformKey(rendition.platform) !== 'youtube');
		const base: FocusedComposerFields = {};
		if (publication.source_url) base.linkUrl = publication.source_url;
		if (mode === 'link_share') base.postText = social?.body || publication.source_text;
		if (mode === 'image_post' || mode === 'carousel' || mode === 'story') {
			base.caption = social?.body || youtube?.body || publication.source_text;
		}
		if (mode === 'short_video' || mode === 'long_video') {
			base.videoTitle = youtube?.title || publication.title;
			base.videoDescription = youtube?.description || youtube?.body || publication.source_text;
			base.caption = social?.body || '';
		}
		return base;
	}

	function hydrateSchedule(scheduledAt?: string) {
		if (!scheduledAt || scheduledAt === '0001-01-01T00:00:00Z') {
			selectedDate = undefined;
			selectedTime = null;
			return;
		}
		const schedule = workspaceScheduleFromISO(scheduledAt, scheduleTimezoneLabel);
		selectedDate = schedule?.date;
		selectedTime = schedule?.time ?? null;
	}

	async function changeWorkspace(workspaceId: string) {
		const changeSequence = ++workspaceChangeSequence;
		selectedWorkspaceId = workspaceId;
		resetWorkspaceScopedState();
		try {
			const settingsReady = await loadSchedulingSettings(workspaceId);
			if (
				!settingsReady ||
				changeSequence !== workspaceChangeSequence ||
				selectedWorkspaceId !== workspaceId
			) {
				return;
			}
			await Promise.all([loadAccounts(workspaceId), loadProviderReadiness(workspaceId)]);
		} catch (err) {
			if (changeSequence === workspaceChangeSequence && selectedWorkspaceId === workspaceId) {
				error = err instanceof Error ? err.message : m.compose_load_workspace_settings_failed();
			}
		}
	}

	function resetWorkspaceScopedState() {
		accountsRequestSequence += 1;
		readinessRequestSequence += 1;
		mediaUploadRequestSequence += 1;
		thumbnailUploadRequestSequence += 1;
		publicationId = '';
		accounts = [];
		selectedAccountIds = [];
		providerReadiness = [];
		providerReadinessWorkspaceId = '';
		providerReadinessLoading = false;
		providerReadinessError = '';
		customizeAccountId = '';
		media = [];
		thumbnailMedia = null;
		thumbnailMediaId = '';
		settingsByAccount = {};
		selectedDate = undefined;
		selectedTime = null;
		showSchedulePopover = false;
		validationIssues = [];
		schedulingSettings = defaultFocusedSchedulingSettings();
		schedulingSettingsWorkspaceId = '';
		accountsLoading = false;
		uploading = false;
		uploadingThumbnail = false;
		error = '';
		success = '';
	}

	function selectAllAccounts() {
		selectedAccountIds = compatibleAccounts.map((account) => account.id);
		settingsByAccount = normalizeAllAccountSettings(settingsByAccount);
		validationIssues = [];
	}

	function clearAllAccounts() {
		selectedAccountIds = [];
		customizeAccountId = '';
		validationIssues = [];
	}

	function normalizeSelectedAccounts() {
		const compatible = accounts.filter(isAccountCompatible).map((account) => account.id);
		const preserved = selectedAccountIds.filter((id) => compatible.includes(id));
		if (preserved.length > 0) {
			selectedAccountIds = preserved;
		} else if (!initialPublication || selectedAccountIds.length === 0) {
			selectedAccountIds = compatible;
		} else {
			selectedAccountIds = [];
		}
		if (customizeAccountId && !selectedAccountIds.includes(customizeAccountId))
			customizeAccountId = '';
	}

	function toggleAccount(account: SocialAccount) {
		if (!isAccountCompatible(account)) return;
		selectedAccountIds = selectedAccountIds.includes(account.id)
			? selectedAccountIds.filter((id) => id !== account.id)
			: [...selectedAccountIds, account.id];
		settingsByAccount = normalizeAllAccountSettings(settingsByAccount);
		validationIssues = [];
	}

	function isAccountCompatible(account: SocialAccount): boolean {
		return isAccountCompatibleWithMode(mode, account, capabilities);
	}

	function capabilityForAccount(account: SocialAccount): Capability | null {
		const provider = getPlatformKey(account.platform);
		return (
			capabilities.find(
				(capability) => capability.provider === provider && capability.profile === mode
			) ?? null
		);
	}

	function readinessForAccount(account: SocialAccount): ProviderReadinessItem | null {
		const provider = getPlatformKey(account.platform);
		return providerReadiness.find((item) => item.provider === provider) ?? null;
	}

	function readinessBlockers(account: SocialAccount): string[] {
		return readinessForAccount(account)?.blocking_issues ?? [];
	}

	function accountBlockerText(account: SocialAccount): string {
		return readinessBlockers(account)
			.map((item) => item.replaceAll('_', ' '))
			.join(', ');
	}

	function updateField(key: FocusedFieldKey, value: string) {
		fields = { ...fields, [key]: value };
		validationIssues = [];
	}

	function fieldValue(key: FocusedFieldKey): string {
		return fields[key] ?? '';
	}

	async function handleFiles(files: FileList | null) {
		if (!files || !selectedWorkspaceId) return;
		const selected = Array.from(files).filter(isSupportedMediaFile);
		if (selected.length === 0) return;
		const workspaceId = selectedWorkspaceId;
		const requestSequence = ++mediaUploadRequestSequence;
		uploading = true;
		error = '';
		try {
			for (const file of selected) {
				const uploaded = await uploadMediaFile({ workspaceId, file });
				if (requestSequence !== mediaUploadRequestSequence || selectedWorkspaceId !== workspaceId) {
					return;
				}
				media = [
					...media,
					{
						id: uploaded.id,
						mime_type: uploaded.mime_type,
						url: uploaded.url,
						size: uploaded.size,
						filename: file.name
					}
				];
			}
			validationIssues = [];
		} catch (err) {
			if (requestSequence === mediaUploadRequestSequence && selectedWorkspaceId === workspaceId) {
				error = err instanceof Error ? err.message : m.compose_upload_failed();
			}
		} finally {
			if (requestSequence === mediaUploadRequestSequence && selectedWorkspaceId === workspaceId) {
				uploading = false;
			}
		}
	}

	async function handleThumbnail(files: FileList | null) {
		if (!files?.[0] || !selectedWorkspaceId) return;
		const file = files[0];
		if (!file.type.startsWith('image/')) {
			error = m.compose_choose_thumbnail();
			return;
		}
		const workspaceId = selectedWorkspaceId;
		const requestSequence = ++thumbnailUploadRequestSequence;
		uploadingThumbnail = true;
		error = '';
		try {
			const uploaded = await uploadMediaFile({ workspaceId, file });
			if (
				requestSequence !== thumbnailUploadRequestSequence ||
				selectedWorkspaceId !== workspaceId
			) {
				return;
			}
			thumbnailMedia = {
				id: uploaded.id,
				mime_type: uploaded.mime_type,
				url: uploaded.url,
				size: uploaded.size,
				filename: file.name,
				role: 'thumbnail'
			};
			thumbnailMediaId = uploaded.id;
			validationIssues = [];
		} catch (err) {
			if (
				requestSequence === thumbnailUploadRequestSequence &&
				selectedWorkspaceId === workspaceId
			) {
				error = err instanceof Error ? err.message : m.compose_thumbnail_upload_failed();
			}
		} finally {
			if (
				requestSequence === thumbnailUploadRequestSequence &&
				selectedWorkspaceId === workspaceId
			) {
				uploadingThumbnail = false;
			}
		}
	}

	function removeMedia(mediaId: string) {
		media = media.filter((item) => item.id !== mediaId);
		validationIssues = [];
	}

	function clearThumbnail() {
		thumbnailMedia = null;
		thumbnailMediaId = '';
		validationIssues = [];
	}

	function getScheduledAt(): string | undefined {
		if (!selectedWorkspaceSettingsReady || !selectedDate || !selectedTime) return undefined;
		return workspaceScheduleToISO(selectedDate, selectedTime, scheduleTimezoneLabel);
	}

	function scheduleLabel(): string {
		if (!selectedDate || !selectedTime) return m.compose_schedule();
		const date = selectedDate.toDate(scheduleTimezoneLabel);
		return `${date.toLocaleDateString(getLocaleTag(), {
			month: 'short',
			day: 'numeric',
			timeZone: scheduleTimezoneLabel
		})} ${selectedTime}`;
	}

	function clearSchedule() {
		selectedDate = undefined;
		selectedTime = null;
		showSchedulePopover = false;
	}

	function formBlockers(): string[] {
		const blockers: string[] = [];
		if (!selectedWorkspaceId) blockers.push(m.compose_choose_workspace_blocker());
		if (selectedAccounts.length === 0) blockers.push(m.compose_choose_account_blocker());
		for (const field of roleFields) {
			if (field.required && !fieldValue(field.key).trim()) {
				blockers.push(m.compose_field_required({ field: field.label }));
			}
		}
		const mediaMin = Math.max(
			0,
			...selectedCapabilities.map((capability) => capability.media.min_count)
		);
		if (mediaMin > 0 && media.length < mediaMin) {
			blockers.push(
				mediaMin === 1
					? m.compose_add_media_singular()
					: m.compose_add_media_plural({ count: mediaMin })
			);
		}
		for (const account of selectedAccounts) {
			const blockersForAccount = readinessBlockers(account);
			if (blockersForAccount.length > 0) {
				blockers.push(
					m.compose_provider_blocked({
						platform: getPlatformName(account.platform),
						reason: accountBlockerText(account)
					})
				);
			}
		}
		return blockers;
	}

	function normalizeAllAccountSettings(
		current: Record<string, Record<string, unknown>>
	): Record<string, Record<string, unknown>> {
		const next = { ...current };
		for (const account of selectedAccounts) {
			next[account.id] = normalizeSettings(account, next[account.id] ?? {});
		}
		return next;
	}

	function normalizeSettings(
		account: SocialAccount,
		current: Record<string, unknown>
	): Record<string, unknown> {
		const next = { ...current };
		for (const field of capabilityForAccount(account)?.settings ?? []) {
			if (next[field.key] !== undefined) continue;
			if (field.key === 'privacy') next[field.key] = 'private';
			else if (field.key === 'content_posting_method') next[field.key] = 'DIRECT_POST';
			else if (field.key === 'privacy_level') next[field.key] = 'SELF_ONLY';
			else if (field.key === 'post_type') next[field.key] = mode === 'story' ? 'story' : 'post';
			else if (field.key === 'is_reel') next[field.key] = mode === 'short_video';
			else if (field.type === 'boolean') next[field.key] = false;
			else next[field.key] = '';
		}
		return next;
	}

	function settingsForAccount(account: SocialAccount): Record<string, unknown> {
		return normalizeSettings(account, settingsByAccount[account.id] ?? {});
	}

	function updateAccountSetting(account: SocialAccount, key: string, value: unknown) {
		const current = settingsForAccount(account);
		settingsByAccount = {
			...settingsByAccount,
			[account.id]: { ...current, [key]: value }
		};
		validationIssues = [];
	}

	function visibleSettings(account: SocialAccount): SettingField[] {
		const provider = getPlatformKey(account.platform);
		return (capabilityForAccount(account)?.settings ?? []).filter((field) => {
			if (
				provider === 'youtube' &&
				['title', 'description', 'thumbnail_media_id'].includes(field.key)
			) {
				return false;
			}
			if (
				mode === 'link_share' &&
				['url', 'link_url', 'link_title', 'link_description'].includes(field.key) &&
				!field.required
			) {
				return false;
			}
			return true;
		});
	}

	function selectedSettingsInput(): Record<string, Record<string, unknown>> {
		return Object.fromEntries(
			selectedAccounts.map((account) => [account.id, settingsForAccount(account)])
		);
	}

	function publicationPayload() {
		return buildFocusedPublicationPayload({
			mode,
			workspaceId: selectedWorkspaceId,
			accounts: selectedAccounts.map((account) => ({
				id: account.id,
				platform: account.platform,
				account_username: account.account_username
			})),
			fields,
			mediaIds: media.map((item) => item.id),
			scheduledAt: getScheduledAt(),
			thumbnailMediaId,
			settingsByAccount: selectedSettingsInput()
		});
	}

	async function persistPublication(): Promise<string> {
		const payload = publicationPayload();
		if (publicationId) {
			const { error: updateError } = await client.PUT('/publications/{id}', {
				params: { path: { id: publicationId } },
				body: {
					title: payload.title,
					content_profile: payload.content_profile,
					source_text: payload.source_text,
					source_url: payload.source_url ?? '',
					...(payload.scheduled_at
						? { scheduled_at: payload.scheduled_at }
						: { clear_schedule: true }),
					metadata: payload.metadata
				}
			});
			if (updateError) throw new Error(updateError.detail || m.compose_save_publication_failed());
			const { error: renditionError } = await client.PUT('/publications/{id}/renditions', {
				params: { path: { id: publicationId } },
				body: { renditions: payload.renditions }
			});
			if (renditionError) throw new Error(renditionError.detail || m.compose_save_outputs_failed());
			return publicationId;
		}

		const { data, error: createError } = await client.POST('/publications', {
			body: payload
		});
		if (createError) throw new Error(createError.detail || m.compose_create_publication_failed());
		publicationId = data.id;
		return data.id;
	}

	async function validatePublication(id: string): Promise<ValidationIssue[]> {
		const { data, error: err } = await client.POST('/publications/{id}/validate', {
			params: { path: { id } }
		});
		if (err) throw new Error(err.detail || m.compose_validation_failed());
		validationIssues = data?.issues ?? [];
		return validationIssues;
	}

	async function runAction(action: 'draft' | 'validate' | 'schedule' | 'publish') {
		if (action === 'schedule' && !selectedWorkspaceSettingsReady) {
			error = m.compose_load_workspace_settings_failed();
			success = '';
			return;
		}
		if (action !== 'draft' && !selectedProviderReadinessReady) {
			error = providerReadinessError || m.compose_load_readiness_failed();
			success = '';
			return;
		}
		if (action !== 'draft' && localBlockers.length > 0) {
			error = localBlockers[0];
			success = '';
			return;
		}
		if (action === 'schedule') {
			const scheduledAt = getScheduledAt();
			if (!scheduledAt) {
				error = m.compose_choose_schedule();
				success = '';
				return;
			}
			if (!isFutureSchedule(scheduledAt)) {
				error = m.compose_schedule_future();
				success = '';
				return;
			}
		}
		saving = true;
		error = '';
		success = '';
		try {
			const id = await persistPublication();
			if (action === 'validate') {
				await validatePublication(id);
				success = m.compose_validation_complete();
			} else if (action === 'schedule') {
				const issues = await validatePublication(id);
				if (issues.some((issue) => issue.severity === 'error')) {
					throw new Error(m.compose_fix_before_scheduling());
				}
				const { data, error: scheduleError } = await client.POST('/publications/{id}/schedule', {
					params: { path: { id } }
				});
				if (scheduleError) throw new Error(scheduleError.detail || m.compose_schedule_failed());
				success = data?.message ?? m.compose_publication_scheduled();
			} else if (action === 'publish') {
				const issues = await validatePublication(id);
				if (issues.some((issue) => issue.severity === 'error')) {
					throw new Error(m.compose_fix_before_publishing());
				}
				const { data, error: publishError } = await client.POST('/publications/{id}/publish-now', {
					params: { path: { id } }
				});
				if (publishError) throw new Error(publishError.detail || m.compose_publish_failed());
				success = data?.message ?? m.compose_publication_queued();
			} else {
				success = isEditMode ? m.compose_changes_saved() : m.compose_draft_saved();
			}
			ui.triggerRefresh();
			if (isEditMode && action !== 'validate') onSuccess?.();
		} catch (err) {
			error = err instanceof Error ? err.message : m.compose_action_failed();
		} finally {
			saving = false;
		}
	}

	function mediaSummaryToFocusedMedia(item: MediaSummary): FocusedMedia {
		return {
			id: item.id,
			mime_type: item.mime_type,
			url: item.url,
			size: item.size,
			filename: item.original_filename || item.id,
			role: item.role
		};
	}

	function youtubeThumbnailId(renditions: Rendition[]): string {
		const youtube = renditions.find(
			(rendition) => getPlatformKey(rendition.platform) === 'youtube'
		);
		const value = youtube?.settings?.thumbnail_media_id;
		return typeof value === 'string' ? value : '';
	}

	function isVideo(item: FocusedMedia): boolean {
		return item.mime_type.startsWith('video/');
	}

	function isImage(item: FocusedMedia): boolean {
		return item.mime_type.startsWith('image/');
	}

	function previewSrc(item: FocusedMedia): string {
		return getAuthenticatedMediaByID(item.id) || item.url;
	}

	function accountLabel(account: SocialAccount): string {
		return account.account_username || account.slug || getPlatformName(account.platform);
	}

	function settingAsString(account: SocialAccount, key: string): string {
		const value = settingsForAccount(account)[key];
		return typeof value === 'string' || typeof value === 'number' ? String(value) : '';
	}

	function settingAsBoolean(account: SocialAccount, key: string): boolean {
		return Boolean(settingsForAccount(account)[key]);
	}
</script>

<div class="flex min-h-0 flex-1 flex-col bg-background" data-testid="focused-composer">
	{#if !loading}
		<div class="border-b bg-background px-3 py-2 md:px-4 md:py-3">
			<div class="flex flex-wrap items-center justify-between gap-2">
				<div class="flex min-w-0 flex-wrap items-center gap-2">
					{#if modeControl}
						{@render modeControl()}
					{/if}
					{#if onCancel}
						<Button variant="ghost" size="sm" class="h-8" onclick={onCancel} disabled={saving}>
							{m.common_cancel()}
						</Button>
					{/if}
					{#if accounts.length > 0}
						<ComposerAccountMenu
							{accounts}
							{selectedAccountIds}
							compatibleAccountIds={compatibleAccounts.map((account) => account.id)}
							triggerLabel={m.compose_target_accounts()}
							triggerClass="h-11 md:h-8"
							description={m.compose_accounts_compatible({ format: modeMeta.label })}
							onToggle={toggleAccount}
							onSelectAll={selectAllAccounts}
							onClearAll={clearAllAccounts}
						/>
					{/if}
				</div>

				<div class="flex min-w-0 flex-wrap items-center justify-end gap-1.5 md:gap-2">
					<Button
						variant="ghost"
						size="sm"
						class="h-8 gap-1.5"
						disabled={!canSaveDraft}
						onclick={() => runAction('draft')}
					>
						{#if saving}
							<LoaderIcon class="h-3.5 w-3.5 animate-spin" />
						{:else}
							<SaveIcon class="h-3.5 w-3.5" />
						{/if}
						{isEditMode ? m.compose_save_changes() : m.compose_save_draft()}
					</Button>
					<Popover.Root bind:open={showSchedulePopover}>
						<Popover.Trigger>
							{#snippet child({ props })}
								<Button
									{...props}
									size="sm"
									class="h-8 gap-1.5"
									disabled={saving || !selectedWorkspaceSettingsReady}
								>
									<CalendarClockIcon class="h-3.5 w-3.5" />
									{scheduleLabel()}
								</Button>
							{/snippet}
						</Popover.Trigger>
						<Popover.Content class="w-auto max-w-[calc(100vw-2rem)] p-0" align="end">
							<div class="p-3">
								<Calendar
									type="single"
									bind:value={selectedDate}
									minValue={workspaceClock(scheduleTimezoneLabel).date}
									class="bg-transparent p-0 [--cell-size:--spacing(8)]"
									weekdayFormat="short"
									weekStartsOn={schedulingSettings.weekStartsOn}
								/>
								<div class="mt-3 max-h-48 overflow-y-auto">
									<div class="grid grid-cols-3 gap-1.5 sm:grid-cols-4">
										{#each timeSlots as time (time)}
											<Button
												variant={selectedTime === time ? 'default' : 'outline'}
												size="sm"
												class="h-8 text-xs"
												onclick={() => {
													if (!selectedDate) {
														selectedDate = workspaceClock(scheduleTimezoneLabel).date;
													}
													selectedTime = time;
												}}
											>
												{time}
											</Button>
										{/each}
									</div>
								</div>
								{#if selectedDate || selectedTime}
									<div class="mt-3 flex gap-2 border-t pt-3">
										<Button variant="ghost" size="sm" class="flex-1 text-xs" onclick={clearSchedule}
											>{m.compose_clear()}</Button
										>
										<Button
											size="sm"
											class="flex-1 text-xs"
											disabled={!canSchedule || !getScheduledAt()}
											onclick={() => {
												showSchedulePopover = false;
												runAction('schedule');
											}}
										>
											{m.compose_schedule()}
										</Button>
									</div>
								{/if}
							</div>
						</Popover.Content>
					</Popover.Root>
					<Button
						variant="outline"
						size="sm"
						class="h-8 gap-1.5"
						disabled={!canQueue}
						onclick={() => runAction('publish')}
					>
						<SendIcon class="h-3.5 w-3.5" />
						{m.compose_publish_now()}
					</Button>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="min-h-0 flex-1 overflow-y-auto" aria-busy="true">
			<PageLoading layout="composer" label={m.common_loading()} />
		</div>
	{:else}
		<div class="min-h-0 flex-1 overflow-y-auto">
			<div class="mx-auto w-full max-w-3xl space-y-6 px-4 py-5 md:px-6">
				{#if error}
					<InlineNotice tone="error" message={error} />
				{/if}
				{#if success}
					<InlineNotice tone="success" message={success} />
				{/if}

				{#if accounts.length === 0}
					<div class="rounded-lg border bg-muted/20 px-4 py-3 text-sm text-muted-foreground">
						{m.compose_connect_compatible()}
					</div>
				{/if}

				<section class="flex flex-col gap-5">
					<div class="{modeMeta.mediaFirst ? 'order-1' : 'order-2'} space-y-3">
						<div class="rounded-md border border-dashed bg-muted/20 p-4">
							<div class="flex flex-wrap items-center gap-3">
								<label
									class="inline-flex h-11 cursor-pointer items-center gap-2 rounded-md border bg-background px-3 text-sm font-medium"
								>
									{#if uploading}
										<LoaderIcon class="h-4 w-4 animate-spin" />
									{:else}
										<UploadIcon class="h-4 w-4" />
									{/if}
									{m.compose_upload_media()}
									<input
										class="sr-only"
										type="file"
										multiple
										accept="image/*,video/*"
										disabled={uploading || !selectedWorkspaceId}
										onchange={(event) => handleFiles(event.currentTarget.files)}
									/>
								</label>
								<p class="text-sm text-muted-foreground">
									{mode === 'long_video' ? m.compose_add_long_video() : m.compose_add_media()}
								</p>
							</div>
							{#if media.length > 0}
								<div class="mt-4 grid gap-2 sm:grid-cols-2">
									{#each media as item (item.id)}
										<div class="group relative overflow-hidden rounded-md border bg-background">
											{#if isImage(item)}
												<img
													src={previewSrc(item)}
													alt={item.filename || m.compose_uploaded_media()}
													class="aspect-video w-full object-cover"
												/>
											{:else if isVideo(item)}
												<video
													src={previewSrc(item)}
													class="aspect-video w-full object-cover"
													controls
													muted
													playsinline
												></video>
											{:else}
												<div
													class="flex aspect-video items-center justify-center text-sm text-muted-foreground"
												>
													{item.mime_type}
												</div>
											{/if}
											<div
												class="flex items-center justify-between gap-2 border-t bg-background px-2 py-1.5 text-xs"
											>
												<span class="min-w-0 truncate">{item.filename || item.id}</span>
												<button
													type="button"
													class="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
													aria-label={m.compose_remove_media()}
													onclick={() => removeMedia(item.id)}
												>
													<XIcon class="h-3.5 w-3.5" />
												</button>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>

						{#if hasYouTubeTarget && (mode === 'short_video' || mode === 'long_video')}
							<div class="rounded-md border bg-background p-4">
								<div class="flex flex-wrap items-center gap-3">
									<label
										class="inline-flex h-11 cursor-pointer items-center gap-2 rounded-md border bg-background px-3 text-sm font-medium"
									>
										{#if uploadingThumbnail}
											<LoaderIcon class="h-4 w-4 animate-spin" />
										{:else}
											<ImagePlusIcon class="h-4 w-4" />
										{/if}
										{m.compose_thumbnail()}
										<input
											class="sr-only"
											type="file"
											accept="image/*"
											disabled={uploadingThumbnail || !selectedWorkspaceId}
											onchange={(event) => handleThumbnail(event.currentTarget.files)}
										/>
									</label>
									<p class="text-sm text-muted-foreground">{m.compose_thumbnail_youtube()}</p>
									{#if thumbnailMediaId}
										<Button variant="ghost" size="sm" class="h-8 text-xs" onclick={clearThumbnail}>
											{m.compose_clear()}
										</Button>
									{/if}
								</div>
								{#if thumbnailMedia}
									<img
										src={previewSrc(thumbnailMedia)}
										alt={thumbnailMedia.filename || m.compose_thumbnail()}
										class="mt-3 aspect-video max-h-48 rounded-md border object-cover"
									/>
								{:else if thumbnailMediaId}
									<p class="mt-2 text-xs text-muted-foreground">
										{m.compose_existing_thumbnail({ id: thumbnailMediaId })}
									</p>
								{/if}
							</div>
						{/if}
					</div>

					<div class="{modeMeta.mediaFirst ? 'order-2' : 'order-1'} space-y-4">
						{#each roleFields as field (field.key)}
							<div>
								<label class="text-sm font-medium" for="focused-field-{field.key}">
									{field.label}
								</label>
								{#if field.type === 'textarea'}
									<Textarea
										id="focused-field-{field.key}"
										data-testid="composer-field-{field.key}"
										class="mt-1 min-h-32 resize-y text-base"
										rows={field.rows ?? 6}
										value={fieldValue(field.key)}
										oninput={(event) => updateField(field.key, event.currentTarget.value)}
									/>
								{:else}
									<Input
										id="focused-field-{field.key}"
										data-testid="composer-field-{field.key}"
										class="mt-1"
										type={field.type === 'url' ? 'url' : 'text'}
										value={fieldValue(field.key)}
										oninput={(event) => updateField(field.key, event.currentTarget.value)}
									/>
								{/if}
								<p class="mt-1 text-xs text-muted-foreground">{field.hint}</p>
							</div>
						{/each}
						{#if roleFields.length === 0}
							<div class="rounded-md border bg-muted/20 px-3 py-2 text-sm text-muted-foreground">
								{m.compose_media_first_notice()}
							</div>
						{/if}
					</div>
				</section>

				<section class="space-y-3 border-t pt-5">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<div>
							<h2 class="text-sm font-semibold">{m.compose_platform_settings()}</h2>
							<p class="text-xs text-muted-foreground">
								{m.compose_platform_settings_body()}
							</p>
						</div>
					</div>
					<div class="flex flex-wrap gap-2">
						{#each selectedAccounts as account (account.id)}
							<Button
								type="button"
								variant={customizeAccountId === account.id ? 'default' : 'outline'}
								size="sm"
								class="h-8 gap-1.5 text-xs"
								onclick={() =>
									(customizeAccountId = customizeAccountId === account.id ? '' : account.id)}
							>
								<Settings2Icon class="h-3.5 w-3.5" />
								{m.compose_customize_output({ platform: getPlatformName(account.platform) })}
							</Button>
						{/each}
					</div>

					{#if customizeAccountId}
						{@const account = selectedAccounts.find((item) => item.id === customizeAccountId)}
						{#if account}
							{@const settings = visibleSettings(account)}
							<div class="rounded-md border bg-background p-4">
								<div class="mb-3 flex items-center gap-2">
									<PlatformIcon platform={account.platform} class="h-4 w-4" />
									<h3 class="text-sm font-semibold">
										{m.compose_customize_output({ platform: getPlatformName(account.platform) })}
									</h3>
								</div>
								{#if settings.length === 0}
									<p class="text-sm text-muted-foreground">
										{m.compose_platform_fields_main({
											platform: getPlatformName(account.platform)
										})}
									</p>
								{:else}
									<div class="grid gap-3 sm:grid-cols-2">
										{#each settings as setting (setting.key)}
											<div class={setting.type === 'textarea' ? 'sm:col-span-2' : ''}>
												<label class="text-xs font-medium" for="setting-{account.id}-{setting.key}">
													{setting.label}
												</label>
												{#if setting.type === 'boolean'}
													<label class="mt-2 flex items-center gap-2 text-sm">
														<input
															type="checkbox"
															class="size-4 rounded border"
															checked={settingAsBoolean(account, setting.key)}
															onchange={(event) =>
																updateAccountSetting(
																	account,
																	setting.key,
																	event.currentTarget.checked
																)}
														/>
														<span>{setting.label}</span>
													</label>
												{:else if setting.type === 'select'}
													<select
														id="setting-{account.id}-{setting.key}"
														class="mt-1 h-9 w-full rounded-md border bg-background px-2 text-sm"
														value={settingAsString(account, setting.key)}
														onchange={(event) =>
															updateAccountSetting(account, setting.key, event.currentTarget.value)}
													>
														{#each setting.options ?? [] as option (option)}
															<option value={option}>{option}</option>
														{/each}
													</select>
												{:else if setting.type === 'textarea'}
													<Textarea
														id="setting-{account.id}-{setting.key}"
														class="mt-1 min-h-24"
														value={settingAsString(account, setting.key)}
														oninput={(event) =>
															updateAccountSetting(account, setting.key, event.currentTarget.value)}
													/>
												{:else}
													<Input
														id="setting-{account.id}-{setting.key}"
														class="mt-1"
														type={setting.type === 'number' ? 'number' : 'text'}
														value={settingAsString(account, setting.key)}
														oninput={(event) =>
															updateAccountSetting(account, setting.key, event.currentTarget.value)}
													/>
												{/if}
												{#if setting.help}
													<p class="mt-1 text-xs text-muted-foreground">{setting.help}</p>
												{/if}
											</div>
										{/each}
									</div>
								{/if}
							</div>
						{/if}
					{/if}
				</section>

				{#if localBlockers.length > 0 || blockingIssues.length > 0 || warningIssues.length > 0}
					<section
						class="space-y-2 border-t pt-5 text-sm"
						aria-label={m.compose_publishing_issues()}
					>
						<h2 class="font-semibold">{m.compose_check_before_publishing()}</h2>
						{#each localBlockers as blocker (blocker)}
							<div class="flex gap-2 text-destructive">
								<AlertCircleIcon class="mt-0.5 size-4 shrink-0" /><span>{blocker}</span>
							</div>
						{/each}
						{#each blockingIssues as issue (`error-${issue.code}-${issue.field}-${issue.media_id}`)}
							<div class="flex gap-2 text-destructive">
								<AlertCircleIcon class="mt-0.5 size-4 shrink-0" /><span>{issue.message}</span>
							</div>
						{/each}
						{#each warningIssues as issue (`warning-${issue.code}-${issue.field}-${issue.media_id}`)}
							<div class="flex gap-2 text-amber-700 dark:text-amber-300">
								<AlertCircleIcon class="mt-0.5 size-4 shrink-0" /><span>{issue.message}</span>
							</div>
						{/each}
					</section>
				{/if}
			</div>
		</div>
	{/if}
</div>
