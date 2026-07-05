<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { client, type Post } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Tabs, TabsList, TabsTrigger, TabsContent } from '$lib/components/ui/tabs';
	import CommentInbox from '$lib/components/comment-inbox.svelte';
	import PageContainer from '$lib/components/page-container.svelte';
	import EmptyState from '$lib/components/empty-state.svelte';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import ClockIcon from 'lucide-svelte/icons/clock';
	import CheckCircleIcon from 'lucide-svelte/icons/circle-check';
	import XCircleIcon from 'lucide-svelte/icons/circle-x';
	import AlertCircleIcon from 'lucide-svelte/icons/alert-circle';
	import RefreshIcon from 'lucide-svelte/icons/refresh-cw';
	import CalendarIcon from 'lucide-svelte/icons/calendar';
	import FileTextIcon from 'lucide-svelte/icons/file-text';
	import CpuIcon from 'lucide-svelte/icons/cpu';
	import PencilIcon from 'lucide-svelte/icons/pencil';
	import PackageIcon from 'lucide-svelte/icons/package';
	import ScrollTextIcon from 'lucide-svelte/icons/scroll-text';
	import MessageCircleIcon from 'lucide-svelte/icons/message-circle';
	import { m } from '$lib/paraglide/messages';
	import { getLocaleTag } from '$lib/i18n';

	type PublicationResponse = components['schemas']['PublicationResponse'];
	type RenditionResponse = components['schemas']['RenditionResponse'];

	type JobLog = {
		id: string;
		type: string;
		status: string;
		payload?: string;
		run_at: string;
		attempts: number;
		max_attempts: number;
		last_error?: string;
		locked_at?: string;
	};

	type PublishedRendition = RenditionResponse & {
		publicationTitle: string;
		publicationSourceText: string;
	};

	const jobsPageSize = 50;

	let posts = $state<Post[]>([]);
	let publications = $state<PublicationResponse[]>([]);
	let scheduledPosts = $state<Post[]>([]);
	let drafts = $state<Post[]>([]);
	let jobs = $state<JobLog[]>([]);
	let jobsTotal = $state(0);
	let jobsNextOffset = $state(0);
	let jobsHasMore = $state(false);
	let jobsLoadingMore = $state(false);
	let loading = $state(true);
	let error = $state('');
	let activeTab = $state('schedule');
	let selectedRenditionId = $state('');

	onMount(() => {
		loadData();
	});

	async function loadData() {
		loading = true;
		error = '';
		try {
			await Promise.all([
				loadPosts(),
				loadPublications(),
				loadScheduled(),
				loadDrafts(),
				loadJobs()
			]);
		} catch (e) {
			error = (e as Error).message || m.activity_failed_load();
		} finally {
			loading = false;
		}
	}

	async function loadPosts() {
		const { data, error: err } = await client.GET('/posts', {
			params: { query: { limit: 100 } }
		});
		if (err || !data) throw new Error(m.activity_failed_posts());
		posts = data.sort(
			(a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
		);
	}

	async function loadPublications() {
		if (!workspaceCtx.currentWorkspace) {
			await workspaceCtx.initialize();
		}
		const workspaceId = workspaceCtx.currentWorkspace?.id;
		if (!workspaceId) {
			publications = [];
			selectedRenditionId = '';
			return;
		}

		const { data, error: err } = await (client as any).GET('/publications', {
			params: { query: { workspace_id: workspaceId, status: 'published', limit: 100 } }
		});
		if (err || !data) throw new Error(readProblem(err, 'Failed to load publications'));
		publications = data;
		if (!publishedRenditions.some((rendition) => rendition.id === selectedRenditionId)) {
			selectedRenditionId = publishedRenditions[0]?.id ?? '';
		}
	}

	async function loadScheduled() {
		const { data, error: err } = await client.GET('/posts', {
			params: { query: { status: 'scheduled', limit: 100 } }
		});
		if (err || !data) throw new Error(m.activity_failed_scheduled());
		scheduledPosts = data.sort(
			(a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime()
		);
	}

	async function loadDrafts() {
		const { data, error: err } = await client.GET('/posts', {
			params: { query: { status: 'draft', limit: 100 } }
		});
		if (err || !data) throw new Error(m.activity_failed_drafts());
		drafts = data.sort(
			(a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
		);
	}

	async function loadJobs(offset = 0, append = false) {
		const {
			data,
			error: err,
			response
		} = await client.GET('/jobs', {
			params: { query: { limit: jobsPageSize, offset } }
		});
		if (err || !data) throw new Error(m.activity_failed_jobs());
		jobs = append ? [...jobs, ...data] : data;
		jobsTotal = readIntHeader(response.headers.get('X-Total-Count'), jobs.length);
		jobsNextOffset = readIntHeader(response.headers.get('X-Next-Offset'), jobs.length);
		jobsHasMore = response.headers.get('X-Has-More') === 'true';
	}

	function readProblem(value: unknown, fallback: string): string {
		if (value && typeof value === 'object' && 'detail' in value) {
			const detail = (value as { detail?: unknown }).detail;
			if (typeof detail === 'string' && detail.length > 0) return detail;
		}
		return fallback;
	}

	async function loadMoreJobs() {
		if (jobsLoadingMore || !jobsHasMore) return;
		jobsLoadingMore = true;
		error = '';
		try {
			await loadJobs(jobsNextOffset, true);
		} catch (e) {
			error = (e as Error).message || m.activity_failed_jobs();
		} finally {
			jobsLoadingMore = false;
		}
	}

	function readIntHeader(value: string | null, fallback: number): number {
		if (!value) return fallback;
		const parsed = Number.parseInt(value, 10);
		return Number.isFinite(parsed) ? parsed : fallback;
	}

	function formatRelative(iso: string): string {
		if (!iso) return '-';
		const d = new Date(iso);
		const now = new Date();
		const diffMs = d.getTime() - now.getTime();
		const diffMins = Math.round(diffMs / 60000);
		const diffHours = Math.round(diffMs / 3600000);
		const diffDays = Math.round(diffMs / 86400000);

		if (Math.abs(diffMins) < 1) return m.activity_just_now();
		if (diffMins > 0 && diffMins < 60) return m.activity_in_minutes({ count: diffMins });
		if (diffMins < 0 && diffMins > -60)
			return m.activity_minutes_ago({ count: Math.abs(diffMins) });
		if (diffHours > 0 && diffHours < 24) return m.activity_in_hours({ count: diffHours });
		if (diffHours < 0 && diffHours > -24)
			return m.activity_hours_ago({ count: Math.abs(diffHours) });
		if (diffDays > 0 && diffDays < 7) return m.activity_in_days({ count: diffDays });
		if (diffDays < 0 && diffDays > -7) return m.activity_days_ago({ count: Math.abs(diffDays) });
		return d.toLocaleDateString(getLocaleTag(), { month: 'short', day: 'numeric' });
	}

	function formatDateTime(iso: string): string {
		if (!iso) return '-';
		const d = new Date(iso);
		return d.toLocaleString(getLocaleTag(), {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getPostStatusMeta(status: string) {
		switch (status) {
			case 'published':
				return {
					icon: CheckCircleIcon,
					label: m.activity_status_published(),
					class:
						'bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-950/30 dark:text-emerald-400'
				};
			case 'failed':
				return {
					icon: XCircleIcon,
					label: m.activity_status_failed(),
					class: 'bg-red-50 text-red-700 ring-red-600/20 dark:bg-red-950/30 dark:text-red-400'
				};
			case 'scheduled':
				return {
					icon: ClockIcon,
					label: m.activity_status_scheduled(),
					class:
						'bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-950/30 dark:text-amber-400'
				};
			case 'publishing':
				return {
					icon: RefreshIcon,
					label: m.activity_status_publishing(),
					class: 'bg-blue-50 text-blue-700 ring-blue-600/20 dark:bg-blue-950/30 dark:text-blue-400'
				};
			default:
				return {
					icon: AlertCircleIcon,
					label: status,
					class:
						'bg-slate-50 text-slate-700 ring-slate-600/20 dark:bg-slate-950/30 dark:text-slate-400'
				};
		}
	}

	function getJobStatusMeta(status: string) {
		switch (status) {
			case 'completed':
				return {
					icon: CheckCircleIcon,
					label: m.activity_status_completed(),
					class:
						'bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-950/30 dark:text-emerald-400'
				};
			case 'failed':
				return {
					icon: XCircleIcon,
					label: m.activity_status_failed(),
					class: 'bg-red-50 text-red-700 ring-red-600/20 dark:bg-red-950/30 dark:text-red-400'
				};
			case 'processing':
				return {
					icon: RefreshIcon,
					label: m.activity_status_processing(),
					class: 'bg-blue-50 text-blue-700 ring-blue-600/20 dark:bg-blue-950/30 dark:text-blue-400'
				};
			case 'pending':
				return {
					icon: ClockIcon,
					label: m.activity_status_pending(),
					class:
						'bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-950/30 dark:text-amber-400'
				};
			default:
				return {
					icon: AlertCircleIcon,
					label: status,
					class:
						'bg-slate-50 text-slate-700 ring-slate-600/20 dark:bg-slate-950/30 dark:text-slate-400'
				};
		}
	}

	function truncate(str: string, max: number = 100): string {
		if (!str) return '';
		if (str.length <= max) return str;
		return str.slice(0, max).trim() + '...';
	}

	function extractPostIdFromPayload(payload: string): string {
		try {
			const p = JSON.parse(payload);
			return p.post_id || payload;
		} catch {
			return payload;
		}
	}

	const stats = $derived([
		{
			label: m.activity_stat_scheduled(),
			value: scheduledPosts.length,
			icon: CalendarIcon,
			color: 'text-amber-600'
		},
		{
			label: m.activity_stat_drafts(),
			value: drafts.length,
			icon: FileTextIcon,
			color: 'text-slate-500'
		},
		{
			label: m.activity_stat_pending_jobs(),
			value: jobs.filter((j) => j.status === 'pending').length,
			icon: CpuIcon,
			color: 'text-blue-600'
		}
	]);

	const publishedRenditions = $derived(
		publications.flatMap((publication) =>
			(publication.renditions ?? [])
				.filter((rendition) => rendition.status === 'published' && rendition.external_id)
				.map((rendition) => ({
					...rendition,
					publicationTitle: publication.title,
					publicationSourceText: publication.source_text
				}))
		)
	);

	const selectedRendition = $derived(
		publishedRenditions.find((rendition) => rendition.id === selectedRenditionId)
	);
</script>

<svelte:head>
	<title>{m.activity_title()} — {m.common_openpost()}</title>
</svelte:head>

<PageContainer
	title={m.activity_title()}
	description={m.activity_description()}
	icon={ScrollTextIcon}
	{loading}
>
	{#snippet actions()}
		<Button variant="outline" size="sm" onclick={loadData} disabled={loading}>
			<RefreshIcon class="mr-1.5 h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" />
			{m.common_refresh()}
		</Button>
	{/snippet}

	{#if error}
		<div
			class="mb-6 flex items-center gap-3 rounded-lg border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive"
		>
			<AlertCircleIcon class="h-4 w-4 shrink-0" />
			{error}
		</div>
	{/if}

	<!-- Stats -->
	<div class="mb-8 grid grid-cols-3 gap-3">
		{#each stats as stat (stat.label)}
			<div class="rounded-xl border bg-card p-4">
				<div class="flex items-center gap-3">
					<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
						<stat.icon class="h-4 w-4 {stat.color}" />
					</div>
					<div>
						<p class="text-xl leading-none font-semibold">{stat.value}</p>
						<p class="mt-1 text-xs text-muted-foreground">{stat.label}</p>
					</div>
				</div>
			</div>
		{/each}
	</div>

	<Tabs bind:value={activeTab}>
		<TabsList variant="line" class="mb-6">
			<TabsTrigger value="schedule">{m.activity_tab_scheduled()}</TabsTrigger>
			<TabsTrigger value="comments">Comments</TabsTrigger>
			<TabsTrigger value="drafts">{m.activity_tab_drafts()}</TabsTrigger>
			<TabsTrigger value="jobs">{m.activity_tab_jobs()}</TabsTrigger>
		</TabsList>

		<!-- SCHEDULED -->
		<TabsContent value="schedule">
			{#if scheduledPosts.length === 0}
				<EmptyState
					icon={CalendarIcon}
					title={m.activity_empty_scheduled_title()}
					description={m.activity_empty_scheduled_description()}
					variant="muted"
				/>
			{:else}
				<div class="space-y-3">
					{#each scheduledPosts as post (post.id)}
						{@const meta = getPostStatusMeta(post.status)}
						<div
							class="group relative flex items-start gap-4 rounded-xl border bg-card p-4 transition-colors hover:bg-accent/40"
						>
							<div class="relative mt-1.5 flex flex-col items-center">
								<div
									class="h-2.5 w-2.5 rounded-full ring-2 {post.status === 'scheduled'
										? 'bg-amber-500 ring-amber-500/30'
										: post.status === 'published'
											? 'bg-emerald-500 ring-emerald-500/30'
											: 'bg-slate-400 ring-slate-400/30'}"
								></div>
							</div>

							<div class="min-w-0 flex-1">
								<div class="flex flex-wrap items-center gap-2">
									<span
										class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 {meta.class}"
									>
										<meta.icon class="h-3 w-3" />
										{meta.label}
									</span>
									<span class="text-[11px] text-muted-foreground">
										{formatRelative(post.scheduled_at)}
										<span class="text-muted-foreground/50">·</span>
										{formatDateTime(post.scheduled_at)}
									</span>
								</div>

								<p class="mt-2 text-sm leading-relaxed text-foreground/90">
									{truncate(post.content, 160)}
								</p>

								{#if post.destinations && post.destinations.length > 0}
									<div class="mt-3 flex flex-wrap items-center gap-2">
										{#each post.destinations as dest (dest.social_account_id)}
											<div
												class="inline-flex items-center gap-1.5 rounded-md bg-muted px-2 py-1 text-[11px] text-muted-foreground"
											>
												<PlatformIcon platform={dest.platform} class="h-3 w-3" />
												<span class="capitalize">{dest.platform}</span>
											</div>
										{/each}
									</div>
								{/if}
							</div>

							<Button
								size="sm"
								variant="ghost"
								class="shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
								onclick={() => goto(`/posts/${post.id}`)}
							>
								<PencilIcon class="h-3.5 w-3.5" />
							</Button>
						</div>
					{/each}
				</div>
			{/if}
		</TabsContent>

		<!-- COMMENTS -->
		<TabsContent value="comments">
			{#if publishedRenditions.length === 0}
				<EmptyState
					icon={MessageCircleIcon}
					title="No published conversations"
					description="Published renditions with provider IDs will appear here."
					variant="muted"
				/>
			{:else}
				<div class="grid gap-4 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.2fr)]">
					<div class="space-y-2">
						{#each publishedRenditions as rendition (rendition.id)}
							<button
								type="button"
								class="w-full rounded-lg border bg-card p-3 text-left transition-colors hover:bg-accent/40 {selectedRenditionId ===
								rendition.id
									? 'border-primary/40 bg-primary/5'
									: ''}"
								onclick={() => (selectedRenditionId = rendition.id)}
							>
								<div class="flex items-center justify-between gap-3">
									<div class="min-w-0">
										<div class="flex items-center gap-2">
											<PlatformIcon platform={rendition.platform} class="h-4 w-4" />
											<span class="truncate text-sm font-medium">{rendition.publicationTitle}</span>
										</div>
										<p class="mt-1 line-clamp-2 text-xs text-muted-foreground">
											{truncate(rendition.body || rendition.publicationSourceText, 110)}
										</p>
									</div>
									<span
										class="shrink-0 rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] text-emerald-700 ring-1 ring-emerald-600/20"
									>
										Published
									</span>
								</div>
							</button>
						{/each}
					</div>

					{#if selectedRendition}
						{#key selectedRendition.id}
							<CommentInbox
								renditionId={selectedRendition.id}
								platform={selectedRendition.platform}
							/>
						{/key}
					{/if}
				</div>
			{/if}
		</TabsContent>

		<!-- DRAFTS -->
		<TabsContent value="drafts">
			{#if drafts.length === 0}
				<EmptyState
					icon={FileTextIcon}
					title={m.activity_empty_drafts_title()}
					description={m.activity_empty_drafts_description()}
					variant="muted"
				/>
			{:else}
				<div class="grid gap-3">
					{#each drafts as post (post.id)}
						<div
							class="group flex items-start justify-between gap-4 rounded-xl border bg-card p-4 transition-colors hover:bg-accent/40"
						>
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2">
									<span
										class="inline-flex items-center gap-1 rounded-full bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-600 ring-1 ring-slate-600/10 dark:bg-slate-950/30 dark:text-slate-400"
									>
										<AlertCircleIcon class="h-3 w-3" />
										{m.activity_status_draft()}
									</span>
									<span class="text-[11px] text-muted-foreground">
										{formatRelative(post.created_at)}
									</span>
								</div>
								<p class="mt-2 text-sm leading-relaxed text-foreground/90">
									{truncate(post.content, 160)}
								</p>
							</div>
							<Button
								size="sm"
								variant="ghost"
								class="shrink-0"
								onclick={() => goto(`/posts/${post.id}`)}
							>
								<PencilIcon class="mr-1 h-3.5 w-3.5" />
								{m.common_edit()}
							</Button>
						</div>
					{/each}
				</div>
			{/if}
		</TabsContent>

		<!-- JOBS -->
		<TabsContent value="jobs">
			{#if jobs.length === 0}
				<EmptyState
					icon={CpuIcon}
					title={m.activity_empty_jobs_title()}
					description={m.activity_empty_jobs_description()}
					variant="muted"
				/>
			{:else}
				<div class="space-y-2">
					{#each jobs as job (job.id)}
						{@const meta = getJobStatusMeta(job.status)}
						<div class="rounded-xl border bg-card p-4" data-testid="job-row">
							<div class="flex items-start justify-between gap-4">
								<div class="flex items-center gap-3">
									<span
										class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 {meta.class}"
									>
										<meta.icon class="h-3 w-3" />
										{meta.label}
									</span>
									<span class="text-xs font-medium text-muted-foreground">
										{job.type.replace(/_/g, ' ')}
									</span>
								</div>
								<span class="shrink-0 text-[11px] text-muted-foreground">
									{formatRelative(job.run_at)}
								</span>
							</div>

							<div
								class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-muted-foreground"
							>
								<span>{m.activity_run_at({ time: formatDateTime(job.run_at) })}</span>
								<span class="hidden sm:inline">·</span>
								<span>{m.activity_attempts({ attempts: job.attempts, max: job.max_attempts })}</span
								>
								{#if job.locked_at}
									<span class="hidden sm:inline">·</span>
									<span>{m.activity_locked({ time: formatRelative(job.locked_at) })}</span>
								{/if}
							</div>

							{#if job.last_error}
								<div class="mt-2 rounded-md bg-destructive/5 px-3 py-2 text-xs text-destructive">
									{job.last_error}
								</div>
							{/if}

							{#if job.payload}
								<div
									class="mt-2 inline-flex items-center gap-1.5 font-mono text-[11px] text-muted-foreground/70"
								>
									<PackageIcon class="h-3 w-3" />
									{extractPostIdFromPayload(job.payload)}
								</div>
							{/if}
						</div>
					{/each}
					{#if jobsHasMore}
						<div class="flex justify-center pt-2">
							<Button
								variant="outline"
								size="sm"
								onclick={loadMoreJobs}
								disabled={jobsLoadingMore}
								data-testid="jobs-load-more"
							>
								<RefreshIcon class="mr-1.5 h-3.5 w-3.5 {jobsLoadingMore ? 'animate-spin' : ''}" />
								{m.activity_load_more_jobs({
									count: Math.max(jobsTotal - jobs.length, 0)
								})}
							</Button>
						</div>
					{/if}
				</div>
			{/if}
		</TabsContent>
	</Tabs>
</PageContainer>
