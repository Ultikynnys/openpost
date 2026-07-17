<script lang="ts">
	import { onMount } from 'svelte';
	import { client, type SocialAccount, type Workspace } from '$lib/api/client';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { isSupportedMediaFile, uploadMediaFile } from '$lib/media-upload-client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import { getPlatformName } from '$lib/utils';
	import CalendarClockIcon from 'lucide-svelte/icons/calendar-clock';
	import CheckCircle2Icon from 'lucide-svelte/icons/check-circle-2';
	import CircleAlertIcon from 'lucide-svelte/icons/circle-alert';
	import FilmIcon from 'lucide-svelte/icons/film';
	import ImagesIcon from 'lucide-svelte/icons/images';
	import LinkIcon from 'lucide-svelte/icons/link';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import MessageSquareIcon from 'lucide-svelte/icons/message-square';
	import PanelsTopLeftIcon from 'lucide-svelte/icons/panels-top-left';
	import SendIcon from 'lucide-svelte/icons/send';
	import TextIcon from 'lucide-svelte/icons/text';
	import UploadIcon from 'lucide-svelte/icons/upload';
	import VideoIcon from 'lucide-svelte/icons/video';
	import XIcon from 'lucide-svelte/icons/x';
	import { m } from '$lib/paraglide/messages';

	type Profile = {
		key: string;
		name: string;
		description: string;
	};

	type SettingField = {
		key: string;
		label: string;
		type: string;
		required?: boolean;
		options?: string[] | null;
		help?: string;
	};

	type Capability = {
		provider: string;
		profile: string;
		label: string;
		text_limit?: number;
		title_required?: boolean;
		requires_app_review?: boolean;
		requires_public_media?: boolean;
		media: {
			min_count: number;
			max_count: number;
			allowed_mimes?: string[] | null;
			aspect_ratios?: string[] | null;
			max_duration_seconds?: number;
			requires_public_url?: boolean;
		};
		settings?: SettingField[] | null;
		caveats?: string[] | null;
	};

	type UploadedMedia = {
		id: string;
		mime_type: string;
		url?: string;
		size?: number;
		filename?: string;
	};

	type RenditionDraft = {
		accountId: string;
		platform: string;
		profile: string;
		body: string;
		title: string;
		description: string;
		settings: Record<string, unknown>;
		mediaIds: string[];
	};

	type ValidationIssue = {
		severity: string;
		code: string;
		message: string;
		provider?: string;
		profile?: string;
		media_id?: string;
		field?: string;
	};

	type ProviderReadinessItem = {
		provider: string;
		configured_app_state: string;
		connected_accounts: number;
		blocking_issues?: string[] | null;
		next_actions?: string[] | null;
		public_media_health?: {
			status: string;
			failing_count: number;
			last_failure?: string;
		};
	};

	const profileIcons: Record<string, typeof TextIcon> = {
		short_text: TextIcon,
		thread: MessageSquareIcon,
		link_share: LinkIcon,
		image_post: ImagesIcon,
		carousel: PanelsTopLeftIcon,
		story: FilmIcon,
		short_video: VideoIcon,
		long_video: VideoIcon
	};

	let workspaces = $state<Workspace[]>([]);
	let selectedWorkspaceId = $state('');
	let accounts = $state<SocialAccount[]>([]);
	let profiles = $state<Profile[]>([]);
	let capabilities = $state<Capability[]>([]);
	let media = $state<UploadedMedia[]>([]);
	let selectedProfile = $state('short_text');
	let selectedAccountIds = $state<string[]>([]);
	let renditions = $state<RenditionDraft[]>([]);
	let activeRenditionAccountId = $state('');
	let title = $state('');
	let sourceText = $state('');
	let sourceURL = $state('');
	let goal = $state('');
	let audience = $state('');
	let scheduleAt = $state('');
	let validationIssues = $state<ValidationIssue[]>([]);
	let providerReadiness = $state.raw<ProviderReadinessItem[]>([]);
	let loading = $state(true);
	let accountsLoading = $state(false);
	let uploading = $state(false);
	let saving = $state(false);
	let error = $state('');
	let success = $state('');

	let compatibleCapabilities = $derived(
		capabilities.filter((capability) => capability.profile === selectedProfile)
	);
	let selectedAccounts = $derived(
		accounts.filter((account) => selectedAccountIds.includes(account.id))
	);
	let activeRendition = $derived(
		renditions.find((rendition) => rendition.accountId === activeRenditionAccountId) ??
			renditions[0] ??
			null
	);
	let selectedProfileMeta = $derived(
		profiles.find((profile) => profile.key === selectedProfile) ?? profiles[0] ?? null
	);
	let canPublish = $derived(
		selectedWorkspaceId !== '' &&
			selectedProfile !== '' &&
			renditions.length > 0 &&
			selectedAccounts.every((account) => readinessBlockers(account).length === 0) &&
			!saving
	);
	let blockingIssues = $derived(validationIssues.filter((issue) => issue.severity === 'error'));
	let warningIssues = $derived(validationIssues.filter((issue) => issue.severity === 'warning'));

	onMount(async () => {
		await loadInitialData();
	});

	async function loadInitialData() {
		loading = true;
		error = '';
		try {
			const [{ data: workspaceData }, { data: capabilityData, error: capabilityError }] =
				await Promise.all([client.GET('/workspaces', {}), client.GET('/capabilities', {})]);
			if (capabilityError) throw new Error(capabilityError.detail || 'Failed to load capabilities');
			workspaces = workspaceData ?? [];
			profiles = capabilityData?.profiles ?? [];
			capabilities = capabilityData?.capabilities ?? [];
			selectedWorkspaceId = workspaceCtx.currentWorkspace?.id ?? workspaces[0]?.id ?? '';
			if (profiles.length > 0 && !profiles.some((profile) => profile.key === selectedProfile)) {
				selectedProfile = profiles[0].key;
			}
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
			accounts = data ?? [];
			selectedAccountIds = selectedAccountIds.filter((id) =>
				accounts.some((account) => account.id === id)
			);
			syncRenditions();
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

	function selectProfile(profileKey: string) {
		selectedProfile = profileKey;
		selectedAccountIds = selectedAccountIds.filter((accountId) => {
			const account = accounts.find((item) => item.id === accountId);
			return account ? isAccountCompatible(account) : false;
		});
		syncRenditions();
		validationIssues = [];
	}

	function toggleAccount(account: SocialAccount) {
		if (!isAccountCompatible(account)) return;
		selectedAccountIds = selectedAccountIds.includes(account.id)
			? selectedAccountIds.filter((id) => id !== account.id)
			: [...selectedAccountIds, account.id];
		syncRenditions();
	}

	function isAccountCompatible(account: SocialAccount): boolean {
		return compatibleCapabilities.some((capability) => capability.provider === account.platform);
	}

	function capabilityForAccount(account: SocialAccount | RenditionDraft): Capability | null {
		return (
			capabilities.find(
				(capability) =>
					capability.provider === account.platform && capability.profile === selectedProfile
			) ?? null
		);
	}

	function disabledReason(account: SocialAccount): string {
		if (isAccountCompatible(account)) return '';
		return `${getPlatformName(account.platform)} does not support ${selectedProfileMeta?.name ?? selectedProfile}`;
	}

	function readinessForAccount(account: SocialAccount): ProviderReadinessItem | null {
		return providerReadiness.find((item) => item.provider === account.platform) ?? null;
	}

	function readinessBlockers(account: SocialAccount): string[] {
		return readinessForAccount(account)?.blocking_issues ?? [];
	}

	function accountBlockerText(account: SocialAccount): string {
		const blockers = readinessBlockers(account);
		if (blockers.length === 0) return '';
		return blockers.map((item) => item.replaceAll('_', ' ')).join(', ');
	}

	function syncRenditions() {
		const next: RenditionDraft[] = [];
		for (const account of accounts.filter((item) => selectedAccountIds.includes(item.id))) {
			const existing = renditions.find((rendition) => rendition.accountId === account.id);
			next.push({
				accountId: account.id,
				platform: account.platform,
				profile: selectedProfile,
				body: existing?.body ?? sourceText,
				title: existing?.title ?? title,
				description: existing?.description ?? '',
				settings: normalizeSettings(existing?.settings ?? {}, capabilityForAccount(account)),
				mediaIds: existing?.mediaIds ?? media.map((item) => item.id)
			});
		}
		renditions = next;
		activeRenditionAccountId = next.some(
			(rendition) => rendition.accountId === activeRenditionAccountId
		)
			? activeRenditionAccountId
			: (next[0]?.accountId ?? '');
	}

	function normalizeSettings(
		current: Record<string, unknown>,
		capability: Capability | null
	): Record<string, unknown> {
		const next = { ...current };
		for (const field of capability?.settings ?? []) {
			if (next[field.key] !== undefined) continue;
			if (field.key === 'privacy') next[field.key] = 'private';
			else if (field.key === 'content_posting_method') next[field.key] = 'DIRECT_POST';
			else if (field.key === 'privacy_level') next[field.key] = 'SELF_ONLY';
			else if (field.key === 'post_type')
				next[field.key] = selectedProfile === 'story' ? 'story' : 'post';
			else if (field.type === 'boolean') next[field.key] = false;
			else next[field.key] = '';
		}
		return next;
	}

	function updateActiveRendition(patch: Partial<RenditionDraft>) {
		if (!activeRendition) return;
		renditions = renditions.map((rendition) =>
			rendition.accountId === activeRendition.accountId ? { ...rendition, ...patch } : rendition
		);
	}

	function updateActiveSetting(key: string, value: unknown) {
		if (!activeRendition) return;
		updateActiveRendition({ settings: { ...activeRendition.settings, [key]: value } });
	}

	function syncSourceToRenditions() {
		renditions = renditions.map((rendition) => ({
			...rendition,
			body: sourceText,
			title: title || rendition.title,
			mediaIds: media.map((item) => item.id)
		}));
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
			renditions = renditions.map((rendition) => ({
				...rendition,
				mediaIds: media.map((item) => item.id)
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Upload failed';
		} finally {
			uploading = false;
		}
	}

	function removeMedia(mediaId: string) {
		media = media.filter((item) => item.id !== mediaId);
		renditions = renditions.map((rendition) => ({
			...rendition,
			mediaIds: rendition.mediaIds.filter((id) => id !== mediaId)
		}));
	}

	async function createPublication(action: 'draft' | 'validate' | 'schedule' | 'publish') {
		if (!canPublish) return;
		saving = true;
		error = '';
		success = '';
		try {
			const body = publicationPayload();
			const { data, error: createError } = await client.POST('/publications', { body });
			if (createError) throw new Error(createError.detail || 'Failed to create publication');
			const publicationId = data.id;
			if (action === 'validate') {
				await validatePublication(publicationId);
				success = 'Validation complete';
			} else if (action === 'schedule') {
				const { data: scheduleData, error: scheduleError } = await client.POST(
					'/publications/{id}/schedule',
					{ params: { path: { id: publicationId } } }
				);
				if (scheduleError) throw new Error(scheduleError.detail || 'Failed to schedule');
				success = scheduleData?.message ?? 'Publication scheduled';
			} else if (action === 'publish') {
				const { data: publishData, error: publishError } = await client.POST(
					'/publications/{id}/publish-now',
					{ params: { path: { id: publicationId } } }
				);
				if (publishError) throw new Error(publishError.detail || 'Failed to publish');
				success = publishData?.message ?? 'Publication queued';
			} else {
				success = 'Draft saved';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Composer action failed';
		} finally {
			saving = false;
		}
	}

	async function validatePublication(publicationId: string) {
		const { data, error: err } = await client.POST('/publications/{id}/validate', {
			params: { path: { id: publicationId } }
		});
		if (err) throw new Error(err.detail || 'Validation failed');
		validationIssues = data?.issues ?? [];
	}

	function publicationPayload() {
		return {
			workspace_id: selectedWorkspaceId,
			title: title || selectedProfileMeta?.name || 'Publication',
			content_profile: selectedProfile,
			source_text: sourceText,
			source_url: sourceURL,
			goal,
			audience,
			...(scheduleAt ? { scheduled_at: new Date(scheduleAt).toISOString() } : {}),
			metadata: {
				composer: 'format-first',
				created_from: 'web'
			},
			media: media.map((item, index) => ({
				media_id: item.id,
				role: 'attachment',
				display_order: index
			})),
			renditions: renditions.map((rendition) => ({
				social_account_id: rendition.accountId,
				profile: selectedProfile,
				body: rendition.body,
				title: rendition.title,
				description: rendition.description,
				settings: rendition.settings,
				media: rendition.mediaIds.map((mediaId) => ({ media_id: mediaId, role: 'attachment' }))
			}))
		};
	}
</script>

<div class="flex min-h-0 flex-1 flex-col bg-background">
	{#if loading}
		<div class="flex flex-1 items-center justify-center text-sm text-muted-foreground">
			<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
			Loading composer
		</div>
	{:else}
		<div class="border-b bg-background px-4 py-3">
			<div class="flex flex-wrap items-center gap-3">
				<select
					class="h-9 rounded-md border bg-background px-3 text-sm"
					bind:value={selectedWorkspaceId}
					onchange={async () => {
						await loadAccounts();
						await loadProviderReadiness();
					}}
				>
					{#each workspaces as workspace (workspace.id)}
						<option value={workspace.id}>{workspace.name}</option>
					{/each}
				</select>
				<div class="flex min-w-0 flex-1 items-center gap-2 text-sm text-muted-foreground">
					<span class="font-medium text-foreground">{m.compose_format_first()}</span>
					<span>{selectedProfileMeta?.description}</span>
				</div>
				<Button
					variant="outline"
					size="sm"
					disabled={saving}
					onclick={() => createPublication('draft')}
				>
					Save draft
				</Button>
				<Button
					variant="outline"
					size="sm"
					disabled={saving || !canPublish}
					onclick={() => createPublication('validate')}
				>
					<CheckCircle2Icon class="mr-2 h-4 w-4" />
					Validate
				</Button>
				<Button
					variant="outline"
					size="sm"
					disabled={saving || !scheduleAt || !canPublish}
					onclick={() => createPublication('schedule')}
				>
					<CalendarClockIcon class="mr-2 h-4 w-4" />
					Schedule
				</Button>
				<Button
					size="sm"
					disabled={saving || !canPublish}
					onclick={() => createPublication('publish')}
				>
					{#if saving}
						<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
					{:else}
						<SendIcon class="mr-2 h-4 w-4" />
					{/if}
					{m.compose_publish()}
				</Button>
			</div>
		</div>

		<div
			class="grid min-h-0 flex-1 grid-cols-1 overflow-hidden lg:grid-cols-[280px_minmax(0,1fr)_360px]"
		>
			<aside class="border-r bg-muted/20 p-4 lg:overflow-y-auto">
				<div class="space-y-3">
					<div>
						<p class="text-xs font-semibold text-muted-foreground uppercase">
							{m.compose_profile()}
						</p>
						<div class="mt-2 grid gap-2">
							{#each profiles as profile (profile.key)}
								{@const Icon = profileIcons[profile.key] ?? TextIcon}
								<button
									type="button"
									class="rounded-md border p-3 text-left transition hover:border-primary/60 {selectedProfile ===
									profile.key
										? 'border-primary bg-primary/5'
										: 'bg-background'}"
									onclick={() => selectProfile(profile.key)}
								>
									<div class="flex items-center gap-2">
										<Icon class="h-4 w-4" />
										<span class="text-sm font-medium">{profile.name}</span>
									</div>
									<p class="mt-1 line-clamp-2 text-xs text-muted-foreground">
										{profile.description}
									</p>
								</button>
							{/each}
						</div>
					</div>
				</div>
			</aside>

			<main class="min-h-0 overflow-y-auto p-4">
				<div class="mx-auto max-w-5xl space-y-5">
					{#if error}
						<div
							class="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
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
						<div class="flex items-center justify-between">
							<div>
								<h2 class="text-base font-semibold">{m.compose_destinations()}</h2>
								<p class="text-sm text-muted-foreground">
									Only accounts compatible with {selectedProfileMeta?.name} are selectable.
								</p>
							</div>
							{#if accountsLoading}
								<LoaderIcon class="h-4 w-4 animate-spin text-muted-foreground" />
							{/if}
						</div>
						<div class="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
							{#each accounts as account (account.id)}
								{@const compatible = isAccountCompatible(account)}
								{@const blockers = readinessBlockers(account)}
								<button
									type="button"
									class="rounded-md border p-3 text-left transition {selectedAccountIds.includes(
										account.id
									)
										? 'border-primary bg-primary/5'
										: 'bg-background'} {compatible && blockers.length === 0
										? 'hover:border-primary/60'
										: 'opacity-55'}"
									disabled={!compatible || blockers.length > 0}
									title={disabledReason(account) || accountBlockerText(account)}
									onclick={() => toggleAccount(account)}
								>
									<div class="flex items-center gap-2">
										<PlatformIcon platform={account.platform} class="h-4 w-4" />
										<div class="min-w-0">
											<p class="truncate text-sm font-medium">
												{account.account_username || account.slug}
											</p>
											<p class="text-xs text-muted-foreground">
												{getPlatformName(account.platform)}
											</p>
										</div>
									</div>
									{#if !compatible}
										<p class="mt-2 text-xs text-muted-foreground">{disabledReason(account)}</p>
									{/if}
									{#if blockers.length > 0}
										<p class="mt-2 text-xs text-destructive">
											Blocked: {accountBlockerText(account)}
										</p>
									{:else if readinessForAccount(account)?.next_actions?.[0]}
										<p class="mt-2 text-xs text-muted-foreground">
											{readinessForAccount(account)?.next_actions?.[0]}
										</p>
									{/if}
								</button>
							{/each}
						</div>
					</section>

					<section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
						<div class="space-y-3">
							<div class="grid gap-3 md:grid-cols-2">
								<div>
									<label class="text-xs font-medium text-muted-foreground" for="publication-title"
										>{m.compose_publication_title()}</label
									>
									<Input
										id="publication-title"
										bind:value={title}
										oninput={syncSourceToRenditions}
										placeholder={m.compose_title_placeholder()}
									/>
								</div>
								<div>
									<label class="text-xs font-medium text-muted-foreground" for="source-url"
										>{m.compose_source_url()}</label
									>
									<Input id="source-url" bind:value={sourceURL} placeholder="https://..." />
								</div>
							</div>
							<div>
								<label class="text-xs font-medium text-muted-foreground" for="source-text"
									>{m.compose_source_text()}</label
								>
								<Textarea
									id="source-text"
									class="min-h-44 resize-y"
									bind:value={sourceText}
									oninput={syncSourceToRenditions}
									placeholder={m.compose_source_text_placeholder()}
								/>
							</div>
							<div class="grid gap-3 md:grid-cols-3">
								<Input bind:value={goal} placeholder={m.compose_goal()} />
								<Input bind:value={audience} placeholder={m.compose_audience()} />
								<Input type="datetime-local" bind:value={scheduleAt} />
							</div>
							<div class="rounded-md border border-dashed p-4">
								<div class="flex flex-wrap items-center gap-3">
									<label
										class="inline-flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-2 text-sm font-medium"
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
											disabled={uploading}
											onchange={(event) => handleFiles(event.currentTarget.files)}
										/>
									</label>
									<p class="text-sm text-muted-foreground">
										Media rules come from the selected profile and provider capabilities.
									</p>
								</div>
								{#if media.length > 0}
									<div class="mt-3 flex flex-wrap gap-2">
										{#each media as item (item.id)}
											<div
												class="flex items-center gap-2 rounded-md border bg-background px-2 py-1 text-xs"
											>
												<span class="max-w-40 truncate">{item.filename || item.id}</span>
												<span class="text-muted-foreground">{item.mime_type}</span>
												<button
													type="button"
													class="text-muted-foreground hover:text-foreground"
													onclick={() => removeMedia(item.id)}
												>
													<XIcon class="h-3.5 w-3.5" />
												</button>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						</div>

						<div class="rounded-md border bg-background">
							<div class="border-b px-3 py-2">
								<h3 class="text-sm font-semibold">{m.compose_validation()}</h3>
							</div>
							<div class="space-y-2 p-3 text-sm">
								{#if validationIssues.length === 0}
									<p class="text-muted-foreground">
										Run validation before scheduling or publishing.
									</p>
								{:else}
									{#each blockingIssues as issue (`error-${issue.code}-${issue.field}-${issue.media_id}`)}
										<div class="flex gap-2 text-destructive">
											<CircleAlertIcon class="mt-0.5 h-4 w-4" />
											<span>{issue.message}</span>
										</div>
									{/each}
									{#each warningIssues as issue (`warning-${issue.code}-${issue.field}-${issue.media_id}`)}
										<div class="flex gap-2 text-amber-700 dark:text-amber-300">
											<CircleAlertIcon class="mt-0.5 h-4 w-4" />
											<span>{issue.message}</span>
										</div>
									{/each}
									{#if blockingIssues.length === 0}
										<div class="flex gap-2 text-emerald-700 dark:text-emerald-300">
											<CheckCircle2Icon class="mt-0.5 h-4 w-4" />
											<span>{m.compose_no_blocking_issues()}</span>
										</div>
									{/if}
								{/if}
							</div>
						</div>
					</section>
				</div>
			</main>

			<aside class="min-h-0 border-l bg-muted/20 lg:overflow-y-auto">
				<div class="border-b p-4">
					<h2 class="text-base font-semibold">{m.compose_renditions()}</h2>
					<p class="text-sm text-muted-foreground">{m.compose_renditions_body()}</p>
				</div>
				<div class="space-y-3 p-4">
					{#if renditions.length === 0}
						<p class="text-sm text-muted-foreground">
							Select compatible accounts to create renditions.
						</p>
					{:else}
						<div class="grid gap-2">
							{#each renditions as rendition (rendition.accountId)}
								{@const account = accounts.find((item) => item.id === rendition.accountId)}
								{@const capability = capabilityForAccount(rendition)}
								<button
									type="button"
									class="rounded-md border p-3 text-left {activeRenditionAccountId ===
									rendition.accountId
										? 'border-primary bg-primary/5'
										: 'bg-background'}"
									onclick={() => (activeRenditionAccountId = rendition.accountId)}
								>
									<div class="flex items-center gap-2">
										<PlatformIcon platform={rendition.platform} class="h-4 w-4" />
										<div class="min-w-0 flex-1">
											<p class="truncate text-sm font-medium">
												{account?.account_username || account?.slug}
											</p>
											<p class="text-xs text-muted-foreground">{capability?.label}</p>
										</div>
									</div>
								</button>
							{/each}
						</div>

						{#if activeRendition}
							{@const capability = capabilityForAccount(activeRendition)}
							<div class="space-y-3 rounded-md border bg-background p-3">
								<div>
									<label class="text-xs font-medium text-muted-foreground" for="rendition-title"
										>{m.compose_rendition_title()}</label
									>
									<Input
										id="rendition-title"
										value={activeRendition.title}
										oninput={(event) => updateActiveRendition({ title: event.currentTarget.value })}
									/>
								</div>
								<div>
									<label class="text-xs font-medium text-muted-foreground" for="rendition-body"
										>{m.compose_body()}</label
									>
									<Textarea
										id="rendition-body"
										class="min-h-32"
										value={activeRendition.body}
										maxlength={capability?.text_limit}
										oninput={(event) => updateActiveRendition({ body: event.currentTarget.value })}
									/>
									{#if capability?.text_limit}
										<p class="mt-1 text-right text-xs text-muted-foreground">
											{activeRendition.body.length}/{capability.text_limit}
										</p>
									{/if}
								</div>
								<div>
									<label
										class="text-xs font-medium text-muted-foreground"
										for="rendition-description">{m.compose_description()}</label
									>
									<Textarea
										id="rendition-description"
										class="min-h-24"
										value={activeRendition.description}
										oninput={(event) =>
											updateActiveRendition({ description: event.currentTarget.value })}
									/>
								</div>

								{#if capability?.settings?.length}
									<div class="space-y-2">
										<p class="text-xs font-semibold text-muted-foreground uppercase">
											Provider settings
										</p>
										{#each capability.settings as setting (setting.key)}
											<div>
												<label
													class="text-xs font-medium text-muted-foreground"
													for={`setting-${setting.key}`}>{setting.label}</label
												>
												{#if setting.type === 'select'}
													<select
														id={`setting-${setting.key}`}
														class="h-9 w-full rounded-md border bg-background px-2 text-sm"
														value={String(activeRendition.settings[setting.key] ?? '')}
														onchange={(event) =>
															updateActiveSetting(setting.key, event.currentTarget.value)}
													>
														{#each setting.options ?? [] as option (option)}
															<option value={option}>{option}</option>
														{/each}
													</select>
												{:else if setting.type === 'boolean'}
													<label class="mt-1 flex items-center gap-2 text-sm">
														<input
															type="checkbox"
															checked={Boolean(activeRendition.settings[setting.key])}
															onchange={(event) =>
																updateActiveSetting(setting.key, event.currentTarget.checked)}
														/>
														Enabled
													</label>
												{:else if setting.type === 'textarea' || setting.type === 'json'}
													<Textarea
														id={`setting-${setting.key}`}
														value={String(activeRendition.settings[setting.key] ?? '')}
														oninput={(event) =>
															updateActiveSetting(setting.key, event.currentTarget.value)}
													/>
												{:else}
													<Input
														id={`setting-${setting.key}`}
														type={setting.type === 'number' ? 'number' : 'text'}
														value={String(activeRendition.settings[setting.key] ?? '')}
														oninput={(event) =>
															updateActiveSetting(setting.key, event.currentTarget.value)}
													/>
												{/if}
												{#if setting.help}
													<p class="mt-1 text-xs text-muted-foreground">{setting.help}</p>
												{/if}
											</div>
										{/each}
									</div>
								{/if}

								<div class="rounded-md border bg-muted/30 p-3">
									<p class="text-xs font-semibold text-muted-foreground uppercase">
										{m.compose_preview()}
									</p>
									<p class="mt-2 text-sm font-medium">{activeRendition.title || title}</p>
									<p class="mt-2 text-sm whitespace-pre-wrap text-muted-foreground">
										{activeRendition.body || 'No body text yet.'}
									</p>
									{#if activeRendition.mediaIds.length > 0}
										<p class="mt-2 text-xs text-muted-foreground">
											{activeRendition.mediaIds.length} media item{activeRendition.mediaIds
												.length === 1
												? ''
												: 's'} attached
										</p>
									{/if}
								</div>
							</div>
						{/if}
					{/if}
				</div>
			</aside>
		</div>
	{/if}
</div>
