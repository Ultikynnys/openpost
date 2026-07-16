<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { client, type Post } from '$lib/api/client';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Tabs, TabsList, TabsTrigger, TabsContent } from '$lib/components/ui/tabs';
	import PageContainer from '$lib/components/page-container.svelte';
	import EmptyState from '$lib/components/empty-state.svelte';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import CalendarIcon from 'lucide-svelte/icons/calendar-days';
	import CheckCircleIcon from 'lucide-svelte/icons/circle-check';
	import XCircleIcon from 'lucide-svelte/icons/circle-x';
	import AlertCircleIcon from 'lucide-svelte/icons/alert-circle';
	import RefreshIcon from 'lucide-svelte/icons/refresh-cw';
	import FileTextIcon from 'lucide-svelte/icons/file-text';
	import PencilIcon from 'lucide-svelte/icons/pencil';
	import PostsIcon from 'lucide-svelte/icons/files';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import ClockIcon from 'lucide-svelte/icons/clock';
	import { m } from '$lib/paraglide/messages';
	import { getLocaleTag } from '$lib/i18n';
	import { getDraftPresentation } from '$lib/components/compose/draft-utils';

	type JobLog = {
		id: string;
		type: string;
		status: string;
		payload?: string;
		run_at: string;
		last_error?: string;
	};

	let posts = $state.raw<Post[]>([]);
	let failedJobs = $state.raw<JobLog[]>([]);
	let loading = $state(true);
	let error = $state('');
	let activeTab = $state(page.url.searchParams.get('tab') === 'drafts' ? 'drafts' : 'scheduled');

	const scheduledPosts = $derived(
		posts
			.filter((post) => post.status === 'scheduled')
			.toSorted((a, b) => timestamp(a.scheduled_at) - timestamp(b.scheduled_at))
	);
	const publishedPosts = $derived(
		posts
			.filter((post) => post.status === 'published')
			.toSorted((a, b) => timestamp(b.created_at) - timestamp(a.created_at))
	);
	const failedPosts = $derived(
		posts
			.filter((post) => post.status === 'failed')
			.toSorted((a, b) => timestamp(b.created_at) - timestamp(a.created_at))
	);
	const drafts = $derived(
		posts
			.filter((post) => post.status === 'draft')
			.toSorted((a, b) => timestamp(b.created_at) - timestamp(a.created_at))
	);

	onMount(() => {
		void loadData();
	});

	async function loadData() {
		loading = true;
		error = '';
		try {
			if (!workspaceCtx.currentWorkspace) await workspaceCtx.initialize();
			const workspaceId = workspaceCtx.currentWorkspace?.id;
			const [postsResponse, jobsResponse] = await Promise.all([
				client.GET('/posts', {
					params: { query: { ...(workspaceId ? { workspace_id: workspaceId } : {}), limit: 200 } }
				}),
				client.GET('/jobs', { params: { query: { limit: 100, offset: 0 } } })
			]);

			if (postsResponse.error || !postsResponse.data) {
				throw new Error(m.activity_failed_posts());
			}
			posts = postsResponse.data;
			failedJobs = (jobsResponse.data ?? []).filter((job) => job.status === 'failed');
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.activity_failed_load();
		} finally {
			loading = false;
		}
	}

	function timestamp(value?: string) {
		return value ? new Date(value).getTime() : 0;
	}

	function formatDateTime(value?: string) {
		if (!value) return '';
		return new Date(value).toLocaleString(getLocaleTag(), {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function postText(post: Post) {
		if (post.status !== 'draft') return post.content || 'Untitled post';
		const presentation = getDraftPresentation(post);
		return presentation.isThread
			? `${presentation.title} · ${presentation.postCount} posts`
			: presentation.title;
	}

	function truncate(value: string, max = 180) {
		return value.length > max ? `${value.slice(0, max).trim()}…` : value;
	}

	function failedJobPostID(job: JobLog) {
		if (!job.payload) return '';
		try {
			return JSON.parse(job.payload).post_id ?? '';
		} catch {
			return '';
		}
	}

	function statusLabel(post: Post) {
		switch (post.status) {
			case 'scheduled':
				return m.activity_status_scheduled();
			case 'published':
				return m.activity_status_published();
			case 'failed':
				return m.activity_status_failed();
			default:
				return m.activity_status_draft();
		}
	}

	function statusIcon(post: Post) {
		switch (post.status) {
			case 'scheduled':
				return ClockIcon;
			case 'published':
				return CheckCircleIcon;
			case 'failed':
				return XCircleIcon;
			default:
				return FileTextIcon;
		}
	}

	function statusClass(post: Post) {
		switch (post.status) {
			case 'scheduled':
				return 'text-amber-700 dark:text-amber-300';
			case 'published':
				return 'text-emerald-700 dark:text-emerald-300';
			case 'failed':
				return 'text-destructive';
			default:
				return 'text-muted-foreground';
		}
	}
</script>

{#snippet postList(items: Post[], emptyTitle: string, emptyDescription: string)}
	{#if items.length === 0}
		<EmptyState
			icon={FileTextIcon}
			title={emptyTitle}
			description={emptyDescription}
			variant="muted"
		/>
	{:else}
		<div class="divide-y border-y">
			{#each items as post (post.id)}
				{@const StatusIcon = statusIcon(post)}
				<article class="group flex items-start gap-3 py-4 sm:gap-4">
					<div class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-muted">
						<StatusIcon class={`size-4 ${statusClass(post)}`} />
					</div>
					<div class="min-w-0 flex-1">
						<div class="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs">
							<span class={['font-medium', statusClass(post)]}>{statusLabel(post)}</span>
							<span class="text-muted-foreground">
								{formatDateTime(post.scheduled_at || post.created_at)}
							</span>
						</div>
						<p class="mt-1.5 max-w-[72ch] text-sm leading-6 text-foreground/92">
							{truncate(postText(post))}
						</p>
						{#if post.destinations?.length}
							<div class="mt-2 flex flex-wrap gap-2">
								{#each post.destinations as destination (destination.social_account_id)}
									<span class="inline-flex items-center gap-1 text-xs text-muted-foreground">
										<PlatformIcon platform={destination.platform} class="size-3.5" />
										<span class="capitalize">{destination.platform}</span>
									</span>
								{/each}
							</div>
						{/if}
					</div>
					<Button
						variant="ghost"
						size="sm"
						class="min-h-10 shrink-0"
						onclick={() => goto(resolve('/posts/[id]', { id: post.id }))}
						aria-label={`Edit ${truncate(postText(post), 40)}`}
					>
						<PencilIcon class="size-4 sm:mr-1.5" />
						<span class="hidden sm:inline">{m.common_edit()}</span>
					</Button>
				</article>
			{/each}
		</div>
	{/if}
{/snippet}

<svelte:head>
	<title>{m.activity_title()} — {m.common_openpost()}</title>
</svelte:head>

<PageContainer
	title={m.activity_title()}
	description={m.activity_description()}
	icon={PostsIcon}
	{loading}
>
	{#snippet actions()}
		<Button variant="outline" size="sm" onclick={loadData} disabled={loading}>
			<RefreshIcon class={`mr-1.5 size-3.5 ${loading ? 'animate-spin' : ''}`} />
			{m.common_refresh()}
		</Button>
		<Button size="sm" onclick={() => goto(resolve('/'))}>
			<PlusIcon class="mr-1.5 size-3.5" />
			New post
		</Button>
	{/snippet}

	{#if error}
		<div
			class="mb-6 flex items-center gap-3 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive"
		>
			<AlertCircleIcon class="size-4 shrink-0" />
			<span>{error}</span>
		</div>
	{/if}

	<Tabs bind:value={activeTab}>
		<TabsList variant="line" class="mb-6 w-full justify-start overflow-x-auto">
			<TabsTrigger value="scheduled"
				>Scheduled <span class="text-muted-foreground">{scheduledPosts.length}</span></TabsTrigger
			>
			<TabsTrigger value="published"
				>Published <span class="text-muted-foreground">{publishedPosts.length}</span></TabsTrigger
			>
			<TabsTrigger value="failed"
				>Failed <span class="text-muted-foreground">{failedPosts.length + failedJobs.length}</span
				></TabsTrigger
			>
			<TabsTrigger value="drafts"
				>Drafts <span class="text-muted-foreground">{drafts.length}</span></TabsTrigger
			>
		</TabsList>

		<TabsContent value="scheduled">
			{@render postList(
				scheduledPosts,
				'Nothing scheduled',
				'Choose a time from the composer when your next post is ready.'
			)}
		</TabsContent>
		<TabsContent value="published">
			{@render postList(
				publishedPosts,
				'No published posts yet',
				'Posts you publish will be kept here for reference.'
			)}
		</TabsContent>
		<TabsContent value="failed">
			{@render postList(
				failedPosts,
				'No failed posts',
				'Publishing problems will appear here with a clear path to edit and retry.'
			)}
			{#if failedJobs.length > 0}
				<details class="mt-6 border-t pt-4">
					<summary
						class="cursor-pointer text-sm font-medium text-destructive focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
					>
						Technical details for {failedJobs.length} failed {failedJobs.length === 1
							? 'delivery'
							: 'deliveries'}
					</summary>
					<div class="mt-3 divide-y rounded-md bg-muted/35 px-4">
						{#each failedJobs as job (job.id)}
							<div class="flex items-start justify-between gap-4 py-3 text-sm">
								<div>
									<p class="font-medium">{job.type.replaceAll('_', ' ')}</p>
									<p class="mt-1 text-xs text-destructive">
										{job.last_error || 'Delivery failed.'}
									</p>
								</div>
								{#if failedJobPostID(job)}
									<Button
										variant="ghost"
										size="sm"
										onclick={() => goto(resolve('/posts/[id]', { id: failedJobPostID(job) }))}
										>Open post</Button
									>
								{/if}
							</div>
						{/each}
					</div>
				</details>
			{/if}
		</TabsContent>
		<TabsContent value="drafts">
			{@render postList(
				drafts,
				'No drafts yet',
				'Start writing and OpenPost will keep your work here.'
			)}
		</TabsContent>
	</Tabs>
</PageContainer>
