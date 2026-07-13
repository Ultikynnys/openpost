<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { client } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import ComposeFocusedPublication from '$lib/components/compose-focused-publication.svelte';
	import { ui } from '$lib/stores/ui.svelte';
	import { COMPOSER_MODE_KEYS, type ComposerModeKey } from '$lib/components/compose/modes';

	type Publication = components['schemas']['PublicationResponse'];

	let publication = $state<Publication | null>(null);
	let hasLoaded = $state(false);
	let error = $state('');
	let requestedPublicationId = $state('');

	const publicationId = $derived($page.params.id);

	async function loadPublication(id: string) {
		hasLoaded = false;
		error = '';
		try {
			const { data, error: err } = await client.GET('/publications/{id}', {
				params: { path: { id } }
			});
			if (err) throw new Error((err as any)?.detail || 'Failed to load publication');
			publication = data;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load publication';
			publication = null;
		} finally {
			hasLoaded = true;
		}
	}

	function publicationMode(item: Publication): ComposerModeKey {
		return COMPOSER_MODE_KEYS.includes(item.content_profile as ComposerModeKey)
			? (item.content_profile as ComposerModeKey)
			: 'link_share';
	}

	async function handleSuccess() {
		ui.triggerRefresh();
		goto(resolve('/'));
	}

	function handleCancel() {
		goto(resolve('/'));
	}

	$effect(() => {
		if (publicationId && publicationId !== requestedPublicationId) {
			requestedPublicationId = publicationId;
			loadPublication(publicationId);
		}
	});
</script>

<svelte:head>
	<title>{publication ? 'Edit Publication' : 'Loading...'} - OpenPost</title>
</svelte:head>

{#if !hasLoaded}
	<div class="mx-auto w-full max-w-3xl space-y-4 p-6">
		<Skeleton class="h-9 w-full rounded-lg" />
		<Skeleton class="h-64 w-full rounded-lg" />
	</div>
{:else if error && !publication}
	<div class="mx-auto w-full max-w-6xl px-4 py-6 lg:px-8">
		<div class="rounded-lg border border-destructive/20 bg-destructive/10 p-6 text-center">
			<p class="mb-3 text-destructive">{error}</p>
			<Button variant="outline" onclick={() => goto(resolve('/'))}>Back</Button>
		</div>
	</div>
{:else if publication}
	<div class="flex flex-1 flex-col overflow-hidden">
		<ComposeFocusedPublication
			mode={publicationMode(publication)}
			initialPublication={publication}
			onSuccess={handleSuccess}
			onCancel={handleCancel}
		/>
	</div>
{/if}
