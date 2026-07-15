<script lang="ts">
	import { onMount } from 'svelte';
	import { client, type SocialAccount, type Workspace } from '$lib/api/client';
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
	import PlatformIcon from './platform-icon.svelte';
	import { getPlatformKey, getPlatformName } from '$lib/utils';
	import { CalendarDate, getLocalTimeZone, isEqualDay, today } from '@internationalized/date';
	import {
		buildFocusedPublicationPayload,
		composerMode,
		roleFieldsForMode,
		type ComposerModeKey,
		type FocusedComposerFields,
		type FocusedFieldKey
	} from './compose/modes';
	import AlertCircleIcon from 'lucide-svelte/icons/alert-circle';
	import CalendarClockIcon from 'lucide-svelte/icons/calendar-clock';
	import ImagePlusIcon from 'lucide-svelte/icons/image-plus';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import SaveIcon from 'lucide-svelte/icons/save';
	import SendIcon from 'lucide-svelte/icons/send';
	import Settings2Icon from 'lucide-svelte/icons/settings-2';
	import UploadIcon from 'lucide-svelte/icons/upload';
	import XIcon from 'lucide-svelte/icons/x';

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
	}

	let { mode, initialPublication = null, onSuccess, onCancel }: Props = $props();

	let publicationId = $state('');
	let hydratedPublicationId = $state('');
	let workspaces = $state<Workspace[]>([]);
	let selectedWorkspaceId = $state('');
	let accounts = $state<SocialAccount[]>([]);
	let selectedAccountIds = $state<string[]>([]);
	let capabilities = $state<Capability[]>([]);
	let providerReadiness = $state<ProviderReadinessItem[]>([]);
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

	const modeMeta = $derived(composerMode(mode));
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
	const canQueue = $derived(canSaveDraft && localBlockers.length === 0);
	const isEditMode = $derived(Boolean(initialPublication?.id));
	const isToday = $derived(
		selectedDate ? isEqualDay(selectedDate, today(getLocalTimeZone())) : false
	);
	const allTimeSlots = $derived.by(() => {
		const start = workspaceCtx.settings.slot_start_hour;
		const end = workspaceCtx.settings.slot_end_hour;
		const interval = workspaceCtx.settings.slot_interval_minutes;
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
		if (!isToday) return allTimeSlots;
		const now = new Date();
		const currentMinutes = now.getHours() * 60 + now.getMinutes();
		return allTimeSlots.filter((slot) => {
			const [hour, minute] = slot.split(':').map(Number);
			return hour * 60 + minute > currentMinutes;
		});
	});

	onMount(async () => {
		if (initialPublication) hydrateInitialPublication(initialPublication);
		await loadInitialData();
	});

	$effect(() => {
		if (initialPublication && initialPublication.id !== hydratedPublicationId) {
			hydrateInitialPublication(initialPublication);
		}
	});

	async function loadInitialData() {
		loading = true;
		error = '';
		try {
			const [
				{ data: workspaceData, error: workspaceError },
				{ data: capabilityData, error: capError }
			] = await Promise.all([client.GET('/workspaces', {}), client.GET('/capabilities', {})]);
			if (workspaceError) throw new Error(workspaceError.detail || 'Failed to load workspaces');
			if (capError) throw new Error(capError.detail || 'Failed to load capabilities');
			workspaces = workspaceData ?? [];
			capabilities = capabilityData?.capabilities ?? [];
			selectedWorkspaceId =
				selectedWorkspaceId || workspaceCtx.currentWorkspace?.id || workspaces[0]?.id || '';
			if (selectedWorkspaceId) {
				await loadAccounts();
				await loadProviderReadiness();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load composer';
		} finally {
			loading = false;
		}
	}

	async function loadAccounts() {
		if (!selectedWorkspaceId) return;
		accountsLoading = true;
		try {
			const { data, error: err } = await client.GET('/accounts', {
				params: { query: { workspace_id: selectedWorkspaceId } }
			});
			if (err) throw new Error(err.detail || 'Failed to load accounts');
			accounts = (data ?? []).filter((account) => account.is_active);
			normalizeSelectedAccounts();
			settingsByAccount = normalizeAllAccountSettings(settingsByAccount);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load accounts';
		} finally {
			accountsLoading = false;
		}
	}

	async function loadProviderReadiness() {
		if (!selectedWorkspaceId) return;
		try {
			const { data, error: err } = await client.GET('/provider-readiness', {
				params: { query: { workspace_id: selectedWorkspaceId } }
			});
			if (err) throw new Error(err.detail || 'Failed to load provider readiness');
			providerReadiness = data?.providers ?? [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load provider readiness';
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
		const date = new Date(scheduledAt);
		selectedDate = new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate());
		selectedTime = `${date.getHours().toString().padStart(2, '0')}:${date
			.getMinutes()
			.toString()
			.padStart(2, '0')}`;
	}

	async function changeWorkspace() {
		selectedAccountIds = [];
		customizeAccountId = '';
		await loadAccounts();
		await loadProviderReadiness();
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
		if (capabilities.length === 0) return true;
		const provider = getPlatformKey(account.platform);
		return capabilities.some(
			(capability) => capability.provider === provider && capability.profile === mode
		);
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

	function disabledReason(account: SocialAccount): string {
		if (!isAccountCompatible(account)) {
			return `${getPlatformName(account.platform)} does not support ${modeMeta.label}`;
		}
		return accountBlockerText(account);
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
		uploading = true;
		error = '';
		try {
			for (const file of selected) {
				const uploaded = await uploadMediaFile({ workspaceId: selectedWorkspaceId, file });
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
			error = err instanceof Error ? err.message : 'Upload failed';
		} finally {
			uploading = false;
		}
	}

	async function handleThumbnail(files: FileList | null) {
		if (!files?.[0] || !selectedWorkspaceId) return;
		const file = files[0];
		if (!file.type.startsWith('image/')) {
			error = 'Choose an image thumbnail';
			return;
		}
		uploadingThumbnail = true;
		error = '';
		try {
			const uploaded = await uploadMediaFile({ workspaceId: selectedWorkspaceId, file });
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
			error = err instanceof Error ? err.message : 'Thumbnail upload failed';
		} finally {
			uploadingThumbnail = false;
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
		if (!selectedDate || !selectedTime) return undefined;
		const [hours, minutes] = selectedTime.split(':').map(Number);
		const date = selectedDate.toDate(getLocalTimeZone());
		date.setHours(hours, minutes, 0, 0);
		return date.toISOString();
	}

	function scheduleLabel(): string {
		if (!selectedDate || !selectedTime) return 'Schedule';
		const date = selectedDate.toDate(getLocalTimeZone());
		return `${date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })} ${selectedTime}`;
	}

	function clearSchedule() {
		selectedDate = undefined;
		selectedTime = null;
		showSchedulePopover = false;
	}

	function formBlockers(): string[] {
		const blockers: string[] = [];
		if (!selectedWorkspaceId) blockers.push('Choose a workspace.');
		if (selectedAccounts.length === 0) blockers.push('Choose at least one account.');
		for (const field of roleFields) {
			if (field.required && !fieldValue(field.key).trim()) {
				blockers.push(`${field.label} is required.`);
			}
		}
		const mediaMin = Math.max(
			0,
			...selectedCapabilities.map((capability) => capability.media.min_count)
		);
		if (mediaMin > 0 && media.length < mediaMin) {
			blockers.push(`Add at least ${mediaMin} media item${mediaMin === 1 ? '' : 's'}.`);
		}
		for (const account of selectedAccounts) {
			const blockersForAccount = readinessBlockers(account);
			if (blockersForAccount.length > 0) {
				blockers.push(
					`${getPlatformName(account.platform)} is blocked: ${accountBlockerText(account)}.`
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
					scheduled_at: payload.scheduled_at,
					metadata: payload.metadata
				}
			});
			if (updateError) throw new Error(updateError.detail || 'Failed to save publication');
			const { error: renditionError } = await client.PUT('/publications/{id}/renditions', {
				params: { path: { id: publicationId } },
				body: { renditions: payload.renditions }
			});
			if (renditionError) throw new Error(renditionError.detail || 'Failed to save outputs');
			return publicationId;
		}

		const { data, error: createError } = await client.POST('/publications', {
			body: payload
		});
		if (createError) throw new Error(createError.detail || 'Failed to create publication');
		publicationId = data.id;
		return data.id;
	}

	async function validatePublication(id: string): Promise<ValidationIssue[]> {
		const { data, error: err } = await client.POST('/publications/{id}/validate', {
			params: { path: { id } }
		});
		if (err) throw new Error(err.detail || 'Validation failed');
		validationIssues = data?.issues ?? [];
		return validationIssues;
	}

	async function runAction(action: 'draft' | 'validate' | 'schedule' | 'publish') {
		if (action !== 'draft' && localBlockers.length > 0) {
			error = localBlockers[0];
			success = '';
			return;
		}
		if (action === 'schedule' && !getScheduledAt()) {
			error = 'Choose a date and time before scheduling.';
			success = '';
			return;
		}
		saving = true;
		error = '';
		success = '';
		try {
			const id = await persistPublication();
			if (action === 'validate') {
				await validatePublication(id);
				success = 'Validation complete';
			} else if (action === 'schedule') {
				const issues = await validatePublication(id);
				if (issues.some((issue) => issue.severity === 'error')) {
					throw new Error('Fix blocking issues before scheduling.');
				}
				const { data, error: scheduleError } = await client.POST('/publications/{id}/schedule', {
					params: { path: { id } }
				});
				if (scheduleError) throw new Error(scheduleError.detail || 'Failed to schedule');
				success = data?.message ?? 'Publication scheduled';
			} else if (action === 'publish') {
				const issues = await validatePublication(id);
				if (issues.some((issue) => issue.severity === 'error')) {
					throw new Error('Fix blocking issues before publishing.');
				}
				const { data, error: publishError } = await client.POST('/publications/{id}/publish-now', {
					params: { path: { id } }
				});
				if (publishError) throw new Error(publishError.detail || 'Failed to publish');
				success = data?.message ?? 'Publication queued';
			} else {
				success = isEditMode ? 'Changes saved' : 'Draft saved';
			}
			ui.triggerRefresh();
			if (isEditMode && action !== 'validate') onSuccess?.();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Composer action failed';
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
	{#if loading}
		<div class="flex flex-1 items-center justify-center text-sm text-muted-foreground">
			<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
			Loading composer
		</div>
	{:else}
		<div class="border-b bg-background px-3 py-2 md:px-4">
			<div class="flex flex-wrap items-center gap-2">
				<select
					class="h-8 max-w-52 rounded-md border bg-background px-2 text-xs"
					bind:value={selectedWorkspaceId}
					onchange={changeWorkspace}
				>
					{#each workspaces as workspace (workspace.id)}
						<option value={workspace.id}>{workspace.name}</option>
					{/each}
				</select>
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-medium">{modeMeta.label}</p>
					<p class="truncate text-xs text-muted-foreground">{modeMeta.description}</p>
				</div>
				{#if onCancel}
					<Button variant="ghost" size="sm" class="h-8" onclick={onCancel} disabled={saving}>
						Cancel
					</Button>
				{/if}
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
					{isEditMode ? 'Save changes' : 'Save draft'}
				</Button>
				<Popover.Root bind:open={showSchedulePopover}>
					<Popover.Trigger>
						{#snippet child({ props })}
							<Button {...props} size="sm" class="h-8 gap-1.5" disabled={saving}>
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
								minValue={today(getLocalTimeZone())}
								class="bg-transparent p-0 [--cell-size:--spacing(8)]"
								weekdayFormat="short"
								weekStartsOn={workspaceCtx.weekStartsOn}
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
													const date = today(getLocalTimeZone());
													selectedDate = new CalendarDate(date.year, date.month, date.day);
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
										>Clear</Button
									>
									<Button
										size="sm"
										class="flex-1 text-xs"
										disabled={!canQueue || !getScheduledAt()}
										onclick={() => {
											showSchedulePopover = false;
											runAction('schedule');
										}}
									>
										Schedule
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
					Publish now
				</Button>
			</div>
		</div>

		<main class="min-h-0 flex-1 overflow-y-auto">
			<div class="mx-auto w-full max-w-3xl space-y-6 px-4 py-5 md:px-6">
				{#if error}
					<div
						class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
					>
						{error}
					</div>
				{/if}
				{#if success}
					<div
						class="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-700 dark:text-emerald-300"
					>
						{success}
					</div>
				{/if}

				<section class="space-y-3">
					<div class="flex items-center justify-between gap-3">
						<div>
							<h2 class="text-sm font-semibold">Accounts</h2>
						</div>
						{#if accountsLoading}
							<LoaderIcon class="h-4 w-4 animate-spin text-muted-foreground" />
						{/if}
					</div>
					<div class="flex flex-wrap gap-2">
						{#each accounts as account (account.id)}
							{@const selected = selectedAccountIds.includes(account.id)}
							{@const disabled = !isAccountCompatible(account)}
							<button
								type="button"
								class="flex max-w-full items-center gap-2 rounded-md border px-2.5 py-2 text-left text-sm transition {selected
									? 'border-primary bg-primary/5 text-foreground'
									: 'bg-background text-muted-foreground'} {disabled
									? 'cursor-not-allowed opacity-50'
									: 'hover:border-primary/60 hover:text-foreground'}"
								{disabled}
								title={disabledReason(account)}
								onclick={() => toggleAccount(account)}
							>
								<PlatformIcon platform={account.platform} class="h-4 w-4 shrink-0" />
								<span class="min-w-0 truncate">{accountLabel(account)}</span>
								<span class="shrink-0 text-xs text-muted-foreground">
									{getPlatformName(account.platform)}
								</span>
							</button>
						{/each}
					</div>
					{#if accounts.length === 0}
						<p class="text-sm text-muted-foreground">Connect an account before using this mode.</p>
					{/if}
				</section>

				<section class="flex flex-col gap-5">
					<div class="{modeMeta.mediaFirst ? 'order-1' : 'order-2'} space-y-3">
						<div class="rounded-md border border-dashed bg-muted/20 p-4">
							<div class="flex flex-wrap items-center gap-3">
								<label
									class="inline-flex h-9 cursor-pointer items-center gap-2 rounded-md border bg-background px-3 text-sm font-medium"
								>
									{#if uploading}
										<LoaderIcon class="h-4 w-4 animate-spin" />
									{:else}
										<UploadIcon class="h-4 w-4" />
									{/if}
									Upload media
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
									{mode === 'long_video'
										? 'Add the video file after the title and description are clear.'
										: 'Add the images or videos this post will publish.'}
								</p>
							</div>
							{#if media.length > 0}
								<div class="mt-4 grid gap-2 sm:grid-cols-2">
									{#each media as item (item.id)}
										<div class="group relative overflow-hidden rounded-md border bg-background">
											{#if isImage(item)}
												<img
													src={previewSrc(item)}
													alt={item.filename || 'Uploaded media'}
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
													aria-label="Remove media"
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
										class="inline-flex h-9 cursor-pointer items-center gap-2 rounded-md border bg-background px-3 text-sm font-medium"
									>
										{#if uploadingThumbnail}
											<LoaderIcon class="h-4 w-4 animate-spin" />
										{:else}
											<ImagePlusIcon class="h-4 w-4" />
										{/if}
										Thumbnail
										<input
											class="sr-only"
											type="file"
											accept="image/*"
											disabled={uploadingThumbnail || !selectedWorkspaceId}
											onchange={(event) => handleThumbnail(event.currentTarget.files)}
										/>
									</label>
									<p class="text-sm text-muted-foreground">Thumbnail · YouTube thumbnail</p>
									{#if thumbnailMediaId}
										<Button variant="ghost" size="sm" class="h-8 text-xs" onclick={clearThumbnail}>
											Clear
										</Button>
									{/if}
								</div>
								{#if thumbnailMedia}
									<img
										src={previewSrc(thumbnailMedia)}
										alt={thumbnailMedia.filename || 'Thumbnail'}
										class="mt-3 aspect-video max-h-48 rounded-md border object-cover"
									/>
								{:else if thumbnailMediaId}
									<p class="mt-2 text-xs text-muted-foreground">
										Existing thumbnail selected: {thumbnailMediaId}
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
								This mode is media-first for the selected accounts.
							</div>
						{/if}
					</div>
				</section>

				<section class="space-y-3 border-t pt-5">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<div>
							<h2 class="text-sm font-semibold">Platform settings</h2>
							<p class="text-xs text-muted-foreground">
								Open a platform only when you need its specific output options.
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
								Customize {getPlatformName(account.platform)} output
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
										Customize {getPlatformName(account.platform)} output
									</h3>
								</div>
								{#if settings.length === 0}
									<p class="text-sm text-muted-foreground">
										The important {getPlatformName(account.platform)} fields are already shown in the
										main composer.
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
					<section class="space-y-2 border-t pt-5 text-sm" aria-label="Publishing issues">
						<h2 class="font-semibold">Check before publishing</h2>
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
		</main>
	{/if}
</div>
