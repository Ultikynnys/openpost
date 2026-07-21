<script lang="ts">
	import { onMount, tick, type Snippet } from 'svelte';
	import { MediaQuery, SvelteMap, SvelteSet } from 'svelte/reactivity';
	import { client, type SocialAccount, type Workspace, getToken } from '$lib/api/client';
	import { getApiBase } from '$lib/stores/instance.svelte';
	import { getAuthenticatedMediaByID } from '$lib/media-url';
	import { isSupportedMediaFile, uploadMediaFile } from '$lib/media-upload-client';
	import {
		mediaCapabilityItemsFromIds,
		providerMediaWarningMessages
	} from '$lib/media-capabilities';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Calendar } from '$lib/components/ui/calendar';
	import { Input } from '$lib/components/ui/input';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Select from '$lib/components/ui/select';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import PlatformIcon from './platform-icon.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { getPlatformKey, getPlatformName } from '$lib/utils';
	import { CalendarDate, isEqualDay } from '@internationalized/date';
	import ArrowRightIcon from 'lucide-svelte/icons/arrow-right';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import XIcon from 'lucide-svelte/icons/x';
	import LightbulbIcon from 'lucide-svelte/icons/lightbulb';
	import ShuffleIcon from 'lucide-svelte/icons/shuffle';
	import ImageIcon from 'lucide-svelte/icons/image';
	import ChevronDownIcon from 'lucide-svelte/icons/chevron-down';
	import UnlinkIcon from 'lucide-svelte/icons/unlink';
	import Link2Icon from 'lucide-svelte/icons/link-2';
	import GripVerticalIcon from 'lucide-svelte/icons/grip-vertical';
	import Trash2Icon from 'lucide-svelte/icons/trash-2';
	import TypeIcon from 'lucide-svelte/icons/type';
	import MoreHorizontalIcon from 'lucide-svelte/icons/ellipsis';
	import CalendarClockIcon from 'lucide-svelte/icons/calendar-clock';
	import { ui } from '$lib/stores/ui.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { ReorderableList } from 'svelte-reorderable-list';
	import { m } from '$lib/paraglide/messages';
	import { getLocaleTag } from '$lib/i18n';
	import {
		type PostItem,
		makeEmptyPost,
		encodeThreadDraft,
		isThreadDraft,
		decodeThreadDraft,
		getDraftSnapshot,
		hasAnyContent,
		type VariantPost
	} from './compose/draft-utils';
	import { minimumAccountCharacterLimit, uniquePlatformLimits } from './compose/platform-limits';
	import { editorAccountIdAfterVariantLoad } from './compose/editor-target';
	import { parseNaturalScheduleInput } from './compose/schedule-language';
	import {
		workspaceClock,
		workspaceScheduleFromISO,
		workspaceScheduleToISO
	} from './compose/schedule-timezone';
	import { soundPreferences } from '$lib/stores/sound-preferences.svelte';
	import InlineNotice from './inline-notice.svelte';
	import DestructiveConfirmDialog from './destructive-confirm-dialog.svelte';

	// --------------------------------------------------------------------------
	// Types
	// --------------------------------------------------------------------------
	interface InitialPost {
		id: string;
		workspace_id: string;
		content: string;
		thread_draft?: string | null;
		status: string;
		scheduled_at: string;
		random_delay_minutes?: number;
		media: Array<{ media_id: string; mime_type?: string; alt_text?: string }>;
		destinations: Array<{ social_account_id: string; platform: string }>;
	}

	type PersistedVariant = {
		social_account_id: string;
		content: string;
		media_ids: string;
		is_unsynced: boolean;
	};

	interface Props {
		initialPost?: InitialPost;
		initialScheduleDate?: string | null;
		initialWorkspaceId?: string | null;
		onSuccess?: () => void;
		onDeleted?: () => void;
		onDraftCreated?: (id: string) => void;
		onThreadStateChange?: (isThread: boolean) => void;
		modeControl?: Snippet;
	}

	// --------------------------------------------------------------------------
	// Props & core state
	// --------------------------------------------------------------------------
	let {
		initialPost,
		initialScheduleDate = null,
		initialWorkspaceId = null,
		onSuccess,
		onDeleted,
		onDraftCreated,
		onThreadStateChange,
		modeControl
	}: Props = $props();
	let isEditMode = $derived(!!initialPost);

	let posts = $state<PostItem[]>([makeEmptyPost()]);
	let activePostIndex = $state(0);
	let draftId = $state<string | null>(null);
	let lastInitializedPostId = $state<string | null>(null);
	let isSaving = $state(false);
	let isSubmitting = $state(false);
	let isDeleting = $state(false);
	let showDeleteConfirm = $state(false);
	let error = $state('');
	let success = $state('');

	let workspaces = $state<Workspace[]>([]);
	let selectedWorkspaceId = $state<string>('');
	let accounts = $state<SocialAccount[]>([]);
	let selectedAccountIds = $state<string[]>([]);
	let loadingWorkspaces = $state(true);
	let loadingAccounts = $state(false);
	let workspaceLoadError = $state('');
	let workspaceSettingsError = $state('');
	let workspaceChangeNotice = $state('');
	let accountLoadError = $state('');
	let accountsWorkspaceId = $state('');
	let accountRetryIds: string[] | undefined = undefined;
	let workspaceRequestSequence = 0;
	let accountRequestSequence = 0;
	let nextSlotRequestSequence = 0;
	let saveGeneration = 0;

	let selectedDate = $state<CalendarDate | undefined>(undefined);
	let selectedTime = $state<string | null>(null);
	let suggestingSlot = $state(false);
	let showScheduleDialog = $state(false);
	let scheduleInput = $state('');
	let scheduleInputError = $state('');
	let randomDelayOverride = $state<string>('default');

	let showPromptCard = $state(false);
	let currentPrompt = $state<{ text: string; category: string } | null>(null);
	let loadingPrompt = $state(false);

	let variants = $state<Map<string, Record<string, VariantPost>>>(new Map());
	let activeVariantAccountId = $state<string | null>(null);

	let isDraggingFile = $state(false);
	let isUploading = $state(false);

	let mediaAltTexts = $state<Map<string, string>>(new Map());
	let mediaMimeTypes = $state<Map<string, string>>(new Map());
	let mediaSizes = $state<Map<string, number>>(new Map());
	let editingAltMediaId = $state<string | null>(null);

	let autoSaveTimer: ReturnType<typeof setTimeout> | null = null;
	let lastSavedSnapshot = $state('');
	let appliedInitialContextKey = $state('');
	const textareaRefs = new SvelteMap<number, HTMLTextAreaElement>();
	const randomDelayOptions = [0, 5, 10, 15, 30, 45, 60];
	const desktopComposerControls = new MediaQuery('min-width: 768px');
	const accountControlLoading = $derived(loadingWorkspaces || loadingAccounts);
	const selectedWorkspaceSettingsReady = $derived(
		Boolean(selectedWorkspaceId) &&
			workspaceCtx.currentWorkspace?.id === selectedWorkspaceId &&
			workspaceCtx.settingsReady
	);

	// --------------------------------------------------------------------------
	// Constants & derived values
	// --------------------------------------------------------------------------
	const scheduleTimezoneLabel = $derived(workspaceCtx.settings.timezone || 'UTC');

	// Generate time slots dynamically from workspace settings
	const allTimeSlots = $derived.by(() => {
		const start = workspaceCtx.settings.slot_start_hour;
		const end = workspaceCtx.settings.slot_end_hour;
		const interval = workspaceCtx.settings.slot_interval_minutes;
		const slots: string[] = [];
		for (let hour = start; hour <= end; hour++) {
			for (let min = 0; min < 60; min += interval) {
				if (hour === end && min > 0) break;
				slots.push(`${hour.toString().padStart(2, '0')}:${min.toString().padStart(2, '0')}`);
			}
		}
		return slots;
	});

	const timeSlots = $derived(selectedDate ? slotsForDate(selectedDate) : allTimeSlots);

	function slotsForDate(date: CalendarDate): string[] {
		const validSlots = allTimeSlots.filter((slot) =>
			Boolean(workspaceScheduleToISO(date, slot, scheduleTimezoneLabel))
		);
		if (!isEqualDay(date, workspaceClock(scheduleTimezoneLabel).date)) return validSlots;
		const currentMinutes = workspaceClock(scheduleTimezoneLabel).minutes;
		return validSlots.filter((slot) => {
			const [hours, minutes] = slot.split(':').map(Number);
			return hours * 60 + minutes > currentMinutes;
		});
	}

	const activePost = $derived(posts[activePostIndex] ?? posts[0]);
	const hasContent = $derived(hasAnyContent(posts));
	const totalChars = $derived(posts.reduce((sum, p) => sum + p.content.length, 0));
	const isThread = $derived(posts.length > 1);
	const autoSavesDraft = $derived(!isEditMode || initialPost?.status === 'draft');
	const selectedAccounts = $derived(accounts.filter((a) => selectedAccountIds.includes(a.id)));
	const syncedLinkedInThreadAccounts = $derived.by(() => {
		if (!isThread) return [];
		return selectedAccounts.filter(
			(account) => getPlatformKey(account.platform) === 'linkedin' && !variants.has(account.id)
		);
	});
	const mediaCapabilityWarnings = $derived.by(() => {
		const warnings: string[] = [];
		const sourcePosts = isThread ? posts : activePost ? [activePost] : [];

		for (const account of selectedAccounts) {
			const platform = getPlatformKey(account.platform);
			for (const post of sourcePosts) {
				const mediaIds = getVariantMediaIds(account.id, post.key) ?? post.mediaIds;
				warnings.push(
					...providerMediaWarningMessages(
						platform,
						mediaCapabilityItemsFromIds(mediaIds, mediaMimeTypes, mediaSizes)
					)
				);
			}
		}

		return Array.from(new Set(warnings));
	});
	const activeVariantAccount = $derived(
		activeVariantAccountId ? (accounts.find((a) => a.id === activeVariantAccountId) ?? null) : null
	);
	const activeVariantIsUnsynced = $derived(
		activeVariantAccountId ? variants.has(activeVariantAccountId) : false
	);
	const activeEditorContent = $derived(
		activeVariantAccountId
			? (getVariantContent(activeVariantAccountId, activePost.key) ?? activePost.content)
			: activePost.content
	);

	const editorTargetAccounts = $derived.by(() => {
		if (activeVariantAccountId) {
			const activeAccount = accounts.find((a) => a.id === activeVariantAccountId);
			return activeAccount ? [activeAccount] : [];
		}

		return selectedAccounts.filter((account) => !variants.has(account.id));
	});

	const editorLimitAccounts = $derived(editorTargetAccounts);

	const editorPlatformLimits = $derived.by(() => {
		return uniquePlatformLimits(editorLimitAccounts);
	});

	const editorMaxChars = $derived.by(() => {
		return minimumAccountCharacterLimit(editorLimitAccounts);
	});
	const effectiveRandomDelayMinutes = $derived.by(() => {
		if (randomDelayOverride === 'default') return workspaceCtx.settings.random_delay_minutes;
		const value = Number(randomDelayOverride);
		return Number.isFinite(value)
			? Math.max(0, Math.round(value))
			: workspaceCtx.settings.random_delay_minutes;
	});
	const randomDelaySelectOptions = $derived.by(() => {
		const options = new SvelteSet(randomDelayOptions);
		const selected = Number(randomDelayOverride);
		if (randomDelayOverride !== 'default' && Number.isFinite(selected)) {
			options.add(selected);
		}
		return Array.from(options).sort((a, b) => a - b);
	});

	const editorResizeSignature = $derived.by(() =>
		posts
			.map((post, index) => {
				const mediaIds = getEditorMediaIdsForPost(post);
				return `${post.key}:${index}:${getEditorContentForPost(post).length}:${mediaIds.join(',')}`;
			})
			.join('|')
	);

	// --------------------------------------------------------------------------
	// Helpers
	// --------------------------------------------------------------------------
	function getCharCounterColor(count: number, max: number): string {
		const pct = count / max;
		if (pct >= 1) return 'text-red-500';
		if (pct >= 0.8) return 'text-amber-500';
		return 'text-muted-foreground';
	}

	function formatRandomDelay(minutes: number): string {
		if (!Number.isFinite(minutes) || minutes <= 0) return m.compose_exact_time();
		if (minutes === 1) return m.compose_random_delay_one_minute();
		if (minutes === 60) return m.compose_random_delay_one_hour();
		return m.compose_random_delay_minutes({ minutes });
	}

	function normalizeRandomDelayValue(value: number | null | undefined): string {
		if (value === undefined || value === null || !Number.isFinite(value)) return 'default';
		return String(Math.max(0, Math.round(value)));
	}

	function parseScheduleDateParam(value: string | null): CalendarDate | undefined {
		const match = value?.match(/^(\d{4})-(\d{2})-(\d{2})$/);
		if (!match) return undefined;
		const year = Number(match[1]);
		const month = Number(match[2]);
		const day = Number(match[3]);
		const parsed = new Date(year, month - 1, day);
		if (
			parsed.getFullYear() !== year ||
			parsed.getMonth() + 1 !== month ||
			parsed.getDate() !== day
		) {
			return undefined;
		}
		return new CalendarDate(year, month, day);
	}

	function applyInitialScheduleDate(dateParam: string | null) {
		const date = parseScheduleDateParam(dateParam);
		if (!date) return;
		if (date.compare(workspaceClock(scheduleTimezoneLabel).date) < 0) {
			error = m.compose_schedule_future();
			return;
		}
		selectedDate = date;
		selectedTime = slotsForDate(date)[0] ?? allTimeSlots[0] ?? '09:00';
		scheduleInput = '';
		scheduleInputError = '';
	}

	async function ensureComposerWorkspace(workspaceId: string) {
		const workspace = workspaces.find((candidate) => candidate.id === workspaceId);
		if (!workspace) throw new Error(m.compose_load_workspaces_failed());

		if (workspaceCtx.currentWorkspace?.id !== workspaceId) {
			await workspaceCtx.setWorkspace(workspace);
		} else if (!workspaceCtx.settingsReady) {
			await workspaceCtx.loadSettings(workspaceId);
		}

		if (!workspaceCtx.settingsReady || workspaceCtx.currentWorkspace?.id !== workspaceId) {
			throw new Error(m.compose_load_workspace_settings_failed());
		}
		workspaceSettingsError = '';
	}

	async function retryComposerWorkspaceSettings() {
		if (!selectedWorkspaceId || workspaceCtx.currentWorkspace?.id !== selectedWorkspaceId) return;
		workspaceSettingsError = '';
		await workspaceCtx.loadSettings(selectedWorkspaceId);
		if (!workspaceCtx.settingsReady) {
			workspaceSettingsError = m.compose_load_workspace_settings_failed();
		}
	}

	async function applyInitialComposerContext(
		dateParam: string | null,
		workspaceParam: string | null
	) {
		if (!dateParam && !workspaceParam) {
			appliedInitialContextKey = '';
			return;
		}

		const contextKey = `${dateParam ?? ''}|${workspaceParam ?? ''}`;
		if (contextKey === appliedInitialContextKey) return;
		appliedInitialContextKey = contextKey;

		const nextWorkspaceId =
			workspaceParam && workspaces.some((workspace) => workspace.id === workspaceParam)
				? workspaceParam
				: '';
		if (
			nextWorkspaceId &&
			(nextWorkspaceId !== selectedWorkspaceId ||
				workspaceCtx.currentWorkspace?.id !== nextWorkspaceId ||
				!workspaceCtx.settingsReady)
		) {
			try {
				await ensureComposerWorkspace(nextWorkspaceId);
			} catch (cause) {
				appliedInitialContextKey = '';
				workspaceLoadError =
					cause instanceof Error ? cause.message : m.compose_load_workspace_settings_failed();
				return;
			}
			if (nextWorkspaceId !== selectedWorkspaceId) {
				selectedWorkspaceId = nextWorkspaceId;
				variants = new Map();
				activeVariantAccountId = null;
				await loadAccounts(nextWorkspaceId);
			} else if (accountsWorkspaceId !== nextWorkspaceId) {
				await loadAccounts(nextWorkspaceId);
			}
		}

		applyInitialScheduleDate(dateParam);
	}

	function arraysEqual(left: string[], right: string[]): boolean {
		if (left.length !== right.length) return false;
		return left.every((value, index) => value === right[index]);
	}

	function sanitizeSelectedAccounts(validAccounts: SocialAccount[]) {
		const validIds = new Set(validAccounts.map((account) => account.id));
		const nextSelectedIds = selectedAccountIds.filter((id) => validIds.has(id));
		if (!arraysEqual(nextSelectedIds, selectedAccountIds)) {
			selectedAccountIds = nextSelectedIds;
		}

		const nextVariants = new SvelteMap<string, Record<string, VariantPost>>();
		for (const [accountID, value] of variants.entries()) {
			if (validIds.has(accountID)) {
				nextVariants.set(accountID, value);
			}
		}
		if (nextVariants.size !== variants.size) {
			variants = nextVariants;
			activeVariantAccountId = editorAccountIdAfterVariantLoad(
				activeVariantAccountId,
				selectedAccountIds,
				nextVariants.keys()
			);
		}

		if (activeVariantAccountId && !validIds.has(activeVariantAccountId)) {
			activeVariantAccountId = null;
		}
	}

	function getCharCounterStrokeColor(count: number, max: number): string {
		const pct = count / max;
		if (pct >= 1) return '#ef4444';
		if (pct >= 0.8) return '#f59e0b';
		return 'currentColor';
	}

	function autoResize(el: HTMLTextAreaElement) {
		el.style.height = 'auto';
		el.style.height = el.scrollHeight + 'px';
	}

	function textareaAttachment(index: number) {
		return (el: HTMLTextAreaElement) => {
			textareaRefs.set(index, el);
			autoResize(el);
			return () => textareaRefs.delete(index);
		};
	}

	function getScheduledAt(): string | undefined {
		if (!selectedDate || !selectedTime) return undefined;
		return workspaceScheduleToISO(selectedDate, selectedTime, scheduleTimezoneLabel);
	}

	function getSaveSnapshot(): string {
		const variantEntries = Array.from(variants.entries())
			.sort(([a], [b]) => a.localeCompare(b))
			.map(([accountId, values]) => [
				accountId,
				Object.fromEntries(Object.entries(values).sort(([a], [b]) => a.localeCompare(b)))
			]);
		const selectedAccountsSnapshot = [...selectedAccountIds].sort();
		return JSON.stringify({
			draft: getDraftSnapshot(posts),
			selectedAccounts: selectedAccountsSnapshot,
			variants: variantEntries,
			scheduledDate: selectedDate?.toString() ?? null,
			selectedTime,
			randomDelayOverride,
			selectedWorkspaceId
		});
	}

	function getVariantPost(accountId: string, postKey: string): VariantPost | null {
		const values = variants.get(accountId);
		if (!values) return null;
		return values[postKey] ?? null;
	}

	function getVariantContent(accountId: string, postKey: string): string | null {
		const variant = getVariantPost(accountId, postKey);
		if (!variant) return null;
		return variant.content;
	}

	function getVariantMediaIds(accountId: string, postKey: string): string[] | null {
		const variant = getVariantPost(accountId, postKey);
		if (!variant) return null;
		return variant.mediaIds;
	}

	function getVariantPayloadForSave(): Record<string, Record<string, VariantPost>> {
		return Object.fromEntries(
			Array.from(variants.entries()).map(([accountId, values]) => [accountId, values])
		);
	}

	function getPersistedVariantPayload(
		sourceVariants: Map<string, Record<string, VariantPost>>,
		sourcePosts: PostItem[]
	): PersistedVariant[] {
		const firstPost = sourcePosts[0];
		if (!firstPost) return [];
		return Array.from(sourceVariants.entries()).map(([accountId, values]) => ({
			social_account_id: accountId,
			content: values[firstPost.key]?.content ?? firstPost.content,
			media_ids: JSON.stringify(values[firstPost.key]?.mediaIds ?? firstPost.mediaIds),
			is_unsynced: true
		}));
	}

	function makeVariantRecord(sourcePosts: PostItem[]): Record<string, VariantPost> {
		return Object.fromEntries(
			sourcePosts.map((post) => [
				post.key,
				{
					content: post.content,
					mediaIds: [...post.mediaIds]
				}
			])
		);
	}

	function normalizeVariantRecord(
		record: Record<string, VariantPost> | undefined,
		sourcePosts: PostItem[]
	): Record<string, VariantPost> {
		return Object.fromEntries(
			sourcePosts.map((post) => {
				const value = record?.[post.key];
				return [
					post.key,
					{
						content: value?.content ?? post.content,
						mediaIds: value?.mediaIds ? [...value.mediaIds] : [...post.mediaIds]
					}
				];
			})
		);
	}

	function variantRecordEquals(
		left: Record<string, VariantPost> | undefined,
		right: Record<string, VariantPost>,
		sourcePosts: PostItem[]
	): boolean {
		if (Object.keys(left ?? {}).length !== Object.keys(right).length) return false;
		return sourcePosts.every((post) => {
			const leftValue = left?.[post.key];
			const rightValue = right[post.key];
			return (
				(leftValue?.content ?? post.content) === rightValue.content &&
				arraysEqual(leftValue?.mediaIds ?? post.mediaIds, rightValue.mediaIds)
			);
		});
	}

	function getEditorContentForPost(post: PostItem): string {
		if (!activeVariantAccountId) return post.content;
		return getVariantContent(activeVariantAccountId, post.key) ?? post.content;
	}

	function getEditorMediaIdsForPost(post: PostItem): string[] {
		if (!activeVariantAccountId) return post.mediaIds;
		return getVariantMediaIds(activeVariantAccountId, post.key) ?? post.mediaIds;
	}

	function clearAutoSaveTimer() {
		if (autoSaveTimer) {
			clearTimeout(autoSaveTimer);
			autoSaveTimer = null;
		}
	}

	function isVideoMedia(mediaId: string): boolean {
		return mediaMimeTypes.get(mediaId)?.startsWith('video/') ?? false;
	}

	function mergeMediaIds(current: string[], incoming: string[]): string[] {
		const seen = new SvelteSet<string>();
		const merged: string[] = [];
		for (const id of [...current, ...incoming]) {
			const clean = id.trim();
			if (!clean || seen.has(clean)) continue;
			seen.add(clean);
			merged.push(clean);
			if (merged.length >= 4) break;
		}
		return merged;
	}

	function normalizeVariantsMap(
		nextVariants: Map<string, Record<string, VariantPost>>,
		sourcePosts: PostItem[] = posts
	): Map<string, Record<string, VariantPost>> {
		const normalized = new SvelteMap<string, Record<string, VariantPost>>();
		for (const accountId of selectedAccountIds) {
			const values = nextVariants.get(accountId);
			if (values) {
				normalized.set(accountId, normalizeVariantRecord(values, sourcePosts));
			}
		}
		return normalized;
	}

	// --------------------------------------------------------------------------
	// Initialization
	// --------------------------------------------------------------------------
	async function initializeFromPost(post: InitialPost | undefined) {
		clearAutoSaveTimer();
		if (!post) {
			draftId = null;
			lastInitializedPostId = null;
			posts = [makeEmptyPost()];
			activePostIndex = 0;
			lastSavedSnapshot = '';
			variants = new Map();
			activeVariantAccountId = null;
			selectedAccountIds = [];
			mediaAltTexts = new Map();
			mediaMimeTypes = new Map();
			mediaSizes = new Map();
			selectedDate = undefined;
			selectedTime = null;
			randomDelayOverride = 'default';
			if (workspaces.length > 0) {
				selectedWorkspaceId = workspaceCtx.currentWorkspace?.id ?? workspaces[0].id;
				await ensureComposerWorkspace(selectedWorkspaceId);
				await loadAccounts(selectedWorkspaceId);
			}
			return;
		}

		await ensureComposerWorkspace(post.workspace_id);
		draftId = post.id;
		lastInitializedPostId = post.id;
		selectedWorkspaceId = post.workspace_id;
		selectedAccountIds = post.destinations?.map((d) => d.social_account_id) ?? [];
		randomDelayOverride = normalizeRandomDelayValue(post.random_delay_minutes);

		// Load alt texts from media
		const newAlts = new SvelteMap<string, string>();
		const newMimeTypes = new SvelteMap<string, string>();
		post.media?.forEach((m) => {
			if (m.alt_text) newAlts.set(m.media_id, m.alt_text);
			if (m.mime_type) newMimeTypes.set(m.media_id, m.mime_type);
		});
		mediaAltTexts = newAlts;
		mediaMimeTypes = newMimeTypes;
		mediaSizes = new Map();
		if (post.media?.length) {
			await hydrateMediaMetadata(
				post.workspace_id,
				post.media.map((m) => m.media_id).filter(Boolean)
			);
		}

		// Read the thread state. Prefer the explicit `thread_draft`
		// field on the post (new P0.2 representation, set by the
		// backend when the post is a thread draft). Fall back to the
		// legacy `__openpost_thread__:` blob inside `content` for posts
		// that were saved before the migration ran. If neither is
		// present, treat it as a single-post draft.
		const threadSource: string | null = post.thread_draft ?? null;
		const legacySource: string | null = isThreadDraft(post.content) ? post.content : null;
		const source = threadSource ?? legacySource;
		if (source) {
			const threadData = decodeThreadDraft(source);
			if (threadData && threadData.posts.length > 0) {
				posts = threadData.posts.map((item) => ({
					key: item.key,
					content: item.content,
					mediaIds: item.mediaIds
				}));
				variants = normalizeVariantsMap(new Map(Object.entries(threadData.variants)), posts);
			} else {
				posts = [makeEmptyPost()];
				variants = new Map();
			}
		} else {
			posts = [
				{
					key: makeEmptyPost().key,
					content: post.content,
					mediaIds: post.media?.map((m) => m.media_id) ?? []
				}
			];
			variants = new Map();
		}
		activePostIndex = 0;
		activeVariantAccountId = null;

		if (post.scheduled_at && post.scheduled_at !== '0001-01-01T00:00:00Z') {
			const schedule = workspaceScheduleFromISO(post.scheduled_at, scheduleTimezoneLabel);
			selectedDate = schedule?.date;
			selectedTime = schedule?.time ?? null;
		} else {
			selectedDate = undefined;
			selectedTime = null;
		}

		await loadAccounts(selectedWorkspaceId, selectedAccountIds);
		if (!source) {
			await loadVariants(post.id);
		}
		lastSavedSnapshot = getSaveSnapshot();
	}

	async function initializeComposer() {
		const requestSequence = ++workspaceRequestSequence;
		loadingWorkspaces = true;
		workspaceLoadError = '';

		try {
			if (workspaceCtx.workspaces.length === 0) {
				await workspaceCtx.initialize();
			}
			if (requestSequence !== workspaceRequestSequence) return;
			workspaces = [...workspaceCtx.workspaces];
			await initializeFromPost(initialPost);
		} catch (e) {
			console.error('Failed to load workspaces:', e);
			if (requestSequence === workspaceRequestSequence) {
				workspaceLoadError = m.compose_load_workspaces_failed();
			}
		} finally {
			if (requestSequence === workspaceRequestSequence) {
				loadingWorkspaces = false;
			}
		}
	}

	onMount(() => {
		void initializeComposer();
	});

	$effect(() => {
		const post = initialPost;
		if (!loadingWorkspaces && post && lastInitializedPostId !== post.id) {
			initializeFromPost(post);
		}
	});

	$effect(() => {
		const dateParam = initialScheduleDate;
		const workspaceParam = initialWorkspaceId;
		if (!loadingWorkspaces && !isEditMode) {
			void applyInitialComposerContext(dateParam, workspaceParam);
		}
	});

	$effect(() => {
		const workspaceId = workspaceCtx.currentWorkspace?.id ?? '';
		if (
			!isEditMode &&
			!initialWorkspaceId &&
			workspaceId &&
			workspaceId !== selectedWorkspaceId &&
			(!loadingWorkspaces || Boolean(selectedWorkspaceId))
		) {
			void handleWorkspaceChange(workspaceId);
		}
	});

	$effect(() => {
		const workspaceId = workspaceCtx.currentWorkspace?.id ?? '';
		const settingsFailed = workspaceCtx.settingsError;
		if (workspaceId && workspaceId === selectedWorkspaceId && settingsFailed) {
			workspaceSettingsError = m.compose_load_workspace_settings_failed();
		} else if (workspaceCtx.settingsReady && workspaceId === selectedWorkspaceId) {
			workspaceSettingsError = '';
		}
	});

	$effect(() => {
		String(editorResizeSignature);
		tick().then(() => {
			textareaRefs.forEach((el) => {
				if (el) autoResize(el);
			});
		});
	});

	$effect(() => {
		const text = ui.promptText;
		if (text && !initialPost && !loadingWorkspaces) {
			posts = [{ ...makeEmptyPost(), content: text }];
			activePostIndex = 0;
			ui.clearPrompt();
		}
	});

	$effect(() => {
		const selected = new Set(selectedAccountIds);
		let changed = false;
		const nextVariants = new SvelteMap<string, Record<string, VariantPost>>();
		for (const [accountId, value] of variants.entries()) {
			if (selected.has(accountId)) {
				const normalized = normalizeVariantRecord(value, posts);
				nextVariants.set(accountId, normalized);
				if (!variantRecordEquals(value, normalized, posts)) changed = true;
			} else {
				changed = true;
			}
		}
		if (changed) {
			variants = nextVariants;
		}
		if (activeVariantAccountId && !selected.has(activeVariantAccountId)) {
			activeVariantAccountId = null;
		}
	});

	// --------------------------------------------------------------------------
	// Data loading
	// --------------------------------------------------------------------------
	async function hydrateMediaMetadata(workspaceId: string, mediaIds: string[]) {
		const missingIds = Array.from(new Set(mediaIds.filter(Boolean))).filter(
			(id) => !mediaMimeTypes.has(id) || !mediaSizes.has(id)
		);
		if (!workspaceId || missingIds.length === 0) return;

		try {
			const token = getToken();
			const resp = await fetch(
				`${getApiBase()}/media/metadata?workspace_id=${encodeURIComponent(
					workspaceId
				)}&media_ids=${encodeURIComponent(missingIds.join(','))}`,
				{
					credentials: 'include',
					headers: token ? { Authorization: `Bearer ${token}` } : {}
				}
			);
			if (!resp.ok) return;

			const mediaData = await resp.json();
			const nextMimeTypes = new SvelteMap(mediaMimeTypes);
			const nextAltTexts = new SvelteMap(mediaAltTexts);
			const nextSizes = new SvelteMap(mediaSizes);
			for (const media of mediaData.media ?? []) {
				if (media.mime_type) {
					nextMimeTypes.set(media.id, media.mime_type);
				}
				if (typeof media.size === 'number') {
					nextSizes.set(media.id, media.size);
				}
				if (media.alt_text) {
					nextAltTexts.set(media.id, media.alt_text);
				}
			}
			mediaMimeTypes = nextMimeTypes;
			mediaAltTexts = nextAltTexts;
			mediaSizes = nextSizes;
		} catch (e) {
			console.error('Failed to load media metadata:', e);
		}
	}

	async function loadAccounts(
		workspaceId: string,
		preferredAccountIds: string[] | undefined = undefined
	) {
		const requestSequence = ++accountRequestSequence;
		if (!workspaceId) {
			accountsWorkspaceId = '';
			accounts = [];
			selectedAccountIds = [];
			accountLoadError = '';
			loadingAccounts = false;
			return;
		}

		const workspaceChanged = accountsWorkspaceId !== workspaceId;
		const selectionToPreserve = preferredAccountIds
			? [...preferredAccountIds]
			: workspaceChanged
				? undefined
				: [...selectedAccountIds];
		accountRetryIds = selectionToPreserve;
		accountsWorkspaceId = workspaceId;
		accountLoadError = '';
		loadingAccounts = true;

		if (workspaceChanged) {
			accounts = [];
			selectedAccountIds = [];
			activeVariantAccountId = null;
			if (preferredAccountIds === undefined) {
				variants = new Map();
			}
		}

		try {
			const { data, error: err } = await client.GET('/accounts', {
				params: { query: { workspace_id: workspaceId } }
			});
			if (err) throw new Error(err.detail || m.compose_load_accounts_failed());
			if (requestSequence !== accountRequestSequence || selectedWorkspaceId !== workspaceId) {
				return;
			}

			const nextAccounts = data ?? [];
			accounts = nextAccounts;
			if (selectionToPreserve && selectionToPreserve.length > 0) {
				const validIds = nextAccounts.map((account) => account.id);
				selectedAccountIds = selectionToPreserve.filter((id) => validIds.includes(id));
				if (selectedAccountIds.length === 0) {
					selectedAccountIds = nextAccounts.map((account) => account.id);
				}
			} else {
				selectedAccountIds = nextAccounts.map((account) => account.id);
			}
			sanitizeSelectedAccounts(nextAccounts);
		} catch (e) {
			console.error('Failed to load accounts:', e);
			if (requestSequence !== accountRequestSequence || selectedWorkspaceId !== workspaceId) {
				return;
			}
			accountLoadError = m.compose_load_accounts_failed();
		} finally {
			if (requestSequence === accountRequestSequence && selectedWorkspaceId === workspaceId) {
				loadingAccounts = false;
			}
		}
	}

	function handleWorkspaceChange(value: string) {
		if (!value || value === selectedWorkspaceId) return;
		const resetWorkspaceState = Boolean(
			draftId ||
			hasContent ||
			selectedDate ||
			selectedTime ||
			posts.some((post) => post.mediaIds.length > 0)
		);
		clearAutoSaveTimer();
		saveGeneration += 1;
		nextSlotRequestSequence += 1;
		suggestingSlot = false;
		draftId = null;
		lastSavedSnapshot = '';
		isSaving = false;
		showDeleteConfirm = false;
		selectedDate = undefined;
		selectedTime = null;
		showScheduleDialog = false;
		scheduleInput = '';
		scheduleInputError = '';
		randomDelayOverride = 'default';
		posts = posts.map((post) => ({ ...post, mediaIds: [] }));
		mediaAltTexts = new Map();
		mediaMimeTypes = new Map();
		mediaSizes = new Map();
		selectedWorkspaceId = value;
		accounts = [];
		selectedAccountIds = [];
		variants = new Map();
		activeVariantAccountId = null;
		accountLoadError = '';
		workspaceChangeNotice = resetWorkspaceState ? m.compose_workspace_context_reset() : '';
		void loadAccounts(value);
	}

	function toggleAccount(id: string) {
		if (selectedAccountIds.includes(id)) {
			selectedAccountIds = selectedAccountIds.filter((a) => a !== id);
			if (variants.has(id)) {
				const nextVariants = new SvelteMap(variants);
				nextVariants.delete(id);
				variants = nextVariants;
			}
			if (activeVariantAccountId === id) {
				activeVariantAccountId = null;
			}
		} else {
			selectedAccountIds = [...selectedAccountIds, id];
		}
		scheduleAutoSave();
	}

	function selectAllAccounts() {
		selectedAccountIds = accounts.map((a) => a.id);
		scheduleAutoSave();
	}

	function clearAllAccounts() {
		selectedAccountIds = [];
		scheduleAutoSave();
	}

	// --------------------------------------------------------------------------
	// Draft saving
	// --------------------------------------------------------------------------
	function scheduleAutoSave() {
		if (!autoSavesDraft) return;
		if (autoSaveTimer) clearTimeout(autoSaveTimer);
		autoSaveTimer = setTimeout(() => {
			if (!hasContent) return;
			const snapshot = getSaveSnapshot();
			if (snapshot !== lastSavedSnapshot) {
				saveDraft();
			}
		}, 2000);
	}

	async function saveDraft() {
		if (!selectedWorkspaceId || !hasContent) return;
		const generation = saveGeneration;
		const workspaceId = selectedWorkspaceId;
		const startingDraftId = draftId;
		const snapshot = getSaveSnapshot();
		isSaving = true;
		error = '';

		try {
			// Threads: store the encoded draft in the new dedicated
			// `thread_draft` field, and put the first post's text in
			// `content` so the parent row still carries a meaningful
			// preview. The backend is the source of truth on the
			// server side; the prefix is no longer stored on the
			// post row.
			const isThreadDraft_ = isThread;
			const sourcePosts = posts.map((post) => ({ ...post, mediaIds: [...post.mediaIds] }));
			const sourceVariants = new Map(variants);
			const threadDraft = isThreadDraft_
				? encodeThreadDraft(sourcePosts, getVariantPayloadForSave())
				: null;
			const draftContent = sourcePosts[0].content;
			const draftMediaIds = isThreadDraft_
				? sourcePosts.flatMap((post) => post.mediaIds)
				: sourcePosts[0].mediaIds;
			const variantPayload = getPersistedVariantPayload(sourceVariants, sourcePosts);

			const defaultDelay = effectiveRandomDelayMinutes;
			const body = {
				workspace_id: workspaceId,
				content: draftContent,
				social_account_ids: [...selectedAccountIds],
				media_ids: draftMediaIds,
				random_delay_minutes: defaultDelay,
				...(threadDraft ? { thread_draft: threadDraft } : {})
			};
			let savedDraftId = startingDraftId;
			if (startingDraftId) {
				const { error: patchErr } = await client.PATCH('/posts/{id}', {
					params: { path: { id: startingDraftId } },
					body: {
						content: draftContent,
						scheduled_at: '',
						social_account_ids: body.social_account_ids,
						media_ids: draftMediaIds,
						random_delay_minutes: defaultDelay,
						...(threadDraft ? { thread_draft: threadDraft } : {})
					}
				});
				if (patchErr) throw new Error((patchErr as any).detail || m.compose_update_draft_failed());
			} else {
				const { data, error: postErr } = await client.POST('/posts', { body });
				if (postErr) throw new Error((postErr as any).detail || m.compose_save_draft_failed());
				savedDraftId = data?.id ?? null;
			}

			if (savedDraftId && !isThreadDraft_) {
				await persistVariants(savedDraftId, variantPayload);
			}

			if (
				generation !== saveGeneration ||
				selectedWorkspaceId !== workspaceId ||
				draftId !== startingDraftId
			) {
				return;
			}
			const createdDraftId = startingDraftId ? null : savedDraftId;
			if (createdDraftId) draftId = createdDraftId;
			lastSavedSnapshot = snapshot;
			ui.triggerRefresh();
			if (createdDraftId) onDraftCreated?.(createdDraftId);
		} catch (e) {
			console.error('Failed to auto-save draft:', e);
			if (generation === saveGeneration && selectedWorkspaceId === workspaceId) {
				error = (e as Error).message || m.compose_save_draft_failed();
			}
		} finally {
			if (generation === saveGeneration && selectedWorkspaceId === workspaceId) {
				isSaving = false;
			}
		}
	}

	async function deleteDraft() {
		if (!draftId || isDeleting) return;
		clearAutoSaveTimer();
		isDeleting = true;
		error = '';
		try {
			const { error: deleteErr } = await client.DELETE('/posts/{id}', {
				params: { path: { id: draftId } }
			});
			if (deleteErr) throw new Error((deleteErr as any).detail || m.compose_delete_post_failed());

			ui.triggerRefresh();
			posts = [makeEmptyPost()];
			activePostIndex = 0;
			draftId = null;
			lastSavedSnapshot = '';
			variants = new Map();
			activeVariantAccountId = null;
			selectedDate = undefined;
			selectedTime = null;
			randomDelayOverride = 'default';
			showDeleteConfirm = false;
			onDeleted?.();
		} catch (e) {
			error = (e as Error).message || m.compose_delete_post_failed();
			soundPreferences.play('error');
		} finally {
			isDeleting = false;
		}
	}

	async function saveEditedPost() {
		if (!draftId || !initialPost) return;
		error = '';
		success = '';

		if (!selectedWorkspaceId) {
			error = m.compose_please_select_workspace();
			return;
		}
		if (!hasContent) {
			error = m.compose_please_enter_content();
			return;
		}
		if (selectedAccountIds.length === 0) {
			error = m.compose_select_account();
			return;
		}
		if ((selectedDate && !selectedTime) || (!selectedDate && selectedTime)) {
			error = m.compose_select_date_time();
			return;
		}
		if (selectedDate && selectedTime && !selectedWorkspaceSettingsReady) {
			error = m.compose_load_workspace_settings_failed();
			workspaceSettingsError = error;
			return;
		}
		const scheduledAt = getScheduledAt();
		if (selectedDate && selectedTime && !scheduledAt) {
			error = m.compose_invalid_timezone_time();
			return;
		}
		const isThreadDraft_ = isThread;
		const threadDraft = isThreadDraft_ ? encodeThreadDraft(posts, getVariantPayloadForSave()) : '';
		const draftContent = posts[0]?.content ?? '';
		const mediaIds = isThreadDraft_
			? posts.flatMap((post) => post.mediaIds)
			: (posts[0]?.mediaIds ?? []);
		const randomDelay = scheduledAt ? effectiveRandomDelayMinutes : 0;

		isSaving = true;
		try {
			if (isThreadDraft_ && scheduledAt) {
				await scheduleEditedThread(scheduledAt, randomDelay);
				lastSavedSnapshot = getSaveSnapshot();
				success = m.compose_changes_saved();
				soundPreferences.play('success');
				ui.triggerRefresh();
				if (onSuccess) {
					setTimeout(() => onSuccess(), 500);
				}
				return;
			}

			const { error: patchErr } = await client.PATCH('/posts/{id}', {
				params: { path: { id: draftId } },
				body: {
					content: draftContent,
					scheduled_at: isThreadDraft_ ? '' : (scheduledAt ?? ''),
					social_account_ids: selectedAccountIds,
					media_ids: mediaIds,
					random_delay_minutes: randomDelay,
					thread_draft: threadDraft
				}
			});
			if (patchErr) throw new Error((patchErr as any).detail || m.compose_save_changes_failed());

			if (!isThreadDraft_) {
				await persistVariants(draftId);
			}

			lastSavedSnapshot = getSaveSnapshot();
			success = m.compose_changes_saved();
			soundPreferences.play('success');
			ui.triggerRefresh();

			if (onSuccess) {
				setTimeout(() => onSuccess(), 500);
			}
		} catch (e) {
			error = (e as Error).message || m.compose_save_changes_failed();
			soundPreferences.play('error');
		} finally {
			isSaving = false;
		}
	}

	async function scheduleEditedThread(scheduledAt: string, randomDelay: number) {
		if (!draftId) return;
		const validPosts = posts.filter(
			(post) => post.content.trim().length > 0 || post.mediaIds.length > 0
		);
		if (validPosts.length < 2) {
			throw new Error(m.compose_thread_minimum());
		}
		if (syncedLinkedInThreadAccounts.length > 0) {
			activeVariantAccountId = syncedLinkedInThreadAccounts[0].id;
			throw new Error(m.compose_linkedin_thread_replies_unsupported());
		}

		const { data, error: createErr } = await client.POST('/posts/thread' as any, {
			body: {
				workspace_id: selectedWorkspaceId,
				social_account_ids: selectedAccountIds,
				scheduled_at: scheduledAt,
				random_delay_minutes: randomDelay,
				posts: validPosts.map((post) => ({
					content: post.content,
					media_ids: post.mediaIds
				}))
			}
		});
		if (createErr) throw new Error((createErr as any).detail || m.compose_schedule_thread_failed());
		if (data?.post_ids && variants.size > 0) {
			await persistThreadVariants(data.post_ids, validPosts);
		}

		const { error: deleteErr } = await client.DELETE('/posts/{id}', {
			params: { path: { id: draftId } }
		});
		if (deleteErr) {
			console.error('Scheduled thread but failed to delete original draft:', deleteErr);
		}
	}

	// --------------------------------------------------------------------------
	// Publishing
	// --------------------------------------------------------------------------
	async function publish(publishNow: boolean = false) {
		clearAutoSaveTimer();
		error = '';
		success = '';

		if (!selectedWorkspaceId) {
			error = m.compose_please_select_workspace();
			return;
		}
		if (!hasContent) {
			error = m.compose_please_enter_content();
			return;
		}
		if (selectedAccountIds.length === 0) {
			error = m.compose_select_account();
			return;
		}

		let scheduledAt: string | undefined;
		if (publishNow) {
			scheduledAt = new Date().toISOString();
		} else {
			if (!selectedWorkspaceSettingsReady) {
				error = m.compose_load_workspace_settings_failed();
				workspaceSettingsError = error;
				return;
			}
			scheduledAt = getScheduledAt();
			if (!scheduledAt) {
				error =
					selectedDate && selectedTime
						? m.compose_invalid_timezone_time()
						: m.compose_select_date_time();
				return;
			}
			if (new Date(scheduledAt).getTime() <= Date.now()) {
				error = m.compose_schedule_future();
				return;
			}
		}

		const randomDelay = publishNow ? 0 : effectiveRandomDelayMinutes;
		isSubmitting = true;

		try {
			if (isThread) {
				const validPosts = posts.filter(
					(p) => p.content.trim().length > 0 || p.mediaIds.length > 0
				);
				if (validPosts.length < 2) {
					error = m.compose_thread_minimum();
					isSubmitting = false;
					return;
				}
				if (syncedLinkedInThreadAccounts.length > 0) {
					error = m.compose_linkedin_thread_replies_unsupported();
					activeVariantAccountId = syncedLinkedInThreadAccounts[0].id;
					isSubmitting = false;
					return;
				}

				const { data, error: err } = await client.POST('/posts/thread' as any, {
					body: {
						workspace_id: selectedWorkspaceId,
						social_account_ids: selectedAccountIds,
						scheduled_at: scheduledAt,
						random_delay_minutes: randomDelay,
						posts: validPosts.map((p) => ({
							content: p.content,
							media_ids: p.mediaIds
						}))
					}
				});
				if (err) throw new Error((err as any).detail || m.compose_create_thread_failed());
				if (data?.post_ids && variants.size > 0) {
					await persistThreadVariants(data.post_ids, validPosts);
				}
			} else {
				const postId = draftId;
				// Always explicitly clear any leftover thread_draft on the
				// server side. We send an empty string so the backend
				// upsert removes the row. This covers the case where a
				// user had a multi-post thread draft, removed posts until
				// only one was left, and is now publishing a single post.
				const clearThreadDraft = isThread ? undefined : '';
				if (postId) {
					const { error: patchErr } = await client.PATCH('/posts/{id}', {
						params: { path: { id: postId } },
						body: {
							content: posts[0].content,
							scheduled_at: scheduledAt,
							social_account_ids: selectedAccountIds,
							media_ids: posts[0].mediaIds,
							random_delay_minutes: randomDelay,
							...(clearThreadDraft !== undefined ? { thread_draft: clearThreadDraft } : {})
						}
					});
					if (patchErr) throw new Error((patchErr as any).detail || m.compose_update_post_failed());
				} else {
					const { data, error: postErr } = await client.POST('/posts', {
						body: {
							workspace_id: selectedWorkspaceId,
							content: posts[0].content,
							social_account_ids: selectedAccountIds,
							scheduled_at: scheduledAt,
							media_ids: posts[0].mediaIds,
							random_delay_minutes: randomDelay
						}
					});
					if (postErr) throw new Error((postErr as any).detail || m.compose_create_post_failed());
					if (data?.id) draftId = data.id;
				}

				if (draftId) {
					await persistVariants(draftId);
				}
			}

			success = publishNow ? m.compose_publishing_now() : m.compose_scheduled_success();
			soundPreferences.play('success');
			ui.triggerRefresh();

			if (isEditMode && onSuccess) {
				setTimeout(() => onSuccess(), 800);
			} else {
				posts = [makeEmptyPost()];
				activePostIndex = 0;
				draftId = null;
				lastSavedSnapshot = '';
				variants = new Map();
				activeVariantAccountId = null;
				selectedDate = undefined;
				selectedTime = null;
				randomDelayOverride = 'default';
				onSuccess?.();
				setTimeout(() => (success = ''), 3000);
			}
		} catch (e) {
			error = (e as Error).message || m.compose_publish_failed();
			soundPreferences.play('error');
		} finally {
			isSubmitting = false;
		}
	}

	// --------------------------------------------------------------------------
	// Thread management
	// --------------------------------------------------------------------------
	function addPost() {
		const newIndex = activePostIndex + 1;
		posts = [...posts.slice(0, newIndex), makeEmptyPost(), ...posts.slice(newIndex)];
		variants = normalizeVariantsMap(variants, posts);
		activePostIndex = newIndex;
		onThreadStateChange?.(true);
		scheduleAutoSave();
		tick().then(() => {
			document.getElementById(`post-textarea-${newIndex}`)?.focus();
		});
	}

	function removePost(index: number) {
		if (posts.length <= 1) return;
		posts = posts.filter((_, i) => i !== index);
		variants = normalizeVariantsMap(variants, posts);
		if (activePostIndex >= posts.length) {
			activePostIndex = posts.length - 1;
		}
		onThreadStateChange?.(posts.length > 1);
		scheduleAutoSave();
	}

	function handleReorder(newItems: PostItem[]) {
		posts = newItems;
		variants = normalizeVariantsMap(variants, newItems);
		activePostIndex = Math.min(activePostIndex, newItems.length - 1);
		scheduleAutoSave();
	}

	// --------------------------------------------------------------------------
	// Media
	// --------------------------------------------------------------------------
	async function handleFileUpload(
		files: FileList | File[],
		targetPostIndex: number = activePostIndex
	) {
		if (!selectedWorkspaceId || isSubmitting) return;

		isUploading = true;
		let uploadedCount = 0;
		try {
			for (const file of Array.from(files)) {
				if (!isSupportedMediaFile(file)) continue;
				const targetPost = posts[targetPostIndex];
				if (!targetPost) break;
				if (getEditorMediaIdsForPost(targetPost).length >= 4) break;

				const data = await uploadMediaFile({ workspaceId: selectedWorkspaceId, file });
				if (data.mime_type) {
					const nextMimeTypes = new SvelteMap(mediaMimeTypes);
					nextMimeTypes.set(data.id, data.mime_type);
					mediaMimeTypes = nextMimeTypes;
				}
				if (typeof data.size === 'number') {
					const nextSizes = new SvelteMap(mediaSizes);
					nextSizes.set(data.id, data.size);
					mediaSizes = nextSizes;
				}
				addMediaToPost(targetPostIndex, data.id);
				uploadedCount++;
				scheduleAutoSave();
			}
			if (uploadedCount > 0) soundPreferences.play('success');
		} catch (e) {
			console.error('Failed to upload media:', e);
			error = (e as Error).message || m.compose_upload_failed();
			soundPreferences.play('error');
		} finally {
			isUploading = false;
		}
	}

	function handlePaste(e: ClipboardEvent, postIndex: number = activePostIndex) {
		const items = e.clipboardData?.items;
		if (!items) return;

		const files: File[] = [];
		for (const item of Array.from(items)) {
			if (item.kind === 'file') {
				const file = item.getAsFile();
				if (file) files.push(file);
			}
		}
		if (files.length > 0) {
			e.preventDefault();
			handleFileUpload(files, postIndex);
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		isDraggingFile = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		isDraggingFile = false;
	}

	function handleDrop(e: DragEvent, postIndex: number = activePostIndex) {
		e.preventDefault();
		isDraggingFile = false;
		const files = e.dataTransfer?.files;
		if (files && files.length > 0) {
			handleFileUpload(files, postIndex);
		}
	}

	function removeMedia(postIndex: number, mediaIndex: number) {
		const post = posts[postIndex];
		if (!post) return;
		const mediaIds = getEditorMediaIdsForPost(post);
		const mediaId = mediaIds[mediaIndex];
		if (activeVariantAccountId && activeVariantIsUnsynced) {
			setVariantMediaIds(
				activeVariantAccountId,
				postIndex,
				mediaIds.filter((_, mi) => mi !== mediaIndex)
			);
		} else {
			posts = posts.map((p, i) =>
				i === postIndex ? { ...p, mediaIds: p.mediaIds.filter((_, mi) => mi !== mediaIndex) } : p
			);
		}
		if (mediaId) {
			const newAlts = new SvelteMap(mediaAltTexts);
			newAlts.delete(mediaId);
			mediaAltTexts = newAlts;
			const newMimeTypes = new SvelteMap(mediaMimeTypes);
			newMimeTypes.delete(mediaId);
			mediaMimeTypes = newMimeTypes;
			const newSizes = new SvelteMap(mediaSizes);
			newSizes.delete(mediaId);
			mediaSizes = newSizes;
		}
		scheduleAutoSave();
	}

	function addMediaToPost(postIndex: number, mediaId: string) {
		const post = posts[postIndex];
		if (!post) return;
		if (activeVariantAccountId && activeVariantIsUnsynced) {
			setVariantMediaIds(activeVariantAccountId, postIndex, [
				...getEditorMediaIdsForPost(post),
				mediaId
			]);
			return;
		}
		posts = posts.map((p, i) =>
			i === postIndex ? { ...p, mediaIds: [...p.mediaIds, mediaId] } : p
		);
	}

	function setVariantMediaIds(accountId: string, index: number, mediaIds: string[]) {
		const postKey = posts[index]?.key;
		if (!postKey) return;
		const newVariants = new SvelteMap(variants);
		const current = {
			...normalizeVariantRecord(newVariants.get(accountId), posts),
			[postKey]: {
				content: getVariantContent(accountId, postKey) ?? posts[index].content,
				mediaIds
			}
		};
		newVariants.set(accountId, current);
		variants = newVariants;
	}

	function setMediaAltText(mediaId: string, alt: string) {
		const newAlts = new SvelteMap(mediaAltTexts);
		if (alt.trim()) {
			newAlts.set(mediaId, alt.trim());
		} else {
			newAlts.delete(mediaId);
		}
		mediaAltTexts = newAlts;

		// Persist to backend
		client
			.PATCH('/media/{id}', {
				params: { path: { id: mediaId } },
				body: { alt_text: alt.trim() }
			})
			.catch((e: any) => {
				console.error('Failed to save alt text:', e);
			});
	}

	// --------------------------------------------------------------------------
	// Prompts
	// --------------------------------------------------------------------------
	async function fetchRandomPrompt() {
		if (!selectedWorkspaceId) return;
		loadingPrompt = true;
		try {
			const { data, error: err } = await client.GET('/prompts/random', {
				params: { query: { workspace_id: selectedWorkspaceId } }
			});
			if (err) throw err;
			if (data) {
				currentPrompt = { text: data.text, category: data.category };
				showPromptCard = true;
			}
		} catch (e) {
			console.error('Failed to fetch prompt:', e);
		} finally {
			loadingPrompt = false;
		}
	}

	function dismissPrompt() {
		showPromptCard = false;
		currentPrompt = null;
	}

	// --------------------------------------------------------------------------
	// Variants
	// --------------------------------------------------------------------------
	function handleVariantChange(accountId: string, index: number, value: string) {
		const newVariants = new SvelteMap(variants);
		const postKey = posts[index]?.key;
		if (!postKey) return;
		const current = {
			...normalizeVariantRecord(newVariants.get(accountId), posts),
			[postKey]: {
				content: value,
				mediaIds: getVariantMediaIds(accountId, postKey) ?? [...posts[index].mediaIds]
			}
		};
		newVariants.set(accountId, current);
		variants = newVariants;
		scheduleAutoSave();
	}

	async function loadVariants(postId: string) {
		try {
			const { data, error: err } = await client.GET('/posts/{id}/variants', {
				params: { path: { id: postId } }
			});
			if (err) throw err;
			const nextVariants = new SvelteMap<string, Record<string, VariantPost>>();
			const variantMediaIds = new SvelteSet<string>();
			for (const variant of data?.variants ?? []) {
				if (variant.is_unsynced) {
					let mediaIds = [...(posts[0]?.mediaIds ?? [])];
					if (typeof variant.media_ids === 'string' && variant.media_ids !== '') {
						try {
							const parsed = JSON.parse(variant.media_ids);
							if (Array.isArray(parsed)) {
								mediaIds = parsed.map(String);
							}
						} catch (e) {
							console.error('Failed to parse variant media IDs:', e);
						}
					}
					for (const id of mediaIds) {
						variantMediaIds.add(id);
					}
					nextVariants.set(variant.social_account_id, {
						[posts[0]?.key ?? makeEmptyPost().key]: {
							content: variant.content,
							mediaIds
						}
					});
				}
			}
			variants = nextVariants;
			activeVariantAccountId = editorAccountIdAfterVariantLoad(
				activeVariantAccountId,
				selectedAccountIds,
				nextVariants.keys()
			);

			// Fetch metadata for variant-only media IDs not already hydrated.
			const missingIds = [...variantMediaIds].filter(
				(id) => !mediaMimeTypes.has(id) || !mediaSizes.has(id)
			);
			if (missingIds.length > 0) {
				await hydrateMediaMetadata(initialPost?.workspace_id ?? '', missingIds);
			}
		} catch (e) {
			console.error('Failed to load variants:', e);
			variants = new Map();
		}
	}

	async function persistVariants(
		postId: string,
		variantPayload = getPersistedVariantPayload(variants, posts)
	) {
		const { error: deleteErr } = await client.DELETE('/posts/{id}/variants', {
			params: { path: { id: postId } }
		});
		if (deleteErr) {
			throw new Error((deleteErr as any).detail || m.compose_reset_variants_failed());
		}

		if (variantPayload.length === 0) return;
		const { error: upsertErr } = await client.PUT('/posts/{id}/variants', {
			params: { path: { id: postId } },
			body: { variants: variantPayload }
		});
		if (upsertErr) {
			throw new Error((upsertErr as any).detail || m.compose_save_variants_failed());
		}
	}

	function activateVariantTab(accountId: string | null) {
		activeVariantAccountId = accountId;
	}

	function unsyncAccount(accountId: string) {
		if (!variants.has(accountId)) {
			variants = new Map([...variants, [accountId, makeVariantRecord(posts)]]);
		}
		activeVariantAccountId = accountId;
		scheduleAutoSave();
	}

	function resyncAccount(accountId: string) {
		if (!variants.has(accountId)) return;
		const nextVariants = new SvelteMap(variants);
		nextVariants.delete(accountId);
		variants = nextVariants;
		activeVariantAccountId = null;
		scheduleAutoSave();
	}

	async function persistThreadVariants(postIds: string[], sourcePosts: PostItem[]) {
		for (let index = 0; index < postIds.length; index++) {
			const postKey = sourcePosts[index]?.key;
			if (!postKey) continue;
			const payload = Array.from(variants.entries()).map(([accountId, values]) => ({
				social_account_id: accountId,
				content: values[postKey]?.content ?? sourcePosts[index]?.content ?? '',
				media_ids: JSON.stringify(values[postKey]?.mediaIds ?? sourcePosts[index]?.mediaIds ?? []),
				is_unsynced: true
			}));
			if (payload.length === 0) continue;
			const { error: upsertErr } = await client.PUT('/posts/{id}/variants', {
				params: { path: { id: postIds[index] } },
				body: { variants: payload }
			});
			if (upsertErr) {
				throw new Error((upsertErr as any).detail || m.compose_save_thread_variants_failed());
			}
		}
	}

	// --------------------------------------------------------------------------
	// Scheduling
	// --------------------------------------------------------------------------
	function applyScheduleInput(): boolean {
		const trimmed = scheduleInput.trim();
		if (!trimmed) {
			scheduleInputError = '';
			return true;
		}

		const parsed = parseNaturalScheduleInput(trimmed, new Date(), scheduleTimezoneLabel);
		if (!parsed) {
			scheduleInputError = m.compose_parse_time_failed();
			return false;
		}
		if (!workspaceScheduleToISO(parsed.date, parsed.time, scheduleTimezoneLabel)) {
			scheduleInputError = m.compose_invalid_timezone_time();
			return false;
		}

		selectedDate = parsed.date;
		selectedTime = parsed.time;
		scheduleInputError = '';
		return true;
	}

	function closeScheduleDialog() {
		scheduleInputError = '';
		showScheduleDialog = false;
	}

	function applyScheduleInputAndClose() {
		if (!applyScheduleInput()) return;
		closeScheduleDialog();
	}

	function openScheduleDialog() {
		if (!selectedWorkspaceSettingsReady) {
			error = m.compose_load_workspace_settings_failed();
			workspaceSettingsError = error;
			return;
		}
		scheduleInput = '';
		scheduleInputError = '';
		showScheduleDialog = true;
	}

	async function scheduleWithSelectedTime() {
		if (showScheduleDialog && !applyScheduleInput()) return;
		if (!selectedDate || !selectedTime) {
			openScheduleDialog();
			return;
		}
		showScheduleDialog = false;
		await publish(false);
	}

	async function fillNextSlot(showComposerError = false): Promise<boolean> {
		if (!selectedWorkspaceId) return false;
		if (!selectedWorkspaceSettingsReady) {
			workspaceSettingsError = m.compose_load_workspace_settings_failed();
			if (showComposerError) error = workspaceSettingsError;
			return false;
		}
		const requestSequence = ++nextSlotRequestSequence;
		const workspaceId = selectedWorkspaceId;
		const timeZone = scheduleTimezoneLabel;
		suggestingSlot = true;
		try {
			const { data, error: err } = await client.GET('/posting-schedules/next-slot', {
				params: {
					query: { workspace_id: workspaceId }
				}
			});
			if (
				requestSequence !== nextSlotRequestSequence ||
				selectedWorkspaceId !== workspaceId ||
				scheduleTimezoneLabel !== timeZone
			) {
				return false;
			}
			if (err) throw err;
			if (data?.slot_time) {
				// Parse date directly from ISO string to avoid timezone conversion issues
				const iso = data.slot_time as string;
				const [datePart, timePart] = iso.split('T');
				const [year, month, day] = datePart.split('-').map(Number);
				const rawHours = parseInt(timePart.split(':')[0], 10);
				const rawMinutes = parseInt(timePart.split(':')[1], 10);

				selectedDate = new CalendarDate(year, month, day);
				selectedTime = `${rawHours.toString().padStart(2, '0')}:${rawMinutes.toString().padStart(2, '0')}`;

				// Guard: if the slot is in the past, advance by one day
				const slotInstant = workspaceScheduleToISO(selectedDate, selectedTime, timeZone);
				if (!slotInstant) {
					scheduleInputError = m.compose_invalid_timezone_time();
					if (showComposerError) error = scheduleInputError;
					return false;
				}
				if (slotInstant && new Date(slotInstant).getTime() <= Date.now()) {
					selectedDate = selectedDate.add({ days: 1 });
				}
				scheduleInput = '';
				scheduleInputError = '';
				return true;
			}
			scheduleInputError = m.compose_no_free_slot();
			if (showComposerError) {
				error = scheduleInputError;
			}
		} catch (e) {
			if (requestSequence !== nextSlotRequestSequence || selectedWorkspaceId !== workspaceId) {
				return false;
			}
			console.error('Failed to get next available slot:', e);
			scheduleInputError = m.compose_next_free_slot_failed();
			if (showComposerError) {
				error = scheduleInputError;
			}
		} finally {
			if (requestSequence === nextSlotRequestSequence) suggestingSlot = false;
		}
		return false;
	}

	async function suggestNextSlot() {
		await fillNextSlot(false);
	}

	async function scheduleNextFreeSlot() {
		if (!selectedWorkspaceId) {
			error = m.compose_please_select_workspace();
			return;
		}
		if (!hasContent) {
			error = m.compose_please_enter_content();
			return;
		}
		if (selectedAccountIds.length === 0) {
			error = m.compose_select_account();
			return;
		}

		if (selectedDate && selectedTime) {
			showScheduleDialog = false;
			await publish(false);
			return;
		}

		const didApplySlot = await fillNextSlot(true);
		if (!didApplySlot) return;
		showScheduleDialog = false;
		await publish(false);
	}

	function formatScheduledDisplay(): string {
		if (!selectedDate) return m.compose_schedule();
		const now = workspaceClock(scheduleTimezoneLabel).date;
		const diffDays = selectedDate.compare(now);
		const timeSuffix = selectedTime ? ` ${selectedTime}` : '';

		if (diffDays === 0) return `${m.common_today()}${timeSuffix}`;
		if (diffDays === 1) return `${m.common_tomorrow()}${timeSuffix}`;
		const date = selectedDate.toDate(scheduleTimezoneLabel);
		return `${date.toLocaleDateString(getLocaleTag(), {
			month: 'short',
			day: 'numeric',
			timeZone: scheduleTimezoneLabel
		})}${timeSuffix}`;
	}

	function scheduleButtonLabel(): string {
		return m.compose_schedule_post({ schedule: formatScheduledDisplay() });
	}

	// --------------------------------------------------------------------------
	// Snippets
	// --------------------------------------------------------------------------
	function setPostContent(index: number, value: string) {
		posts = posts.map((p, pi) => (pi === index ? { ...p, content: value } : p));
		scheduleAutoSave();
	}

	function setEditorContent(index: number, value: string) {
		if (activeVariantAccountId && activeVariantIsUnsynced) {
			handleVariantChange(activeVariantAccountId, index, value);
			return;
		}
		setPostContent(index, value);
	}

	function setActivePost(index: number) {
		activePostIndex = index;
	}
</script>

<!-- ====================================================================== -->
<!-- Top Bar -->
<!-- ====================================================================== -->
<div class="flex flex-1 flex-col overflow-hidden">
	{#if !desktopComposerControls.current}
		<div
			class="sticky top-0 z-20 border-b bg-background/94 px-3 py-2 backdrop-blur-md"
			data-testid="mobile-composer-controls"
		>
			<div class="flex min-w-0 items-center gap-1.5">
				{#if modeControl}
					<div
						class="min-w-0 flex-1 [&_[data-testid=composer-mode-select]]:h-11 [&_[data-testid=composer-mode-select]]:w-full [&_[data-testid=composer-mode-select]]:max-w-none"
					>
						{@render modeControl()}
					</div>
				{/if}
				{#if accountControlLoading}
					<Button
						type="button"
						variant="outline"
						size="icon"
						class="size-11 shrink-0"
						disabled
						aria-label={m.compose_accounts_loading()}
						data-testid="composer-account-loading"
					>
						<LoaderIcon class="size-4 animate-spin" />
					</Button>
				{:else if accounts.length > 0}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Button
									{...props}
									data-testid="composer-account-control"
									variant="outline"
									size="sm"
									class="h-11 shrink-0 gap-1 px-2.5 text-xs"
									aria-label={m.compose_publish_to()}
								>
									<span class="tabular-nums">{selectedAccountIds.length}/{accounts.length}</span>
									<ChevronDownIcon class="size-3" />
								</Button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content class="w-72 max-w-[calc(100vw-1rem)]" align="start">
							<div class="flex items-center justify-between px-2 py-1.5">
								<span class="text-sm font-medium text-muted-foreground"
									>{m.compose_publish_to()}</span
								>
								<div class="flex gap-1">
									<Button variant="ghost" size="sm" onclick={selectAllAccounts}
										>{m.common_all()}</Button
									>
									<Button variant="ghost" size="sm" onclick={clearAllAccounts}
										>{m.common_none()}</Button
									>
								</div>
							</div>
							<DropdownMenu.Separator />
							{#each accounts as account (account.id)}
								<DropdownMenu.CheckboxItem
									class="min-h-11 gap-2"
									checked={selectedAccountIds.includes(account.id)}
									onCheckedChange={() => toggleAccount(account.id)}
								>
									<PlatformIcon platform={getPlatformKey(account.platform)} class="size-4" />
									<span class="min-w-0 truncate"
										>{getPlatformName(account.platform)}{account.account_username
											? ` @${account.account_username}`
											: ''}</span
									>
								</DropdownMenu.CheckboxItem>
							{/each}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				{/if}
				{#if isEditMode && !autoSavesDraft}
					<Button
						type="button"
						variant="outline"
						size="icon"
						class="ml-auto size-11 shrink-0"
						title={scheduleButtonLabel()}
						aria-label={scheduleButtonLabel()}
						onclick={openScheduleDialog}
						disabled={isSubmitting || isSaving}
					>
						<CalendarClockIcon class="size-4" />
					</Button>
					<Button
						size="sm"
						class="h-11 shrink-0 px-3"
						onclick={saveEditedPost}
						disabled={isSaving || isSubmitting || !hasContent || selectedAccountIds.length === 0}
					>
						{#if isSaving}<LoaderIcon class="size-3.5 animate-spin" />{/if}
						<span>{isSaving ? m.compose_saving_changes() : m.compose_save_changes()}</span>
					</Button>
				{:else}
					<Button
						type="button"
						variant="outline"
						size="icon"
						class="ml-auto size-11 shrink-0"
						title={scheduleButtonLabel()}
						aria-label={scheduleButtonLabel()}
						onclick={openScheduleDialog}
						disabled={isSubmitting || isSaving}
					>
						<CalendarClockIcon class="size-4" />
					</Button>
					<Button
						type="button"
						size="sm"
						class="h-11 shrink-0 px-3"
						onclick={() => publish(true)}
						disabled={isSubmitting || !hasContent || selectedAccountIds.length === 0}
					>
						{#if isSubmitting}<LoaderIcon class="size-3.5 animate-spin" />{/if}
						{m.compose_publish()}
					</Button>
				{/if}
			</div>
			<div class="mt-2 flex min-w-0 items-center gap-1.5 border-t border-border/70 pt-2">
				<div
					class="flex min-w-0 flex-1 [scrollbar-width:none] items-center gap-2 overflow-x-auto overflow-y-hidden px-0.5 py-0.5 [&::-webkit-scrollbar]:hidden"
					data-testid="mobile-rendition-scroller"
				>
					{#if selectedAccounts.length > 0}
						<button
							type="button"
							class={[
								'flex aspect-square size-11 min-h-11 max-w-11 min-w-11 flex-none shrink-0 items-center justify-center rounded-full border',
								activeVariantAccountId === null
									? 'border-foreground bg-foreground text-background'
									: 'border-border bg-background text-foreground'
							]}
							onclick={() => activateVariantTab(null)}
							title={m.compose_all_synced()}
							aria-label={m.compose_all_synced()}
							data-testid="mobile-rendition-all"
						>
							<Link2Icon class="size-4" />
						</button>
						{#each selectedAccounts as account (account.id)}
							{@const isUnsynced = variants.has(account.id)}
							<button
								type="button"
								class={[
									'relative flex aspect-square size-11 min-h-11 max-w-11 min-w-11 flex-none shrink-0 items-center justify-center rounded-full border',
									activeVariantAccountId === account.id
										? isUnsynced
											? 'border-amber-500/70 bg-amber-500/12 text-amber-700'
											: 'border-foreground bg-foreground text-background'
										: 'border-border bg-background text-foreground'
								]}
								onclick={() => activateVariantTab(account.id)}
								title={`${getPlatformName(account.platform)}${account.account_username ? ` @${account.account_username}` : ''}`}
								aria-label={`${getPlatformName(account.platform)}${account.account_username ? ` @${account.account_username}` : ''} · ${isUnsynced ? m.compose_custom_state() : m.compose_synced_state()}`}
								data-testid="mobile-rendition-account"
							>
								<PlatformIcon platform={getPlatformKey(account.platform)} class="size-4" />
								{#if isUnsynced}<span
										class="absolute -right-0.5 -bottom-0.5 flex size-4 items-center justify-center rounded-full bg-amber-500 text-white ring-2 ring-background"
										><UnlinkIcon class="size-2.5" /></span
									>{/if}
							</button>
						{/each}
					{/if}
				</div>
				{#if activeVariantAccount}
					<button
						type="button"
						class="flex aspect-square size-11 min-h-11 min-w-11 flex-none items-center justify-center rounded-full border bg-background"
						onclick={() =>
							activeVariantIsUnsynced
								? resyncAccount(activeVariantAccount.id)
								: unsyncAccount(activeVariantAccount.id)}
						title={activeVariantIsUnsynced ? m.compose_sync_back() : m.compose_unsync()}
						aria-label={activeVariantIsUnsynced ? m.compose_sync_back() : m.compose_unsync()}
					>
						{#if activeVariantIsUnsynced}<Link2Icon class="size-4" />{:else}<UnlinkIcon
								class="size-4"
							/>{/if}
					</button>
				{/if}
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<button
								{...props}
								type="button"
								class="flex aspect-square size-11 min-h-11 min-w-11 flex-none items-center justify-center rounded-full border bg-background"
								aria-label={m.sidebar_more()}><MoreHorizontalIcon class="size-4" /></button
							>
						{/snippet}
					</DropdownMenu.Trigger>
					<DropdownMenu.Content class="w-56" align="end">
						<DropdownMenu.Item
							class="min-h-11"
							onclick={() => (showPromptCard ? dismissPrompt() : fetchRandomPrompt())}
							><LightbulbIcon class="mr-2 size-4" />{showPromptCard
								? m.compose_dismiss_inspiration()
								: m.compose_need_inspiration()}</DropdownMenu.Item
						>
						{#if draftId}<DropdownMenu.Item
								class="min-h-11 text-destructive focus:text-destructive"
								onclick={() => (showDeleteConfirm = true)}
								disabled={isDeleting || isSaving || isSubmitting}
								><Trash2Icon class="mr-2 size-4" />{m.common_delete()}</DropdownMenu.Item
							>{/if}
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</div>
		</div>
	{:else}
		<div
			class="flex flex-wrap items-center justify-between gap-2 border-b px-4 py-3"
			data-testid="desktop-composer-controls"
		>
			<div class="flex flex-wrap items-center gap-2">
				{#if modeControl}
					{@render modeControl()}
				{/if}

				<!-- Account selector -->
				{#if accountControlLoading}
					<Button
						type="button"
						variant="ghost"
						size="sm"
						class="gap-1.5 text-xs"
						disabled
						data-testid="composer-account-loading"
					>
						<LoaderIcon class="size-3.5 animate-spin" />
						{m.compose_accounts_loading()}
					</Button>
				{:else if accounts.length > 0}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Button
									{...props}
									variant="ghost"
									size="sm"
									class="gap-1.5 text-xs"
									aria-label={m.compose_publish_to()}
									data-testid="composer-account-control"
								>
									<span class="hidden text-muted-foreground sm:inline">
										{selectedAccountIds.length === accounts.length
											? m.compose_all_accounts()
											: m.compose_account_count({ count: selectedAccountIds.length })}
									</span>
									<span class="text-muted-foreground sm:hidden"
										>{selectedAccountIds.length}/{accounts.length}</span
									>
									<ChevronDownIcon class="h-3 w-3" />
								</Button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content class="w-64" align="start">
							<div class="flex items-center justify-between px-2 py-1.5">
								<span class="text-sm font-medium text-muted-foreground"
									>{m.compose_publish_to()}</span
								>
								<div class="flex gap-1">
									<Button variant="ghost" size="xs" onclick={selectAllAccounts} class="h-5 text-xs"
										>{m.common_all()}</Button
									>
									<Button variant="ghost" size="xs" onclick={clearAllAccounts} class="h-5 text-xs"
										>{m.common_none()}</Button
									>
								</div>
							</div>
							<DropdownMenu.Separator />
							{#each accounts as account (account.id)}
								{@const isSelected = selectedAccountIds.includes(account.id)}
								{@const isUnsynced = variants.has(account.id)}
								<DropdownMenu.CheckboxItem
									checked={isSelected}
									onCheckedChange={() => toggleAccount(account.id)}
									class="gap-2"
								>
									<PlatformIcon platform={getPlatformKey(account.platform)} class="h-4 w-4" />
									<div class="flex flex-1 items-center gap-1.5">
										<span class="text-sm">{getPlatformName(account.platform)}</span>
										{#if account.account_username}<span class="text-xs text-muted-foreground"
												>@{account.account_username}</span
											>{/if}
									</div>
									{#if isUnsynced}<span class="text-xs text-amber-500">{m.compose_custom()}</span
										>{/if}
								</DropdownMenu.CheckboxItem>
							{/each}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				{/if}
			</div>

			<div class="flex flex-wrap items-center gap-1.5 md:gap-2">
				<!-- Per-account customization tabs -->
				{#if selectedAccounts.length > 0}
					<div
						class="flex max-w-[min(62vw,30rem)] [scrollbar-width:none] items-center gap-1 overflow-x-auto overflow-y-hidden py-1 pr-2 pl-1 [-ms-overflow-style:none] sm:max-w-[min(58vw,34rem)] lg:max-w-[40rem] lg:pr-3 [&::-webkit-scrollbar]:hidden"
					>
						<Tooltip.Root>
							<Tooltip.Trigger>
								{#snippet child({ props })}
									<button
										{...props}
										type="button"
										class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border transition-colors {activeVariantAccountId ===
										null
											? 'border-foreground bg-foreground text-background'
											: 'border-border bg-background text-foreground hover:border-foreground/30'}"
										onclick={() => activateVariantTab(null)}
										title={m.compose_all_synced()}
										aria-label={m.compose_all_synced()}
									>
										<Link2Icon class="h-3.5 w-3.5" />
									</button>
								{/snippet}
							</Tooltip.Trigger>
							<Tooltip.Content><p class="text-sm">{m.compose_all_synced()}</p></Tooltip.Content>
						</Tooltip.Root>

						{#each selectedAccounts as account (account.id)}
							{@const isUnsynced = variants.has(account.id)}
							<Tooltip.Root>
								<Tooltip.Trigger>
									{#snippet child({ props })}
										<button
											{...props}
											type="button"
											class="relative z-0 flex h-8 w-8 shrink-0 items-center justify-center overflow-visible rounded-full border transition-colors {activeVariantAccountId ===
											account.id
												? isUnsynced
													? 'border-amber-500/70 bg-amber-500/12 text-amber-700'
													: 'border-foreground bg-foreground text-background'
												: 'border-border bg-background text-foreground hover:border-foreground/30'}"
											onclick={() => activateVariantTab(account.id)}
											title={getPlatformName(account.platform)}
											aria-label={`${getPlatformName(account.platform)}${account.account_username ? ` @${account.account_username}` : ''} · ${isUnsynced ? m.compose_custom_state() : m.compose_synced_state()}`}
										>
											<PlatformIcon
												platform={getPlatformKey(account.platform)}
												class="h-3.5 w-3.5"
											/>
											{#if isUnsynced}
												<span
													class="absolute -right-1 -bottom-1 z-10 flex h-3.5 w-3.5 items-center justify-center rounded-full bg-amber-500 text-white shadow-sm ring-2 ring-background"
												>
													<UnlinkIcon class="h-2 w-2" />
												</span>
											{/if}
										</button>
									{/snippet}
								</Tooltip.Trigger>
								<Tooltip.Content>
									<p class="text-sm">
										{getPlatformName(account.platform)}{account.account_username
											? ` @${account.account_username}`
											: ''}{isUnsynced
											? ` · ${m.compose_custom_state()}`
											: ` · ${m.compose_synced_state()}`}
									</p>
								</Tooltip.Content>
							</Tooltip.Root>
						{/each}
					</div>

					{#if activeVariantAccount}
						<Tooltip.Root>
							<Tooltip.Trigger>
								{#snippet child({ props })}
									<button
										{...props}
										type="button"
										class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border bg-background text-foreground transition-colors hover:border-foreground/30"
										onclick={() =>
											activeVariantIsUnsynced
												? resyncAccount(activeVariantAccount.id)
												: unsyncAccount(activeVariantAccount.id)}
										title={activeVariantIsUnsynced ? m.compose_sync_back() : m.compose_unsync()}
										aria-label={activeVariantIsUnsynced
											? m.compose_sync_back()
											: m.compose_unsync()}
									>
										{#if activeVariantIsUnsynced}
											<Link2Icon class="h-3.5 w-3.5" />
										{:else}
											<UnlinkIcon class="h-3.5 w-3.5" />
										{/if}
									</button>
								{/snippet}
							</Tooltip.Trigger>
							<Tooltip.Content>
								<p class="text-sm">
									{activeVariantIsUnsynced ? m.compose_sync_back() : m.compose_unsync()}
								</p>
							</Tooltip.Content>
						</Tooltip.Root>
					{/if}
				{/if}

				<!-- Prompt -->
				<Tooltip.Root>
					<Tooltip.Trigger>
						{#snippet child({ props })}
							<Button
								{...props}
								variant="ghost"
								size="icon"
								class={showPromptCard ? 'text-amber-500' : ''}
								onclick={() => (showPromptCard ? dismissPrompt() : fetchRandomPrompt())}
								title={showPromptCard
									? m.compose_dismiss_inspiration()
									: m.compose_need_inspiration()}
								aria-label={showPromptCard
									? m.compose_dismiss_inspiration()
									: m.compose_need_inspiration()}
							>
								<LightbulbIcon class="h-4 w-4" />
							</Button>
						{/snippet}
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p class="text-sm">
							{showPromptCard ? m.compose_dismiss_inspiration() : m.compose_need_inspiration()}
						</p>
					</Tooltip.Content>
				</Tooltip.Root>

				{#if draftId}
					<Tooltip.Root>
						<Tooltip.Trigger>
							{#snippet child({ props })}
								<Button
									{...props}
									variant="ghost"
									size="icon"
									class="h-8 w-8 text-muted-foreground hover:text-destructive"
									aria-label={m.common_delete()}
									data-testid="composer-delete"
									onclick={() => (showDeleteConfirm = true)}
									disabled={isDeleting || isSaving || isSubmitting}
								>
									<Trash2Icon class="size-4" />
								</Button>
							{/snippet}
						</Tooltip.Trigger>
						<Tooltip.Content><p class="text-sm">{m.common_delete()}</p></Tooltip.Content>
					</Tooltip.Root>
				{/if}

				{#if isEditMode && !autoSavesDraft}
					<Button
						size="sm"
						class="gap-1.5"
						onclick={saveEditedPost}
						disabled={isSaving || isSubmitting || !hasContent || selectedAccountIds.length === 0}
					>
						{#if isSaving}<LoaderIcon class="h-3.5 w-3.5 animate-spin" />{/if}
						<span>{isSaving ? m.compose_saving_changes() : m.compose_save_changes()}</span>
					</Button>
				{:else}
					<div
						class="inline-flex overflow-hidden rounded-md border bg-background"
						title={formatScheduledDisplay()}
					>
						<Button
							type="button"
							variant="ghost"
							size="sm"
							class="h-8 rounded-none border-r border-border px-3 shadow-none"
							onclick={openScheduleDialog}
							disabled={isSubmitting || isSaving}
						>
							{formatScheduledDisplay()}
						</Button>
						<Tooltip.Root>
							<Tooltip.Trigger>
								{#snippet child({ props })}
									<Button
										{...props}
										type="button"
										variant="ghost"
										size="icon-sm"
										class="h-8 w-8 rounded-none shadow-none"
										aria-label={m.compose_schedule_next_slot()}
										onclick={scheduleNextFreeSlot}
										disabled={suggestingSlot ||
											isSubmitting ||
											!hasContent ||
											selectedAccountIds.length === 0}
									>
										{#if suggestingSlot || isSubmitting}
											<LoaderIcon class="h-3.5 w-3.5 animate-spin" />
										{:else}
											<ArrowRightIcon class="h-3.5 w-3.5" />
										{/if}
									</Button>
								{/snippet}
							</Tooltip.Trigger>
							<Tooltip.Content>
								<p class="text-sm">{m.compose_schedule_next_slot()}</p>
							</Tooltip.Content>
						</Tooltip.Root>
					</div>

					<Button
						type="button"
						size="sm"
						class="h-8 px-3"
						onclick={() => publish(true)}
						disabled={isSubmitting || !hasContent || selectedAccountIds.length === 0}
					>
						{#if isSubmitting}<LoaderIcon class="h-3.5 w-3.5 animate-spin" />{/if}
						{m.compose_publish()}
					</Button>
				{/if}
			</div>
		</div>
	{/if}

	<Dialog.Root bind:open={showScheduleDialog}>
		<Dialog.Content
			data-testid="schedule-dialog-shell"
			class="flex max-h-[calc(100dvh-1rem)] flex-col gap-0 overflow-hidden p-0 sm:max-w-3xl"
		>
			<Dialog.Header
				class="shrink-0 border-b px-5 pt-[max(1.25rem,env(safe-area-inset-top))] pb-4 text-center"
			>
				<Dialog.Title class="text-2xl font-semibold">{m.compose_schedule()}</Dialog.Title>
				<Dialog.Description class="text-sm text-muted-foreground">
					{m.compose_schedule_timezone({ timezone: scheduleTimezoneLabel })}
				</Dialog.Description>
			</Dialog.Header>

			<div data-testid="schedule-dialog-body" class="min-h-0 flex-1 space-y-4 overflow-y-auto p-5">
				<form
					class="space-y-2"
					onsubmit={(event) => {
						event.preventDefault();
						applyScheduleInputAndClose();
					}}
				>
					<Input
						bind:value={scheduleInput}
						placeholder={m.compose_schedule_input_placeholder()}
						class="h-10 bg-muted/40 text-base"
						aria-label={m.compose_schedule_time()}
					/>
					{#if scheduleInputError}
						<p class="px-1 text-xs text-destructive">{scheduleInputError}</p>
					{/if}
				</form>

				<div class="grid gap-2 sm:grid-cols-3">
					<Button
						type="button"
						variant="secondary"
						class="h-10 justify-center gap-2"
						onclick={suggestNextSlot}
						disabled={suggestingSlot}
					>
						{#if suggestingSlot}
							<LoaderIcon class="h-4 w-4 animate-spin" />
						{:else}
							<ArrowRightIcon class="h-4 w-4" />
						{/if}
						{m.compose_next_free_slot()}
					</Button>
					<Button
						type="button"
						variant="secondary"
						class="h-10 justify-center"
						onclick={() => {
							const next = workspaceClock(scheduleTimezoneLabel).date.add({ days: 1 });
							selectedDate = new CalendarDate(next.year, next.month, next.day);
							selectedTime = '09:00';
							scheduleInput = '';
							scheduleInputError = '';
						}}
					>
						{m.compose_tomorrow_time({ time: '09:00' })}
					</Button>
					<Button
						type="button"
						variant="secondary"
						class="h-10 justify-center"
						onclick={() => {
							const parsed = parseNaturalScheduleInput(
								'in 3 hours',
								new Date(),
								scheduleTimezoneLabel
							);
							if (!parsed) return;
							selectedDate = parsed.date;
							selectedTime = parsed.time;
							scheduleInput = '';
							scheduleInputError = '';
						}}
					>
						{m.compose_in_three_hours()}
					</Button>
				</div>

				<div class="overflow-hidden rounded-lg border bg-muted/15 md:grid md:grid-cols-[1fr_10rem]">
					<div class="flex justify-center p-3 md:p-4">
						<Calendar
							type="single"
							bind:value={selectedDate}
							minValue={workspaceClock(scheduleTimezoneLabel).date}
							class="bg-transparent p-0 [--cell-size:--spacing(9)]"
							weekdayFormat="short"
							weekStartsOn={workspaceCtx.weekStartsOn}
						/>
					</div>
					<div class="border-t md:border-t-0 md:border-l">
						<div class="border-b px-3 py-2 text-center text-sm font-medium">
							{m.compose_time()}
						</div>
						<div data-testid="schedule-dialog-time-list" class="p-2 md:max-h-72 md:overflow-y-auto">
							{#if timeSlots.length === 0}
								<p class="px-2 py-6 text-center text-xs text-muted-foreground">
									{m.compose_no_remaining_slots_today()}
								</p>
							{:else}
								<div class="grid grid-cols-2 gap-1.5 md:grid-cols-1">
									{#each timeSlots as time (time)}
										<Button
											type="button"
											variant={selectedTime === time ? 'default' : 'ghost'}
											size="sm"
											onclick={() => {
												if (!selectedDate) {
													const date = workspaceClock(scheduleTimezoneLabel).date;
													selectedDate = new CalendarDate(date.year, date.month, date.day);
												}
												selectedTime = time;
												scheduleInput = '';
												scheduleInputError = '';
											}}
											class="h-9 justify-center text-sm tabular-nums"
										>
											{time}
										</Button>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				</div>

				<div class="rounded-lg border bg-muted/10 p-3">
					<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<div class="space-y-1">
							<div class="text-sm font-medium">{m.compose_randomize_time()}</div>
							<div class="text-xs text-muted-foreground">
								{m.compose_workspace_default({
									delay: formatRandomDelay(workspaceCtx.settings.random_delay_minutes)
								})}.
								{m.compose_delay_applies_to_post()}
							</div>
						</div>
						<Select.Root
							type="single"
							value={randomDelayOverride}
							onValueChange={(value) => (randomDelayOverride = value || 'default')}
						>
							<Select.Trigger class="w-full sm:w-52">
								{#if randomDelayOverride === 'default'}
									{m.compose_workspace_default({
										delay: formatRandomDelay(workspaceCtx.settings.random_delay_minutes)
									})}
								{:else}
									{formatRandomDelay(effectiveRandomDelayMinutes)}
								{/if}
							</Select.Trigger>
							<Select.Content>
								<Select.Item value="default">
									{m.compose_workspace_default({
										delay: formatRandomDelay(workspaceCtx.settings.random_delay_minutes)
									})}
								</Select.Item>
								{#each randomDelaySelectOptions as minutes (minutes)}
									<Select.Item value={String(minutes)}>{formatRandomDelay(minutes)}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
				</div>

				<div class="flex flex-wrap items-center justify-between gap-3 text-sm">
					<div class="text-muted-foreground">
						{#if selectedDate && selectedTime}
							{m.compose_selected_schedule({ schedule: formatScheduledDisplay() })}
						{:else}
							{m.compose_select_date_time()}
						{/if}
					</div>
					{#if selectedDate || selectedTime}
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onclick={() => {
								selectedDate = undefined;
								selectedTime = null;
								scheduleInput = '';
								scheduleInputError = '';
							}}
						>
							{m.compose_clear_schedule()}
						</Button>
					{/if}
				</div>
			</div>

			<Dialog.Footer class="shrink-0 border-t px-5 pt-4 pb-[max(1rem,env(safe-area-inset-bottom))]">
				<Button type="button" variant="outline" onclick={closeScheduleDialog}
					>{m.common_cancel()}</Button
				>
				<Button type="button" variant="secondary" onclick={applyScheduleInputAndClose}
					>{m.common_done()}</Button
				>
				<Button
					type="button"
					onclick={scheduleWithSelectedTime}
					disabled={isSubmitting || !hasContent || selectedAccountIds.length === 0}
				>
					{#if isSubmitting}<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />{/if}
					{m.compose_schedule()}
				</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>

	<!-- ====================================================================== -->
	<!-- Messages -->
	<!-- ====================================================================== -->
	{#if workspaceLoadError}
		<div class="contents" data-testid="composer-workspaces-load-error">
			<InlineNotice tone="error" message={workspaceLoadError} class="mx-3 mt-2 md:mx-4 md:mt-3">
				{#snippet actions()}
					<Button
						variant="outline"
						size="sm"
						onclick={initializeComposer}
						disabled={loadingWorkspaces}
					>
						{m.common_retry()}
					</Button>
				{/snippet}
			</InlineNotice>
		</div>
	{:else if workspaceSettingsError}
		<div class="contents" data-testid="composer-workspace-settings-error">
			<InlineNotice tone="error" message={workspaceSettingsError} class="mx-3 mt-2 md:mx-4 md:mt-3">
				{#snippet actions()}
					<Button
						variant="outline"
						size="sm"
						onclick={retryComposerWorkspaceSettings}
						disabled={workspaceCtx.settingsLoading}
					>
						{m.common_retry()}
					</Button>
				{/snippet}
			</InlineNotice>
		</div>
	{:else if accountLoadError}
		<div class="contents" data-testid="composer-accounts-load-error">
			<InlineNotice tone="error" message={accountLoadError} class="mx-3 mt-2 md:mx-4 md:mt-3">
				{#snippet actions()}
					<Button
						variant="outline"
						size="sm"
						onclick={() => loadAccounts(selectedWorkspaceId, accountRetryIds)}
						disabled={loadingAccounts}
					>
						{m.common_retry()}
					</Button>
				{/snippet}
			</InlineNotice>
		</div>
	{/if}
	{#if workspaceChangeNotice}
		<InlineNotice
			tone="info"
			message={workspaceChangeNotice}
			class="mx-3 mt-2 md:mx-4 md:mt-3"
			onDismiss={() => (workspaceChangeNotice = '')}
			dismissLabel={m.common_dismiss()}
		/>
	{/if}
	{#if error}
		<InlineNotice
			tone="error"
			message={error}
			class="mx-3 mt-2 md:mx-4 md:mt-3"
			onDismiss={() => (error = '')}
			dismissLabel={m.common_dismiss()}
		/>
	{/if}
	{#if success}
		<InlineNotice
			tone="success"
			message={success}
			class="mx-3 mt-2 md:mx-4 md:mt-3"
			onDismiss={() => (success = '')}
			dismissLabel={m.common_dismiss()}
		/>
	{/if}
	{#if mediaCapabilityWarnings.length > 0}
		<div
			class="mx-3 mt-2 rounded-md border border-amber-500/20 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 md:mx-4 md:mt-3 dark:text-amber-300"
		>
			<div class="font-medium">{m.compose_media_compatibility_warning()}</div>
			<ul class="mt-1 list-disc space-y-0.5 pl-4">
				{#each mediaCapabilityWarnings as warning (warning)}
					<li>{warning}</li>
				{/each}
			</ul>
		</div>
	{/if}

	<!-- ====================================================================== -->
	<!-- Main Content Area -->
	<!-- ====================================================================== -->
	<div class="flex flex-1 overflow-hidden">
		<!-- Compose Column -->
		<div class="flex flex-1 flex-col overflow-y-auto">
			<div class="mx-auto w-full max-w-2xl px-3 py-4 md:px-6 md:py-6">
				<!-- Prompt Card -->
				{#if showPromptCard}
					<div class="relative mb-5 rounded border bg-muted/30 p-4 pr-24">
						<div class="absolute top-2 right-2 flex items-center gap-1">
							<Button
								variant="ghost"
								size="icon"
								class="text-muted-foreground"
								onclick={fetchRandomPrompt}
								disabled={loadingPrompt}
								title={m.compose_shuffle()}
								aria-label={m.compose_shuffle()}
							>
								<ShuffleIcon class="size-4" />
							</Button>
							<Button
								variant="ghost"
								size="icon"
								class="text-muted-foreground"
								onclick={dismissPrompt}
								title={m.compose_close()}
								aria-label={m.compose_close()}
							>
								<XIcon class="size-4" />
							</Button>
						</div>
						{#if loadingPrompt}
							<div class="space-y-2 py-2">
								<Skeleton class="h-3 w-full" />
								<Skeleton class="h-3 w-3/4" />
							</div>
						{:else if currentPrompt}
							<p class="text-sm leading-relaxed text-foreground/80">{currentPrompt.text}</p>
						{:else}
							<p class="text-sm text-muted-foreground">{m.compose_no_prompts()}</p>
						{/if}
					</div>
				{/if}

				<!-- Posts -->
				<div class="space-y-0">
					<ReorderableList
						items={posts}
						getKey={(post) => post.key}
						onUpdate={handleReorder}
						cssSelectorHandle=".drag-handle"
						direction="vertical"
					>
						{#snippet item(post, i)}
							<div
								class="group/post relative {isDraggingFile && activePostIndex === i
									? 'bg-primary/5'
									: ''}"
								role="region"
								aria-label={m.compose_drop_zone({ number: i + 1 })}
								ondragover={handleDragOver}
								ondragleave={handleDragLeave}
								ondrop={(e) => handleDrop(e, i)}
							>
								{#if isThread && i < posts.length - 1}
									<div class="absolute top-0 bottom-0 left-3 w-px bg-border"></div>
								{/if}

								<div class="relative flex gap-3 {isThread ? 'pl-7' : ''}">
									{#if isThread}
										<div class="relative flex flex-col items-center pt-3">
											<button
												type="button"
												class="drag-handle -ml-6 flex size-10 cursor-grab items-center justify-center rounded-md text-muted-foreground opacity-70 transition-opacity hover:bg-muted hover:opacity-100 active:cursor-grabbing md:-ml-4 md:size-6 md:opacity-0 md:group-hover/post:opacity-60"
												title={m.compose_drag_to_reorder()}
												aria-label={m.compose_drag_to_reorder()}
											>
												<GripVerticalIcon class="h-4 w-4" />
											</button>
										</div>
									{/if}

									<div class="min-w-0 flex-1">
										<div class="relative">
											<textarea
												id="post-textarea-{i}"
												aria-label={m.compose_post_text()}
												{@attach textareaAttachment(i)}
												value={getEditorContentForPost(post)}
												oninput={(e) => {
													const target = e.target as HTMLTextAreaElement;
													setEditorContent(i, target.value);
													autoResize(target);
												}}
												onpaste={(e) => handlePaste(e, i)}
												onfocus={() => setActivePost(i)}
												placeholder={activeVariantAccountId
													? activeVariantIsUnsynced
														? m.compose_write_custom_version({
																platform: getPlatformName(activeVariantAccount?.platform ?? '')
															})
														: m.compose_unsync_to_edit_placeholder()
													: i === 0
														? m.compose_whats_on_your_mind()
														: m.compose_add_to_thread()}
												class="w-full resize-none border-0 bg-transparent py-2 pr-3 text-base leading-relaxed text-foreground placeholder:text-muted-foreground/50 focus:ring-0 focus:outline-none md:py-3 md:pr-4 md:text-lg"
												style="min-height: {i === 0 ? '120px' : '56px'};"
												disabled={isSubmitting ||
													(!!activeVariantAccountId && !activeVariantIsUnsynced)}
											></textarea>

											{#if activeVariantAccountId && activePostIndex === i && !activeVariantIsUnsynced}
												<div class="absolute inset-x-0 bottom-0 px-1 pb-2">
													<div
														class="rounded-xl border border-dashed border-border/80 bg-background/95 px-3 py-2 text-xs text-muted-foreground shadow-sm"
													>
														<div class="flex flex-wrap items-center justify-between gap-2">
															<span>{m.compose_editor_locked_synced()}</span>
															<Button
																variant="outline"
																size="sm"
																class="h-7 gap-1 text-xs"
																onclick={() =>
																	activeVariantAccountId && unsyncAccount(activeVariantAccountId)}
															>
																<UnlinkIcon class="h-3.5 w-3.5" />
																{m.compose_unsync_to_edit()}
															</Button>
														</div>
													</div>
												</div>
											{/if}

											{#if isUploading && activePostIndex === i}
												<div
													class="absolute inset-0 flex items-center justify-center bg-background/80"
												>
													<LoaderIcon class="h-5 w-5 animate-spin text-primary" />
												</div>
											{/if}
										</div>

										<!-- Media grid -->
										{#if getEditorMediaIdsForPost(post).length > 0}
											<div
												class="mb-3 {getEditorMediaIdsForPost(post).length === 1
													? ''
													: 'grid grid-cols-2 gap-1.5'}"
											>
												{#each getEditorMediaIdsForPost(post) as mediaId, mi (mediaId)}
													{@const isFirstOfThree =
														getEditorMediaIdsForPost(post).length === 3 && mi === 0}
													<div
														class="group/media relative overflow-hidden rounded-lg {isFirstOfThree
															? 'col-span-2'
															: ''}"
													>
														{#if isVideoMedia(mediaId)}
															<video
																src={getAuthenticatedMediaByID(mediaId)}
																class="{getEditorMediaIdsForPost(post).length === 1
																	? 'aspect-video'
																	: 'aspect-square'} w-full object-cover"
																controls
																muted
																playsinline
															></video>
														{:else}
															<img
																src={getAuthenticatedMediaByID(mediaId)}
																alt={mediaAltTexts.get(mediaId) || ''}
																class="{getEditorMediaIdsForPost(post).length === 1
																	? 'aspect-video'
																	: 'aspect-square'} w-full object-cover"
															/>
														{/if}
														<div
															class="absolute top-2 right-2 flex items-center gap-1 opacity-100 md:opacity-0 md:transition-opacity md:group-focus-within/media:opacity-100 md:group-hover/media:opacity-100"
															data-testid="composer-media-actions"
														>
															<button
																type="button"
																class={[
																	'flex size-10 items-center justify-center rounded-md bg-black/75 text-white shadow-sm backdrop-blur-sm transition-colors hover:bg-black/90 md:size-7',
																	mediaAltTexts.get(mediaId)
																		? 'ring-2 ring-primary/80 ring-offset-1 ring-offset-transparent'
																		: ''
																]}
																aria-label={mediaAltTexts.get(mediaId)
																	? m.media_alt_text()
																	: m.media_add_alt_text()}
																title={mediaAltTexts.get(mediaId)
																	? m.media_alt_text()
																	: m.media_add_alt_text()}
																onclick={(e) => {
																	e.stopPropagation();
																	editingAltMediaId =
																		editingAltMediaId === mediaId ? null : mediaId;
																}}
															>
																<TypeIcon class="size-4 md:size-3.5" />
															</button>
															<button
																type="button"
																class="flex size-10 items-center justify-center rounded-md bg-black/75 text-white shadow-sm backdrop-blur-sm transition-colors hover:bg-red-600 md:size-7"
																aria-label={m.compose_remove_media()}
																title={m.compose_remove_media()}
																onclick={(e) => {
																	e.stopPropagation();
																	removeMedia(i, mi);
																}}
															>
																<XIcon class="size-4 md:size-3.5" />
															</button>
														</div>
														{#if editingAltMediaId === mediaId}
															<div
																class="absolute inset-x-0 bottom-0 bg-black/70 p-2 backdrop-blur-sm"
															>
																<textarea
																	value={mediaAltTexts.get(mediaId) || ''}
																	oninput={(e) =>
																		setMediaAltText(
																			mediaId,
																			(e.target as HTMLTextAreaElement).value
																		)}
																	placeholder={m.compose_alt_text_placeholder()}
																	rows={2}
																	class="w-full resize-none rounded bg-white/10 px-2 py-2 text-base text-white placeholder:text-white/60 focus:ring-2 focus:ring-white/70 focus:outline-none md:py-1 md:text-xs"
																	aria-label={m.media_alt_text()}
																></textarea>
																<div class="mt-1 flex justify-end gap-1">
																	<button
																		type="button"
																		class="text-[10px] text-white/70 hover:text-white"
																		onclick={() => (editingAltMediaId = null)}
																		>{m.common_done()}</button
																	>
																</div>
															</div>
														{/if}
													</div>
												{/each}
											</div>
										{/if}

										<!-- Bottom bar -->
										<div
											class="flex items-center gap-2 pb-2 transition-opacity {activePostIndex === i
												? 'opacity-100'
												: 'pointer-events-none opacity-0'}"
										>
											{#if isThread}<span
													class="text-[10px] font-medium text-muted-foreground/60 tabular-nums"
													>#{i + 1}</span
												>{/if}

											<label
												class={activeVariantAccountId && !activeVariantIsUnsynced
													? 'cursor-not-allowed opacity-50'
													: 'cursor-pointer'}
											>
												<input
													type="file"
													accept="image/*,video/*"
													multiple
													disabled={!!activeVariantAccountId && !activeVariantIsUnsynced}
													class="hidden"
													onchange={(e) => {
														const target = e.target as HTMLInputElement;
														if (target.files) handleFileUpload(target.files, i);
													}}
												/>
												<div
													class="flex size-11 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground md:size-7"
												>
													<ImageIcon class="h-3.5 w-3.5" />
												</div>
											</label>

											<Tooltip.Root>
												<Tooltip.Trigger>
													{#snippet child({ props })}
														<div {...props} class="flex cursor-default items-center gap-1.5">
															<svg
																class="h-4 w-4 {getCharCounterColor(
																	getEditorContentForPost(post).length,
																	editorMaxChars
																)}"
																viewBox="0 0 20 20"
															>
																<circle
																	cx="10"
																	cy="10"
																	r="8"
																	fill="none"
																	stroke="currentColor"
																	stroke-width="2.5"
																	opacity="0.15"
																/>
																<circle
																	cx="10"
																	cy="10"
																	r="8"
																	fill="none"
																	stroke={getCharCounterStrokeColor(
																		getEditorContentForPost(post).length,
																		editorMaxChars
																	)}
																	stroke-width="2.5"
																	stroke-linecap="round"
																	stroke-dasharray={50.27}
																	stroke-dashoffset={50.27 *
																		Math.max(
																			0,
																			1 - getEditorContentForPost(post).length / editorMaxChars
																		)}
																	transform="rotate(-90 10 10)"
																/>
															</svg>
															<span class="text-[10px] text-muted-foreground/60 tabular-nums"
																>{getEditorContentForPost(post).length}/{editorMaxChars}</span
															>
														</div>
													{/snippet}
												</Tooltip.Trigger>
												<Tooltip.Content>
													<div class="space-y-1">
														<p class="text-xs font-medium text-muted-foreground">
															{m.compose_character_limits()}
														</p>
														{#each editorPlatformLimits as pl (pl.key)}
															<div class="flex items-center justify-between gap-2 text-xs">
																<div class="flex items-center gap-1.5">
																	<PlatformIcon platform={pl.key} class="h-3 w-3" /><span
																		>{pl.platform}</span
																	>
																</div>
																<span
																	class="tabular-nums {getEditorContentForPost(post).length >
																	pl.limit
																		? 'text-red-500'
																		: 'text-muted-foreground'}"
																	>{getEditorContentForPost(post).length}/{pl.limit}</span
																>
															</div>
														{/each}
													</div>
												</Tooltip.Content>
											</Tooltip.Root>

											<button
												type="button"
												class="-mx-2 flex min-h-11 items-center gap-1.5 px-2 text-xs text-muted-foreground/60 transition-colors hover:text-foreground md:mx-0 md:min-h-7 md:px-0"
												onclick={addPost}
											>
												<PlusIcon class="h-3 w-3" />{m.compose_add_post()}
											</button>
										</div>

										{#if isThread}
											<button
												type="button"
												class="absolute top-1 right-0 flex size-10 items-center justify-center rounded-md text-muted-foreground opacity-70 transition-opacity hover:bg-muted hover:text-destructive md:top-3 md:size-7 md:opacity-0 md:group-hover/post:opacity-100"
												onclick={() => removePost(i)}
												title={m.compose_remove_post()}
												aria-label={m.compose_remove_post()}
											>
												<Trash2Icon class="h-3.5 w-3.5" />
											</button>
										{/if}
									</div>
								</div>
							</div>
						{/snippet}
					</ReorderableList>
				</div>
			</div>
		</div>
	</div>
</div>

<DestructiveConfirmDialog
	bind:open={showDeleteConfirm}
	title={m.sidebar_delete_draft_confirm()}
	description={m.compose_delete_draft_body()}
	onConfirm={deleteDraft}
/>
