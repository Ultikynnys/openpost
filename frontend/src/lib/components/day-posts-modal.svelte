<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Button } from '$lib/components/ui/button';
	import { client, type Post } from '$lib/api/client';
	import { ui } from '$lib/stores/ui.svelte';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { getLocalTimeZone, today, type DateValue } from '@internationalized/date';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import CalendarIcon from 'lucide-svelte/icons/calendar-days';
	import TrashIcon from 'lucide-svelte/icons/trash-2';
	import PencilIcon from 'lucide-svelte/icons/pencil';
	import MoreIcon from 'lucide-svelte/icons/ellipsis';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { getStatusColor } from '$lib/utils';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { m } from '$lib/paraglide/messages';
	import { getLocaleTag } from '$lib/i18n';

	let posts = $state.raw<Post[]>([]);
	let loading = $state(false);
	let error = $state('');
	let open = $state(false);

	const currentDate = $derived<DateValue | undefined>(ui.dayPostsDate);
	const dateStr = $derived(currentDate ? currentDate.toString() : '');
	const isFutureDay = $derived.by(() => {
		if (!currentDate) return false;
		return currentDate.compare(today(getLocalTimeZone())) >= 0;
	});
	const formattedDate = $derived.by(() => {
		if (!currentDate) return '';
		return currentDate.toDate(getLocalTimeZone()).toLocaleDateString(getLocaleTag(), {
			weekday: 'long',
			month: 'long',
			day: 'numeric'
		});
	});

	$effect(() => {
		open = ui.isDayPostsOpen;
		if (open && dateStr) void loadPosts(dateStr);
	});

	function handleOpenChange(isOpen: boolean) {
		open = isOpen;
		if (!isOpen) ui.closeDayPosts();
	}

	async function loadPosts(date: string) {
		loading = true;
		error = '';
		try {
			const workspaceId = workspaceCtx.currentWorkspace?.id;
			const { data, error: responseError } = await client.GET('/posts', {
				params: { query: { date, ...(workspaceId ? { workspace_id: workspaceId } : {}) } }
			});
			if (responseError) throw new Error(m.day_posts_load_failed());
			posts = (data ?? []).toSorted(
				(a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime()
			);
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.day_posts_load_failed();
			posts = [];
		} finally {
			loading = false;
		}
	}

	function getTime(value: string) {
		return new Date(value).toLocaleTimeString(getLocaleTag(), {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: workspaceCtx.settings.timezone || 'UTC'
		});
	}

	function postExcerpt(post: Post) {
		const text = post.content || 'Untitled post';
		return text.length > 130 ? `${text.slice(0, 130).trim()}…` : text;
	}

	function handleNewPost() {
		ui.closeDayPosts();
		const params = new URLSearchParams();
		if (dateStr) params.set('date', dateStr);
		if (workspaceCtx.currentWorkspace?.id)
			params.set('workspace_id', workspaceCtx.currentWorkspace.id);
		const target = `/?${params.toString()}`;
		goto(resolve(target as '/'));
	}

	function handleEdit(postId: string) {
		ui.closeDayPosts();
		goto(resolve(`/posts/${postId}` as '/'));
	}

	async function handleDelete(postId: string) {
		if (!confirm(m.day_posts_delete_confirm())) return;
		try {
			const { error: responseError } = await client.DELETE('/posts/{id}', {
				params: { path: { id: postId } }
			});
			if (responseError) throw new Error(responseError.detail || m.day_posts_delete_failed());
			await loadPosts(dateStr);
			ui.triggerRefresh();
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.day_posts_delete_failed();
		}
	}
</script>

<Sheet.Root {open} onOpenChange={handleOpenChange}>
	<Sheet.Content side="right" class="w-full p-0 sm:max-w-md">
		<Sheet.Header class="border-b p-5 pr-14">
			<div class="flex items-start justify-between gap-4">
				<div>
					<Sheet.Title class="flex items-center gap-2 text-base font-semibold">
						<CalendarIcon class="size-4 text-primary" />
						{formattedDate}
					</Sheet.Title>
					<Sheet.Description class="mt-1 text-sm">
						{m.day_posts_scheduled_count({ count: posts.length })}
					</Sheet.Description>
				</div>
				{#if isFutureDay}
					<Button size="sm" onclick={handleNewPost}>
						<PlusIcon class="mr-1.5 size-4" />
						New post
					</Button>
				{/if}
			</div>
		</Sheet.Header>

		<div class="min-h-0 flex-1 overflow-y-auto px-5 py-2">
			{#if loading}
				<div class="divide-y">
					{#each [1, 2, 3] as placeholder (placeholder)}
						<div class="flex gap-3 py-4">
							<Skeleton class="h-8 w-12 rounded" />
							<div class="flex flex-1 flex-col gap-2">
								<Skeleton class="h-3 w-full" />
								<Skeleton class="h-3 w-2/3" />
							</div>
						</div>
					{/each}
				</div>
			{:else if error}
				<div class="my-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">{error}</div>
			{:else if posts.length === 0}
				<div class="flex min-h-64 flex-col items-center justify-center gap-3 text-center">
					<CalendarIcon class="size-8 text-muted-foreground/45" />
					<div>
						<p class="text-sm font-medium">Nothing scheduled</p>
						<p class="mt-1 text-sm text-muted-foreground">This day is open.</p>
					</div>
					{#if isFutureDay}<Button variant="outline" size="sm" onclick={handleNewPost}
							>Schedule a post</Button
						>{/if}
				</div>
			{:else}
				<div class="divide-y">
					{#each posts as post (post.id)}
						<article class="flex items-start gap-3 py-4">
							<time
								class="w-12 shrink-0 pt-0.5 font-mono text-xs font-medium text-muted-foreground"
							>
								{getTime(post.scheduled_at)}
							</time>
							<button
								type="button"
								class="min-w-0 flex-1 text-left focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
								onclick={() => handleEdit(post.id)}
							>
								<p class="text-sm leading-6">{postExcerpt(post)}</p>
								<div class="mt-2 flex flex-wrap items-center gap-2">
									<span
										class={[
											'rounded-sm px-1.5 py-0.5 text-[11px] font-medium capitalize',
											getStatusColor(post.status)
										]}>{post.status}</span
									>
									{#each post.destinations ?? [] as destination (destination.social_account_id)}
										<span class="inline-flex items-center gap-1 text-xs text-muted-foreground">
											<PlatformIcon platform={destination.platform} class="size-3.5" />
											<span class="capitalize">{destination.platform}</span>
										</span>
									{/each}
								</div>
							</button>
							<DropdownMenu.Root>
								<DropdownMenu.Trigger>
									{#snippet child({ props })}
										<Button {...props} variant="ghost" size="icon" aria-label="Post actions"
											><MoreIcon class="size-4" /></Button
										>
									{/snippet}
								</DropdownMenu.Trigger>
								<DropdownMenu.Content align="end">
									<DropdownMenu.Item onclick={() => handleEdit(post.id)}
										><PencilIcon class="mr-2 size-4" />Edit post</DropdownMenu.Item
									>
									<DropdownMenu.Separator />
									<DropdownMenu.Item class="text-destructive" onclick={() => handleDelete(post.id)}
										><TrashIcon class="mr-2 size-4" />Delete post</DropdownMenu.Item
									>
								</DropdownMenu.Content>
							</DropdownMenu.Root>
						</article>
					{/each}
				</div>
			{/if}
		</div>
	</Sheet.Content>
</Sheet.Root>
