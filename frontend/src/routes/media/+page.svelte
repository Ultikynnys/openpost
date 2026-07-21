<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { client, type Workspace } from '$lib/api/client';
	import { getAuthenticatedMediaURL } from '$lib/media-url';
	import { videoProviderSupportDetail, videoProviderSupportLabel } from '$lib/media-capabilities';
	import { isSupportedMediaFile, uploadMediaFiles } from '$lib/media-upload-client';
	import { clampMediaPage } from '$lib/media-pagination';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Select from '$lib/components/ui/select';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Tabs, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import PageContainer from '$lib/components/page-container.svelte';
	import PageLoading from '$lib/components/page-loading.svelte';
	import EmptyState from '$lib/components/empty-state.svelte';
	import InlineNotice from '$lib/components/inline-notice.svelte';
	import AppToast from '$lib/components/app-toast.svelte';
	import DestructiveConfirmDialog from '$lib/components/destructive-confirm-dialog.svelte';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import ImageIcon from 'lucide-svelte/icons/image';
	import VideoIcon from 'lucide-svelte/icons/video';
	import HeartIcon from 'lucide-svelte/icons/heart';
	import TrashIcon from 'lucide-svelte/icons/trash-2';
	import UploadIcon from 'lucide-svelte/icons/upload';
	import DownloadIcon from 'lucide-svelte/icons/download';
	import ExternalLinkIcon from 'lucide-svelte/icons/external-link';
	import CheckIcon from 'lucide-svelte/icons/check';
	import ChevronLeftIcon from 'lucide-svelte/icons/chevron-left';
	import ChevronRightIcon from 'lucide-svelte/icons/chevron-right';
	import Grid2X2Icon from 'lucide-svelte/icons/grid-2x2';
	import MoreHorizontalIcon from 'lucide-svelte/icons/ellipsis';
	import { m } from '$lib/paraglide/messages';
	import { getLocaleTag } from '$lib/i18n';
	import { soundPreferences } from '$lib/stores/sound-preferences.svelte';

	interface MediaItem {
		id: string;
		workspace_id: string;
		mime_type: string;
		size: number;
		original_filename: string;
		width: number;
		height: number;
		alt_text: string;
		is_favorite: boolean;
		created_at: string;
		url: string;
		thumbnail_url: string;
		usage_count: number;
		can_delete: boolean;
		processing_status: string;
	}

	interface MediaUsage {
		post_id: string;
		content: string;
		status: string;
		scheduled_at: string;
	}

	interface BatchDeleteResult {
		deleted: number;
		failed_ids: string[];
	}

	type MediaDeletionRequest =
		{ kind: 'single'; media: MediaItem } | { kind: 'batch'; ids: string[] };

	let workspaces = $derived<Workspace[]>(workspaceCtx.workspaces);
	let selectedWorkspaceId = $derived(workspaceCtx.currentWorkspace?.id ?? '');
	let loading = $state(true);
	let error = $state('');
	let toastMessage = $state('');
	let toastTone = $state<'neutral' | 'success' | 'error'>('neutral');

	let mediaItems = $state<MediaItem[]>([]);
	let mediaLoading = $state(false);
	let mediaRequestSequence = 0;
	let totalCount = $state(0);
	let currentPage = $state(0);
	const pageSize = 40;

	let filter = $state<string>('all');
	let sort = $state<string>('newest');

	let uploadDialogOpen = $state(false);
	let uploadLoading = $state(false);
	let uploadError = $state('');
	let uploadProgress = $state('');

	let usageDialogOpen = $state(false);
	let selectedMedia = $state<MediaItem | null>(null);
	let mediaUsage = $state<MediaUsage[]>([]);
	let usageLoading = $state(false);
	let usageError = $state('');
	let usageRequestSequence = 0;

	let deleteDialogOpen = $state(false);
	let deletionRequest = $state.raw<MediaDeletionRequest | null>(null);

	const selectedMediaIds = new SvelteSet<string>();
	let isSelectionMode = $state(false);

	function notify(message: string, tone: 'neutral' | 'success' | 'error' = 'neutral') {
		toastMessage = message;
		toastTone = tone;
	}

	function selectedCountLabel(count: number) {
		return count === 1 ? m.media_selected_one() : m.media_selected_many({ count });
	}

	function deletedCountLabel(count: number) {
		return count === 1 ? m.media_deleted_one() : m.media_deleted_many({ count });
	}

	function uploadedCountLabel(count: number) {
		return count === 1 ? m.media_uploaded_one() : m.media_uploaded_many({ count });
	}

	function filesCountLabel(count: number) {
		return count === 1 ? m.media_files_one() : m.media_files_many({ count });
	}

	function usedInCountLabel(count: number) {
		return count === 1 ? m.media_used_in_one() : m.media_used_in_many({ count });
	}

	function usageSummaryLabel(count: number) {
		return count === 1 ? m.media_usage_summary_one() : m.media_usage_summary_many({ count });
	}

	function mediaUsageStatusLabel(status: string) {
		switch (status.toLowerCase()) {
			case 'published':
			case 'success':
				return m.activity_status_published();
			case 'failed':
				return m.activity_status_failed();
			case 'scheduled':
				return m.activity_status_scheduled();
			case 'publishing':
				return m.activity_status_publishing();
			case 'completed':
				return m.activity_status_completed();
			case 'processing':
				return m.activity_status_processing();
			case 'pending':
				return m.activity_status_pending();
			case 'draft':
				return m.activity_status_draft();
			default:
				return status;
		}
	}

	function mediaViewKey() {
		return `${selectedWorkspaceId}:${filter}:${sort}:${currentPage}`;
	}

	async function loadWorkspaces() {
		try {
			if (workspaceCtx.workspaces.length === 0 || !workspaceCtx.currentWorkspace) {
				await workspaceCtx.initialize();
			}
		} catch (e) {
			console.error('Failed to load workspaces:', e);
			error = m.media_load_failed();
		} finally {
			loading = false;
		}
	}

	async function loadMedia(workspaceID = selectedWorkspaceId) {
		if (!workspaceID) {
			mediaRequestSequence++;
			mediaLoading = false;
			mediaItems = [];
			totalCount = 0;
			return;
		}
		const requestSequence = ++mediaRequestSequence;
		const isCurrentRequest = () =>
			requestSequence === mediaRequestSequence && selectedWorkspaceId === workspaceID;
		mediaLoading = true;
		error = '';
		selectedMediaIds.clear();
		isSelectionMode = false;
		try {
			const { data, error: err } = await client.GET('/media', {
				params: {
					query: {
						workspace_id: workspaceID,
						filter: filter,
						sort: sort,
						limit: pageSize,
						offset: currentPage * pageSize
					}
				}
			});
			if (err) throw new Error(err.detail || m.media_load_failed());
			if (!isCurrentRequest()) return;
			const nextTotalCount = data?.total ?? 0;
			const clampedPage = clampMediaPage(currentPage, nextTotalCount, pageSize);
			if (clampedPage !== currentPage) {
				currentPage = clampedPage;
				await loadMedia(workspaceID);
				return;
			}
			mediaItems = (data?.media ?? []) as unknown as MediaItem[];
			totalCount = nextTotalCount;
		} catch (e) {
			if (!isCurrentRequest()) return;
			error = (e as Error).message;
			mediaItems = [];
		} finally {
			if (isCurrentRequest()) mediaLoading = false;
		}
	}

	async function toggleFavorite(mediaId: string) {
		try {
			const { data, error: err } = await client.PATCH('/media/{id}/favorite', {
				params: { path: { id: mediaId } }
			});
			if (err) throw new Error(err.detail || m.media_favorite_failed());
			const item = mediaItems.find((m) => m.id === mediaId);
			if (item) {
				item.is_favorite = data?.is_favorite ?? !item.is_favorite;
			}
		} catch (e) {
			notify((e as Error).message, 'error');
		}
	}

	async function toggleFavoriteBatch() {
		const ids = Array.from(selectedMediaIds);
		for (const id of ids) {
			await toggleFavorite(id);
		}
		selectedMediaIds.clear();
		isSelectionMode = false;
	}

	function requestDeleteMedia(media: MediaItem) {
		deletionRequest = { kind: 'single', media };
		deleteDialogOpen = true;
	}

	function requestDeleteSelectedBatch() {
		const ids = [...selectedDeletableIds];
		if (ids.length === 0) return;
		deletionRequest = { kind: 'batch', ids };
		deleteDialogOpen = true;
	}

	async function deleteMedia(mediaId: string) {
		const requestViewKey = mediaViewKey();
		try {
			const { error: err } = await client.DELETE('/media/{id}', {
				params: { path: { id: mediaId } }
			});
			if (err) throw new Error(err.detail || m.media_delete_failed());
			if (requestViewKey === mediaViewKey()) {
				const nextTotalCount = Math.max(0, totalCount - 1);
				const clampedPage = clampMediaPage(currentPage, nextTotalCount, pageSize);
				totalCount = nextTotalCount;
				if (clampedPage !== currentPage) {
					currentPage = clampedPage;
					await loadMedia();
				} else {
					mediaItems = mediaItems.filter((m) => m.id !== mediaId);
				}
			} else {
				await loadMedia();
			}
			notify(deletedCountLabel(1), 'success');
		} catch (e) {
			notify((e as Error).message, 'error');
		}
	}

	async function deleteSelectedBatch(ids: string[]) {
		if (ids.length === 0) return;
		try {
			const { data, error: err } = await client.POST('/media/batch-delete', {
				body: { media_ids: ids }
			});
			if (err) throw new Error(err.detail || m.media_delete_failed());

			const result = (data ?? { deleted: 0, failed_ids: ids }) as BatchDeleteResult;
			await loadMedia();

			const failedCount = Math.max(result.failed_ids?.length ?? 0, ids.length - result.deleted);
			if (result.deleted === 0) {
				notify(m.media_deleted_none(), 'error');
			} else if (failedCount > 0) {
				notify(
					m.media_deleted_partial({ deleted: result.deleted, failed: failedCount }),
					'neutral'
				);
			} else {
				notify(deletedCountLabel(result.deleted), 'success');
			}
		} catch (e) {
			notify((e as Error).message, 'error');
		}
	}

	async function confirmMediaDeletion() {
		const request = deletionRequest;
		if (!request) return;
		if (request.kind === 'single') {
			await deleteMedia(request.media.id);
			return;
		}
		await deleteSelectedBatch(request.ids);
	}

	async function downloadMedia(media: MediaItem) {
		try {
			const response = await fetch(getAuthenticatedMediaURL(media.url), { credentials: 'include' });
			if (!response.ok) throw new Error(m.media_download_failed());

			const blob = await response.blob();
			const objectURL = URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = objectURL;
			link.download = media.original_filename || `${media.id}.${extensionForMime(media.mime_type)}`;
			document.body.appendChild(link);
			link.click();
			link.remove();
			URL.revokeObjectURL(objectURL);
		} catch (e) {
			notify((e as Error).message, 'error');
		}
	}

	async function showUsage(media: MediaItem) {
		const mediaID = media.id;
		const requestSequence = ++usageRequestSequence;
		const isCurrentRequest = () =>
			requestSequence === usageRequestSequence && usageDialogOpen && selectedMedia?.id === mediaID;
		selectedMedia = media;
		usageDialogOpen = true;
		usageLoading = true;
		usageError = '';
		mediaUsage = [];
		try {
			const { data, error: err } = await client.GET('/media/{id}/usage', {
				params: { path: { id: media.id } }
			});
			if (err) throw new Error(err.detail || m.media_usage_load_failed());
			if (!isCurrentRequest()) return;
			mediaUsage = (data?.usage ?? []) as unknown as MediaUsage[];
		} catch (e) {
			if (!isCurrentRequest()) return;
			usageError = (e as Error).message;
		} finally {
			if (isCurrentRequest()) usageLoading = false;
		}
	}

	function handleUsageDialogOpenChange(nextOpen: boolean) {
		usageDialogOpen = nextOpen;
		if (nextOpen) return;
		usageRequestSequence++;
		usageLoading = false;
		usageError = '';
		mediaUsage = [];
		selectedMedia = null;
	}

	async function handleUpload() {
		if (!selectedWorkspaceId) return;
		uploadLoading = true;
		uploadError = '';

		const fileInput = document.getElementById('file-upload') as HTMLInputElement;
		const batchFileInput = document.getElementById('batch-file-upload') as HTMLInputElement;
		const selectedFiles = fileInput?.files?.length
			? Array.from(fileInput.files)
			: Array.from(batchFileInput?.files ?? []);
		const files = selectedFiles.filter(isSupportedMediaFile);
		if (files.length === 0) {
			uploadError = m.media_select_file_error();
			uploadLoading = false;
			return;
		}
		if (files.length > 10) {
			uploadError = m.media_max_files_error();
			uploadLoading = false;
			return;
		}

		try {
			uploadProgress =
				files.length === 1 ? m.media_uploading() : m.media_uploading_count({ count: files.length });
			const uploaded = await uploadMediaFiles(selectedWorkspaceId, files, (done, total) => {
				uploadProgress =
					total === 1 ? m.media_finalizing() : m.media_uploaded_progress({ done, total });
			});

			uploadDialogOpen = false;
			fileInput.value = '';
			if (batchFileInput) batchFileInput.value = '';
			notify(uploadedCountLabel(uploaded.length), 'success');
			soundPreferences.play('success');
			await loadMedia();
		} catch (e) {
			uploadError = (e as Error).message;
			soundPreferences.play('error');
		} finally {
			uploadLoading = false;
			uploadProgress = '';
		}
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString(getLocaleTag(), {
			month: 'short',
			day: 'numeric',
			timeZone: workspaceCtx.settings.timezone || 'UTC'
		});
	}

	function isImage(mimeType: string): boolean {
		return mimeType.startsWith('image/');
	}

	function isVideo(mimeType: string): boolean {
		return mimeType.startsWith('video/');
	}

	function canDeleteMedia(media: MediaItem): boolean {
		return media.can_delete ?? media.usage_count === 0;
	}

	function extensionForMime(mimeType: string): string {
		if (mimeType === 'image/jpeg') return 'jpg';
		if (mimeType === 'image/png') return 'png';
		if (mimeType === 'image/webp') return 'webp';
		if (mimeType === 'image/gif') return 'gif';
		if (mimeType === 'video/mp4') return 'mp4';
		if (mimeType === 'video/webm') return 'webm';
		return 'bin';
	}

	function toggleSelection(mediaId: string) {
		if (selectedMediaIds.has(mediaId)) {
			selectedMediaIds.delete(mediaId);
		} else {
			selectedMediaIds.add(mediaId);
		}
		isSelectionMode = selectedMediaIds.size > 0;
	}

	function selectAll() {
		const deletableMedia = mediaItems.filter(canDeleteMedia);
		if (deletableMedia.every((media) => selectedMediaIds.has(media.id))) {
			deletableMedia.forEach((media) => selectedMediaIds.delete(media.id));
		} else {
			deletableMedia.forEach((m) => selectedMediaIds.add(m.id));
		}
		isSelectionMode = selectedMediaIds.size > 0;
	}

	function cancelSelection() {
		selectedMediaIds.clear();
		isSelectionMode = false;
	}

	async function changeWorkspace(value: string) {
		if (!value || value === selectedWorkspaceId) return;
		const workspace = workspaces.find((candidate) => candidate.id === value);
		if (!workspace) return;
		currentPage = 0;
		await workspaceCtx.setWorkspace(workspace);
	}

	function changeFilter(value: string) {
		if (!value || value === filter) return;
		filter = value;
		currentPage = 0;
		void loadMedia();
	}

	function changeSort(value: string) {
		if (!value || value === sort) return;
		sort = value;
		currentPage = 0;
		void loadMedia();
	}

	function nextPage() {
		if ((currentPage + 1) * pageSize < totalCount) {
			currentPage++;
			loadMedia();
		}
	}

	function prevPage() {
		if (currentPage > 0) {
			currentPage--;
			loadMedia();
		}
	}

	onMount(() => {
		void loadWorkspaces();
	});

	$effect(() => {
		const workspaceID = selectedWorkspaceId;
		untrack(() => void loadMedia(workspaceID));
	});

	const filterTabs = $derived([
		{ value: 'all', label: m.media_filter_all() },
		{ value: 'used', label: m.media_filter_used() },
		{ value: 'unused', label: m.media_filter_unused() },
		{ value: 'favorites', label: m.media_filter_favorites() }
	]);

	const totalPages = $derived(Math.ceil(totalCount / pageSize));
	const deletableCount = $derived(mediaItems.filter(canDeleteMedia).length);
	const selectedDeletableIds = $derived(
		mediaItems
			.filter((media) => selectedMediaIds.has(media.id) && canDeleteMedia(media))
			.map((media) => media.id)
	);

	const descriptionText = $derived.by(() => {
		if (totalCount > 0) {
			let text: string = filesCountLabel(totalCount);
			if (filter === 'unused') {
				text += ` (${m.media_unused_suffix({ count: totalCount })})`;
			}
			return text;
		}
		return m.media_library_description();
	});
</script>

<svelte:head>
	<title>{m.media_library_title()} - {m.common_openpost()}</title>
</svelte:head>

{#if toastMessage}
	<AppToast
		message={toastMessage}
		tone={toastTone}
		dismissLabel={m.common_close()}
		onDismiss={() => (toastMessage = '')}
	/>
{/if}

<PageContainer
	title={m.media_library_title()}
	description={descriptionText}
	icon={ImageIcon}
	loading={loading || (mediaLoading && mediaItems.length === 0)}
	loadingMessage={m.common_loading()}
	loadingLayout="gallery"
>
	{#snippet actions()}
		{#if workspaces && workspaces.length > 1}
			<Select.Root type="single" value={selectedWorkspaceId} onValueChange={changeWorkspace}>
				<Select.Trigger class="w-[160px]">
					{workspaces.find((w) => w.id === selectedWorkspaceId)?.name || m.sidebar_workspace()}
				</Select.Trigger>
				<Select.Content>
					{#each workspaces as workspace (workspace.id)}
						<Select.Item value={workspace.id}>{workspace.name}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		{/if}
		<Button onclick={() => (uploadDialogOpen = true)} class="gap-2">
			<UploadIcon class="size-4" />
			{m.media_upload_action()}
		</Button>
	{/snippet}

	{#if error && mediaItems.length > 0}
		<InlineNotice
			tone="error"
			message={error}
			dismissLabel={m.common_close()}
			onDismiss={() => (error = '')}
		/>
	{/if}

	<!-- Filter Tabs + Sort -->
	<div class="flex flex-wrap items-center gap-3">
		<Tabs value={filter} onValueChange={changeFilter} class="max-w-full min-w-0">
			<TabsList class="max-w-full justify-start overflow-x-auto">
				{#each filterTabs as tab (tab.value)}
					<TabsTrigger value={tab.value}>{tab.label}</TabsTrigger>
				{/each}
			</TabsList>
		</Tabs>

		<div class="ml-auto flex items-center gap-2">
			<Select.Root type="single" value={sort} onValueChange={changeSort}>
				<Select.Trigger class="h-9 w-[120px] text-sm">
					{sort === 'newest'
						? m.media_sort_newest()
						: sort === 'oldest'
							? m.media_sort_oldest()
							: m.media_sort_size()}
				</Select.Trigger>
				<Select.Content>
					<Select.Item value="newest">{m.media_sort_newest()}</Select.Item>
					<Select.Item value="oldest">{m.media_sort_oldest()}</Select.Item>
					<Select.Item value="size">{m.media_sort_size()}</Select.Item>
				</Select.Content>
			</Select.Root>
		</div>
	</div>

	<!-- Selection Toolbar -->
	{#if isSelectionMode}
		<div class="flex flex-col gap-3 rounded-lg border bg-muted/50 p-3 sm:flex-row sm:items-center">
			<div class="flex flex-wrap items-center gap-2">
				<span class="text-sm font-medium">
					{selectedCountLabel(selectedMediaIds.size)}
				</span>
				{#if deletableCount > 0}
					<Button variant="outline" size="sm" onclick={selectAll}>
						{deletableCount === selectedDeletableIds.length
							? m.media_deselect_all()
							: m.media_select_all_deletable()}
					</Button>
				{/if}
			</div>
			<div class="flex flex-wrap items-center gap-2 sm:ml-auto sm:justify-end">
				<Button variant="outline" size="sm" onclick={toggleFavoriteBatch}>
					<HeartIcon class="mr-1 size-4" />
					{m.media_toggle_favorite()}
				</Button>
				{#if selectedDeletableIds.length > 0}
					<Button variant="destructive" size="sm" onclick={requestDeleteSelectedBatch}>
						<TrashIcon class="mr-1 size-4" />
						{m.media_delete_selected()}
					</Button>
				{/if}
				<Button variant="ghost" size="sm" onclick={cancelSelection}>{m.common_cancel()}</Button>
			</div>
		</div>
	{/if}

	<!-- Media Grid -->
	{#if mediaLoading}
		<PageLoading layout="gallery" label={m.common_loading()} items={10} />
	{:else if error && mediaItems.length === 0}
		<InlineNotice tone="error" message={error} class="my-2">
			{#snippet actions()}
				<Button variant="outline" size="sm" onclick={() => loadMedia()}>
					{m.common_retry()}
				</Button>
			{/snippet}
		</InlineNotice>
	{:else if mediaItems.length === 0}
		{#if filter !== 'all'}
			<EmptyState
				icon={ImageIcon}
				title={m.media_empty_title()}
				description={m.media_empty_filtered_body()}
				actionLabel={m.media_show_all()}
				onAction={() => changeFilter('all')}
				variant="dashed"
				size="lg"
			/>
		{:else}
			<EmptyState
				icon={ImageIcon}
				title={m.media_empty_title()}
				description={m.media_empty_library_body()}
				actionLabel={m.media_upload_action()}
				onAction={() => (uploadDialogOpen = true)}
				variant="dashed"
				size="lg"
			/>
		{/if}
	{:else}
		<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
			{#each mediaItems as media (media.id)}
				<div
					class="group relative overflow-hidden rounded-lg border bg-card transition-all hover:shadow-sm {selectedMediaIds.has(
						media.id
					)
						? 'ring-2 ring-primary'
						: ''}"
				>
					<div class="relative aspect-square overflow-hidden bg-muted/30">
						{#if isVideo(media.mime_type)}
							<video
								src={getAuthenticatedMediaURL(media.url)}
								class="size-full object-cover"
								muted
								playsinline
								preload="metadata"
							></video>
							<div class="pointer-events-none absolute inset-0 flex items-center justify-center">
								<div
									class="flex size-10 items-center justify-center rounded-full bg-background/80 backdrop-blur-sm"
								>
									<VideoIcon class="size-5 text-foreground" />
								</div>
							</div>
						{:else if isImage(media.mime_type)}
							<img
								src={getAuthenticatedMediaURL(media.thumbnail_url || media.url)}
								alt={media.alt_text || media.original_filename || m.media_library_title()}
								loading="lazy"
								class="size-full object-cover transition-transform group-hover:scale-105"
							/>
						{:else}
							<div class="flex size-full items-center justify-center">
								<VideoIcon class="size-10 text-muted-foreground/40" />
							</div>
						{/if}

						<!-- Selection checkbox -->
						<button
							type="button"
							class="media-card-control absolute top-2 left-2 z-10 flex items-center justify-center rounded-md border bg-background/90 shadow-sm backdrop-blur-sm transition-colors hover:bg-background"
							onclick={(e) => {
								e.stopPropagation();
								toggleSelection(media.id);
							}}
							aria-label={selectedMediaIds.has(media.id)
								? m.media_deselect_item({ name: media.original_filename || media.id })
								: m.media_select_item({ name: media.original_filename || media.id })}
							aria-pressed={selectedMediaIds.has(media.id)}
						>
							{#if selectedMediaIds.has(media.id)}
								<CheckIcon class="size-4 text-primary" />
							{:else}
								<div class="size-4 rounded-sm border-2 border-muted-foreground"></div>
							{/if}
						</button>

						<div class="absolute top-2 right-2 z-10">
							<DropdownMenu.Root>
								<DropdownMenu.Trigger>
									{#snippet child({ props })}
										<Button
											{...props}
											variant="outline"
											size="icon-sm"
											class="media-card-control bg-background/90 shadow-sm backdrop-blur-sm"
											aria-label={m.media_actions_for({
												name: media.original_filename || media.id
											})}
										>
											<MoreHorizontalIcon class="size-4" />
										</Button>
									{/snippet}
								</DropdownMenu.Trigger>
								<DropdownMenu.Content align="end" class="w-48">
									<DropdownMenu.Item onclick={() => showUsage(media)} class="gap-2">
										<ExternalLinkIcon class="size-4" />
										{m.media_view_usage()}
									</DropdownMenu.Item>
									<DropdownMenu.Item onclick={() => downloadMedia(media)} class="gap-2">
										<DownloadIcon class="size-4" />
										{m.media_download()}
									</DropdownMenu.Item>
									<DropdownMenu.Item onclick={() => toggleFavorite(media.id)} class="gap-2">
										<HeartIcon class="size-4" fill={media.is_favorite ? 'currentColor' : 'none'} />
										{media.is_favorite ? m.media_unfavorite() : m.media_favorite()}
									</DropdownMenu.Item>
									{#if canDeleteMedia(media)}
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											class="gap-2 text-destructive"
											onclick={() => requestDeleteMedia(media)}
										>
											<TrashIcon class="size-4" />
											{m.common_delete()}
										</DropdownMenu.Item>
									{/if}
								</DropdownMenu.Content>
							</DropdownMenu.Root>
						</div>

						{#if media.is_favorite}
							<div class="absolute bottom-2 left-2 rounded-full bg-background/90 p-1.5 shadow-sm">
								<HeartIcon class="size-3.5 fill-red-500 text-red-500" />
							</div>
						{/if}
					</div>

					<div class="p-2.5">
						{#if media.original_filename}
							<p class="truncate text-sm font-medium" title={media.original_filename}>
								{media.original_filename}
							</p>
						{/if}
						<p class="truncate text-sm text-muted-foreground">
							{formatSize(media.size)} · {formatDate(media.created_at)}
							{#if media.width && media.height}
								· {media.width}×{media.height}
							{/if}
						</p>
						<div class="mt-1.5">
							{#if videoProviderSupportLabel(media.mime_type)}
								<span
									class="mr-1 inline-flex items-center rounded-full bg-amber-500/10 px-2 py-0.5 text-xs font-medium text-amber-700 dark:text-amber-300"
									title={videoProviderSupportDetail(media.mime_type) ?? ''}
								>
									{videoProviderSupportLabel(media.mime_type)}
								</span>
							{/if}
							{#if media.usage_count > 0}
								<span
									class="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary"
								>
									{usedInCountLabel(media.usage_count)}
								</span>
							{:else}
								<span
									class="inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground"
								>
									{m.media_unused_label()}
								</span>
							{/if}
							{#if canDeleteMedia(media) && media.usage_count > 0}
								<span
									class="ml-1 inline-flex items-center rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:text-emerald-300"
								>
									{m.media_deletable_label()}
								</span>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="mt-6 flex flex-wrap items-center justify-center gap-2 sm:gap-4">
				<Button variant="outline" size="sm" onclick={prevPage} disabled={currentPage === 0}>
					<ChevronLeftIcon class="mr-1 h-4 w-4" />
					{m.media_previous_page()}
				</Button>
				<span
					class="order-first w-full text-center text-sm text-muted-foreground sm:order-none sm:w-auto"
				>
					{m.media_page_count({ current: currentPage + 1, total: totalPages })}
				</span>
				<Button
					variant="outline"
					size="sm"
					onclick={nextPage}
					disabled={currentPage >= totalPages - 1}
				>
					{m.media_next_page()}
					<ChevronRightIcon class="ml-1 h-4 w-4" />
				</Button>
			</div>
		{/if}
	{/if}
</PageContainer>

<DestructiveConfirmDialog
	bind:open={deleteDialogOpen}
	title={deletionRequest?.kind === 'batch' ? m.media_delete_batch_title() : m.media_delete_title()}
	description={deletionRequest?.kind === 'batch'
		? deletionRequest.ids.length === 1
			? m.media_delete_batch_body_one()
			: m.media_delete_batch_body_many({ count: deletionRequest.ids.length })
		: m.media_delete_body()}
	onConfirm={confirmMediaDeletion}
/>

<!-- Upload Dialog -->
<Dialog.Root bind:open={uploadDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>{m.media_upload_title()}</Dialog.Title>
			<Dialog.Description>{m.media_upload_description()}</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<span class="text-sm font-medium">{m.media_single_upload()}</span>
				<input id="file-upload" type="file" accept="image/*,video/*" class="peer sr-only" />
				<label
					class="flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-muted-foreground/25 p-6 transition-colors peer-focus-visible:ring-2 peer-focus-visible:ring-ring peer-focus-visible:ring-offset-2 peer-focus-visible:outline-none hover:border-primary/50"
					for="file-upload"
				>
					<UploadIcon class="mb-2 h-8 w-8 text-muted-foreground/40" />
					<p class="text-sm font-medium">{m.media_select_file()}</p>
					<p class="text-sm text-muted-foreground">{m.media_file_limits()}</p>
				</label>
			</div>

			<div class="relative">
				<div class="absolute inset-0 flex items-center">
					<div class="w-full border-t"></div>
				</div>
				<div class="relative flex justify-center text-xs uppercase">
					<span class="bg-background px-2 text-muted-foreground">{m.media_or()}</span>
				</div>
			</div>

			<div class="space-y-2">
				<span class="text-sm font-medium">{m.media_batch_upload()}</span>
				<input
					id="batch-file-upload"
					type="file"
					accept="image/*,video/*"
					multiple
					class="peer sr-only"
				/>
				<label
					class="flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-muted-foreground/25 p-6 transition-colors peer-focus-visible:ring-2 peer-focus-visible:ring-ring peer-focus-visible:ring-offset-2 peer-focus-visible:outline-none hover:border-primary/50"
					for="batch-file-upload"
				>
					<Grid2X2Icon class="mb-2 h-8 w-8 text-muted-foreground/40" />
					<p class="text-sm font-medium">{m.media_select_files()}</p>
					<p class="text-sm text-muted-foreground">{m.media_images_or_videos()}</p>
				</label>
			</div>

			{#if uploadError}
				<InlineNotice
					tone="error"
					message={uploadError}
					dismissLabel={m.common_dismiss()}
					onDismiss={() => (uploadError = '')}
				/>
			{/if}

			<div
				class="rounded-md border border-amber-500/20 bg-amber-500/10 p-3 text-sm text-amber-700 dark:text-amber-300"
			>
				{m.media_video_limits()}
			</div>

			{#if uploadProgress}
				<p class="text-sm text-muted-foreground">{uploadProgress}</p>
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (uploadDialogOpen = false)}
				>{m.common_cancel()}</Button
			>
			<Button onclick={handleUpload} disabled={uploadLoading}>
				{#if uploadLoading}
					<LoaderIcon class="mr-2 size-4 animate-spin" />
				{/if}
				{m.media_upload_action()}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Usage Dialog -->
<Dialog.Root open={usageDialogOpen} onOpenChange={handleUsageDialogOpenChange}>
	<Dialog.Content class="sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>{m.media_usage_title()}</Dialog.Title>
			<Dialog.Description>
				{#if selectedMedia}
					{usageSummaryLabel(selectedMedia.usage_count)}
				{/if}
			</Dialog.Description>
		</Dialog.Header>

		<div class="max-h-[400px] space-y-2 overflow-y-auto py-4">
			{#if usageLoading}
				<div class="py-4">
					<PageLoading layout="list" label={m.common_loading()} items={3} />
				</div>
			{:else if usageError}
				<InlineNotice tone="error" message={usageError}>
					{#snippet actions()}
						<Button
							variant="outline"
							size="sm"
							onclick={() => selectedMedia && showUsage(selectedMedia)}
						>
							{m.common_retry()}
						</Button>
					{/snippet}
				</InlineNotice>
			{:else if mediaUsage.length === 0}
				<p class="py-8 text-center text-sm text-muted-foreground">
					{m.media_usage_empty()}
				</p>
			{:else}
				{#each mediaUsage as usage (usage.post_id)}
					<div class="rounded-lg border p-3">
						<p class="line-clamp-2 text-sm">{usage.content}</p>
						<div class="mt-2 flex items-center gap-3 text-sm text-muted-foreground">
							<span class="rounded-full bg-muted px-2 py-0.5 text-xs"
								>{mediaUsageStatusLabel(usage.status)}</span
							>
							{#if usage.scheduled_at}
								<span
									>{new Date(usage.scheduled_at).toLocaleString(getLocaleTag(), {
										timeZone: workspaceCtx.settings.timezone || 'UTC'
									})}</span
								>
							{/if}
						</div>
					</div>
				{/each}
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => handleUsageDialogOpenChange(false)}
				>{m.common_close()}</Button
			>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<style>
	.media-card-control {
		width: 2rem;
		height: 2rem;
	}

	@media (pointer: coarse) {
		.media-card-control {
			width: 2.75rem;
			height: 2.75rem;
		}
	}
</style>
