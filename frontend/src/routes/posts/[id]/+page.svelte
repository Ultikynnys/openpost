<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { client } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import ComposeSimple from '$lib/components/compose-simple.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';

	type PostDetailResponse = components['schemas']['PostDetailResponse'];
	type PostDetail = Omit<PostDetailResponse, 'media' | 'destinations'> & {
		media: NonNullable<PostDetailResponse['media']>;
		destinations: NonNullable<PostDetailResponse['destinations']>;
	};

	let post = $state<PostDetail | null>(null);
	let hasLoaded = $state(false);
	let error = $state('');
	let requestedPostId = $state('');

	const postId = $derived($page.params.id);

	async function loadPost(id: string) {
		hasLoaded = false;
		error = '';
		try {
			const { data, error: err } = await client.GET('/posts/{id}', {
				params: { path: { id } }
			});
			if (err) throw new Error(err.detail || 'Failed to load post');
			post = data
				? { ...data, media: data.media ?? [], destinations: data.destinations ?? [] }
				: null;
		} catch (e) {
			error = (e as Error).message;
			if (!hasLoaded) post = null;
		} finally {
			hasLoaded = true;
		}
	}

	$effect(() => {
		if (postId && postId !== requestedPostId) {
			requestedPostId = postId;
			loadPost(postId);
		}
	});

	async function handleSuccess() {
		goto(resolve('/'));
	}
</script>

<svelte:head>
	<title>{post ? 'Edit Post' : 'Loading...'} - OpenPost</title>
</svelte:head>

{#if !hasLoaded}
	<div class="mx-auto w-full max-w-2xl space-y-4 p-6">
		<Skeleton class="h-9 w-full rounded-lg" />
		<Skeleton class="h-64 w-full rounded-lg" />
	</div>
{:else if error && !post}
	<div class="mx-auto w-full max-w-6xl px-4 py-6 lg:px-8">
		<div class="rounded-lg border border-destructive/20 bg-destructive/10 p-6 text-center">
			<p class="mb-3 text-destructive">{error}</p>
			<Button variant="outline" onclick={() => goto(resolve('/'))}>Back</Button>
		</div>
	</div>
{:else if post}
	<div class="flex flex-1 flex-col overflow-hidden">
		{#if error}
			<div
				class="mx-4 mt-3 rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive"
			>
				{error}
			</div>
		{/if}

		<ComposeSimple initialPost={post} onSuccess={handleSuccess} onDeleted={handleSuccess} />
	</div>
{/if}
