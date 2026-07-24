<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import CameraCapture from './camera-capture.svelte';
	import { getAuthenticatedMediaURL } from '$lib/media-url';
	import { uploadMediaFile } from '$lib/media-upload-client';
	import { listStudioMedia } from '$lib/studio/api';
	import type { StudioMediaItem } from '$lib/studio/types';
	import SearchIcon from 'lucide-svelte/icons/search';
	import UploadIcon from 'lucide-svelte/icons/upload';
	import CameraIcon from 'lucide-svelte/icons/camera';
	import ImageIcon from 'lucide-svelte/icons/image';
	import PaletteIcon from 'lucide-svelte/icons/palette';
	import CheckIcon from 'lucide-svelte/icons/check';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import { m } from '$lib/paraglide/messages';

	let {
		open = $bindable(false),
		workspaceId,
		currentSelection = [],
		accept = ['image/*', 'video/*'],
		maxSelection = 4,
		multiple = true,
		title = m.media_picker_add_media(),
		purpose = 'post_media',
		showCreate = true,
		onConfirm,
		onCreate
	}: {
		open?: boolean;
		workspaceId: string;
		currentSelection?: string[];
		accept?: string[];
		maxSelection?: number;
		multiple?: boolean;
		title?: string;
		purpose?: string;
		showCreate?: boolean;
		onConfirm: (mediaIDs: string[], media: StudioMediaItem[]) => void | Promise<void>;
		onCreate?: () => void | Promise<void>;
	} = $props();

	let tab = $state('library');
	let media = $state<StudioMediaItem[]>([]);
	let selectedIDs = $state.raw<string[]>([]);
	let search = $state('');
	let loading = $state(false);
	let actionLoading = $state(false);
	let error = $state('');
	let uploadInput = $state<HTMLInputElement>();
	let loadedForWorkspace = $state('');

	function attachUploadInput(node: HTMLInputElement) {
		uploadInput = node;
		return () => {
			if (uploadInput === node) uploadInput = undefined;
		};
	}

	function initializePicker() {
		selectedIDs = [...currentSelection];
		if (loadedForWorkspace !== workspaceId) void loadMedia();
	}

	async function loadMedia(): Promise<void> {
		if (!workspaceId) return;
		loading = true;
		error = '';
		try {
			media = (await listStudioMedia(workspaceId, search)).filter((item) =>
				mimeAllowed(item.mime_type)
			);
			loadedForWorkspace = workspaceId;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.media_picker_load_failed();
		} finally {
			loading = false;
		}
	}

	function mimeAllowed(mime: string): boolean {
		return accept.some((accepted) =>
			accepted.endsWith('/*') ? mime.startsWith(accepted.slice(0, -1)) : mime === accepted
		);
	}

	function toggleMedia(id: string): void {
		if (selectedIDs.includes(id)) {
			selectedIDs = selectedIDs.filter((item) => item !== id);
			return;
		}
		if (!multiple) {
			selectedIDs = [id];
			return;
		}
		if (selectedIDs.length >= maxSelection) {
			error = m.media_picker_selection_limit({ maximum: maxSelection });
			return;
		}
		selectedIDs = [...selectedIDs, id];
	}

	async function confirm(): Promise<void> {
		actionLoading = true;
		error = '';
		try {
			await onConfirm(
				selectedIDs,
				selectedIDs
					.map((id) => media.find((item) => item.id === id))
					.filter((item): item is StudioMediaItem => Boolean(item))
			);
			open = false;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.media_picker_add_failed();
		} finally {
			actionLoading = false;
		}
	}

	async function uploadFiles(files: FileList | null): Promise<void> {
		if (!files?.length) return;
		actionLoading = true;
		error = '';
		try {
			const available = Math.max(0, maxSelection - selectedIDs.length);
			const candidates = Array.from(files)
				.filter((file) => mimeAllowed(file.type))
				.slice(0, available);
			for (const file of candidates) {
				const uploaded = await uploadMediaFile({ workspaceId, file, source: 'upload' });
				selectedIDs = [...selectedIDs, uploaded.id];
			}
			await loadMedia();
			tab = 'library';
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.media_picker_upload_failed();
		} finally {
			actionLoading = false;
			if (uploadInput) uploadInput.value = '';
		}
	}

	async function capturePhoto(file: File): Promise<void> {
		actionLoading = true;
		error = '';
		try {
			const uploaded = await uploadMediaFile({ workspaceId, file, source: 'camera' });
			selectedIDs = multiple ? [...selectedIDs, uploaded.id].slice(0, maxSelection) : [uploaded.id];
			await loadMedia();
			tab = 'library';
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.media_picker_photo_failed();
		} finally {
			actionLoading = false;
		}
	}

	async function createDesign(): Promise<void> {
		actionLoading = true;
		error = '';
		try {
			if (onCreate) {
				await onCreate();
				open = false;
				return;
			}
			await goto(
				resolve(
					`/studio/new?workspace=${encodeURIComponent(workspaceId)}&purpose=${encodeURIComponent(purpose)}` as '/'
				)
			);
			open = false;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.media_picker_studio_failed();
		} finally {
			actionLoading = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="flex h-[min(760px,calc(100dvh-2rem))] max-w-5xl flex-col gap-0 p-0">
		<div class="contents" {@attach initializePicker}></div>
		<Dialog.Header class="border-b px-4 py-3 pr-14">
			<Dialog.Title>{title}</Dialog.Title>
			<Dialog.Description>{m.media_picker_description()}</Dialog.Description>
		</Dialog.Header>

		<Tabs.Root bind:value={tab} class="min-h-0 flex-1 gap-0">
			<Tabs.List class={`mx-4 mt-3 grid ${showCreate ? 'grid-cols-4' : 'grid-cols-3'}`}>
				<Tabs.Trigger value="library"><ImageIcon /> {m.media_picker_library()}</Tabs.Trigger>
				<Tabs.Trigger value="upload"><UploadIcon /> {m.media_picker_upload()}</Tabs.Trigger>
				<Tabs.Trigger value="camera"><CameraIcon /> {m.studio_camera()}</Tabs.Trigger>
				{#if showCreate}<Tabs.Trigger value="create"
						><PaletteIcon /> {m.media_picker_create()}</Tabs.Trigger
					>{/if}
			</Tabs.List>

			<Tabs.Content value="library" class="min-h-0 flex-1 overflow-y-auto px-4 py-3">
				<form
					class="mb-3 flex gap-2"
					onsubmit={(event) => {
						event.preventDefault();
						void loadMedia();
					}}
				>
					<div class="relative flex-1">
						<SearchIcon
							class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
						/>
						<Input bind:value={search} class="pl-9" placeholder={m.media_picker_search()} />
					</div>
					<Button variant="outline" type="submit">{m.media_picker_search_action()}</Button>
				</form>
				{#if loading}
					<div class="flex min-h-48 items-center justify-center text-muted-foreground">
						<LoaderIcon class="mr-2 size-5 animate-spin" />
						{m.media_picker_loading()}
					</div>
				{:else if media.length === 0}
					<div
						class="flex min-h-48 flex-col items-center justify-center rounded-xl border border-dashed text-center"
					>
						<ImageIcon class="mb-3 size-8 text-muted-foreground" />
						<p class="font-medium">{m.media_picker_no_match()}</p>
						<p class="mt-1 text-sm text-muted-foreground">{m.media_picker_no_match_body()}</p>
					</div>
				{:else}
					<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
						{#each media as item (item.id)}
							<button
								type="button"
								class="group relative aspect-square overflow-hidden rounded-lg border bg-muted text-left focus-visible:ring-2 focus-visible:ring-ring {selectedIDs.includes(
									item.id
								)
									? 'ring-2 ring-primary'
									: ''}"
								onclick={() => toggleMedia(item.id)}
								aria-pressed={selectedIDs.includes(item.id)}
								aria-label={m.media_picker_select_item({ name: item.original_filename })}
							>
								<img
									src={getAuthenticatedMediaURL(item.thumbnail_url || item.url)}
									alt={item.alt_text || item.original_filename}
									class="size-full object-cover transition-transform group-hover:scale-[1.02]"
									loading="lazy"
								/>
								{#if selectedIDs.includes(item.id)}
									<span
										class="absolute top-2 right-2 flex size-7 items-center justify-center rounded-full bg-primary text-primary-foreground shadow"
									>
										<CheckIcon class="size-4" />
									</span>
									<span
										class="absolute right-2 bottom-2 rounded bg-background/90 px-2 py-1 text-xs font-medium"
									>
										{selectedIDs.indexOf(item.id) + 1}
									</span>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</Tabs.Content>

			<Tabs.Content value="upload" class="flex flex-1 items-center justify-center p-4">
				<label
					class="flex min-h-64 w-full cursor-pointer flex-col items-center justify-center rounded-xl border border-dashed bg-muted/20 p-6 text-center hover:bg-muted/40"
				>
					<UploadIcon class="mb-3 size-9 text-muted-foreground" />
					<span class="font-medium">{m.media_picker_choose_files()}</span>
					<span class="mt-1 text-sm text-muted-foreground">{m.media_picker_provider_limits()}</span>
					<input
						{@attach attachUploadInput}
						type="file"
						{multiple}
						accept={accept.join(',')}
						class="sr-only"
						onchange={(event) => uploadFiles(event.currentTarget.files)}
					/>
				</label>
			</Tabs.Content>

			<Tabs.Content value="camera" class="min-h-0 flex-1 overflow-y-auto p-4">
				<CameraCapture onCapture={capturePhoto} />
			</Tabs.Content>

			{#if showCreate}
				<Tabs.Content value="create" class="flex flex-1 items-center justify-center p-4">
					<div class="max-w-md text-center">
						<div
							class="mx-auto mb-4 flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary"
						>
							<PaletteIcon class="size-7" />
						</div>
						<h3 class="text-lg font-semibold">{m.media_picker_create_title()}</h3>
						<p class="mt-2 text-sm text-muted-foreground">{m.media_picker_create_body()}</p>
						<Button class="mt-5" onclick={createDesign}>{m.media_picker_open_studio()}</Button>
					</div>
				</Tabs.Content>
			{/if}
		</Tabs.Root>

		{#if error}
			<div
				class="mx-4 mb-2 rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive"
				role="alert"
			>
				{error}
			</div>
		{/if}
		<Dialog.Footer class="border-t px-4 py-3">
			<div class="mr-auto self-center text-sm text-muted-foreground">
				{m.media_picker_selected_count({
					selected: selectedIDs.length,
					maximum: maxSelection
				})}
			</div>
			<Button variant="ghost" onclick={() => (open = false)}>{m.common_cancel()}</Button>
			<Button onclick={confirm} disabled={actionLoading || selectedIDs.length === 0}>
				{#if actionLoading}<LoaderIcon class="animate-spin" />{/if}
				{m.media_picker_add_media()}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
