<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { SvelteSet } from 'svelte/reactivity';
	import { client, type Workspace } from '$lib/api/client';
	import { getAuthenticatedMediaURL } from '$lib/media-url';
	import { videoProviderSupportDetail, videoProviderSupportLabel } from '$lib/media-capabilities';
	import { isSupportedMediaFile, uploadMediaFiles } from '$lib/media-upload-client';
	import { uploadMediaFile } from '$lib/media-upload-client';
	import {
		deleteStudioTemplate,
		listStudioDesigns,
		listStudioTemplates,
		loadStudioBrandKit,
		loadStudioConfig
	} from '$lib/studio/api';
	import type { StudioBrandKit, StudioDesignSummary, StudioTemplate } from '$lib/studio/types';
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
	import CameraCapture from '$lib/components/camera-capture.svelte';
	import MediaOrganizationDialog from '$lib/components/media-organization-dialog.svelte';
	import BrandKitEditor from '$lib/studio/components/brand-kit-editor.svelte';
	import TemplatePreview from '$lib/studio/components/template-preview.svelte';
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
	import CameraIcon from 'lucide-svelte/icons/camera';
	import PaletteIcon from 'lucide-svelte/icons/palette';
	import SearchIcon from 'lucide-svelte/icons/search';
	import TagIcon from 'lucide-svelte/icons/tag';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import ListIcon from 'lucide-svelte/icons/list';
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
		source: string;
		asset_kind: string;
		parent_media_id?: string;
		design_document_id?: string;
		design_page_id?: string;
		collections: string[];
		tags: string[];
	}

	interface MediaCollection {
		id: string;
		workspace_id: string;
		name: string;
		color: string;
		item_count: number;
	}

	interface MediaTag {
		id: string;
		workspace_id: string;
		name: string;
		item_count: number;
	}

	interface MediaUsage {
		kind: string;
		id: string;
		label: string;
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
	let search = $state('');
	let mediaType = $state('all');
	let source = $state('all');
	let collectionID = $state('');
	let tagID = $state('');
	let aspect = $state('all');
	let minWidth = $state(0);
	let minHeight = $state(0);
	let maxWidth = $state(0);
	let maxHeight = $state(0);
	let dateFrom = $state('');
	let dateTo = $state('');
	let layoutMode = $state<'grid' | 'list'>('grid');
	let hubView = $derived<'assets' | 'designs' | 'templates' | 'brand'>(
		(['assets', 'designs', 'templates', 'brand'].includes($page.url.searchParams.get('view') ?? '')
			? $page.url.searchParams.get('view')
			: 'assets') as 'assets' | 'designs' | 'templates' | 'brand'
	);
	let designs = $state<StudioDesignSummary[]>([]);
	let templates = $state<StudioTemplate[]>([]);
	let brandKit = $state<StudioBrandKit | null>(null);
	let collections = $state<MediaCollection[]>([]);
	let tags = $state<MediaTag[]>([]);
	let hubLoading = $state(false);
	let cameraDialogOpen = $state(false);
	let cameraUploading = $state(false);
	let organizationDialogOpen = $state(false);
	let batchCollectionID = $state('');
	let batchTagID = $state('');
	let organizationSaving = $state(false);
	let templateToDelete = $state<StudioTemplate | null>(null);
	let templateDeleteDialogOpen = $state(false);
	let storageUsage = $state({ used_bytes: 0, asset_count: 0, internal_bytes: 0, limit_bytes: 0 });
	let studioEnabled = $state(true);
	let mediaCanEdit = $state(false);

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
	let detailAltText = $state('');
	let detailSaving = $state(false);

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
		return `${selectedWorkspaceId}:${filter}:${sort}:${search}:${mediaType}:${source}:${collectionID}:${tagID}:${aspect}:${minWidth}:${minHeight}:${maxWidth}:${maxHeight}:${dateFrom}:${dateTo}:${currentPage}`;
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
						search,
						type: mediaType,
						source,
						collection_id: collectionID,
						tag_id: tagID,
						aspect,
						min_width: minWidth,
						min_height: minHeight,
						max_width: maxWidth,
						max_height: maxHeight,
						date_from: dateFrom,
						date_to: dateTo,
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

	async function loadStudioHub(workspaceID = selectedWorkspaceId) {
		if (!workspaceID) return;
		hubLoading = true;
		try {
			const config = await loadStudioConfig();
			studioEnabled = config.enabled;
			const [collectionResult, tagResult, storageResult] = await Promise.all([
				client.GET('/media/collections', {
					params: { query: { workspace_id: workspaceID } }
				}),
				client.GET('/media/tags', { params: { query: { workspace_id: workspaceID } } }),
				client.GET('/media/storage', { params: { query: { workspace_id: workspaceID } } })
			]);
			collections = (collectionResult.data?.collections ?? []) as MediaCollection[];
			tags = (tagResult.data?.tags ?? []) as MediaTag[];
			mediaCanEdit = Boolean(collectionResult.data?.can_edit);
			if (storageResult.data) storageUsage = storageResult.data;
			if (studioEnabled) {
				const [designResult, templateResult, brandResult] = await Promise.all([
					listStudioDesigns(workspaceID),
					listStudioTemplates(workspaceID),
					loadStudioBrandKit(workspaceID)
				]);
				designs = designResult.designs;
				templates = templateResult;
				brandKit = brandResult;
			} else {
				designs = [];
				templates = [];
				brandKit = null;
				if (hubView !== 'assets') changeHubView('assets');
			}
		} catch (cause) {
			notify(cause instanceof Error ? cause.message : m.media_hub_load_failed(), 'error');
		} finally {
			hubLoading = false;
		}
	}

	function changeHubView(view: 'assets' | 'designs' | 'templates' | 'brand') {
		const next = new URL($page.url);
		if (view === 'assets') next.searchParams.delete('view');
		else next.searchParams.set('view', view);
		void goto(resolve(`${next.pathname}${next.search}` as '/'), {
			replaceState: true,
			noScroll: true
		});
	}

	function requestTemplateDeletion(template: StudioTemplate): void {
		if (template.built_in) return;
		templateToDelete = template;
		templateDeleteDialogOpen = true;
	}

	async function confirmTemplateDeletion(): Promise<void> {
		if (!templateToDelete) return;
		try {
			await deleteStudioTemplate(templateToDelete.id);
			templates = templates.filter((template) => template.id !== templateToDelete?.id);
			notify(m.media_template_deleted(), 'success');
		} catch (cause) {
			notify(cause instanceof Error ? cause.message : m.media_template_delete_failed(), 'error');
			throw cause;
		} finally {
			templateToDelete = null;
		}
	}

	function applyAssetFilters() {
		currentPage = 0;
		void loadMedia();
	}

	async function saveCameraPhoto(file: File) {
		cameraUploading = true;
		try {
			await uploadMediaFile({ workspaceId: selectedWorkspaceId, file, source: 'camera' });
			cameraDialogOpen = false;
			await loadMedia();
			notify(m.media_photo_saved(), 'success');
		} catch (cause) {
			notify(cause instanceof Error ? cause.message : m.media_photo_save_failed(), 'error');
		} finally {
			cameraUploading = false;
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

	async function assignSelectedOrganization(
		kind: 'collection' | 'tag',
		mode: 'add' | 'remove' = 'add'
	) {
		const id = kind === 'collection' ? batchCollectionID : batchTagID;
		const mediaIDs = Array.from(selectedMediaIds);
		if (!id || mediaIDs.length === 0) return;
		organizationSaving = true;
		try {
			const result =
				kind === 'collection'
					? await client.PUT('/media/collections/{id}/items', {
							params: { path: { id } },
							body: { media_ids: mediaIDs, mode } as never
						})
					: await client.PUT('/media/tags/{id}/items', {
							params: { path: { id } },
							body: { media_ids: mediaIDs, mode } as never
						});
			if (result.error) throw new Error(result.error.detail);
			await Promise.all([loadMedia(), loadStudioHub()]);
			notify(
				mode === 'remove'
					? m.media_organization_removed({
							count: mediaIDs.length,
							kind:
								kind === 'collection'
									? m.media_organization_collection()
									: m.media_organization_tag()
						})
					: kind === 'collection'
						? m.media_organization_collected({ count: mediaIDs.length })
						: m.media_organization_tagged({ count: mediaIDs.length }),
				'success'
			);
			selectedMediaIds.clear();
			isSelectionMode = false;
			if (kind === 'collection') batchCollectionID = '';
			else batchTagID = '';
		} catch (cause) {
			notify(cause instanceof Error ? cause.message : m.media_assets_organize_failed(), 'error');
		} finally {
			organizationSaving = false;
		}
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

	function openMediaInStudio(media: MediaItem, action = '') {
		const query = new URLSearchParams({
			workspace: selectedWorkspaceId,
			source_media: media.id,
			width: String(media.width || 1080),
			height: String(media.height || 1080)
		});
		if (action) query.set('action', action);
		void goto(resolve(`/studio/new?${query.toString()}` as '/'));
	}

	async function duplicateMedia(media: MediaItem) {
		try {
			const response = await fetch(getAuthenticatedMediaURL(media.url), { credentials: 'include' });
			if (!response.ok) throw new Error(m.media_read_failed());
			const blob = await response.blob();
			const duplicated = new File(
				[blob],
				`copy-${media.original_filename || `${media.id}.${extensionForMime(media.mime_type)}`}`,
				{ type: media.mime_type }
			);
			await uploadMediaFile({
				workspaceId: selectedWorkspaceId,
				file: duplicated,
				source: 'studio_edit',
				parentMediaId: media.id
			});
			await loadMedia();
			notify(m.media_duplicated(), 'success');
		} catch (cause) {
			notify(cause instanceof Error ? cause.message : m.media_duplicate_failed(), 'error');
		}
	}

	async function showUsage(media: MediaItem) {
		const mediaID = media.id;
		const requestSequence = ++usageRequestSequence;
		const isCurrentRequest = () =>
			requestSequence === usageRequestSequence && usageDialogOpen && selectedMedia?.id === mediaID;
		selectedMedia = media;
		detailAltText = media.alt_text;
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

	async function saveDetailAltText(): Promise<void> {
		if (!selectedMedia || detailSaving) return;
		detailSaving = true;
		try {
			const { error: updateError } = await client.PATCH('/media/{id}', {
				params: { path: { id: selectedMedia.id } },
				body: { alt_text: detailAltText.trim() }
			});
			if (updateError) throw new Error(updateError.detail || m.media_alt_update_failed());
			selectedMedia.alt_text = detailAltText.trim();
			const item = mediaItems.find((media) => media.id === selectedMedia?.id);
			if (item) item.alt_text = detailAltText.trim();
			notify(m.media_alt_saved(), 'success');
		} catch (cause) {
			notify(cause instanceof Error ? cause.message : m.media_alt_update_failed(), 'error');
		} finally {
			detailSaving = false;
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

	function mediaSourceLabel(value: string): string {
		switch (value) {
			case 'camera':
				return m.media_camera();
			case 'studio_export':
				return m.media_studio_exports();
			case 'studio_edit':
				return m.media_studio_edits();
			case 'background_removal':
				return m.media_background_removal();
			default:
				return m.media_uploads();
		}
	}

	function mediaUsageKindLabel(value: string): string {
		switch (value) {
			case 'post':
				return m.media_usage_post();
			case 'design':
				return m.media_usage_design();
			case 'design_preview':
				return m.media_usage_design_preview();
			case 'design_page_export':
				return m.media_usage_design_page_export();
			case 'template':
				return m.media_usage_template();
			case 'template_preview':
				return m.media_usage_template_preview();
			case 'brand_asset':
				return m.media_usage_brand_asset();
			case 'brand_font':
				return m.media_usage_brand_font();
			default:
				return value.replaceAll('_', ' ');
		}
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
		if (mediaItems.every((media) => selectedMediaIds.has(media.id))) {
			mediaItems.forEach((media) => selectedMediaIds.delete(media.id));
		} else {
			mediaItems.forEach((media) => selectedMediaIds.add(media.id));
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
		untrack(() => {
			void loadMedia(workspaceID);
			void loadStudioHub(workspaceID);
		});
	});

	const filterTabs = $derived([
		{ value: 'all', label: m.media_filter_all() },
		{ value: 'used', label: m.media_filter_used() },
		{ value: 'unused', label: m.media_filter_unused() },
		{ value: 'favorites', label: m.media_filter_favorites() }
	]);

	const totalPages = $derived(Math.ceil(totalCount / pageSize));
	const allMediaSelected = $derived(
		mediaItems.length > 0 && mediaItems.every((media) => selectedMediaIds.has(media.id))
	);
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
	title={m.media_hub_title()}
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
		{#if mediaCanEdit}
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button {...props} class="gap-2">
							<PlusIcon class="size-4" />
							{m.media_create()}
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="end" class="w-48">
					<DropdownMenu.Item onclick={() => (uploadDialogOpen = true)}>
						<UploadIcon />
						{m.media_upload_title()}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => (cameraDialogOpen = true)}>
						<CameraIcon />
						{m.media_take_photo()}
					</DropdownMenu.Item>
					{#if studioEnabled}
						<DropdownMenu.Item
							onclick={() =>
								goto(
									resolve(`/studio/new?workspace=${encodeURIComponent(selectedWorkspaceId)}` as '/')
								)}
						>
							<PaletteIcon />
							{m.media_create_design()}
						</DropdownMenu.Item>
					{/if}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{/if}
	{/snippet}

	<Tabs value={hubView} onValueChange={(value) => changeHubView(value as typeof hubView)}>
		<TabsList class="max-w-full justify-start overflow-x-auto">
			<TabsTrigger value="assets"><ImageIcon /> {m.media_assets()}</TabsTrigger>
			{#if studioEnabled}
				<TabsTrigger value="designs"><PaletteIcon /> {m.media_designs()}</TabsTrigger>
				<TabsTrigger value="templates"><Grid2X2Icon /> {m.media_templates()}</TabsTrigger>
				<TabsTrigger value="brand"><HeartIcon /> {m.media_brand()}</TabsTrigger>
			{/if}
		</TabsList>
	</Tabs>

	{#if hubView === 'assets'}
		<div
			class="flex flex-wrap items-center justify-between gap-2 rounded-xl border bg-card px-3 py-2"
		>
			<div>
				<p class="text-sm font-medium">
					{m.media_stored({ size: formatSize(storageUsage.used_bytes) })}
				</p>
				<p class="text-xs text-muted-foreground">
					{m.media_library_assets({
						count: storageUsage.asset_count,
						suffix: storageUsage.asset_count === 1 ? '' : 's'
					})}
					{#if storageUsage.internal_bytes > 0}
						·
						{m.media_internal_previews({ size: formatSize(storageUsage.internal_bytes) })}
					{/if}
				</p>
			</div>
			{#if storageUsage.limit_bytes > 0}
				<div class="w-full max-w-56">
					<div class="h-2 overflow-hidden rounded-full bg-muted">
						<div
							class="h-full rounded-full bg-primary"
							style:width={`${Math.min(100, (storageUsage.used_bytes / storageUsage.limit_bytes) * 100)}%`}
						></div>
					</div>
				</div>
			{/if}
		</div>
		{#if error && mediaItems.length > 0}
			<InlineNotice
				tone="error"
				message={error}
				dismissLabel={m.common_close()}
				onDismiss={() => (error = '')}
			/>
		{/if}

		<form
			class="grid gap-2 rounded-xl border bg-card p-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-[minmax(14rem,1fr)_9rem_10rem_12rem_12rem_auto]"
			onsubmit={(event) => {
				event.preventDefault();
				applyAssetFilters();
			}}
		>
			<div class="relative">
				<SearchIcon
					class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
				/>
				<input
					class="h-11 w-full rounded-md border border-input bg-background pr-3 pl-9 text-sm sm:h-10"
					bind:value={search}
					placeholder={m.media_search_filename_alt()}
				/>
			</div>
			<select
				class="h-11 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
				bind:value={mediaType}
			>
				<option value="all">{m.media_all_types()}</option>
				<option value="image">{m.media_images()}</option>
				<option value="video">{m.media_videos()}</option>
			</select>
			<select
				class="h-11 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
				bind:value={source}
			>
				<option value="all">{m.media_all_sources()}</option>
				<option value="upload">{m.media_uploads()}</option>
				<option value="camera">{m.media_camera()}</option>
				<option value="studio_export">{m.media_studio_exports()}</option>
				<option value="studio_edit">{m.media_studio_edits()}</option>
				<option value="background_removal">{m.media_background_removal()}</option>
			</select>
			<select
				class="h-11 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
				bind:value={collectionID}
			>
				<option value="">{m.media_all_collections()}</option>
				{#each collections as collection (collection.id)}
					<option value={collection.id}>{collection.name} ({collection.item_count})</option>
				{/each}
			</select>
			<select
				class="h-11 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
				bind:value={tagID}
			>
				<option value="">{m.media_all_tags()}</option>
				{#each tags as tag (tag.id)}
					<option value={tag.id}>{tag.name} ({tag.item_count})</option>
				{/each}
			</select>
			<Button type="submit" variant="outline" class="h-11 sm:h-10">{m.media_apply()}</Button>
		</form>
		<details class="rounded-xl border bg-card">
			<summary class="min-h-11 cursor-pointer px-3 py-3 text-sm font-medium">
				{m.media_dimensions_date()}
			</summary>
			<div class="grid gap-3 border-t p-3 sm:grid-cols-2 lg:grid-cols-4">
				<label class="grid gap-1 text-xs font-medium">
					<span>{m.media_aspect_ratio()}</span>
					<select
						class="h-11 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
						bind:value={aspect}
					>
						<option value="all">{m.media_any_aspect()}</option>
						<option value="square">{m.media_square()}</option>
						<option value="portrait">{m.media_portrait()}</option>
						<option value="landscape">{m.media_landscape()}</option>
					</select>
				</label>
				<div class="grid grid-cols-2 gap-2">
					<label class="grid gap-1 text-xs font-medium">
						<span>{m.media_min_width()}</span>
						<input
							class="h-11 min-w-0 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
							type="number"
							min="0"
							bind:value={minWidth}
						/>
					</label>
					<label class="grid gap-1 text-xs font-medium">
						<span>{m.media_min_height()}</span>
						<input
							class="h-11 min-w-0 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
							type="number"
							min="0"
							bind:value={minHeight}
						/>
					</label>
				</div>
				<div class="grid grid-cols-2 gap-2">
					<label class="grid gap-1 text-xs font-medium">
						<span>{m.media_max_width()}</span>
						<input
							class="h-11 min-w-0 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
							type="number"
							min="0"
							bind:value={maxWidth}
						/>
					</label>
					<label class="grid gap-1 text-xs font-medium">
						<span>{m.media_max_height()}</span>
						<input
							class="h-11 min-w-0 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
							type="number"
							min="0"
							bind:value={maxHeight}
						/>
					</label>
				</div>
				<div class="grid grid-cols-2 gap-2">
					<label class="grid gap-1 text-xs font-medium">
						<span>{m.media_from()}</span>
						<input
							class="h-11 min-w-0 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
							type="date"
							bind:value={dateFrom}
						/>
					</label>
					<label class="grid gap-1 text-xs font-medium">
						<span>{m.media_to()}</span>
						<input
							class="h-11 min-w-0 rounded-md border border-input bg-background px-2 text-sm sm:h-10"
							type="date"
							bind:value={dateTo}
						/>
					</label>
				</div>
				<div class="flex items-end gap-2 lg:col-span-4 lg:justify-end">
					<Button
						variant="ghost"
						onclick={() => {
							aspect = 'all';
							minWidth = 0;
							minHeight = 0;
							maxWidth = 0;
							maxHeight = 0;
							dateFrom = '';
							dateTo = '';
							applyAssetFilters();
						}}
						class="h-11 sm:h-9">{m.media_clear()}</Button
					>
					<Button onclick={applyAssetFilters} class="h-11 sm:h-9">{m.media_apply_filters()}</Button>
				</div>
			</div>
		</details>
		{#if mediaCanEdit}
			<div class="flex justify-end">
				<Button variant="ghost" size="sm" onclick={() => (organizationDialogOpen = true)}>
					<TagIcon />
					{m.media_manage_organization()}
				</Button>
			</div>
		{/if}

		<!-- Filter Tabs + Sort -->
		<div class="flex flex-wrap items-center gap-3">
			<Tabs value={filter} onValueChange={changeFilter} class="max-w-full min-w-0">
				<TabsList class="max-w-full justify-start overflow-x-auto">
					{#each filterTabs as tab (tab.value)}
						<TabsTrigger value={tab.value} class="px-3">{tab.label}</TabsTrigger>
					{/each}
				</TabsList>
			</Tabs>

			<div class="ml-auto flex items-center gap-2">
				<Select.Root type="single" value={sort} onValueChange={changeSort}>
					<Select.Trigger class="h-9 w-[148px] text-sm">
						{sort === 'newest'
							? m.media_sort_newest()
							: sort === 'oldest'
								? m.media_sort_oldest()
								: sort === 'name'
									? m.media_sort_name()
									: sort === 'recently_used'
										? m.media_recently_used()
										: m.media_sort_size()}
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="newest">{m.media_sort_newest()}</Select.Item>
						<Select.Item value="oldest">{m.media_sort_oldest()}</Select.Item>
						<Select.Item value="name">{m.media_sort_name()}</Select.Item>
						<Select.Item value="size">{m.media_sort_size()}</Select.Item>
						<Select.Item value="recently_used">{m.media_recently_used()}</Select.Item>
					</Select.Content>
				</Select.Root>
				<div class="flex rounded-md border p-0.5">
					<Button
						variant={layoutMode === 'grid' ? 'secondary' : 'ghost'}
						size="icon-xs"
						onclick={() => (layoutMode = 'grid')}
						aria-label={m.media_grid_view()}><Grid2X2Icon /></Button
					>
					<Button
						variant={layoutMode === 'list' ? 'secondary' : 'ghost'}
						size="icon-xs"
						onclick={() => (layoutMode = 'list')}
						aria-label={m.media_compact_view()}><ListIcon /></Button
					>
				</div>
			</div>
		</div>

		<!-- Selection Toolbar -->
		{#if isSelectionMode}
			<div
				class="flex flex-col gap-3 rounded-lg border bg-muted/50 p-3 sm:flex-row sm:items-center"
			>
				<div class="flex flex-wrap items-center gap-2">
					<span class="text-sm font-medium">
						{selectedCountLabel(selectedMediaIds.size)}
					</span>
					{#if mediaItems.length > 0}
						<Button variant="outline" size="sm" onclick={selectAll}>
							{allMediaSelected ? m.media_deselect_all() : m.media_select_all()}
						</Button>
					{/if}
				</div>
				<div class="flex flex-wrap items-center gap-2 sm:ml-auto sm:justify-end">
					<div class="flex items-center gap-1">
						<select
							class="h-9 max-w-40 rounded-md border border-input bg-background px-2 text-sm"
							bind:value={batchCollectionID}
							aria-label={m.media_collection()}
						>
							<option value="">{m.media_collection()}…</option>
							{#each collections as collection (collection.id)}
								<option value={collection.id}>{collection.name}</option>
							{/each}
						</select>
						<Button
							variant="outline"
							size="sm"
							disabled={!batchCollectionID || organizationSaving}
							onclick={() => assignSelectedOrganization('collection')}
						>
							{m.media_add()}
						</Button>
						<Button
							variant="ghost"
							size="sm"
							disabled={!batchCollectionID || organizationSaving}
							onclick={() => assignSelectedOrganization('collection', 'remove')}
						>
							{m.media_remove()}
						</Button>
					</div>
					<div class="flex items-center gap-1">
						<select
							class="h-9 max-w-36 rounded-md border border-input bg-background px-2 text-sm"
							bind:value={batchTagID}
							aria-label={m.media_tag()}
						>
							<option value="">{m.media_tag()}…</option>
							{#each tags as tag (tag.id)}
								<option value={tag.id}>{tag.name}</option>
							{/each}
						</select>
						<Button
							variant="outline"
							size="sm"
							disabled={!batchTagID || organizationSaving}
							onclick={() => assignSelectedOrganization('tag')}
						>
							{m.media_add()}
						</Button>
						<Button
							variant="ghost"
							size="sm"
							disabled={!batchTagID || organizationSaving}
							onclick={() => assignSelectedOrganization('tag', 'remove')}
						>
							{m.media_remove()}
						</Button>
					</div>
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
					actionLabel={mediaCanEdit ? m.media_upload_action() : undefined}
					onAction={() => (uploadDialogOpen = true)}
					variant="dashed"
					size="lg"
				/>
			{/if}
		{:else}
			<div
				class={layoutMode === 'grid'
					? 'grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5'
					: 'grid grid-cols-1 gap-2'}
			>
				{#each mediaItems as media (media.id)}
					<div
						class="group relative overflow-hidden rounded-lg border bg-card transition-all hover:shadow-sm {layoutMode ===
						'list'
							? 'grid grid-cols-[6rem_minmax(0,1fr)]'
							: ''} {selectedMediaIds.has(media.id) ? 'ring-2 ring-primary' : ''}"
					>
						<div
							class="relative overflow-hidden bg-muted/30 {layoutMode === 'grid'
								? 'aspect-square'
								: 'aspect-square h-24'}"
						>
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
							{#if mediaCanEdit}
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
							{/if}

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
											{m.media_details()}
										</DropdownMenu.Item>
										{#if isImage(media.mime_type) && mediaCanEdit && studioEnabled}
											<DropdownMenu.Item onclick={() => openMediaInStudio(media)} class="gap-2">
												<PaletteIcon class="size-4" />
												{m.media_edit_studio()}
											</DropdownMenu.Item>
											<DropdownMenu.Item
												onclick={() => openMediaInStudio(media, 'remove-background')}
												class="gap-2"
											>
												<ImageIcon class="size-4" />
												{m.studio_remove_background()}
											</DropdownMenu.Item>
										{/if}
										{#if mediaCanEdit}
											<DropdownMenu.Item onclick={() => duplicateMedia(media)} class="gap-2">
												<Grid2X2Icon class="size-4" />
												{m.studio_duplicate()}
											</DropdownMenu.Item>
										{/if}
										<DropdownMenu.Item onclick={() => downloadMedia(media)} class="gap-2">
											<DownloadIcon class="size-4" />
											{m.media_download()}
										</DropdownMenu.Item>
										{#if mediaCanEdit}
											<DropdownMenu.Item onclick={() => toggleFavorite(media.id)} class="gap-2">
												<HeartIcon
													class="size-4"
													fill={media.is_favorite ? 'currentColor' : 'none'}
												/>
												{media.is_favorite ? m.media_unfavorite() : m.media_favorite()}
											</DropdownMenu.Item>
										{/if}
										{#if mediaCanEdit && canDeleteMedia(media)}
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
								{#if media.source && media.source !== 'upload'}
									<span
										class="ml-1 inline-flex items-center rounded-full bg-orange-500/10 px-2 py-0.5 text-xs font-medium text-orange-700 dark:text-orange-300"
									>
										{mediaSourceLabel(media.source)}
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
	{:else if hubView === 'designs'}
		{#if hubLoading}
			<PageLoading layout="gallery" label={m.media_loading_designs()} items={8} />
		{:else if designs.length === 0}
			<EmptyState
				icon={PaletteIcon}
				title={m.media_no_designs()}
				description={m.media_design_empty_body()}
				actionLabel={mediaCanEdit ? m.media_create_design() : undefined}
				onAction={() =>
					goto(resolve(`/studio/new?workspace=${encodeURIComponent(selectedWorkspaceId)}` as '/'))}
				variant="dashed"
				size="lg"
			/>
		{:else}
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each designs as design (design.id)}
					<a
						href={resolve(`/studio/${design.id}` as '/')}
						class="group overflow-hidden rounded-xl border bg-card transition hover:-translate-y-0.5 hover:shadow-sm"
					>
						<div class="flex aspect-[4/3] items-center justify-center bg-neutral-800 p-5">
							{#if design.cover_preview_media_id}
								<img
									src={getAuthenticatedMediaURL(`/media/${design.cover_preview_media_id}`)}
									alt=""
									class="max-h-full max-w-full shadow"
								/>
							{:else}
								<div
									class="max-h-full max-w-full bg-orange-50 shadow"
									style:aspect-ratio={`${design.width_px}/${design.height_px}`}
									style:height={design.height_px > design.width_px ? '100%' : 'auto'}
									style:width={design.width_px >= design.height_px ? '100%' : 'auto'}
								></div>
							{/if}
						</div>
						<div class="p-3">
							<p class="truncate text-sm font-medium">{design.title}</p>
							<p class="mt-1 text-xs text-muted-foreground">
								{m.media_design_pages({
									count: design.page_count,
									suffix: design.page_count === 1 ? '' : 's'
								})}
								· {design.width_px}×{design.height_px}
							</p>
							<p class="mt-1 text-xs text-muted-foreground">{formatDate(design.updated_at)}</p>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	{:else if hubView === 'templates'}
		{#if hubLoading}
			<PageLoading layout="gallery" label={m.media_loading_templates()} items={8} />
		{:else}
			<div class="mb-4">
				<h2 class="text-base font-semibold">{m.media_templates_heading()}</h2>
				<p class="text-sm text-muted-foreground">{m.media_templates_body()}</p>
			</div>
			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
				{#each templates as template (template.id)}
					<div class="relative rounded-xl border bg-card hover:border-primary/40 hover:shadow-sm">
						<a
							href={resolve(
								`/studio/new?workspace=${encodeURIComponent(selectedWorkspaceId)}&template=${encodeURIComponent(template.id)}` as '/'
							)}
							class="block p-3"
						>
							<div class="mb-3 aspect-[4/3] overflow-hidden rounded-lg border">
								<TemplatePreview document={template.document} label={template.name} />
							</div>
							<p class="truncate pr-8 text-sm font-medium">{template.name}</p>
							<p class="mt-1 truncate text-xs text-muted-foreground">
								{template.built_in ? m.media_openpost_starter() : m.media_workspace_template()} ·
								{template.category}
							</p>
						</a>
						{#if !template.built_in && brandKit?.can_edit}
							<Button
								variant="ghost"
								size="icon-sm"
								class="absolute right-2 bottom-2"
								onclick={() => requestTemplateDeletion(template)}
								aria-label={m.media_delete_named({ name: template.name })}
							>
								<TrashIcon />
							</Button>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	{:else if hubView === 'brand'}
		{#if hubLoading}
			<PageLoading layout="sections" label={m.media_loading_brand()} />
		{:else if brandKit}
			{#if brandKit.can_edit}
				{#key brandKit.revision}
					<BrandKitEditor kit={brandKit} onSaved={(saved) => (brandKit = saved)} />
				{/key}
			{:else}
				<div class="grid gap-5 lg:grid-cols-2">
					<section class="rounded-xl border bg-card p-4">
						<div class="mb-4 flex items-center gap-2">
							<PaletteIcon class="size-4 text-primary" />
							<div>
								<h2 class="font-semibold">{brandKit.name || m.brand_default_name()}</h2>
								<p class="text-xs text-muted-foreground">
									{m.media_brand_revision({ revision: brandKit.revision })}
								</p>
							</div>
						</div>
						<h3 class="mb-2 text-sm font-medium">{m.studio_colors()}</h3>
						{#if brandKit.colors.length}
							<div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
								{#each brandKit.colors as color (color.name)}
									<div class="overflow-hidden rounded-lg border">
										<div class="h-16" style:background={color.value}></div>
										<div class="px-2 py-1.5 text-xs">
											<p class="font-medium">{color.name}</p>
											<p class="text-muted-foreground">{color.value}</p>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">{m.media_no_workspace_colors()}</p>
						{/if}
						<h3 class="mt-5 mb-2 text-sm font-medium">{m.brand_page_backgrounds()}</h3>
						<div class="flex flex-wrap gap-2">
							{#each brandKit.backgrounds as background (background)}
								<span class="size-10 rounded-md border" style:background></span>
							{/each}
						</div>
					</section>
					<section class="rounded-xl border bg-card p-4">
						<h2 class="font-semibold">{m.media_logos_fonts()}</h2>
						<p class="mt-1 text-sm text-muted-foreground">{m.media_brand_assets_body()}</p>
						<div class="mt-4 grid grid-cols-3 gap-3">
							{#each brandKit.assets as asset (asset.id)}
								<div class="rounded-lg border p-2">
									<img
										src={getAuthenticatedMediaURL(`/media/${asset.media_id}/thumb/md`)}
										alt={asset.name || asset.role}
										class="aspect-square w-full object-contain"
									/>
									<p class="mt-1 truncate text-xs">{asset.name || asset.role}</p>
								</div>
							{/each}
						</div>
						<div class="mt-5 space-y-2">
							{#each brandKit.fonts as font (font.id)}
								<div class="flex items-center justify-between rounded-lg border px-3 py-2">
									<div>
										<p class="text-sm font-medium">{font.family}</p>
										<p class="text-xs text-muted-foreground">{font.weight} · {font.style}</p>
									</div>
									<TagIcon class="size-4 text-muted-foreground" />
								</div>
							{/each}
						</div>
						{#if brandKit.assets.length === 0 && brandKit.fonts.length === 0}
							<div
								class="mt-5 rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground"
							>
								{m.studio_brand_empty()}
							</div>
						{/if}
					</section>
				</div>
			{/if}
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

<DestructiveConfirmDialog
	bind:open={templateDeleteDialogOpen}
	title={m.media_delete_template_title()}
	description={m.media_delete_template_body({
		name: templateToDelete?.name ?? m.media_workspace_template()
	})}
	confirmLabel={m.media_delete_template()}
	onConfirm={confirmTemplateDeletion}
/>

<MediaOrganizationDialog
	bind:open={organizationDialogOpen}
	workspaceId={selectedWorkspaceId}
	{collections}
	{tags}
	onChanged={() => loadStudioHub()}
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

<Dialog.Root bind:open={cameraDialogOpen}>
	<Dialog.Content class="sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>{m.media_take_photo()}</Dialog.Title>
			<Dialog.Description>{m.media_camera_body()}</Dialog.Description>
		</Dialog.Header>
		<CameraCapture onCapture={saveCameraPhoto} onCancel={() => (cameraDialogOpen = false)} />
		{#if cameraUploading}
			<p class="text-sm text-muted-foreground" aria-live="polite">{m.media_saving_photo()}</p>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<!-- Usage Dialog -->
<Dialog.Root open={usageDialogOpen} onOpenChange={handleUsageDialogOpenChange}>
	<Dialog.Content class="max-h-[min(820px,calc(100dvh-2rem))] overflow-y-auto sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>{selectedMedia?.original_filename || m.media_details()}</Dialog.Title>
			<Dialog.Description>
				{#if selectedMedia}
					{usageSummaryLabel(selectedMedia.usage_count)}
				{/if}
			</Dialog.Description>
		</Dialog.Header>

		{#if selectedMedia}
			<div class="grid gap-4 py-4 sm:grid-cols-[14rem_1fr]">
				<div class="overflow-hidden rounded-xl border bg-muted/20">
					{#if isImage(selectedMedia.mime_type)}
						<img
							src={getAuthenticatedMediaURL(selectedMedia.url)}
							alt={selectedMedia.alt_text || selectedMedia.original_filename}
							class="aspect-square size-full object-contain"
						/>
					{:else if isVideo(selectedMedia.mime_type)}
						<video
							src={getAuthenticatedMediaURL(selectedMedia.url)}
							class="aspect-square size-full object-contain"
							controls
							muted
							playsinline
						></video>
					{/if}
				</div>
				<dl class="grid grid-cols-[7rem_1fr] content-start gap-x-3 gap-y-2 text-sm">
					<dt class="text-muted-foreground">{m.media_type()}</dt>
					<dd>{selectedMedia.mime_type}</dd>
					<dt class="text-muted-foreground">{m.media_sort_size()}</dt>
					<dd>{formatSize(selectedMedia.size)}</dd>
					<dt class="text-muted-foreground">{m.media_dimensions()}</dt>
					<dd>{selectedMedia.width || '—'} × {selectedMedia.height || '—'}</dd>
					<dt class="text-muted-foreground">{m.media_source()}</dt>
					<dd>{mediaSourceLabel(selectedMedia.source)}</dd>
					<dt class="text-muted-foreground">{m.media_created()}</dt>
					<dd>{formatDate(selectedMedia.created_at)}</dd>
					<dt class="text-muted-foreground">{m.media_alt_text()}</dt>
					<dd class="space-y-2">
						<textarea
							class="min-h-20 w-full rounded-md border border-input bg-background p-2 text-sm"
							bind:value={detailAltText}
							placeholder={m.media_alt_placeholder()}
							disabled={!mediaCanEdit || detailSaving}
						></textarea>
						{#if mediaCanEdit && detailAltText.trim() !== selectedMedia.alt_text}
							<Button
								size="sm"
								variant="outline"
								onclick={saveDetailAltText}
								disabled={detailSaving}
							>
								{#if detailSaving}<LoaderIcon class="animate-spin" />{/if}
								{m.media_save_alt()}
							</Button>
						{/if}
					</dd>
					{#if selectedMedia.parent_media_id}
						<dt class="text-muted-foreground">{m.media_original()}</dt>
						<dd class="font-mono text-xs break-all">{selectedMedia.parent_media_id}</dd>
					{/if}
					{#if selectedMedia.design_document_id}
						<dt class="text-muted-foreground">{m.media_design()}</dt>
						<dd>
							<a
								href={resolve(`/studio/${selectedMedia.design_document_id}` as '/')}
								class="font-medium text-primary hover:underline"
							>
								{m.media_open_design()}
							</a>
						</dd>
					{/if}
					{#if selectedMedia.collections.length}
						<dt class="text-muted-foreground">{m.media_collections()}</dt>
						<dd>
							{selectedMedia.collections
								.map((id) => collections.find((item) => item.id === id)?.name || id)
								.join(', ')}
						</dd>
					{/if}
					{#if selectedMedia.tags.length}
						<dt class="text-muted-foreground">{m.media_tags()}</dt>
						<dd>
							{selectedMedia.tags
								.map((id) => tags.find((item) => item.id === id)?.name || id)
								.join(', ')}
						</dd>
					{/if}
				</dl>
			</div>
			<div class="flex flex-wrap gap-2 border-y py-3">
				{#if isImage(selectedMedia.mime_type) && mediaCanEdit && studioEnabled}
					<Button variant="outline" size="sm" onclick={() => openMediaInStudio(selectedMedia!)}>
						<PaletteIcon />
						{m.media_edit_studio()}
					</Button>
					<Button
						variant="outline"
						size="sm"
						onclick={() => openMediaInStudio(selectedMedia!, 'remove-background')}
					>
						<ImageIcon />
						{m.studio_remove_background()}
					</Button>
				{/if}
				{#if mediaCanEdit}
					<Button variant="outline" size="sm" onclick={() => duplicateMedia(selectedMedia!)}>
						<Grid2X2Icon />
						{m.studio_duplicate()}
					</Button>
				{/if}
				<Button variant="outline" size="sm" onclick={() => downloadMedia(selectedMedia!)}>
					<DownloadIcon />
					{m.studio_download()}
				</Button>
				{#if mediaCanEdit}
					<Button
						variant="destructive"
						size="sm"
						disabled={!canDeleteMedia(selectedMedia)}
						title={!canDeleteMedia(selectedMedia) ? m.media_delete_blocked() : undefined}
						onclick={() => requestDeleteMedia(selectedMedia!)}
					>
						<TrashIcon />
						{m.common_delete()}
					</Button>
				{/if}
				{#if mediaCanEdit && !canDeleteMedia(selectedMedia)}
					<p class="basis-full text-xs text-muted-foreground">
						{m.media_delete_blocked()}
					</p>
				{/if}
			</div>
		{/if}

		<div class="space-y-2 py-4">
			<h3 class="text-sm font-semibold">{m.media_used_by()}</h3>
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
				{#each mediaUsage as usage (`${usage.kind}-${usage.id}`)}
					<div class="rounded-lg border p-3">
						<p class="line-clamp-2 text-sm font-medium">{usage.label || usage.content}</p>
						<p class="mt-1 text-xs text-muted-foreground">{mediaUsageKindLabel(usage.kind)}</p>
						<div class="mt-2 flex items-center gap-3 text-sm text-muted-foreground">
							{#if usage.status}
								<span class="rounded-full bg-muted px-2 py-0.5 text-xs">
									{mediaUsageStatusLabel(usage.status)}
								</span>
							{/if}
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
