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
	import DestinationSettingsDialog from './destination-settings-dialog.svelte';
	import DestructiveConfirmDialog from './destructive-confirm-dialog.svelte';
	import InlineNotice from './inline-notice.svelte';
	import PageLoading from './page-loading.svelte';
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
		type FocusedFieldKey,
		type FocusedMediaInput,
		type FocusedSegmentInput,
		type ResolvedComposerTarget
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
	import PlusIcon from 'lucide-svelte/icons/plus';
	import Trash2Icon from 'lucide-svelte/icons/trash-2';
	import UploadIcon from 'lucide-svelte/icons/upload';
	import XIcon from 'lucide-svelte/icons/x';
	import { m } from '$lib/paraglide/messages';

	type Capability = components['schemas']['Capability'];
	type Publication = components['schemas']['PublicationResponse'];
	type Rendition = components['schemas']['RenditionResponse'];
	type MediaSummary = components['schemas']['MediaSummary'];
	type ValidationIssue = components['schemas']['ValidationIssue'];
	type ProviderReadinessItem = components['schemas']['ProviderReadinessItem'];
	type SettingDefinition = components['schemas']['SettingDefinition'];
	type ResolvedAccountCapability = components['schemas']['ResolvedAccountCapability'];
	type DestinationOption = components['schemas']['DestinationOption'];

	interface FocusedMedia {
		id: string;
		mime_type: string;
		url: string;
		size?: number;
		filename?: string;
		role?: string;
		altText?: string;
		settings?: Record<string, unknown>;
		settingsByAccount?: Record<string, Record<string, unknown>>;
	}

	interface Props {
		mode: ComposerModeKey;
		initialPublication?: Publication | null;
		initialScheduleDate?: string | null;
		initialWorkspaceId?: string | null;
		onSuccess?: () => void;
		onCancel?: () => void;
		modeControl?: Snippet;
	}

	let {
		mode,
		initialPublication = null,
		initialScheduleDate = null,
		initialWorkspaceId = null,
		onSuccess,
		onCancel,
		modeControl
	}: Props = $props();

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
	let segments = $state<FocusedSegmentInput[]>(emptySegmentsForMode('post'));
	let segmentSettingsByAccount = $state<Record<string, Record<string, unknown>>>({});
	let activeSettingsSegmentId = $state('segment-1');
	let thumbnailMedia = $state<FocusedMedia | null>(null);
	let thumbnailMediaId = $state('');
	let settingsByAccount = $state<Record<string, Record<string, unknown>>>({});
	let settingsDialogOpen = $state(false);
	let settingsAccountId = $state('');
	let deleteDestinationDialogOpen = $state(false);
	let deleteDestinationAccount = $state<SocialAccount | null>(null);
	let destinationOptionsByAccount = $state<Record<string, Record<string, DestinationOption[]>>>({});
	let destinationOptionsErrors = $state<Record<string, string>>({});
	let destinationOptionsLoadingAccountId = $state('');
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
	let destinationOptionsRequestSequence = 0;
	let capabilityResolveRequestSequence = 0;
	let publicationContextRequestId = '';
	let resolvedCapabilities = $state<Record<string, ResolvedAccountCapability>>({});
	let capabilityResolveLoading = $state(false);
	let capabilityResolveError = $state('');

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
	const settingsAccount = $derived(
		accounts.find((account) => account.id === settingsAccountId) ?? null
	);
	const settingsDialogFields = $derived(settingsAccount ? visibleSettings(settingsAccount) : []);
	const settingsDialogValues = $derived(
		settingsAccount ? dialogSettingsForAccount(settingsAccount) : {}
	);
	const settingsDialogMedia = $derived(mediaForSettingsDialog());
	const settingsDialogMediaValues = $derived(
		settingsAccount ? mediaSettingsForDialog(settingsAccount) : {}
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
		canSaveDraft &&
			selectedProviderReadinessReady &&
			!capabilityResolveLoading &&
			localBlockers.length === 0
	);
	const accountSummaries = $derived(
		Object.fromEntries(
			accounts.map((account) => [
				account.id,
				resolvedCapabilities[account.id]?.label ?? getPlatformName(account.platform)
			])
		)
	);
	const warningAccountIds = $derived(
		Object.values(resolvedCapabilities)
			.filter((capability) => !capability.compatible || (capability.issues ?? []).length > 0)
			.map((capability) => capability.account_id)
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
		segments = emptySegmentsForMode(mode);
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
			selectedWorkspaceId =
				selectedWorkspaceId || initialWorkspaceId || workspaceCtx.currentWorkspace?.id || '';
			if (selectedWorkspaceId) {
				if (schedulingSettingsWorkspaceId !== selectedWorkspaceId) {
					await loadSchedulingSettings(selectedWorkspaceId);
				}
				await Promise.all([
					loadAccounts(selectedWorkspaceId),
					loadProviderReadiness(selectedWorkspaceId)
				]);
				applyInitialScheduleDate();
				await resolveSelectedCapabilities();
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
			await resolveSelectedCapabilities();
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
		segments = (publication.segments ?? []).map((segment) => ({
			id: segment.id,
			content: segment.body,
			title: segment.title,
			description: segment.description,
			url: segment.url,
			media: (segment.media ?? []).map((item) => ({
				id: item.id,
				mimeType: item.mime_type,
				role: item.role,
				altText: item.alt_text,
				settings: item.settings
			})),
			settingsByAccount: {}
		}));
		if (segments.length === 0) {
			segments = emptySegmentsForMode(mode);
			segments[0].content = publication.source_text;
		}
		activeSettingsSegmentId = segments[0].id;
		selectedAccountIds = (publication.renditions ?? []).map(
			(rendition) => rendition.social_account_id
		);
		settingsByAccount = Object.fromEntries(
			(publication.renditions ?? []).map((rendition) => [
				rendition.social_account_id,
				{ ...(rendition.settings ?? {}) }
			])
		);
		for (const rendition of publication.renditions ?? []) {
			const hydratedSegments = new Set<string>();
			for (const segment of rendition.segments ?? []) {
				const canonical = segments.find((item) => item.id === segment.publication_segment_id);
				if (!canonical) continue;
				if (hydratedSegments.has(canonical.id)) {
					if (segment.body.trim()) {
						canonical.settingsByAccount = {
							...(canonical.settingsByAccount ?? {}),
							[rendition.social_account_id]: {
								...(canonical.settingsByAccount?.[rendition.social_account_id] ?? {}),
								first_comment: segment.body
							}
						};
					}
					continue;
				}
				hydratedSegments.add(canonical.id);
				canonical.settingsByAccount = {
					...(canonical.settingsByAccount ?? {}),
					[rendition.social_account_id]: { ...(segment.settings ?? {}) }
				};
				for (const renditionMedia of segment.media ?? []) {
					const accountMediaSettings = {
						...(renditionMedia.settings ?? {}),
						...(renditionMedia.alt_text ? { alt_text: renditionMedia.alt_text } : {}),
						...(renditionMedia.thumbnail_timestamp_ms
							? { thumbnail_timestamp_ms: renditionMedia.thumbnail_timestamp_ms }
							: {})
					};
					setMediaAccountSettings(
						canonical.media,
						renditionMedia.id,
						rendition.social_account_id,
						accountMediaSettings
					);
					setMediaAccountSettings(
						media,
						renditionMedia.id,
						rendition.social_account_id,
						accountMediaSettings
					);
				}
			}
		}
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
		if (mode === 'post') {
			base.postText = social?.body || publication.source_text;
			base.linkUrl = publication.source_url;
		}
		if (mode === 'story') {
			base.caption = social?.body || youtube?.body || publication.source_text;
		}
		if (mode === 'short_video' || mode === 'video') {
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
			await resolveSelectedCapabilities();
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
		destinationOptionsRequestSequence += 1;
		publicationId = '';
		accounts = [];
		selectedAccountIds = [];
		providerReadiness = [];
		providerReadinessWorkspaceId = '';
		providerReadinessLoading = false;
		providerReadinessError = '';
		settingsDialogOpen = false;
		settingsAccountId = '';
		destinationOptionsByAccount = {};
		destinationOptionsErrors = {};
		destinationOptionsLoadingAccountId = '';
		media = [];
		thumbnailMedia = null;
		thumbnailMediaId = '';
		settingsByAccount = {};
		segmentSettingsByAccount = {};
		segments = emptySegmentsForMode(mode);
		activeSettingsSegmentId = 'segment-1';
		resolvedCapabilities = {};
		capabilityResolveLoading = false;
		capabilityResolveError = '';
		capabilityResolveRequestSequence += 1;
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
		void resolveSelectedCapabilities();
	}

	function clearAllAccounts() {
		selectedAccountIds = [];
		settingsDialogOpen = false;
		settingsAccountId = '';
		validationIssues = [];
		resolvedCapabilities = {};
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
		if (settingsAccountId && !selectedAccountIds.includes(settingsAccountId)) {
			settingsDialogOpen = false;
			settingsAccountId = '';
		}
	}

	function toggleAccount(account: SocialAccount) {
		if (!isAccountCompatible(account)) return;
		selectedAccountIds = selectedAccountIds.includes(account.id)
			? selectedAccountIds.filter((id) => id !== account.id)
			: [...selectedAccountIds, account.id];
		if (!selectedAccountIds.includes(account.id) && settingsAccountId === account.id) {
			settingsDialogOpen = false;
			settingsAccountId = '';
		}
		settingsByAccount = normalizeAllAccountSettings(settingsByAccount);
		validationIssues = [];
		void resolveSelectedCapabilities();
	}

	function isAccountCompatible(account: SocialAccount): boolean {
		return isAccountCompatibleWithMode(mode, account, capabilities);
	}

	function capabilityForAccount(
		account: SocialAccount
	): Capability | ResolvedAccountCapability | null {
		if (resolvedCapabilities[account.id]) return resolvedCapabilities[account.id];
		const provider = getPlatformKey(account.platform);
		return (
			capabilities.find(
				(capability) =>
					capability.provider === provider && (capability.intents ?? []).includes(mode)
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

	async function resolveSelectedCapabilities(): Promise<boolean> {
		const accountIds = selectedAccountIds;
		if (!selectedWorkspaceId || accountIds.length === 0) {
			resolvedCapabilities = {};
			capabilityResolveError = '';
			return true;
		}
		const requestSequence = ++capabilityResolveRequestSequence;
		capabilityResolveLoading = true;
		capabilityResolveError = '';
		const [, localeRegion = 'US'] = getLocaleTag().split('-');
		try {
			const { data, error: resolveError } = await client.POST('/capabilities/resolve', {
				body: {
					account_ids: accountIds,
					intent: mode,
					source_url: fields.linkUrl ?? '',
					locale: getLocaleTag(),
					region: localeRegion,
					account_settings: Object.fromEntries(
						selectedAccounts.map((account) => [account.id, settingsForAccount(account)])
					),
					segments: composerSegments().map((segment) => ({
						id: segment.id,
						content: segment.content,
						title: segment.title ?? '',
						description: segment.description ?? '',
						url: segment.url ?? '',
						media: segment.media.map((item) => ({ media_id: item.id }))
					}))
				}
			});
			if (resolveError) {
				throw new Error(resolveError.detail || m.compose_load_capabilities_failed());
			}
			if (requestSequence !== capabilityResolveRequestSequence) return false;
			resolvedCapabilities = Object.fromEntries(
				(data?.accounts ?? []).map((capability) => [capability.account_id, capability])
			);
			for (const capability of data?.accounts ?? []) {
				const dynamic = capability.dynamic_options ?? {};
				if (Object.keys(dynamic).length === 0) continue;
				destinationOptionsByAccount = {
					...destinationOptionsByAccount,
					[capability.account_id]: {
						...(destinationOptionsByAccount[capability.account_id] ?? {}),
						...Object.fromEntries(
							Object.entries(dynamic).map(([source, options]) => [
								source,
								(options ?? []).map((option) => ({ value: option.value, label: option.label }))
							])
						)
					}
				};
			}
			validationIssues = (data?.accounts ?? []).flatMap((capability) => capability.issues ?? []);
			settingsByAccount = normalizeAllAccountSettings(settingsByAccount);
			return (data?.accounts ?? []).every((capability) => capability.compatible);
		} catch (resolveError) {
			if (requestSequence !== capabilityResolveRequestSequence) return false;
			capabilityResolveError =
				resolveError instanceof Error ? resolveError.message : m.compose_load_capabilities_failed();
			return false;
		} finally {
			if (requestSequence === capabilityResolveRequestSequence) {
				capabilityResolveLoading = false;
			}
		}
	}

	function composerSegments(): FocusedSegmentInput[] {
		if (mode === 'thread') {
			return segments.map((segment, index) => ({
				...segment,
				media: index === 0 && segment.media.length === 0 ? focusedMediaInputs(media) : segment.media
			}));
		}
		return [
			{
				id: segments[0]?.id || 'segment-1',
				content: firstFocusedValue(fields.postText, fields.caption, fields.videoDescription),
				title: fields.videoTitle,
				description: fields.videoDescription,
				url: fields.linkUrl,
				media: focusedMediaInputs(media),
				settingsByAccount: segmentSettingsByAccount
			}
		];
	}

	function focusedMediaInputs(items: FocusedMedia[]) {
		return items.map((item) => ({
			id: item.id,
			mimeType: item.mime_type,
			role: item.role,
			altText: item.altText,
			settings: item.settings,
			settingsByAccount: item.settingsByAccount
		}));
	}

	function firstFocusedValue(...values: Array<string | undefined>): string {
		return values.find((value) => value?.trim())?.trim() ?? '';
	}

	function updateField(key: FocusedFieldKey, value: string) {
		fields = { ...fields, [key]: value };
		validationIssues = [];
	}

	function updateThreadSegment(segmentId: string, content: string) {
		segments = segments.map((segment) =>
			segment.id === segmentId ? { ...segment, content } : segment
		);
		validationIssues = [];
	}

	function addThreadSegment() {
		const id = `segment-${Date.now()}-${segments.length + 1}`;
		segments = [...segments, { id, content: '', media: [], settingsByAccount: {} }];
		activeSettingsSegmentId = id;
		validationIssues = [];
	}

	function removeThreadSegment(segmentId: string) {
		if (segments.length <= 2) return;
		segments = segments.filter((segment) => segment.id !== segmentId);
		if (activeSettingsSegmentId === segmentId) {
			activeSettingsSegmentId = segments[0].id;
		}
		validationIssues = [];
		void resolveSelectedCapabilities();
	}

	function updateMediaAltText(mediaId: string, altText: string) {
		media = media.map((item) => (item.id === mediaId ? { ...item, altText } : item));
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
			await resolveSelectedCapabilities();
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

	async function handleDestinationFile(setting: SettingDefinition, file: File) {
		const account = settingsAccount;
		if (!account || !selectedWorkspaceId) return;
		uploading = true;
		error = '';
		try {
			const uploaded = await uploadMediaFile({ workspaceId: selectedWorkspaceId, file });
			updateAccountSetting(account, setting.key, uploaded.id);
		} catch (uploadError) {
			error = uploadError instanceof Error ? uploadError.message : m.compose_upload_failed();
		} finally {
			uploading = false;
		}
	}

	function requestDeleteDestination(account: SocialAccount) {
		settingsDialogOpen = false;
		deleteDestinationAccount = account;
		deleteDestinationDialogOpen = true;
	}

	async function confirmDeleteDestination() {
		const account = deleteDestinationAccount;
		if (!account || !publicationId) return;
		const { error: deleteError } = await client.DELETE(
			'/publications/{id}/renditions/{account_id}',
			{
				params: {
					path: { id: publicationId, account_id: account.id },
					query: { confirm: true }
				}
			}
		);
		if (deleteError) {
			throw new Error(deleteError.detail || m.compose_save_outputs_failed());
		}
		selectedAccountIds = selectedAccountIds.filter((id) => id !== account.id);
		const nextSettings = { ...settingsByAccount };
		delete nextSettings[account.id];
		settingsByAccount = nextSettings;
		const nextResolved = { ...resolvedCapabilities };
		delete nextResolved[account.id];
		resolvedCapabilities = nextResolved;
		deleteDestinationAccount = null;
		success = m.compose_delete_destination();
		ui.triggerRefresh();
	}

	function removeMedia(mediaId: string) {
		media = media.filter((item) => item.id !== mediaId);
		validationIssues = [];
		void resolveSelectedCapabilities();
	}

	function clearThumbnail() {
		thumbnailMedia = null;
		thumbnailMediaId = '';
		validationIssues = [];
	}

	function validationIssueActionLabel(issue: ValidationIssue): string {
		switch (issue.field) {
			case 'poll_options':
				return m.compose_remove_poll();
			case 'quote_url':
				return m.compose_remove_quote();
			case 'media':
				return m.compose_remove_all_media();
			default:
				return '';
		}
	}

	function resolveValidationIssue(issue: ValidationIssue) {
		const field = issue.field;
		if (field === 'media') {
			media = [];
			segments = segments.map((segment) => ({ ...segment, media: [] }));
		} else if (field === 'poll_options' || field === 'quote_url') {
			const accountIds = selectedAccounts
				.filter((account) => getPlatformKey(account.platform) === issue.provider)
				.map((account) => account.id);
			settingsByAccount = Object.fromEntries(
				Object.entries(settingsByAccount).map(([accountId, values]) => [
					accountId,
					accountIds.includes(accountId) ? { ...values, [field]: '' } : values
				])
			);
			segmentSettingsByAccount = Object.fromEntries(
				Object.entries(segmentSettingsByAccount).map(([accountId, values]) => [
					accountId,
					accountIds.includes(accountId) ? { ...values, [field]: '' } : values
				])
			);
			segments = segments.map((segment) => ({
				...segment,
				settingsByAccount: Object.fromEntries(
					Object.entries(segment.settingsByAccount ?? {}).map(([accountId, values]) => [
						accountId,
						accountIds.includes(accountId) ? { ...values, [field]: '' } : values
					])
				)
			}));
		}
		validationIssues = validationIssues.filter((candidate) => candidate !== issue);
		void resolveSelectedCapabilities();
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

	function applyInitialScheduleDate() {
		if (selectedDate || !initialScheduleDate) return;
		const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(initialScheduleDate);
		if (!match) return;
		selectedDate = new CalendarDate(Number(match[1]), Number(match[2]), Number(match[3]));
	}

	function formBlockers(): string[] {
		const blockers: string[] = [];
		if (!selectedWorkspaceId) blockers.push(m.compose_choose_workspace_blocker());
		if (selectedAccounts.length === 0) blockers.push(m.compose_choose_account_blocker());
		if (capabilityResolveError) blockers.push(capabilityResolveError);
		if (mode === 'thread') {
			if (segments.length < 2) blockers.push(m.compose_thread_minimum());
			for (const [index, segment] of segments.entries()) {
				if (
					!segment.content.trim() &&
					segment.media.length === 0 &&
					!(index === 0 && media.length > 0)
				) {
					blockers.push(m.compose_thread_post_required({ number: index + 1 }));
				}
			}
		}
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
			const resolved = resolvedCapabilities[account.id];
			for (const issue of resolved?.issues ?? []) {
				if (issue.severity === 'error') blockers.push(issue.message);
			}
			for (const setting of visibleSettings(account)) {
				if (!setting.required) continue;
				if (setting.scope === 'media_item') {
					for (const item of mediaForSettingsDialog()) {
						const values = item.settingsByAccount?.[account.id] ?? {};
						if (!settingDependenciesMet(setting, values)) continue;
						const value = values[setting.key];
						if (
							value === undefined ||
							value === null ||
							String(value).trim() === '' ||
							(setting.type === 'boolean' && value !== true)
						) {
							blockers.push(
								m.compose_destination_setting_required({
									setting: setting.label,
									platform: getPlatformName(account.platform)
								})
							);
						}
					}
					continue;
				}
				const values =
					setting.scope === 'segment'
						? segmentSettingsForAccount(account)
						: settingsForAccount(account);
				if (!settingDependenciesMet(setting, values)) continue;
				const value = values[setting.key];
				if (
					value === undefined ||
					value === null ||
					String(value).trim() === '' ||
					(setting.type === 'boolean' && value !== true)
				) {
					blockers.push(
						m.compose_destination_setting_required({
							setting: setting.label,
							platform: getPlatformName(account.platform)
						})
					);
				}
			}
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
			if (field.scope !== 'destination') continue;
			if (next[field.key] !== undefined) continue;
			if (field.default !== undefined) next[field.key] = field.default;
			else if (field.type === 'boolean') next[field.key] = false;
			else next[field.key] = '';
		}
		return next;
	}

	function settingsForAccount(account: SocialAccount): Record<string, unknown> {
		return normalizeSettings(account, settingsByAccount[account.id] ?? {});
	}

	function dialogSettingsForAccount(account: SocialAccount): Record<string, unknown> {
		return {
			...settingsForAccount(account),
			...segmentSettingsForAccount(account)
		};
	}

	function segmentSettingsForAccount(account: SocialAccount): Record<string, unknown> {
		if (mode !== 'thread') return segmentSettingsByAccount[account.id] ?? {};
		const segment = segments.find((item) => item.id === activeSettingsSegmentId) ?? segments[0];
		return segment?.settingsByAccount?.[account.id] ?? {};
	}

	function updateAccountSetting(account: SocialAccount, key: string, value: unknown) {
		const definition = visibleSettings(account).find((field) => field.key === key);
		if (definition?.scope === 'segment') {
			if (mode !== 'thread') {
				segmentSettingsByAccount = {
					...segmentSettingsByAccount,
					[account.id]: {
						...(segmentSettingsByAccount[account.id] ?? {}),
						[key]: value
					}
				};
			} else {
				const segmentIndex = Math.max(
					0,
					segments.findIndex((segment) => segment.id === activeSettingsSegmentId)
				);
				const segment = segments[segmentIndex];
				segments = segments.map((item, index) =>
					index === segmentIndex
						? {
								...item,
								settingsByAccount: {
									...(item.settingsByAccount ?? {}),
									[account.id]: {
										...(segment.settingsByAccount?.[account.id] ?? {}),
										[key]: value
									}
								}
							}
						: item
				);
			}
			validationIssues = [];
			return;
		}
		const current = settingsForAccount(account);
		settingsByAccount = {
			...settingsByAccount,
			[account.id]: { ...current, [key]: value }
		};
		validationIssues = [];
		if (key === 'content_posting_method') {
			void resolveSelectedCapabilities();
		}
	}

	function visibleSettings(account: SocialAccount): SettingDefinition[] {
		const provider = getPlatformKey(account.platform);
		return (capabilityForAccount(account)?.settings ?? []).filter((field) => {
			if (
				provider === 'youtube' &&
				['title', 'description', 'thumbnail_media_id'].includes(field.key)
			) {
				return false;
			}
			if (
				mode === 'post' &&
				['url', 'link_url', 'link_title', 'link_description'].includes(field.key) &&
				!fields.linkUrl?.trim() &&
				!field.required
			)
				return false;
			return true;
		});
	}

	function settingDependenciesMet(
		setting: SettingDefinition,
		values: Record<string, unknown>
	): boolean {
		return (setting.dependencies ?? []).every((condition) => {
			const value = values[condition.key];
			const present = value !== undefined && value !== null && String(value).trim() !== '';
			switch (condition.operator) {
				case 'present':
					return present;
				case 'absent':
					return !present;
				case 'equals':
					return present && String(value) === String(condition.value);
				case 'not_equals':
					return !present || String(value) !== String(condition.value);
				case 'in':
					return Array.isArray(condition.value) && condition.value.includes(value);
			}
		});
	}

	function mediaForSettingsDialog(): FocusedMediaInput[] {
		if (mode !== 'thread') return focusedMediaInputs(media);
		const segment = segments.find((item) => item.id === activeSettingsSegmentId) ?? segments[0];
		return segment?.media.length ? segment.media : focusedMediaInputs(media);
	}

	function mediaSettingsForDialog(account: SocialAccount): Record<string, Record<string, unknown>> {
		return Object.fromEntries(
			mediaForSettingsDialog().map((item) => [
				item.id,
				{ ...(item.settingsByAccount?.[account.id] ?? {}) }
			])
		);
	}

	function updateMediaAccountSetting(
		account: SocialAccount,
		mediaId: string,
		key: string,
		value: unknown
	) {
		media = media.map((item) =>
			item.id === mediaId
				? {
						...item,
						settingsByAccount: {
							...(item.settingsByAccount ?? {}),
							[account.id]: {
								...(item.settingsByAccount?.[account.id] ?? {}),
								[key]: value
							}
						}
					}
				: item
		);
		segments = segments.map((segment) => ({
			...segment,
			media: segment.media.map((item) =>
				item.id === mediaId
					? {
							...item,
							settingsByAccount: {
								...(item.settingsByAccount ?? {}),
								[account.id]: {
									...(item.settingsByAccount?.[account.id] ?? {}),
									[key]: value
								}
							}
						}
					: item
			)
		}));
		validationIssues = [];
	}

	function setMediaAccountSettings(
		items: Array<FocusedMedia | FocusedMediaInput>,
		mediaId: string,
		accountId: string,
		settings: Record<string, unknown>
	) {
		const item = items.find((candidate) => candidate.id === mediaId);
		if (!item) return;
		item.settingsByAccount = {
			...(item.settingsByAccount ?? {}),
			[accountId]: settings
		};
	}

	function openDestinationSettings(account: SocialAccount) {
		settingsByAccount = {
			...settingsByAccount,
			[account.id]: settingsForAccount(account)
		};
		settingsAccountId = account.id;
		settingsDialogOpen = true;
		void loadDestinationOptions(account);
	}

	async function loadDestinationOptions(
		account: SocialAccount,
		force = false,
		onlySource = '',
		search = ''
	) {
		let optionSources = onlySource
			? [onlySource]
			: visibleSettings(account)
					.map((setting) => setting.options_source)
					.filter((source): source is string => Boolean(source));
		if (optionSources.length === 0) return;
		if (!force && !search) {
			optionSources = optionSources.filter(
				(source) => destinationOptionsByAccount[account.id]?.[source] === undefined
			);
			if (optionSources.length === 0) return;
		}

		const requestSequence = ++destinationOptionsRequestSequence;
		destinationOptionsLoadingAccountId = account.id;
		destinationOptionsErrors = { ...destinationOptionsErrors, [account.id]: '' };
		const [, regionCode = 'US'] = getLocaleTag().split('-');
		try {
			const results = await Promise.all(
				optionSources.map(async (source) => {
					const { data, error: loadError } = await client.GET(
						'/accounts/{account_id}/publishing-options/{source}',
						{
							params: {
								path: { account_id: account.id, source },
								query: {
									region: regionCode,
									locale: getLocaleTag(),
									limit: 100,
									search
								}
							}
						}
					);
					if (loadError) {
						throw new Error(loadError.detail || m.compose_load_provider_options_failed());
					}
					return [source, data?.options ?? []] as const;
				})
			);
			if (requestSequence !== destinationOptionsRequestSequence) return;

			const optionGroups: Record<string, DestinationOption[]> = {
				...(destinationOptionsByAccount[account.id] ?? {})
			};
			for (const [source, options] of results) optionGroups[source] = options;
			destinationOptionsByAccount = {
				...destinationOptionsByAccount,
				[account.id]: optionGroups
			};
		} catch (loadError) {
			if (requestSequence !== destinationOptionsRequestSequence) return;
			destinationOptionsErrors = {
				...destinationOptionsErrors,
				[account.id]:
					loadError instanceof Error ? loadError.message : m.compose_load_provider_options_failed()
			};
		} finally {
			if (
				requestSequence === destinationOptionsRequestSequence &&
				destinationOptionsLoadingAccountId === account.id
			) {
				destinationOptionsLoadingAccountId = '';
			}
		}
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
			media: focusedMediaInputs(media),
			segments: composerSegments(),
			scheduledAt: getScheduledAt(),
			thumbnailMediaId,
			settingsByAccount: selectedSettingsInput(),
			resolvedByAccount: Object.fromEntries(
				Object.entries(resolvedCapabilities).map(([accountId, capability]) => [
					accountId,
					{
						profile: capability.profile,
						outputProfile: capability.output_profile,
						revision: capability.capability_revision,
						compatible: capability.compatible
					} satisfies ResolvedComposerTarget
				])
			)
		});
	}

	async function persistPublication(): Promise<string> {
		const payload = publicationPayload();
		if (publicationId) {
			const { error: updateError } = await client.PUT('/publications/{id}', {
				params: { path: { id: publicationId } },
				body: {
					title: payload.title,
					intent: payload.intent,
					content_profile: payload.content_profile,
					source_text: payload.source_text,
					source_url: payload.source_url ?? '',
					...(payload.scheduled_at
						? { scheduled_at: payload.scheduled_at }
						: { clear_schedule: true }),
					metadata: payload.metadata,
					segments: payload.segments
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
			const capabilitiesReady = await resolveSelectedCapabilities();
			if (!capabilitiesReady) {
				throw new Error(
					capabilityResolveError ||
						validationIssues.find((issue) => issue.severity === 'error')?.message ||
						m.compose_fix_before_publishing()
				);
			}
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
			role: item.role,
			altText: item.alt_text,
			settings: item.settings
		};
	}

	function mediaItemLabel(item: FocusedMediaInput): string {
		return media.find((candidate) => candidate.id === item.id)?.filename || item.id;
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

	function emptySegmentsForMode(intent: ComposerModeKey): FocusedSegmentInput[] {
		const count = intent === 'thread' ? 2 : 1;
		return Array.from({ length: count }, (_, index) => ({
			id: `segment-${index + 1}`,
			content: '',
			media: [],
			settingsByAccount: {}
		}));
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
							settingsAccountIds={selectedAccounts
								.filter((account) => visibleSettings(account).length > 0)
								.map((account) => account.id)}
							{accountSummaries}
							{warningAccountIds}
							triggerLabel={m.compose_target_accounts()}
							triggerClass="h-11 md:h-8"
							description={m.compose_accounts_compatible({ format: modeMeta.label })}
							onToggle={toggleAccount}
							onSelectAll={selectAllAccounts}
							onClearAll={clearAllAccounts}
							onSettings={openDestinationSettings}
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
				{#if capabilityResolveError}
					<InlineNotice tone="error" message={capabilityResolveError}>
						{#snippet actions()}
							<Button
								type="button"
								variant="outline"
								size="sm"
								onclick={() => resolveSelectedCapabilities()}
							>
								{m.common_retry()}
							</Button>
						{/snippet}
					</InlineNotice>
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
										accept={mode === 'post'
											? 'image/*,video/*,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation'
											: 'image/*,video/*'}
										disabled={uploading || !selectedWorkspaceId}
										onchange={(event) => handleFiles(event.currentTarget.files)}
									/>
								</label>
								<p class="text-sm text-muted-foreground">
									{mode === 'video' ? m.compose_add_long_video() : m.compose_add_media()}
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
											{#if isImage(item)}
												<div class="border-t px-2 py-2">
													<label class="text-xs font-medium" for="media-alt-{item.id}">
														{m.media_alt_text()}
													</label>
													<Input
														id="media-alt-{item.id}"
														class="mt-1 h-9"
														value={item.altText ?? ''}
														placeholder={m.compose_alt_text_placeholder()}
														oninput={(event) =>
															updateMediaAltText(item.id, event.currentTarget.value)}
													/>
												</div>
											{/if}
										</div>
									{/each}
								</div>
							{/if}
						</div>

						{#if hasYouTubeTarget && (mode === 'short_video' || mode === 'video')}
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
						{#if mode === 'thread'}
							<div class="space-y-3">
								{#each segments as segment, index (segment.id)}
									<article
										class:border-primary={activeSettingsSegmentId === segment.id}
										class="rounded-lg border bg-background p-3 transition-colors"
									>
										<div class="mb-2 flex items-center justify-between gap-2">
											<label class="text-sm font-semibold" for="thread-segment-{segment.id}">
												{m.compose_thread_post({ number: index + 1 })}
											</label>
											{#if segments.length > 2}
												<Button
													type="button"
													variant="ghost"
													size="icon"
													class="size-11 text-muted-foreground hover:text-destructive md:size-9"
													aria-label={m.compose_remove_post()}
													onclick={() => removeThreadSegment(segment.id)}
												>
													<Trash2Icon class="size-4" />
												</Button>
											{/if}
										</div>
										<Textarea
											id="thread-segment-{segment.id}"
											data-testid="composer-thread-segment-{index + 1}"
											class="min-h-28 resize-y text-base"
											rows={5}
											value={segment.content}
											placeholder={m.compose_add_to_thread()}
											onfocus={() => (activeSettingsSegmentId = segment.id)}
											oninput={(event) =>
												updateThreadSegment(segment.id, event.currentTarget.value)}
										/>
									</article>
								{/each}
								<Button
									type="button"
									variant="outline"
									class="h-11 w-full gap-2 border-dashed"
									onclick={addThreadSegment}
								>
									<PlusIcon class="size-4" />
									{m.compose_add_post()}
								</Button>
							</div>
						{/if}
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
						{#if roleFields.length === 0 && mode !== 'thread'}
							<div class="rounded-md border bg-muted/20 px-3 py-2 text-sm text-muted-foreground">
								{m.compose_media_first_notice()}
							</div>
						{/if}
					</div>
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
							<div class="flex flex-wrap items-start gap-2 text-destructive">
								<AlertCircleIcon class="mt-0.5 size-4 shrink-0" />
								<span class="min-w-0 flex-1">{issue.message}</span>
								{#if validationIssueActionLabel(issue)}
									<Button
										type="button"
										variant="outline"
										size="sm"
										class="h-8 border-destructive/30 text-destructive hover:text-destructive"
										onclick={() => resolveValidationIssue(issue)}
									>
										{validationIssueActionLabel(issue)}
									</Button>
								{/if}
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

<DestinationSettingsDialog
	bind:open={settingsDialogOpen}
	account={settingsAccount}
	settings={settingsDialogFields}
	values={settingsDialogValues}
	mediaItems={settingsDialogMedia.map((item) => ({
		id: item.id,
		label: mediaItemLabel(item),
		mimeType: item.mimeType
	}))}
	mediaValues={settingsDialogMediaValues}
	optionGroups={settingsAccount ? (destinationOptionsByAccount[settingsAccount.id] ?? {}) : {}}
	optionsLoading={settingsAccount?.id === destinationOptionsLoadingAccountId}
	optionsError={settingsAccount ? (destinationOptionsErrors[settingsAccount.id] ?? '') : ''}
	scopeLabel={mode === 'thread'
		? m.compose_thread_post({
				number: Math.max(
					1,
					segments.findIndex((segment) => segment.id === activeSettingsSegmentId) + 1
				)
			})
		: ''}
	onChange={(key, value) => {
		if (settingsAccount) updateAccountSetting(settingsAccount, key, value);
	}}
	onMediaChange={(mediaId, key, value) => {
		if (settingsAccount) updateMediaAccountSetting(settingsAccount, mediaId, key, value);
	}}
	onOptionSearch={(setting, search) => {
		if (settingsAccount && setting.options_source) {
			void loadDestinationOptions(settingsAccount, true, setting.options_source, search);
		}
	}}
	onFileChange={handleDestinationFile}
	onRetry={() => {
		if (settingsAccount) void loadDestinationOptions(settingsAccount, true);
	}}
	onRemove={settingsAccount &&
	publicationId &&
	initialPublication?.renditions?.some(
		(rendition) => rendition.social_account_id === settingsAccount.id
	)
		? () => requestDeleteDestination(settingsAccount)
		: undefined}
/>

<DestructiveConfirmDialog
	bind:open={deleteDestinationDialogOpen}
	title={m.compose_delete_destination_title({
		account: deleteDestinationAccount ? accountLabel(deleteDestinationAccount) : ''
	})}
	description={m.compose_delete_destination_body()}
	confirmLabel={m.compose_delete_destination_confirm()}
	onConfirm={confirmDeleteDestination}
/>
