<script lang="ts">
	import { onMount } from 'svelte';
	import { client as apiClient } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import MessageCircleIcon from 'lucide-svelte/icons/message-circle';
	import RefreshIcon from 'lucide-svelte/icons/refresh-cw';
	import AlertCircleIcon from 'lucide-svelte/icons/alert-circle';

	type CommentResponse = components['schemas']['CommentResponse'];
	type CommentListResponse = components['schemas']['CommentListResponse'];
	type CommentActionOutputBody = components['schemas']['CommentActionOutputBody'];

	type CommentClient = {
		listComments: (renditionId: string) => Promise<CommentListResponse>;
		replyToComment: (commentId: string, body: string) => Promise<CommentActionOutputBody>;
		hideComment: (commentId: string) => Promise<CommentActionOutputBody>;
		deleteComment: (commentId: string) => Promise<CommentActionOutputBody>;
	};

	let {
		renditionId,
		platform,
		client = createDefaultClient()
	}: {
		renditionId: string;
		platform?: string;
		client?: CommentClient;
	} = $props();

	let comments = $state<CommentResponse[]>([]);
	let loading = $state(false);
	let error = $state('');
	let actionMessage = $state('');
	let replyFor = $state<string | null>(null);
	let replyBody = $state('');
	let actionBusy = $state<string | null>(null);

	onMount(() => {
		void loadComments();
	});

	function createDefaultClient(): CommentClient {
		return {
			async listComments(id) {
				const { data, error: err } = await apiClient.GET('/renditions/{id}/comments', {
					params: { path: { id } }
				});
				if (err || !data) throw new Error(readProblem(err, 'Failed to load comments'));
				return data;
			},
			async replyToComment(id, body) {
				const { data, error: err } = await apiClient.POST('/comments/{id}/reply', {
					params: { path: { id } },
					body: { body }
				});
				if (err || !data) throw new Error(readProblem(err, 'Failed to send reply'));
				return data;
			},
			async hideComment(id) {
				const { data, error: err } = await apiClient.POST('/comments/{id}/hide', {
					params: { path: { id } }
				});
				if (err || !data) throw new Error(readProblem(err, 'Failed to hide comment'));
				return data;
			},
			async deleteComment(id) {
				const { data, error: err } = await apiClient.DELETE('/comments/{id}', {
					params: { path: { id } }
				});
				if (err || !data) throw new Error(readProblem(err, 'Failed to delete comment'));
				return data;
			}
		};
	}

	function readProblem(value: unknown, fallback: string): string {
		if (value && typeof value === 'object' && 'detail' in value) {
			const detail = (value as { detail?: unknown }).detail;
			if (typeof detail === 'string' && detail.length > 0) return detail;
		}
		return fallback;
	}

	async function loadComments() {
		loading = true;
		error = '';
		actionMessage = '';
		try {
			const data = await client.listComments(renditionId);
			comments = data.comments ?? [];
		} catch (e) {
			error = (e as Error).message || 'Failed to load comments';
			comments = [];
		} finally {
			loading = false;
		}
	}

	async function runAction(commentId: string, action: () => Promise<CommentActionOutputBody>) {
		actionBusy = commentId;
		error = '';
		actionMessage = '';
		try {
			const result = await action();
			actionMessage = result.message;
			await loadComments();
		} catch (e) {
			error = (e as Error).message || 'Comment action failed';
		} finally {
			actionBusy = null;
		}
	}

	async function sendReply(commentId: string) {
		const body = replyBody.trim();
		if (!body) return;
		await runAction(commentId, () => client.replyToComment(commentId, body));
		replyFor = null;
		replyBody = '';
	}

	function cancelReply() {
		replyFor = null;
		replyBody = '';
	}
</script>

<section class="rounded-lg border bg-card p-4" data-testid="comment-inbox">
	<div class="flex flex-wrap items-center justify-between gap-3">
		<div class="flex items-center gap-2">
			<div class="flex h-8 w-8 items-center justify-center rounded-md bg-muted">
				<MessageCircleIcon class="h-4 w-4 text-muted-foreground" />
			</div>
			<div>
				<h3 class="text-sm font-semibold">Comments</h3>
				{#if platform}
					<p class="text-xs text-muted-foreground capitalize">{platform}</p>
				{/if}
			</div>
		</div>
		<Button variant="outline" size="sm" onclick={loadComments} disabled={loading}>
			<RefreshIcon class="mr-1.5 h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" />
			Refresh
		</Button>
	</div>

	{#if error}
		<div
			class="mt-3 flex items-center gap-2 rounded-md border border-destructive/20 bg-destructive/5 px-3 py-2 text-xs text-destructive"
		>
			<AlertCircleIcon class="h-3.5 w-3.5 shrink-0" />
			{error}
		</div>
	{/if}

	{#if actionMessage}
		<div class="mt-3 rounded-md bg-emerald-50 px-3 py-2 text-xs text-emerald-700">
			{actionMessage}
		</div>
	{/if}

	{#if loading}
		<div class="mt-4 space-y-2">
			<div class="h-14 animate-pulse rounded-md bg-muted"></div>
			<div class="h-14 animate-pulse rounded-md bg-muted"></div>
		</div>
	{:else if comments.length === 0}
		<p class="mt-4 rounded-md bg-muted/50 px-3 py-3 text-sm text-muted-foreground">
			No comments yet.
		</p>
	{:else}
		<div class="mt-4 space-y-3">
			{#each comments as comment (comment.id)}
				<article class="rounded-md border bg-background p-3" data-testid="comment-row">
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0">
							<div class="flex flex-wrap items-center gap-2">
								<p class="truncate text-sm font-medium">
									{comment.author_name || 'Unknown author'}
								</p>
								{#if comment.hidden}
									<span
										class="rounded-full bg-amber-50 px-2 py-0.5 text-[11px] text-amber-700 ring-1 ring-amber-600/20"
									>
										Hidden
									</span>
								{/if}
							</div>
							<p class="mt-1 text-sm leading-relaxed text-foreground/90">{comment.text}</p>
						</div>
					</div>

					<div class="mt-3 flex flex-wrap gap-2">
						{#if comment.can_reply}
							<Button
								variant="outline"
								size="sm"
								onclick={() => (replyFor = comment.id)}
								disabled={actionBusy === comment.id}
							>
								Reply
							</Button>
						{/if}
						{#if comment.can_hide && !comment.hidden}
							<Button
								variant="outline"
								size="sm"
								onclick={() => runAction(comment.id, () => client.hideComment(comment.id))}
								disabled={actionBusy === comment.id}
							>
								Hide
							</Button>
						{/if}
						{#if comment.can_delete}
							<Button
								variant="outline"
								size="sm"
								onclick={() => runAction(comment.id, () => client.deleteComment(comment.id))}
								disabled={actionBusy === comment.id}
							>
								Delete
							</Button>
						{/if}
					</div>

					{#if replyFor === comment.id}
						<div class="mt-3 space-y-2">
							<Textarea
								bind:value={replyBody}
								placeholder="Write a reply..."
								aria-label="Reply body"
								class="min-h-20"
							/>
							<div class="flex justify-end gap-2">
								<Button variant="ghost" size="sm" onclick={cancelReply}>Cancel</Button>
								<Button
									size="sm"
									onclick={() => sendReply(comment.id)}
									disabled={actionBusy === comment.id || replyBody.trim().length === 0}
								>
									Send reply
								</Button>
							</div>
						</div>
					{/if}
				</article>
			{/each}
		</div>
	{/if}
</section>
