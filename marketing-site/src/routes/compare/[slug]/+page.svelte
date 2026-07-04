<script lang="ts">
	import { page } from '$app/state';
	import { error } from '@sveltejs/kit';
	import { CheckCircle2, Minus } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { appUrl, getComparison, siteUrl } from '../../_marketing';

	const slug = $derived(page.params.slug ?? '');
	const comparison = $derived.by(() => {
		const found = getComparison(slug);
		if (!found) error(404, 'Comparison not found');
		return found;
	});
</script>

<svelte:head>
	<title>OpenPost vs {comparison.name} - OpenPost</title>
	<meta
		name="description"
		content={`Compare OpenPost with ${comparison.name} for social scheduling and publishing.`}
	/>
	<link rel="canonical" href={`${siteUrl}/compare/${comparison.slug}`} />
</svelte:head>

<section class="section-pad border-b">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl">
			<p class="eyebrow">Comparison</p>
			<h1 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
				OpenPost vs {comparison.name}
			</h1>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">{comparison.openPostAngle}</p>
			<p class="mt-4 text-base leading-7 text-muted-foreground">
				{comparison.name} may be better for: {comparison.bestFor}
			</p>
			<div class="mt-8 flex flex-wrap gap-3">
				<Button href={appUrl} size="lg">Open app</Button>
				<Button href="/compare" variant="outline" size="lg">All comparisons</Button>
			</div>
		</div>
	</div>
</section>

<section class="section-pad">
	<div class="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8">
		<div class="overflow-hidden rounded-xl border bg-card">
			{#each [
				['Focused scheduler', 'OpenPost is built around drafting, variants, preview, queue, and activity.', true],
				['Source-open trust', 'You can inspect the implementation and self-host if needed.', true],
				['Advanced analytics', 'Not a launch feature; do not buy OpenPost for analytics today.', false],
				['Enterprise approvals', 'Not the current focus; OpenPost is for lean teams and operators.', false],
				['Universal video support', 'Provider-dependent and intentionally described with caveats.', false]
			] as row (row[0])}
				<div class="grid gap-3 border-b p-5 last:border-b-0 sm:grid-cols-[1fr_2fr_auto] sm:items-center">
					<h2 class="font-semibold">{row[0]}</h2>
					<p class="text-sm leading-6 text-muted-foreground">{row[1]}</p>
					{#if row[2]}
						<CheckCircle2 class="size-5 text-primary" />
					{:else}
						<Minus class="size-5 text-muted-foreground" />
					{/if}
				</div>
			{/each}
		</div>
	</div>
</section>
