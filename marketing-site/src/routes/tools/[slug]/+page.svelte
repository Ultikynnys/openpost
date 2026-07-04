<script lang="ts">
	import { page } from '$app/state';
	import { error } from '@sveltejs/kit';
	import { ArrowRight } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { appUrl, getTool, siteUrl } from '../../_marketing';

	const slug = $derived(page.params.slug ?? '');
	const tool = $derived.by(() => {
		const found = getTool(slug);
		if (!found) error(404, 'Tool not found');
		return found;
	});
</script>

<svelte:head>
	<title>{tool.name} - OpenPost</title>
	<meta name="description" content={tool.description} />
	<link rel="canonical" href={`${siteUrl}/tools/${tool.slug}`} />
</svelte:head>

<section class="section-pad border-b">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[0.85fr_1.15fr] lg:items-center">
			<div>
				<p class="eyebrow">Free tool</p>
				<h1 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
					{tool.name}
				</h1>
				<p class="mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">{tool.description}</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href={appUrl} size="lg">
						Use OpenPost Cloud
						<ArrowRight data-icon="inline-end" />
					</Button>
					<Button href="/tools" variant="outline" size="lg">All tools</Button>
				</div>
			</div>
			<div class="rounded-xl border bg-card p-5">
				<label for="tool-input" class="text-sm font-medium">Paste a draft</label>
				<textarea
					id="tool-input"
					class="mt-3 min-h-48 w-full rounded-md border bg-background p-4 text-sm"
					placeholder="Paste your social post here..."
				></textarea>
				<div class="mt-4 grid gap-3 sm:grid-cols-3">
					{#each ['X 280', 'Bluesky 300', 'LinkedIn 3000'] as limit (limit)}
						<div class="rounded-md border bg-muted/30 p-3 font-mono text-xs text-muted-foreground">
							{limit}
						</div>
					{/each}
				</div>
				<p class="mt-4 text-xs leading-5 text-muted-foreground">
					This public page is a lightweight lead-in. The full app adds accounts, media, variants,
					scheduling, and queue state.
				</p>
			</div>
		</div>
	</div>
</section>
